package web

import (
	"fmt"
	"net/http"
	"os"

	"github.com/OinkiePie/calc_2/web/internal/router"
)

// StartWebServer запускает HTTP-сервер для обслуживания статических файлов веб-приложения.
//
// Args:
//
//	errChan: chan error - Канал для отправки ошибок, возникающих при работе сервера.
//	port: int - Порт, на котором будет запущен веб-сервер.
//	staticDir: string - Путь к директории, содержащей статические файлы (HTML, CSS, JavaScript, favicon).
//
// Returns:
//
//	*http.Server: Указатель на структуру http.Server, представляющую запущенный сервер агента.
//	              В случае ошибки при запуске сервера, в канал errChan будет отправлена ошибка.
func StartWebServer(errChan chan error, port int, staticDir string) *http.Server {
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		fmt.Println(staticDir)
		errChan <- fmt.Errorf("директория со статическими файлами не найдена")
	}

	addr := fmt.Sprintf("localhost:%d", port)
	router := router.NewWebRouter(staticDir)

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
