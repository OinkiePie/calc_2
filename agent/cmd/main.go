package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/OinkiePie/calc_2/agent/internal/client"
	"github.com/OinkiePie/calc_2/agent/internal/worker"
	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/pkg/initializer"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/shutdown"
)

// Agent представляет собой сервис агента, отвечающий за выполнение задач.
type Agent struct {
	errChan     chan error         // Канал для отправки ошибок, возникающих в сервисе.
	stopWorkers context.CancelFunc // Функция для остановки всех воркеров.
	wokertsCtx  context.Context    // Контекст, используемый воркерами для выполнения задач.
	client      *client.APIClient  // API-клиент для связи с сервисом оркестратора.
	workers     []*worker.Worker   // Список воркеров, выполняющих задачи.
	power       int                // Вычислительная мощность агента (количество воркеров).
	wg          *sync.WaitGroup    // WaitGroup для ожидания завершения всех воркеров.
}

// NewAgent создает новый экземпляр сервиса агента.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при работе сервиса.
//
// Returns:
//
//	*Agent - Указатель на новый экземпляр структуры Agent.
func NewAgent(errChan chan error) *Agent {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	apiClient := client.NewAPIClient(
		fmt.Sprintf("http://%s:%d/internal/task",
			config.Cfg.Server.Orchestrator.ADDR_ORCHESTRATOR,
			config.Cfg.Server.Orchestrator.PORT_ORCHESTRATOR),

		config.Cfg.Middleware.ApiKeyPrefix+config.Cfg.Middleware.Authorization,
		httpClient,
	)

	ctx, cancel := context.WithCancel(context.Background())

	computingPower := config.Cfg.Server.Agent.COMPUTING_POWER
	workers := make([]*worker.Worker, computingPower)

	a := &Agent{
		errChan:     errChan,           // Канал для ошибок
		wokertsCtx:  ctx,               // Контекст для воркеров
		stopWorkers: cancel,            // Функция для отмены контекста
		client:      apiClient,         // API-клиент
		workers:     workers,           // Слайс воркеров
		power:       computingPower,    // Вычислительная мощность
		wg:          &sync.WaitGroup{}, // WaitGroup для ожидания завершения воркеров
	}
	a.initWorkers()
	return a
}

// initWorkers инициализирует воркеров, создавая новые экземпляры Worker
// и добавляя их в слайс workers.
func (a *Agent) initWorkers() {
	logger.Log.Debugf("Инициализация %d работников", a.power)
	for i := range a.power {
		a.workers[i] = worker.NewWorker(i, a.client, a.wg, a.errChan)
	}
}

// Start запускает воркеров, запуская для каждого из них отдельную горутину.
func (a *Agent) Start() {
	logger.Log.Debugf("Запуск %d работников", a.power)
	for i := 1; i <= a.power; i++ {
		go a.workers[i-1].Start(a.wokertsCtx)
	}
}

// Stop останавливает воркеров, отменяя контекст и дожидаясь завершения
// всех горутин воркеров.
func (a *Agent) Stop() {
	a.stopWorkers() //  cancel
	if a.workers != nil {
		a.wg.Wait()
	}
}

// Запуск сервиса агента
func main() {
	// Инициализация конфига и логгера
	initializer.Init()

	errChan := make(chan error, 1)

	// Запуск сервиса агента в отдельной горутине чтобы можно было поймать завершение
	agentService := NewAgent(errChan)
	go func() {
		logger.Log.Debugf("Запуск сервиса Агент...")
		agentService.Start()
		logger.Log.Infof("Сервис Агент запущен")
	}()

	shutdown.WaitForShutdown(errChan, "Agent", agentService)
}
