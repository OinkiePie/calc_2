package task_manager

import (
	"sync"

	"github.com/OinkiePie/calc_2/orchestrator/internal/task_splitter"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/models"
	"github.com/google/uuid"
)

// TaskManager - структура, управляющая списком выражений и задачами.
type TaskManager struct {
	// expressions - Хранилище выражений, где ключ - ID выражения, значение - структура Expression.
	expressions map[string]models.Expression
	// expressionsMu - Mutex для защиты map от конкурентного доступа (чтения и записи).
	expressionsMu sync.RWMutex
}

// NewTaskManager - конструктор для TaskManager. Создает и возвращает новый экземпляр TaskManager.
//
// Args:
//
//	(None) - Функция не принимает аргументов.
//
// Returns:
//
//	*TaskManager - Указатель на новый экземпляр TaskManager.
func NewTaskManager() *TaskManager {
	// Инициализируем map для хранения выражений.
	return &TaskManager{
		expressions: make(map[string]models.Expression),
	}
}

// AddExpression - добавляет новое выражение в TaskManager.
//
// Args:
//
//	expressionString: string - Строка, представляющая арифметическое выражение.
//
// Returns:
//
//	string - ID добавленного выражения.
//	error - Ошибка, если не удалось добавить выражение.
func (tm *TaskManager) AddExpression(expressionString string) (string, error) {
	tm.expressionsMu.Lock()
	defer tm.expressionsMu.Unlock()

	// Генерируем уникальный ID для выражения.
	id := uuid.New().String()

	// Разбираем выражение на задачи с помощью task_splitter.ParseExpression.
	tasks, err := task_splitter.ParseExpression(id, expressionString)
	if err != nil {
		return "", err
	}
	// Создаем структуру Expression.
	expression := models.Expression{
		ID:               id,
		Status:           "pending",
		Result:           nil,
		Tasks:            tasks,
		ExpressionString: expressionString,
	}

	// Добавляем выражение в map выражений.
	tm.expressions[id] = expression

	return id, nil
}

// GetExpressions - возвращает список всех выражений, хранящихся в TaskManager.
//
// Args:
//
//	(None): Функция не принимает аргументов.
//
// Returns:
//
//	[]models.Expression - Срез всех выражений, хранящихся в TaskManager.
func (tm *TaskManager) GetExpressions() []models.Expression {
	tm.expressionsMu.RLock()
	defer tm.expressionsMu.RUnlock()

	// Создаем срез для хранения выражений.
	expressionsList := make([]models.Expression, 0, len(tm.expressions))
	// Копируем все выражения из map в срез.
	for _, expression := range tm.expressions {
		expressionsList = append(expressionsList, expression)
	}

	return expressionsList
}

// GetExpression - возвращает выражение из TaskManager по его ID.
//
// Args:
//
//	id: string - ID выражения, которое необходимо получить.
//
// Returns:
//
//	models.Expression: Выражение с указанным ID.
//	bool: true, если выражение найдено, иначе false.
func (tm *TaskManager) GetExpression(id string) (models.Expression, bool) {
	tm.expressionsMu.Lock()
	defer tm.expressionsMu.Unlock()

	// Получаем выражение из map.
	expression, ok := tm.expressions[id]

	if !ok {
		return models.Expression{}, false
	}

	//Проверяем выполнена ли задача
	if expression.Status != "completed" {
		return expression, true
	}
	// Если выражение забирается пользователем удаляем из списка ожидающих
	delete(tm.expressions, id)

	return expression, true
}

// GetTasks - возвращает список всех задач для заданного выражения.
//
// Args:
//
//	id: string - ID выражения, для которого необходимо получить задачи.
//
// Returns:
//
//	[]models.Task - Срез всех задач для указанного выражения. Если выражение не найдено, возвращается пустой срез.
func (tm *TaskManager) GetTasks(id string) []models.Task {
	tm.expressionsMu.RLock()
	defer tm.expressionsMu.RUnlock()

	expression := models.Expression{}
	// Ищем выражение по ID в map.
	for exprID, expr := range tm.expressions {
		if exprID == id {
			expression = expr
		}
	}

	// Если выражение не найдено, возвращаем пустой срез.
	if expression.ID == "" {
		return []models.Task{}
	}

	return expression.Tasks
}

