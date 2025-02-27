package main

//TODO: вынести завершалку в контекст и передавать внутрь
import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OinkiePie/calc_2/agent"
	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/web"
)

// Храним серверы в глобальных переменных, чтобы их можно было корректно завершить.
var (
	orchestratorService *orchestrator.Orchestrator
	webService          *web.Web
	agentService        *agent.Agent
)

func main() {

	// Инициализация конфига
	err := config.InitConfig()

	logger.InitLogger(logger.Options{
		Level:        logger.Level(config.Cfg.Logger.Level), // Преобразование int в тип Level
		TimeFormat:   config.Cfg.Logger.TimeFormat,
		CallDepth:    config.Cfg.Logger.CallDepth,
		DisableCall:  config.Cfg.Logger.DisableCall,
		DisableTime:  config.Cfg.Logger.DisableTime,
		DisableColor: config.Cfg.Logger.DisableColor,
	})

	if err != nil {
		logger.Log.Errorf(err.Error())
		logger.Log.Warnf("Ошибка при загрузке конфигурации, используется конфигурация по умолчанию")
	}
	logger.Log.Infof("Загружена конфигурация: %s", config.Name)

	// Канал для отслеживания ошибок
	errChan := make(chan error, 3)

	// Запуск сервера оркестратора
	orchestratorService = orchestrator.NewOrchestrator(errChan)
	go func() {
		logger.Log.Debugf("Запуск сервиса оркестратора...")
		orchestratorService.Start()
		logger.Log.Infof("Сервис оркестратора запущен на %s", orchestratorService.Addr)
	}()

	// Запуск сервера агента
	agentService = agent.NewAgent(errChan)
	go func() {
		logger.Log.Debugf("Запуск сервиса агента...")
		agentService.Start()
		logger.Log.Infof("Сервис агента запущен")
	}()

	// Запуск веб сервера агента
	webService = web.NewWeb(errChan)
	go func() {
		logger.Log.Debugf("Запуск веб сервис...")
		webService.Start()
		logger.Log.Infof("Веб сервис запущен на %s", webService.Addr)
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
	// gracefulShutdown()
}

// gracefulShutdown выполняет корректное завершение работы серверов (orchestrator, agent, web).
//
// Args:
//
//	(None): Функция не принимает аргументов.
//
// Returns:
//
//	(None): Функция не возвращает значений
func gracefulShutdown() {
	var wg sync.WaitGroup

	// Остановка сервисов
	if orchestratorService != nil {
		ShutdownSerice(&wg, "Orchestrator", orchestratorService.Stop)
	}

	if webService != nil {
		webService.Stop()
		ShutdownSerice(&wg, "Web", webService.Stop)
	}

	if agentService != nil {
		ShutdownSerice(&wg, "Agent", agentService.Stop)
	}

	wg.Wait()
	logger.Log.Infof("Сервисы успешно завершили работу")
}

// ShutdownService асинхронно останавливает сервис, логируя начало и завершение процесса остановки.
// Функция использует sync.WaitGroup для синхронизации и ожидания завершения горутины,
// в которой выполняется остановка сервиса.
//
// Args:
//
//	wg: *sync.WaitGroup - указатель на sync.WaitGroup, используемый для ожидания завершения всех сервисов.
//	name: string - имя останавливаемого сервиса (используется для логирования).
//	stopFunc: func() - функция, которая выполняет остановку сервиса.
func ShutdownSerice(wg *sync.WaitGroup, name string, stopFunc func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Log.Debugf("Остановка сервиса %s", name)
		stopFunc()
		logger.Log.Debugf("Сервис %s остановлен", name)
	}()
}
