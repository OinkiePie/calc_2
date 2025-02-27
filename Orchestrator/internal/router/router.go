package router

import (
	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator/internal/handlers"
	"github.com/OinkiePie/calc_2/orchestrator/internal/middlewares"
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

	middleware := middlewares.NewOrchestratorMiddlewares(config.Cfg.Middleware.ApiKeyPrefix, config.Cfg.Middleware.Authorization, config.DefaultConfig().Middleware.AllowOrigin)

	router := mux.NewRouter()

	// API endpoints (внешние конечные точки, доступные клиентам)
	router.HandleFunc("/api/v1/calculate", handler.AddExpressionHandler).Methods("POST")
	router.HandleFunc("/api/v1/expressions", handler.GetExpressionsHandler).Methods("GET")
	router.HandleFunc("/api/v1/expressions/{id}", handler.GetExpressionHandler).Methods("GET")

	// Internal endpoints (внутренние конечные точки, используемые агентом)
	// Подмаршрутизатор для Internal endpoints
	internalRouter := router.PathPrefix("/internal").Subrouter()
	internalRouter.Use(middleware.EnableAuthorization) // Применяем аутентификацию

	internalRouter.HandleFunc("/task", handler.GetTaskHandler).Methods("GET")
	internalRouter.HandleFunc("/task", handler.CompleteTaskHandler).Methods("POST")

	// Debug endpoints (конечные точки, используемые только для отладки)
	internalRouter.HandleFunc("/task/{id}", handler.GetTaskIDHandler).Methods("GET")

	router.Use(middleware.EnableCORS)

	return router
}
