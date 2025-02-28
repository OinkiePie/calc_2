package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
)

// Handlers представляет структуру, содержащую методы-обработчики для веб-запросов.
// Хранит путь к директории со статическими файлами.
type Handlers struct {
	staticDir string
	port      int
}

// NewWebHandlers создает новый экземпляр структуры Handlers и инициализирует поле StaticDir.
//
// Args:
//
//	static: string - Путь к директории со статическими файлами.
//
// Returns:
//
//	*Handlers - Указатель на созданный экземпляр структуры Handlers.
func NewWebHandlers(static string, port int) *Handlers {
	return &Handlers{staticDir: static, port: port}
}

// IndexHandler обрабатывает запросы к корневому пути ("/") и возвращает файл index.html.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
func (h *Handlers) IndexHandler(w http.ResponseWriter, r *http.Request) {
	indexFilePath := filepath.Join(h.staticDir, "index.html") // Полный путь к index.html
	http.ServeFile(w, r, indexFilePath)
}

// ScriptHandler обрабатывает запросы к пути "/script.js" и возвращает файл script.js.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
func (h *Handlers) ScriptHandler(w http.ResponseWriter, r *http.Request) {
	scriptFilePath := filepath.Join(h.staticDir, "script.js") // Полный путь к script.js
	http.ServeFile(w, r, scriptFilePath)
}

// StyleHandler обрабатывает запросы к пути "/style.css" и возвращает файл style.css.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
func (h *Handlers) StyleHandler(w http.ResponseWriter, r *http.Request) {
	styleFilePath := filepath.Join(h.staticDir, "style.css") // Полный путь к style.css
	http.ServeFile(w, r, styleFilePath)
}

// FaviconHandler обрабатывает запросы к пути "/favicon.ico" и возвращает файл favicon.ico.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
func (h *Handlers) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	faviconFilePath := filepath.Join(h.staticDir, "favicon.ico") // Полный путь к favicon.ico
	http.ServeFile(w, r, faviconFilePath)
}

// FaviconHandler обрабатывает запросы к пути "/api" и порт сервиса оркестратора
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
func (h *Handlers) ApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, h.port)
}
