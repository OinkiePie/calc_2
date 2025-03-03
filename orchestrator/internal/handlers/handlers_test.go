package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator/internal/handlers"
	"github.com/OinkiePie/calc_2/orchestrator/internal/task_manager"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Отключаем выводы и инициализируем конфиг
	log.SetOutput(io.Discard)
	config.InitConfig()
	logger.InitLogger(logger.Options{Level: 6})
}

func TestAddExpressionHandler(t *testing.T) {
	// Создаем мок для TaskManager
	h := handlers.NewOrchestratorHandlers(task_manager.NewTaskManager())

	t.Run("Successful", func(t *testing.T) {
		requestBody := map[string]string{"expression": "42 + 55"}
		jsonBody, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/calculate", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Bad JSON", func(t *testing.T) {
		requestBody := "bad body"
		jsonBody, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("Empty", func(t *testing.T) {
		requestBody := map[string]string{"expression": ""}
		jsonBody, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("When adding", func(t *testing.T) {
		requestBody := map[string]string{"expression": "+52+"}
		jsonBody, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestGetExpressionsHandler(t *testing.T) {
	// Создаем мок для TaskManager
	h := handlers.NewOrchestratorHandlers(task_manager.NewTaskManager())

	t.Run("Successful", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetExpressionsHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/calculate", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetExpressionsHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestGetExpressionHandler(t *testing.T) {
	// Создаем мок для TaskManager
	h := handlers.NewOrchestratorHandlers(task_manager.NewTaskManager())

	// Успешный запрос невозможно проверить т.к. он получает ID
	// из ссылки благодаря gorilla/mux.

	t.Run("Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/expressions/id42bratuha", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetExpressionHandler(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/expressions/id42bratuha", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetExpressionHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestGetTaskHandler(t *testing.T) {

	// Создаем мок для TaskManager
	h := handlers.NewOrchestratorHandlers(task_manager.NewTaskManager())

	t.Run("Succesful", func(t *testing.T) {
		// Добавляем выражение чтобы потом получать его задачу
		requestBody := map[string]string{"expression": "42+55"}
		jsonBody, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		req, err = http.NewRequest("GET", "/internal/task", nil)
		assert.NoError(t, err)

		rr = httptest.NewRecorder()
		h.GetTaskHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/internal/task", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetTaskHandler(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/internal/task", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetTaskHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestGetTaskIDHandler(t *testing.T) {
	// Создаем мок для TaskManager
	h := handlers.NewOrchestratorHandlers(task_manager.NewTaskManager())

	// Успешный запрос невозможно проверить т.к. он получает ID
	// из ссылки благодаря gorilla/mux.

	t.Run("Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/internal/taskid42bratuha", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetTaskIDHandler(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/internal/task/id42bratuha", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.GetTaskIDHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestCompleteTaskHandler(t *testing.T) {
	// Создаем мок для TaskManager
	h := handlers.NewOrchestratorHandlers(task_manager.NewTaskManager())

	t.Run("Successful", func(t *testing.T) {
		// Добавляем выражение в список выражений чтобы получить реальный ID и таск
		requestBody := map[string]string{"expression": "42+55"}
		jsonBody, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		body, err := io.ReadAll(rr.Body)
		assert.NoError(t, err)

		var expressionBody models.ExpressionResponse

		err = json.Unmarshal(body, &expressionBody)
		assert.NoError(t, err)
		// Сохраняем ID выражения
		expressionID := expressionBody.ID
		// Получаем задачу "для выполнения"
		req, err = http.NewRequest("GET", "/internal/task", nil)
		assert.NoError(t, err)

		rr = httptest.NewRecorder()
		h.GetTaskHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		body, err = io.ReadAll(rr.Body)
		assert.NoError(t, err)

		var taskBody models.TaskResponse

		err = json.Unmarshal(body, &taskBody)
		assert.NoError(t, err)
		// Сохраняем ID задачи
		taskID := taskBody.ID
		// "Выполняем задачу"
		taskCompleted := models.TaskCompleted{
			Expression: expressionID,
			ID:         taskID,
			Result:     97.0, // Посчитали сами
		}
		jsonBody, _ = json.Marshal(taskCompleted)
		req, err = http.NewRequest("POST", "/internal/task", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr = httptest.NewRecorder()
		h.CompleteTaskHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/internal/task", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		taskCompleted := models.TaskCompleted{
			Expression: "ExpressionFakeID42424242",
			ID:         "TaskFakeID42424242",
			Result:     97.0, // Посчитали сами
		}
		jsonBody, _ := json.Marshal(taskCompleted)
		req, err := http.NewRequest("POST", "/internal/task", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.CompleteTaskHandler(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("Bad JSON", func(t *testing.T) {
		requestBody := "bad body"
		jsonBody, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", "/internal/task", bytes.NewBuffer(jsonBody))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		h.AddExpressionHandler(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}
