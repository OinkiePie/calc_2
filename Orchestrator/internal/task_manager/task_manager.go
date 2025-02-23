package task_manager

import (
	"fmt"
	"sync"

	"github.com/OinkiePie/calc_2/orchestrator/internal/models"
	"github.com/OinkiePie/calc_2/orchestrator/internal/task_splitter"
	"github.com/google/uuid"
)

// TaskManager - структура, управляющая списком выражений и задачами
type TaskManager struct {
	expressions   map[string]models.Expression
	expressionsMu sync.RWMutex // Mutex для защиты map от конкурентного доступа
}

// NewTaskManager - конструктор для TaskManager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		expressions: make(map[string]models.Expression),
	}
}

// AddExpression - добавляет новое выражение
func (tm *TaskManager) AddExpression(expressionString string) (string, error) {
	tm.expressionsMu.Lock()
	defer tm.expressionsMu.Unlock()

	id := uuid.New().String()

	tasks, err := task_splitter.ParseExpression(id, expressionString)
	if err != nil {
		return "", err
	}

	expression := models.Expression{
		ID:               id,
		Status:           "pending",
		Result:           nil,
		Tasks:            tasks,
		ExpressionString: expressionString,
	}

	tm.expressions[id] = expression

	fmt.Println("Added expression with ID:", id)

	return id, nil
}

// GetExpressions - возвращает список всех выражений
func (tm *TaskManager) GetExpressions() []models.Expression {
	tm.expressionsMu.RLock()
	defer tm.expressionsMu.RUnlock()

	expressionsList := make([]models.Expression, 0, len(tm.expressions))
	for _, expression := range tm.expressions {
		expressionsList = append(expressionsList, expression)
	}

	return expressionsList
}

// GetExpression - возвращает выражение по ID
func (tm *TaskManager) GetExpression(id string) (models.Expression, bool) {
	tm.expressionsMu.Lock()
	defer tm.expressionsMu.Unlock()

	expression, ok := tm.expressions[id]

	for _, task := range expression.Tasks {
		if task.Status != "completed" {
			return expression, ok
		}
	}
	// Удаляем из списка выражений выполненое
	delete(tm.expressions, id)

	expression.Status = "completed"
	// Сплиттер разделяет задачи так что в конце будет находиться последня операция
	// Если задача имеет зависимости, она будет корневым элементом
	expression.Result = expression.Tasks[len(expression.Tasks)-1].Result

	return expression, ok
}

// GetTasks - возвращает список всех задачь
func (tm *TaskManager) GetTasks(id string) []models.Task {
	tm.expressionsMu.RLock()
	defer tm.expressionsMu.RUnlock()

	expression := models.Expression{}
	for exprID, expr := range tm.expressions {
		if exprID == id {
			expression = expr
		}
	}

	if expression.ID == "" {
		return []models.Task{}
	}

	return expression.Tasks
}

// GetTask - возвращает первую задачу со статусом "pending"
func (tm *TaskManager) GetTask() (models.Task, string, bool) {
	// Получаем блокировку для чтения, чтобы разрешить параллельное чтение выражений.
	tm.expressionsMu.RLock()
	defer tm.expressionsMu.RUnlock() // Не забываем снять блокировку после завершения функции.

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

	for exprID, expr := range tm.expressions {
		if expr.Status == "pending" || expr.Status == "processing" {
			wg.Add(1)
			// Запускаем горутину для обработки текущего выражения.
			go func(exprID string, expr models.Expression) {
				defer wg.Done()

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

						fmt.Println(tm.areDependenciesCompleted(expr.Tasks, task.Dependencies))

						// Проверяем, выполнены ли все зависимости.
						if tm.areDependenciesCompleted(expr.Tasks, task.Dependencies) {
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

							if entry, ok := tm.expressions[exprID]; ok { // Устанавливаем для задачи над таском которой работаем статус "processing"
								entry.Status = "processing"
								tm.expressions[exprID] = entry
							}
							task := *task // Обновляем задачу в выражении
							fmt.Println(task)
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

	wg.Wait()
	close(taskChan)

	// Получаем результат из канала.
	for result := range taskChan {
		// Если задача найдена, сохраняем результат и выходим из цикла.
		if result.found {
			foundTask = result.task
			foundExprID = result.exprID
			found = true
			break // Важно: выходим из цикла, чтобы взять первый результат!
		}
	}

	// Возвращаем результаты.
	return foundTask, foundExprID, found
}

// areDependenciesCompleted проверяет, выполнены ли все зависимости задачи.
// Возвращает: bool: true, если все зависимости выполнены, false в противном случае.
func (tm *TaskManager) areDependenciesCompleted(tasks []models.Task, dependencies []string) bool {
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

// CompleteTask - обновляет статус и результат задачи
func (tm *TaskManager) CompleteTask(expressionID string, taskID string, result float64) bool {
	tm.expressionsMu.Lock()
	defer tm.expressionsMu.Unlock()

	expr, ok := tm.expressions[expressionID]
	if !ok {
		return false
	}

	for i, task := range expr.Tasks {
		if task.ID == taskID {
			res := result
			expr.Tasks[i].Result = &res
			expr.Tasks[i].Status = "completed"
			tm.expressions[expressionID] = expr
			fmt.Printf("Task %s completed for expression %s with result: %f\n", taskID, expressionID, result)
			return true
		}
	}

	return false
}
