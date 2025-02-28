package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/OinkiePie/calc_2/orchestrator/internal/task_manager"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/models"
	"github.com/gorilla/mux"
)

// Handlers - структура для обработчиков запросов, зависит от TaskManager
type Handlers struct {
	taskManager *task_manager.TaskManager
}

// NewOrchestratorHandlers - конструктор для структуры Handlers.
//
// Args:
//
//	tm: *task_manager.TaskManager - Указатель на экземпляр TaskManager.
//	    Необходимо передать уже инициализированный экземпляр TaskManager.
//
// Returns:
//
//	*Handlers - Указатель на новый экземпляр структуры Handlers.
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
// Responses:
//
//	201 Created:
//	{
//	  "id": "уникальный ID созданного выражения"
//	}
//
//	42 Unprocessable Entity:
//	{
//	  "error": "Ошибка прочтение содержания запроса"
//	}
//
//	{
//	  "error": "Ошибка при декодировании JSON"
//	}
//
//	{
//	  "error": "Выражения обязательно"
//	}
//
// 500 Internal Server Error:
//
//	{
//	  "error": "Ошибка при добавлении выражения в TaskManager"
//	}
func (h *Handlers) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnprocessableEntity, "не удалось прочитать запрос") //422
		return
	}

	var requestBody models.ExpressionAdd

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnprocessableEntity, "не удалось декодировать JSON") //422
		return
	}

	// Очищаем для проверки на пустоту и сохраняем в переменную для отправки в
	// *разбиватель на задачи* чтобы не делать это повторно
	trimmedBody := strings.TrimSpace(requestBody.Expression)
	if trimmedBody == "" {
		h.writeErrorResponse(w, http.StatusUnprocessableEntity, "выражения обязательно") //422
		return
	}

	id, err := h.taskManager.AddExpression(trimmedBody)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error()) //500
		return
	}

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(response)

	logger.Log.Debugf("Выражение %s успешно создано", id)
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
// Responses:
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
			Error:  expression.Error,
		}
		expressionResponses = append(expressionResponses, expressionResponse)
	}

	// Создаем map для ответа
	response := map[string][]models.ExpressionResponse{"expressions": expressionResponses}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response) // 200
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500
		return
	}

	logger.Log.Debugf("Список выражений успешно отправлен")
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
// Responses:
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
		h.writeErrorResponse(w, http.StatusNotFound, "выражение не найдено") // 404
		return
	}

	expressionResponse := models.ExpressionResponse{
		ID:     expression.ID,
		Status: expression.Status,
		Result: expression.Result,
		Error:  expression.Error,
	}

	response := map[string]models.ExpressionResponse{"expression": expressionResponse}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response) // 200
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500
		return
	}

	logger.Log.Debugf("Выражение %s успешно добавлено", id)
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
// Responses:
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

	response := models.TaskResponse{
		ID:             task.ID,
		Args:           task.Args,
		Operation:      task.Operation,
		Operation_time: task.Operation_time,
		Expression:     task.Expression,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	logger.Log.Debugf("Задача %s успешно отправлена", task.ID)
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
// Responses:
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
		h.writeErrorResponse(w, http.StatusNotFound, "задача не найдена") // 404
		return
	}

	response := map[string][]models.Task{"tasks": taskMap}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500
		return
	}

	logger.Log.Debugf("Задача %s успешно отправлена", id)
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
// Responses:
//
//	200 OK:
//	(пустой ответ) - В случае успешного завершения.
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
		h.writeErrorResponse(w, http.StatusBadRequest, "не удалось прочитать тело запроса") // 400
		return
	}

	var requestBody models.TaskCompleted

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "не удалось декодировать JSON") // 400
		return
	}

	success := h.taskManager.CompleteTask(requestBody.Expression, requestBody.ID, requestBody.Error, requestBody.Result)
	if !success {
		h.writeErrorResponse(w, http.StatusNotFound, "Задача не найдена") // 404
		return
	}

	w.WriteHeader(http.StatusOK) // 200

	logger.Log.Debugf("Задача %s успешно выполнена", requestBody.ID)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handlers) writeErrorResponse(w http.ResponseWriter, statusCode int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResponse := ErrorResponse{Error: err}
	if encodeErr := json.NewEncoder(w).Encode(errResponse); encodeErr != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError) // Крайний случай
	}
}
