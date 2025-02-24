package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator/internal/models"
	"github.com/OinkiePie/calc_2/orchestrator/internal/task_manager"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/gorilla/mux"
)

// Handlers - структура для обработчиков запросов, зависит от TaskManager
type Handlers struct {
	taskManager *task_manager.TaskManager
}

// NewHandlers - конструктор для структуры Handlers
func NewOrchestratorHandlers(tm *task_manager.TaskManager) *Handlers {
	return &Handlers{taskManager: tm}
}

// AddExpressionHandler обрабатывает POST-запросы на эндпоинт /api/v1/calculate.
//
// Функция принимает JSON-запрос, содержащий математическое выражение в строковом формате,
// передает выражение в TaskManager для обработки и сохранения, и возвращает ID созданного выражения.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
//
// Request body (JSON):
//
//	{
//	  "expression": "строка с математическим выражением"
//	}
//
// Returns (JSON):
//
//	201 Created:
//	{
//	  "id": "уникальный ID созданного выражения"
//	}
//
//	400 Bad Request:
//	{
//	  "error": "Ошибка прочтение содержания запроса"
//	}
//
//	{
//	  "error": "Ошибка при декодировании JSON"
//	}
//
//	422 Unprocessable Entity:
//	{
//	  "error": "Выражения обязательно"
//	}
//
//	500 Internal Server Error:
//	{
//	  "error": "Ошибка при добавлении выражения в TaskManager"
//	}
func (h *Handlers) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest) // 400
		return
	}

	var requestBody models.ExpressionAdd

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest) // 400
		return
	}

	// Очищаем для проверки на пустоту и сохраняем в переменную для отправки в
	// *разбиватель на задачи* чтобы не делать это повторно
	trimmedBody := strings.TrimSpace(requestBody.Expression)
	if trimmedBody == "" {
		http.Error(w, "Expression is required", http.StatusUnprocessableEntity) // 422
		return
	}

	id, err := h.taskManager.AddExpression(trimmedBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(response)

	logger.Log.Debugf("AddExpressionHandler: выражение %s успешно создано", id)
}

// GetExpressionsHandler обрабатывает GET-запросы на эндпоинт /api/v1/expressions.
//
// Функция получает список всех выражений из TaskManager, преобразует их в формат ExpressionResponse
// и возвращает JSON-ответ со списком выражений.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
//
// Returns (JSON):
//
//	200 OK:
//	{
//	  "expressions": [
//	    {
//	      "id": "уникальный ID выражения",
//	      "status": "статус выражения (pending, processing, completed)",
//	      "result": "результат выражения (может отсутствовать, если вычисления не завершены)"
//	    },
//	    ...
//	  ]
//	}
//
//	500 Internal Server Error:
//	{
//	  "error": "Ошибка при кодировании ответа в JSON."
//	}
func (h *Handlers) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	expressionsMap := h.taskManager.GetExpressions()

	// Создаем слайс ExpressionResponse
	var expressionResponses []models.ExpressionResponse

	// Проходим по map и преобразуем Expression в ExpressionResponse
	for _, expression := range expressionsMap {
		expressionResponse := models.ExpressionResponse{
			ID:     expression.ID,
			Status: expression.Status,
			Result: expression.Result,
		}
		expressionResponses = append(expressionResponses, expressionResponse)
	}

	// Создаем map для ответа
	response := map[string][]models.ExpressionResponse{"expressions": expressionResponses}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response) // 200
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	logger.Log.Debugf("GetExpressionsHandler: список выражений успешно отправлен")
}

// GetExpressionHandler обрабатывает GET-запросы на эндпоинт /api/v1/expressions/{id}.
//
// Функция получает выражение по указанному ID из TaskManager, преобразует его в формат ExpressionResponse
// и возвращает JSON-ответ с информацией о выражении.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
//
// Path parameters:
//
//	id: ID выражения, которое нужно получить.
//
// Returns (JSON):
//
//	200 OK:
//	{
//	  "expression": {
//	    "id": "уникальный ID выражения",
//	    "status": "статус выражения (pending, processing, completed)",
//	    "result": "результат выражения (может отсутствовать, если вычисления не завершены)"
//	  }
//	}
//
//	404 Not Found:
//	{
//	  "error": "Выражение не найдено"
//	}
//
//	500 Internal Server Error:
//	{
//	  "error": "Ошибка при кодировании ответа в JSON"
//	}
func (h *Handlers) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	expression, ok := h.taskManager.GetExpression(id)
	if !ok {
		http.Error(w, "Expression not found", http.StatusNotFound) // 404
		return
	}

	expressionResponse := models.ExpressionResponse{
		ID:     expression.ID,
		Status: expression.Status,
		Result: expression.Result,
	}

	response := map[string]models.ExpressionResponse{"expression": expressionResponse}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response) // 200
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	logger.Log.Debugf("GetExpressionHandler: выражение %s успешно добавлено", id)
}

