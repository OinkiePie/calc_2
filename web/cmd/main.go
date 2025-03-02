package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/pkg/initializer"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/shutdown"
	"github.com/OinkiePie/calc_2/web/internal/router"
	"github.com/rs/cors"
)

// Web представляет собой веб-сервис.
type Web struct {
	errChan chan error   // Канал для отправки ошибок, возникающих в сервисе.
	server  *http.Server // Указатель на структуру http.Server, управляющую веб-сервером.
	Addr    string       // Адрес, на котором прослушивает веб-сервер.
}

// NewWeb создает новый экземпляр веб-сервиса.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при инициализации или работе сервиса.
//
// Returns:
//
//	*Web - Указатель на новый экземпляр структуры Web.
func NewWeb(errChan chan error) *Web {
	addr := config.Cfg.Server.Web.Addr
	staticDir := config.Cfg.Server.Web.StaticDir

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		errChan <- fmt.Errorf("директория со статическими файлами не найдена")
		return nil
	}

	router := router.NewWebRouter(staticDir)

	c := cors.New(cors.Options{
		AllowedOrigins:   config.Cfg.Middleware.AllowOrigin,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	routerCORS := c.Handler(router)

	// Создаем экземпляр структуры http.Server, указывая адрес и обработчик
	srv := &http.Server{
		Addr:    addr,
		Handler: routerCORS,
	}

	return &Web{errChan: errChan, server: srv, Addr: addr}
}

// Start запускает веб-сервер в отдельной горутине. Если во время запуска
// возникает ошибка, она отправляется в канал ошибок.
func (w *Web) Start() {
	// Запускаем сервер в отдельной горутине, чтобы не блокировать основной поток выполнения.
	go func() {
		// Запускаем прослушивание входящих соединений на указанном адресе.
		if err := w.server.ListenAndServe(); err != http.ErrServerClosed {
			// Если при запуске сервера произошла ошибка, отправляем её в канал ошибок.
			w.errChan <- err
		}
	}()

}

// Stop останавливает веб-сервер. Он использует контекст с таймаутом, чтобы
// гарантировать, что остановка не займет слишком много времени.
func (w *Web) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := w.server.Shutdown(ctx)
	if err != nil {
		logger.Log.Errorf("Ошибка при остановке сервиса Веб")
	}
}

// Запуск веб сервиса
func main() {
	// Инициализация конфига и логгера
	initializer.Init()

	errChan := make(chan error, 1)

	// Запуск сервиса агента в отдельной горутине чтобы можно было поймать завершение
	webService := NewWeb(errChan)
	if webService != nil {
		go func() {
			logger.Log.Debugf("Запуск сервиса Веб...")
			webService.Start()
			logger.Log.Infof("Сервис Веб запущен на %s", webService.Addr)
		}()
	}

	shutdown.WaitForShutdown(errChan, "Web", webService)
}
