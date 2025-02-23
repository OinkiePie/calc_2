package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OinkiePie/calc_2/orchestrator/internal/router"
	"github.com/OinkiePie/calc_2/pkg/logger"
)

func main() {
	logger.InitLogger(logger.Options{
		// DisableColor: true, раскоментируй если терминал не поддерживает цвета
	})

	router := router.NewRouter()

	// Создаем HTTP-сервер.
	srv := &http.Server{
		Addr:         ":62836",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Создаем канал для сигналов завершения (Ctrl+C, SIGTERM)
	idleConnsClosed := make(chan struct{})

	// Запускаем горутину, которая будет слушать сигналы завершения.
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint // Блокируемся, пока не получим сигнал

		// Получили сигнал завершения
		logger.Log.Infof("Получен сигнал завершения, начинаем остановку сервера...")

		// Создаем контекст с таймаутом чтобы заверишь все запросы
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Таймаут 30 секунд на завершение
		defer cancel()

		// Завершаем работу сервера
		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Errorf("Ошибка при остановке сервера: %v", err)
		}
		close(idleConnsClosed)
	}()

	// Запускаем сервер
	logger.Log.Infof("Запуск сервера на %s", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	// Ожидаем завершения всех соединений
	<-idleConnsClosed
	logger.Log.Infof("Сервер успешно завершил работу")
}
