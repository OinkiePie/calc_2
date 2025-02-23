package agent

import (
	"fmt"
	"net/http"

	"github.com/OinkiePie/calc_2/pkg/logger"
)

// StartAgentServer запускает сервер агента
func StartAgentServer(port int) (*http.Server, error) {
	addr := fmt.Sprintf(":%d", port)

	logger.Log.Infof("Запуск сервера агента на %s", addr)

	srv := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello from Agent Server!")
		}),
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return nil, fmt.Errorf("ошибка при запуске сервера агента: %w", err)
	}
	return srv, nil
}
