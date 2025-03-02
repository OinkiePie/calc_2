package router_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/OinkiePie/calc_2/web/internal/router"
	"github.com/stretchr/testify/assert"
)

func createTempFiles(t *testing.T) string {
	// Создаем временную директорию
	tempDir := t.TempDir()
	// Создаем файл favicon.ico
	faviconPath := filepath.Join(tempDir, "favicon.ico")
	err := os.WriteFile(faviconPath, []byte("favicon content"), 0644)
	if err != nil {
		t.Fatalf("Ошибка при создании favicon.ico: %v", err)
	}

	// Создаем файл index.html
	indexPath := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(indexPath, []byte("<html><body>respect :O</body></html>"), 0644)
	if err != nil {
		t.Fatalf("Ошибка при создании index.html: %v", err)
	}

	// Создаем файл style.css
	stylePath := filepath.Join(tempDir, "style.css")
	err = os.WriteFile(stylePath, []byte("body { background-color: #555555; }"), 0644)
	if err != nil {
		t.Fatalf("Ошибка при создании style.css: %v", err)
	}

	// Создаем файл script.js
	scriptPath := filepath.Join(tempDir, "script.js")
	err = os.WriteFile(scriptPath, []byte("console.log('Hello, yasha lava!');"), 0644)
	if err != nil {
		t.Fatalf("Ошибка при создании script.js: %v", err)
	}

	return tempDir
}

func TestWebRouter(t *testing.T) {
	// Создаем роутер

	router := router.NewWebRouter(createTempFiles(t), "localhost:8080")
	// Тест для /favicon.ico
	t.Run("FaviconHandler", func(t *testing.T) {

		req, err := http.NewRequest("GET", "/favicon.ico", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	// Тест для /script.js
	t.Run("ScriptHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/script.js", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	// Тест для /style.css
	t.Run("StyleHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/style.css", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	// Тест для /api
	t.Run("ApiHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	// Тест для корневого пути "/"
	t.Run("IndexHandler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	// Тест для остальных путей
	t.Run("IndexHandlerOther", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/any/another/adress", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMovedPermanently, rr.Code)
	})
}