// GetTask - возвращает первую задачу со статусом "pending".
//
// Args:
//
//	(None) - Функция не принимает аргументов.
//
// Returns:
//
//	models.Task - Первая задача со статусом "pending" (в момент отправки присвоится "processing"). Если таких задач нет, возвращается пустая задача.
//	string - ID выражения, которому принадлежит найденная задача. Если задача не найдена, возвращается пустая строка.
//	bool - true, если задача найдена, иначе false.
func (tm *TaskManager) GetTask() (models.Task, string, bool) {
	// Устанавливаем блокировку для чтения, чтобы разрешить параллельное чтение выражений.
	tm.expressionsMu.Lock()
	defer tm.expressionsMu.Unlock()

	// Объявляем переменные для хранения результатов и синхронизации.
	var (
		foundTask   models.Task          // Найденная задача.
		foundExprID string               // ID выражения, которому принадлежит задача.
		found       bool                 // Флаг, указывающий, найдена ли задача.
		wg          sync.WaitGroup       // WaitGroup для ожидания завершения всех горутин.
		taskChan    = make(chan struct { // Канал для передачи результатов из горутин.
			task   models.Task
			exprID string
			found  bool
		}, len(tm.expressions)) // Буферизованный канал, размер которого равен количеству выражений. Это предотвращает блокировку горутин при отправке результатов.
	)

	// Итерируемся по всем выражениям.
	for exprID, expr := range tm.expressions {
		if expr.Status == "pending" || expr.Status == "processing" {
			wg.Add(1)
			// Запускаем горутину для обработки текущего выражения.
			go func(exprID string, expr models.Expression) {
				defer wg.Done()

				// Итерируемся по задачам в выражении.
				for i := range expr.Tasks {
					// Получаем указатель на текущую задачу
					task := &expr.Tasks[i]
					if task.Status == "pending" {
						// Проверяем наличие зависимостей.
						if task.Dependencies[0] == "" && task.Dependencies[1] == "" {
							// Задача без зависимостей готова к выполнению.
							task.Status = "processing"                   // Устанавливаем статус "processing"
							if entry, ok := tm.expressions[exprID]; ok { // Устанавливаем для задачи над таском которой работаем статус "processing"
								entry.Status = "processing"
								tm.expressions[exprID] = entry
							}
							task := *task        // Обновляем задачу в выражении
							taskChan <- struct { // Отправляем результат в канал
								task   models.Task
								exprID string
								found  bool
							}{task: task, exprID: exprID, found: true}
							return // Завершаем горутину, так как задача найдена.
						}

						// Проверяем, выполнены ли все зависимости.
						if tm.AreDependenciesCompleted(expr.Tasks, task.Dependencies) {
							// Все зависимости выполнены, задача готова к выполнению
							task.Status = "processing"      // Устанавливаем статус "processing"
							for i, arg := range task.Args { // Если значение nil, то оно находится в зависимостях
								if arg == nil {
									for _, dependency := range tm.expressions[exprID].Tasks {
										if dependency.ID == task.Dependencies[i] {
											task.Args[i] = dependency.Result
										}
									}
								}
							}

							// Устанавливаем для задачи над таском которой работаем статус "processing"
							if entry, ok := tm.expressions[exprID]; ok {
								entry.Status = "processing"
								tm.expressions[exprID] = entry
							}
							task := *task        // Обновляем задачу в выражении
							taskChan <- struct { // Отправляем результат в канал
								task   models.Task
								exprID string
								found  bool
							}{task: task, exprID: exprID, found: true}
							return // Завершаем горутину, так как задача найдена.
						}
					}
				}
			}(exprID, expr) // Передаем значения в горутину.
		}
	}

	// Ожидаем завершения всех горутин.
	wg.Wait()
	// Закрываем канал, чтобы сообщить получателям, что больше не будет данных.
	close(taskChan)

	// Получаем результат из канала.
	for result := range taskChan {
		// Если задача найдена, сохраняем результат и выходим из цикла.
		if result.found {
			foundTask = result.task
			foundExprID = result.exprID
			found = true
			break // Прекращаем поиск после нахождения первой задачи.
		}
	}

	// Возвращаем найденную задачу, ID выражения и флаг, указывающий, была ли задача найдена.
	return foundTask, foundExprID, found
}

