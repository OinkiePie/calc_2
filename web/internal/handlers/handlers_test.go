package handlers_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/web/internal/handlers"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func TestIndexHandler(t *testing.T) {
	// Создаем временную директорию и файл index.html
	tempDir := t.TempDir()
	indexFilePath := filepath.Join(tempDir, "index.html")
	err := os.WriteFile(indexFilePath, []byte("<html><body>Hello, ivan zolo!</body></html>"), 0644)
	assert.NoError(t, err)

	h := handlers.NewWebHandlers(tempDir)

	// Тест для корневого пути
	t.Run("RootPath", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.IndexHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "<html><body>Hello, ivan zolo!</body></html>", rr.Body.String())
	})

	// Тест для не корневого пути (редирект)
	t.Run("NonRootPath", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/ya/ystal/boss", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.IndexHandler(rr, req)

		assert.Equal(t, http.StatusMovedPermanently, rr.Code)
		assert.Equal(t, "/", rr.Header().Get("Location"))
	})
}

func TestScriptHandler(t *testing.T) {
	// Создаем временную директорию и файл script.js
	tempDir := t.TempDir()
	scriptFilePath := filepath.Join(tempDir, "script.js")
	err := os.WriteFile(scriptFilePath, []byte("console.log('mojet v sud podat!');"), 0644)
	assert.NoError(t, err)

	h := handlers.NewWebHandlers(tempDir)

	req, err := http.NewRequest("GET", "/script.js", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	h.ScriptHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "console.log('mojet v sud podat!');", rr.Body.String())
}

func TestStyleHandler(t *testing.T) {
	// Создаем временную директорию и файл style.css
	tempDir := t.TempDir()
	styleFilePath := filepath.Join(tempDir, "style.css")
	err := os.WriteFile(styleFilePath, []byte("body { background-color: #424242; }"), 0644)
	assert.NoError(t, err)

	h := handlers.NewWebHandlers(tempDir)

	req, err := http.NewRequest("GET", "/style.css", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	h.StyleHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "body { background-color: #424242; }", rr.Body.String())
}

func TestFaviconHandler(t *testing.T) {
	// Создаем временную директорию и файл favicon.ico
	tempDir := t.TempDir()
	faviconFilePath := filepath.Join(tempDir, "favicon.ico")
	err := os.WriteFile(faviconFilePath, []byte("favicon content"), 0644)
	assert.NoError(t, err)

	h := handlers.NewWebHandlers(tempDir)

	req, err := http.NewRequest("GET", "/favicon.ico", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	h.FaviconHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "favicon content", rr.Body.String())
}

func TestApiHandler(t *testing.T) {
	err := os.Setenv("ADDR_ORCHESTRATOR", "0.1.0.1")
	assert.NoError(t, err)
	err = os.Setenv("PORT_ORCHESTRATOR", "6666")
	assert.NoError(t, err)

	// Отключаем конфиг
	err = os.Setenv("APP_CFG", "CFG_FALSE")
	assert.NoError(t, err)

	err = config.InitConfig()
	assert.NoError(t, err)

	h := handlers.NewWebHandlers("")

	req, err := http.NewRequest("GET", "/api", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	h.ApiHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fmt.Sprintf("%s:%d\n", config.Cfg.Server.Orchestrator.ADDR_ORCHESTRATOR, config.Cfg.Server.Orchestrator.PORT_ORCHESTRATOR), rr.Body.String())
}
