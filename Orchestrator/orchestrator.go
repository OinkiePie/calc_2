package orchestrator

import (
	"fmt"
	"net/http"

	"github.com/OinkiePie/calc_2/orchestrator/internal/router"
	"github.com/OinkiePie/calc_2/pkg/logger"
)

// StartOrchestratorServer запускает сервер оркестратора
func StartOrchestratorServer(port int) (*http.Server, error) {
	addr := fmt.Sprintf(":%d", port)
	router := router.NewRouter()

	logger.Log.Infof("Запуск сервера оркестратора на %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return nil, fmt.Errorf("ошибка при запуске сервера оркестратора: %w", err)
	}
	return srv, nil
}
