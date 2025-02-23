package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OinkiePie/calc_2/agent"
	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator"
	"github.com/OinkiePie/calc_2/pkg/logger"
)

// Храним серверы в глобальных переменных, чтобы их можно было корректно завершить.
var (
	orchestratorServer *http.Server
	agentServer        *http.Server
)

func main() {
	logger.InitLogger(logger.Options{})

	err := config.InitConfig()
	if err != nil {
		logger.Log.Errorf(err.Error())
		logger.Log.Warnf("Ошибка при загрузке конфигурации, используется конфигурация по умолчанию")
	}

	// Канал для отслеживания ошибок
	errChan := make(chan error, 2)

	// Канал для graceful shutdown
	idleConnsClosed := make(chan struct{})

	// Запускаем сервер оркестратора
	go func() {
		srv, err := orchestrator.StartOrchestratorServer(config.Cfg.Orchestrator.Port)
		if err != nil {
			errChan <- err
			return
		}
		orchestratorServer = srv // Сохраняем для graceful shutdown
	}()

	// Запускаем сервер агента
	go func() {
		srv, err := agent.StartAgentServer(config.Cfg.Agent.Port)
		if err != nil {
			errChan <- err
			return
		}
		agentServer = srv // Сохраняем для graceful shutdown
	}()

	// Обработка сигналов завершения
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Log.Infof("Получен сигнал завершения, начинаем остановку серверов...")

		// Контекст с таймаутом для graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Остановка серверов
		if err := shutdownServer(ctx, orchestratorServer, "orchestrator"); err != nil {
			logger.Log.Errorf("Ошибка при остановке сервера оркестратора: %v", err)
		}
		if err := shutdownServer(ctx, agentServer, "agent"); err != nil {
			logger.Log.Errorf("Ошибка при остановке сервера агента: %v", err)
		}

		close(idleConnsClosed)
	}()

	// Ожидаем завершения работы серверов или возникновения ошибки
	select {
	case err := <-errChan: // Проверяем, была ли ошибка
		logger.Log.Fatalf("Фатальная ошибка: %v", err)
	case <-idleConnsClosed: // Дождались завершения серверов
		logger.Log.Infof("Серверы успешно завершили работу")
	}
}

// Функция для остановки сервера
func shutdownServer(ctx context.Context, srv *http.Server, serverName string) error {
	if srv != nil {
		logger.Log.Infof("Остановка сервера %s...", serverName)
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
