package orchestrator

import (
	"fmt"
	"net/http"

	"github.com/OinkiePie/calc_2/orchestrator/internal/router"
)

// StartOrchestratorServer запускает HTTP-сервер оркестратора.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при работе сервера.
//	port: int - Порт, на котором будет запущен сервер оркестратора.
//
// Returns:
//
//	*http.Server: Указатель на структуру http.Server, представляющую запущенный сервер агента.
//	              В случае ошибки при запуске сервера, в канал errChan будет отправлена ошибка.
func StartOrchestratorServer(errChan chan error, port int) *http.Server {
	addr := fmt.Sprintf("localhost:%d", port)
	router := router.NewOrchestratorRouter()

	// Создаем экземпляр структуры http.Server, указывая адрес и обработчик
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Запускаем сервер в отдельной горутине, чтобы не блокировать основной поток выполнения.
	go func() {
		// Запускаем прослушивание входящих соединений на указанном адресе.
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Если при запуске сервера произошла ошибка, отправляем её в канал ошибок.
			errChan <- err
		}
	}()

	// Возвращаем указатель на созданный и запущенный сервер.
	return srv
}
