package router

import (
	"github.com/OinkiePie/calc_2/web/internal/handlers"
	"github.com/gorilla/mux"
)

// NewWebRouter создает и настраивает новый роутер gorilla/mux для веб-сервера.
//
// Args:
//
//	static: string - Путь к директории со статическими файлами (HTML, CSS, JS, favicon).
//
// Returns:
//
//	*mux.Router - Указатель на созданный и настроенный роутер gorilla/mux.
func NewWebRouter(static string, addr string) *mux.Router {
	handler := handlers.NewWebHandlers(static, addr)
	router := mux.NewRouter()

	router.HandleFunc("/favicon.ico", handler.FaviconHandler)
	router.HandleFunc("/script.js", handler.ScriptHandler)
	router.HandleFunc("/style.css", handler.StyleHandler)

	router.HandleFunc("/api", handler.ApiHandler)
	// Обработчик для всех остальных путей, начиная с "/".
	// Это обеспечивает обслуживание SPA, где для всех путей возвращается index.html.
	router.PathPrefix("/").HandlerFunc(handler.IndexHandler)

	return router
}
