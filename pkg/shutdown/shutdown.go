package shutdown

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/OinkiePie/calc_2/pkg/logger"
)

// Shutdownable определяет интерфейс для сервисов, которые могут быть остановлены.
type Shutdownable interface {
	Stop()
}

// WaitForShutdown ожидает сигнала завершения и корректно завершает работу сервиса.
// Принимает канал для ошибок, название сервиса и функцию остановки сервиса.
func WaitForShutdown(errChan <-chan error, serviceName string, service Shutdownable) {
	// Создаем канал для обработки сигналов завершения (Ctrl+C, SIGTERM)
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	// Ожидаем завершения работы серверов или возникновения ошибки
	select {
	// Ожидаем получения ошибки из канала errChan.
	case err := <-errChan:
		logger.Log.Fatalf("Фатальная ошибка в %s: %v", serviceName, err)
	// Ожидаем получения сигнала из канала sigint (сигнал завершения).
	case <-sigint:
		logger.Log.Debugf("Получен сигнал завершения для %s, начинаем остановку...", serviceName)
	}

	// GracefulShutdown выполняет корректное завершение работы сервиса.
	logger.Log.Infof("Остановка сервиса %s", serviceName)

	service.Stop()

	logger.Log.Infof("Сервис %s завершил работу", serviceName)
}
