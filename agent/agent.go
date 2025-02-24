package agent

import (
	"fmt"
	"net/http"
)

// StartAgentServer запускает HTTP-сервер агента.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при работе сервера.
//	port: int - Порт, на котором будет запущен сервер агента.
//
// Returns:
//
//	*http.Server: Указатель на структуру http.Server, представляющую запущенный сервер агента.
//	              В случае ошибки при запуске сервера, в канал errChan будет отправлена ошибка.
func StartAgentServer(errChan chan error, port int) *http.Server {
	addr := fmt.Sprintf("localhost:%d", port)

	// Создаем экземпляр структуры http.Server, указывая адрес и обработчик
	srv := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello from Agent Server!")
		}),
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
