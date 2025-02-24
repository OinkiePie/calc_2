package router

import (
	"github.com/OinkiePie/calc_2/orchestrator/internal/handlers"
	"github.com/OinkiePie/calc_2/orchestrator/internal/task_manager"
	"github.com/gorilla/mux"
)

// NewOrchestratorRouter создает и настраивает роутер для API оркестратора, используя gorilla/mux.
// Он определяет маршруты для обработки запросов к различным конечным точкам оркестратора.
//
// Args:
//
//	(None): Функция не принимает аргументов.
//
// Returns:
//
//	*mux.Router: Указатель на созданный и настроенный роутер.
func NewOrchestratorRouter() *mux.Router {
	taskManager := task_manager.NewTaskManager()
	handler := handlers.NewOrchestratorHandlers(taskManager)

	router := mux.NewRouter()

	// API endpoints (внешние конечные точки, доступные клиентам)
	router.HandleFunc("/api/v1/calculate", handler.AddExpressionHandler).Methods("POST")
	router.HandleFunc("/api/v1/expressions", handler.GetExpressionsHandler).Methods("GET")
	router.HandleFunc("/api/v1/expressions/{id}", handler.GetExpressionHandler).Methods("GET")

	// Internal endpoints (внутренние конечные точки, используемые агентом)
	router.HandleFunc("/internal/task", handler.GetTaskHandler).Methods("GET")
	router.HandleFunc("/internal/task", handler.CompleteTaskHandler).Methods("POST")

	// Debug endpoints (конечные точки, используемые только для отладки)
	router.HandleFunc("/internal/task/{id}", handler.GetTaskIDHandler).Methods("GET")

	router.Use(handlers.EnableCORS)

	return router
}
