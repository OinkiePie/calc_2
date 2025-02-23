package router

import (
	"github.com/OinkiePie/calc_2/orchestrator/internal/handlers"
	"github.com/OinkiePie/calc_2/orchestrator/internal/task_manager"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	taskManager := task_manager.NewTaskManager()
	handler := handlers.NewHandlers(taskManager)

	router := mux.NewRouter()

	// API endpoints
	router.HandleFunc("/api/v1/calculate", handler.AddExpressionHandler).Methods("POST")
	router.HandleFunc("/api/v1/expressions", handler.GetExpressionsHandler).Methods("GET")
	router.HandleFunc("/api/v1/expressions/{id}", handler.GetExpressionHandler).Methods("GET")

	// Internal endpoints (for agent)
	router.HandleFunc("/internal/task", handler.GetTaskHandler).Methods("GET")
	router.HandleFunc("/internal/task", handler.CompleteTaskHandler).Methods("POST")

	// debug
	router.HandleFunc("/internal/task/{id}", handler.GetTaskIDHandler).Methods("GET")

	router.Use(handlers.EnableCORS)

	return router
}
