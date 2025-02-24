package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OinkiePie/calc_2/agent"
	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/web"
	"github.com/joho/godotenv"
)

// Храним серверы в глобальных переменных, чтобы их можно было корректно завершить.
var (
	orchestratorServer *http.Server
	agentServer        *http.Server
	webServer          *http.Server
)

func main() {
	// Загрузка env переменных из файла .env
	if err := godotenv.Load(); err != nil {
		log.Print("Файл .env не найден")
	}
	// Определение типа приложения - prod или dev
	app, exists := os.LookupEnv("APP_ENV")

	if !exists || app == "" {
		app = "dev" // По умолчанию - разработка
	}

	// Инициализация конфига
	err := config.InitConfig(app)

	logger.InitLogger(logger.Options{
		Level:        logger.Level(config.Cfg.Logger.Level), // Преобразование int в тип Level
		TimeFormat:   config.Cfg.Logger.TimeFormat,
		CallDepth:    config.Cfg.Logger.CallDepth,
		DisableCall:  config.Cfg.Logger.DisableCall,
		DisableTime:  config.Cfg.Logger.DisableTime,
		DisableColor: config.Cfg.Logger.DisableColor,
	})

	logger.Log.Infof("Загружена конфигурация: %s", app)
	if err != nil {
		logger.Log.Errorf(err.Error())
		logger.Log.Warnf("Ошибка при загрузке конфигурации, используется конфигурация по умолчанию")
	}

	// Канал для отслеживания ошибок
	errChan := make(chan error, 3)

	// Запуск сервера оркестратора
	go func() {
		logger.Log.Debugf("Запуск сервера оркестратора...")
		orchestratorServer = orchestrator.StartOrchestratorServer(errChan, config.Cfg.Server.Orchestrator.Port)
		logger.Log.Infof("Сервер оркестратора запущен на %s", orchestratorServer.Addr)
	}()

	// Запуск сервера агента
	go func() {
		logger.Log.Debugf("Запуск сервера агента...")
		agentServer = agent.StartAgentServer(errChan, config.Cfg.Server.Agent.Port)
		logger.Log.Infof("Сервер агента запущен на %s", agentServer.Addr)
	}()

	// Запуск веб сервера агента
	go func() {
		logger.Log.Debugf("Запуск веб сервера...")
		webServer = web.StartWebServer(errChan, config.Cfg.Server.Web.Port, config.Cfg.Server.Web.StaticDir)
		logger.Log.Infof("Веб сервер запущен на %s", webServer.Addr)
	}()

	waitForServers(errChan)
}

// waitForServers ожидает завершения работы серверов (из-за ошибки или сигнала завершения).
//
// Args:
//
//	errChan: <-chan error - Канал, из которого читаются ошибки, возникающие при работе серверов.
//	                    Если в канал поступает ошибка, функция завершает работу и логирует фатальную ошибку.
//
// Returns:
//
//	(None) - Функция ничего не возвращает.
func waitForServers(errChan <-chan error) {
	// Создаем канал для обработки сигналов завершения (Ctrl+C, SIGTERM)
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	// Ожидаем завершения работы серверов или возникновения ошибки
	select {
	// Ожидаем получения ошибки из канала errChan.
	case err := <-errChan:
		logger.Log.Fatalf("Фатальная ошибка: %v", err)
	// Ожидаем получения сигнала из канала sigint (сигнал завершения).
	case <-sigint:
		logger.Log.Debugf("Получен сигнал завершения, начинаем остановку серверов...")
	}

	// Запускаем функцию graceful shutdown в отдельной горутине
	gracefulShutdown()
}

// gracefulShutdown выполняет корректное завершение работы серверов (orchestrator, agent, web) с ограничением по времени.
//
// Args:
//
//	(None): Функция не принимает аргументов.
//
// Returns:
//
//	(None): Функция не возвращает значений
func gracefulShutdown() {
	// Создаем контекст с таймаутом для graceful shutdown
	// Контекст позволяет отменить операцию остановки серверов, если она занимает слишком много времени.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Остановка серверов
	if orchestratorServer != nil {
		err := shutdownServer(ctx, orchestratorServer, "orchestrator")
		if err != nil {
			logger.Log.Errorf("Ошибка при остановке сервера оркестратора: %v", err)
		}
	}
	if agentServer != nil {
		err := shutdownServer(ctx, agentServer, "agent")
		if err != nil {
			logger.Log.Errorf("Ошибка при остановке сервера агента: %v", err)
		}
	}
	if webServer != nil {
		err := shutdownServer(ctx, agentServer, "web")
		if err != nil {
			logger.Log.Errorf("Ошибка при остановке веб сервера: %v", err)
		}
	}

	logger.Log.Infof("Серверы успешно завершили работу")
}

// shutdownServer останавливает HTTP-сервер с использованием заданного контекста и логирует процесс.
//
// Args:
//
//	ctx: context.Context - Контекст с таймаутом или отменой, используемый для graceful shutdown сервера.
//	srv: *http.Server - Указатель на структуру http.Server, представляющую сервер, который необходимо остановить.
//	serverName: string - Имя сервера ("orchestrator", "agent", "web"), используемое для логирования.
//
// Returns:
//
//	error: nil, если сервер успешно остановлен. В противном случае возвращается ошибка, возникшая при остановке сервера.
func shutdownServer(ctx context.Context, srv *http.Server, serverName string) error {
	// Логируем сообщение о начале остановки сервера.
	logger.Log.Debugf("Остановка сервера %s...", serverName)
	// Вызываем метод Shutdown на сервере, передавая контекст для graceful shutdown.
	err := srv.Shutdown(ctx)
	// Проверяем, произошла ли ошибка при остановке сервера.
	if err != nil {
		// Если произошла ошибка, логируем её и возвращаем её.
		logger.Log.Errorf("Ошибка при остановке сервера %s: %v", serverName, err)
		return err
	}
	// Если сервер успешно остановлен, логируем сообщение об этом.
	logger.Log.Debugf("Сервер %s остановлен", serverName)
	// Возвращаем nil, чтобы указать на успешное завершение операции.
	return nil
}