// areDependenciesCompleted проверяет, выполнены ли все зависимости задачи.
//
// Args:
//
//	tasks: []models.Task - Список задач в выражении.
//	dependencies: []string - Список ID задач, от которых зависит текущая задача.
//
// Returns:
//
//	bool - true, если все зависимости выполнены, false в противном случае.
func (tm *TaskManager) AreDependenciesCompleted(tasks []models.Task, dependencies []string) bool {
	for _, dependencyID := range dependencies {
		if dependencyID == "" {
			break
		}
		found := false
		for _, task := range tasks {
			// Проверяем, совпадает ли ID задачи с ID зависимости.
			if task.ID == dependencyID {
				found = true // Зависимость найдена.
				// Если зависимость не выполнена, возвращаем false.
				if task.Status != "completed" {
					return false
				}
				break // Зависимость найдена и проверена, переходим к следующей.
			}
		}
		if !found {
			return false // Зависимость не найдена в списке задач
		}
	}
	// Все зависимости выполнены.
	return true
}

// CompleteTask - обновляет статус и результат задачи. Если все задачи
// выполняются присваивает выражению статус completed.
//
// Args:
//
//	expressionID: string - ID выражения, которому принадлежит задача.
//	taskID: string - ID задачи, которую необходимо завершить.
//	result: float64 - Результат выполнения задачи.
//
// Returns:
//
//	bool - true, если задача успешно завершена и обновлена, false в противном случае.
func (tm *TaskManager) CompleteTask(expressionID, taskID, taskErr string, result float64) bool {
	tm.expressionsMu.Lock()
	defer tm.expressionsMu.Unlock()

	// Пытаемся получить выражение по ID.
	expr, ok := tm.expressions[expressionID]
	if !ok {
		return false
	}

	// Проверяем выполнима ли задача
	if taskErr != "" {
		tm.impossibleTask(expressionID, taskErr)
		return true
	}

	// Итерируемся по задачам в выражении.
	for i, task := range expr.Tasks {
		// Ищем задачу с соответствующим ID.
		if task.ID == taskID {
			// Обновляем результат и статус задачи.
			res := result // Создаем копию результата, чтобы взять указатель на неё.
			expr.Tasks[i].Result = &res
			expr.Tasks[i].Status = "completed"
			// Проверяем все ли задачи выполнены.
			allCompleted := true
			for _, task := range expr.Tasks {
				if task.Status != "completed" {
					allCompleted = false
					break // Нашли незавершенную задачу, дальше проверять нет смысла.
				}
			}
			// Присваиваем статус completed.
			if allCompleted {
				expr.Status = "completed"
				//Меняем статус выражение на "completed"
				// Сплиттер разделяет задачи так, что в конце будет находиться последня операция.
				// Если задача имеет зависимости, она будет корневым элементом
				expr.Result = expr.Tasks[len(expr.Tasks)-1].Result
			}

			tm.expressions[expressionID] = expr // Обновляем выражение в map.
			return true
		}
	}

	//Если задача не найдена возвращаем false
	return false
}

// impossibleTask помечает выражение как невозможное для выполнения,
// сохраняя информацию о задаче, которая привела к невозможности выполнения.
//
// Args:
//
//	expressionID: string - ID выражения, которое невозможно выполнить.
//	taskID: string - ID задачи, ставшее ошибкой.
func (tm *TaskManager) impossibleTask(expressionID, taskErr string) {
	expr := tm.expressions[expressionID]
	expr.Status = "error"
	expr.Error = taskErr
	tm.expressions[expressionID] = expr
	logger.Log.Debugf("Выражение %s невозможно выполнить: %s", expressionID, taskErr)
}
