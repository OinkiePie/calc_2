package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator/internal/router"
	"github.com/OinkiePie/calc_2/pkg/initializer"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/shutdown"
	"github.com/rs/cors"
)

// Orchestrator представляет собой сервис оркестратора.
type Orchestrator struct {
	errChan chan error   // Канал для отправки ошибок, возникающих в сервисе.
	server  *http.Server // Указатель на структуру http.Server, управляющую HTTP-сервером.
	Addr    string       // Адрес, на котором прослушивает HTTP-сервер.
}

// NewOrchestrator создает новый экземпляр сервиса оркестратора.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при инициализации или работе сервиса.
//
// Returns:
//
//	*Orchestrator - Указатель на новый экземпляр структуры Orchestrator.
func NewOrchestrator(errChan chan error) *Orchestrator {
	port := config.Cfg.Server.Orchestrator.Port

	addr := fmt.Sprintf("localhost:%d", port)
	router := router.NewOrchestratorRouter()

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

	return &Orchestrator{errChan: errChan, server: srv, Addr: addr}
}

// Start запускает HTTP-сервер в отдельной горутине. Если во время запуска
// возникает ошибка, она отправляется в канал ошибок.
func (o *Orchestrator) Start() {
	// Запускаем сервер в отдельной горутине, чтобы не блокировать основной поток выполнения.
	go func() {
		// Запускаем прослушивание входящих соединений на указанном адресе.
		if err := o.server.ListenAndServe(); err != http.ErrServerClosed {
			// Если при запуске сервера произошла ошибка, отправляем её в канал ошибок.
			o.errChan <- err
		}
	}()
}

// Stop останавливает HTTP-сервер. Он использует контекст с таймаутом, чтобы
// гарантировать, что остановка не займет слишком много времени.
func (o *Orchestrator) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := o.server.Shutdown(ctx)
	if err != nil {
		logger.Log.Errorf("Ошибка при остановке сервиса Оркестратор")
	}
}

func main() {
	// Инициализация конфига и логгера
	initializer.Init()

	errChan := make(chan error, 1)

	// Запуск сервиса агента в отдельной горутине чтобы можно было поймать завершение
	orchestratorService := NewOrchestrator(errChan)
	go func() {
		logger.Log.Debugf("Запуск веб сервиса...")
		orchestratorService.Start()
		logger.Log.Infof("Веб сервис запущен на %s", orchestratorService.Addr)
	}()

	shutdown.WaitForShutdown(errChan, "Orchestrator", orchestratorService)
}