// GetTaskHandler обрабатывает GET-запросы на эндпоинт /internal/task.
//
// Функция получает задачу для выполнения из TaskManager и возвращает JSON-ответ с информацией о задаче.
// Этот эндпоинт предназначен для внутреннего использования агентом.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
//
// Returns (JSON):
//
//	200 OK:
//	{
//	  "task": {
//	    "id": "уникальный ID задачи",
//	    "operation": "операция, которую нужно выполнить (+, -, *, /, ^, u-)",
//	    "args": [], // 2 числа
//	    "operation_time": "время выполнения задачи",
//	  }
//	}
//
//	404 Not Found:
//	(пустой ответ) - Если нет доступных задач для выполнения
func (h *Handlers) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, _, ok := h.taskManager.GetTask()
	if !ok {
		w.WriteHeader(http.StatusNotFound) // 404
		return
	}

	response := map[string]models.TaskResponse{
		"task": {
			ID:             task.ID,
			Args:           task.Args,
			Operation:      task.Operation,
			Operation_time: task.Operation_time,
			Expression:     task.Expression,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	logger.Log.Debugf("GetTaskHandler: задача %s успешно отправлена", task.ID)

}

// GetTaskIDHandler обрабатывает GET-запросы на эндпоинт /internal/task/{id}.
//
// Функция получает список задач, связанных с определенным идентификатором, из TaskManager и возвращает
// их в формате JSON. Этот эндпоинт предназначен для отладки и проверки.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
//
// Path parameters:
//
//	id: ID, с которым связаны задачи (например, ID выражения).
//
// Returns (JSON):
//
//	200 OK:
//	{
//	  "tasks": [
//	    {
//	    "id": "уникальный ID задачи",
//	    "operation": "операция, которую нужно выполнить (+, -, *, /, ^, u-)",
//	    "args": [] (2 числа или nil'ы, если зависит от иногй задачи)"
//	    "operation_time": "время выполнения задачи",
//	  	}
//	    ...
//	  ]
//	}
//
//	404 OK:
//	{
//	    "error": "Задача с указанным ID не существует"
//	}
//
//	500 Internal Server Error:
//	{
//	  "error": "Ошибка при кодировании ответа в JSON."
//	}
func (h *Handlers) GetTaskIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	taskMap := h.taskManager.GetTasks(id)

	if len(taskMap) == 0 {
		http.Error(w, "task undefined", http.StatusNotFound) // 404
		return
	}

	response := map[string][]models.Task{"tasks": taskMap}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Log.Debugf("GetTasksHandler: задача %s успешно отправлена", id)
}

// CompleteTaskHandler обрабатывает POST-запросы на эндпоинт /internal/task.
//
// Функция принимает JSON-запрос с ID выполненной задачи и результатом ее выполнения,
// обновляет информацию о задаче в TaskManager. Этот эндпоинт предназначен для внутреннего использования агентами.
//
// Args:
//
//	w: http.ResponseWriter - интерфейс для записи HTTP-ответа.
//	r: *http.Request - указатель на структуру, представляющую HTTP-запрос.
//
// Request body (JSON):
//
//	{
//		"expression": "ID корневого выражения"
//		"id": "ID выполненной задачи",
//		"result": "результат выполнения задачи (число)"
//	}
//
// Returns:
//
//	200 OK:
//	- В случае успешного завершения. // пустой ответ
//
//	400 Bad Request:
//	{
//	  "error": "Ошибка при декодировании JSON"
//	}
//
//	404 Not Found:
//	{
//	  "error": "Задача не найдена"
//	}
func (h *Handlers) CompleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var requestBody models.TaskCompleted

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	success := h.taskManager.CompleteTask(requestBody.Expression, requestBody.ID, requestBody.Result)
	if !success {
		http.Error(w, "Task not found", http.StatusNotFound) // 404
		return
	}

	w.WriteHeader(http.StatusOK) // 200

	logger.Log.Debugf("CompleteTaskHandler: задача %s успешно выполнена", requestBody.ID)
}

// EnableCORS - добавляет заголовки CORS  для разрешения запросов с других доменов.
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Проверяем, есть ли origin в списке разрешенных
			allowed := false
			for _, allowedOrigin := range config.Cfg.CORS.AllowOrigin {
				if strings.EqualFold(origin, allowedOrigin) { //Сравнение без учета регистра
					allowed = true
					break
				}
			}
			if allowed {
				// Если origin разрешен, устанавливаем заголовок Access-Control-Allow-Origin
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Дополнительные заголовки CORS
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
			}
		}

		next.ServeHTTP(w, r)
	})
}
