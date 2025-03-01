package task_manager_test

import (
	"io"
	"log"
	"testing"

	"github.com/OinkiePie/calc_2/config"
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

// TestNewTaskManager проверяет создание нового TaskManager.
func TestNewTaskManager(t *testing.T) {
	tm := task_manager.NewTaskManager()
	assert.NotNil(t, tm)
	assert.Empty(t, tm.GetExpressions())
}

// TestAddExpression проверяет добавление нового выражения.
func TestAddExpression(t *testing.T) {
	tm := task_manager.NewTaskManager()

	id, err := tm.AddExpression("2 + 2")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Проверяем, что выражение добавлено
	expressions := tm.GetExpressions()
	assert.Len(t, expressions, 1)
	assert.Equal(t, "2 + 2", expressions[0].ExpressionString)
	assert.Equal(t, "pending", expressions[0].Status)

	// Добавляем некорректное выражение
	_, err = tm.AddExpression("2 + ")
	assert.Error(t, err)
}

// TestGetExpressions проверяет получение всех выражений.
func TestGetExpressions(t *testing.T) {
	tm := task_manager.NewTaskManager()

	id, err := tm.AddExpression("4 + 2")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	id, err = tm.AddExpression("5 + 5")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	id, err = tm.AddExpression("777+-")
	assert.Error(t, err)
	assert.Empty(t, id)

	// Проверяем список выражений
	expressions := tm.GetExpressions()
	assert.Len(t, expressions, 2)
}

// TestGetExpression проверяет получение выражения по ID.
func TestGetExpression(t *testing.T) {
	tm := task_manager.NewTaskManager()

	// Добавляем выражение
	id, err := tm.AddExpression("2 + 2")
	assert.NoError(t, err)

	// Получаем выражение по ID
	expr, found := tm.GetExpression(id)
	assert.True(t, found)
	assert.Equal(t, "2 + 2", expr.ExpressionString)
	assert.Equal(t, "pending", expr.Status)

	// Пытаемся получить несуществующее выражение
	_, found = tm.GetExpression("amogus-sus-id")
	assert.False(t, found)
}

// TestGetTasks проверяет получение задач для выражения.
func TestGetTasks(t *testing.T) {
	tm := task_manager.NewTaskManager()

	// Добавляем выражение
	id, err := tm.AddExpression("2 + 2 * 2")
	assert.NoError(t, err)

	// Получаем задачи для выражения
	tasks := tm.GetTasks(id)
	assert.Len(t, tasks, 2)
	assert.NotEmpty(t, tasks)

	// Пытаемся получить задачи для несуществующего выражения
	tasks = tm.GetTasks("invalid-id")
	assert.Empty(t, tasks)
}

// TestGetTask проверяет получение задачи со статусом "pending".
func TestGetTask(t *testing.T) {
	tm := task_manager.NewTaskManager()

	// Добавляем выражение
	id, err := tm.AddExpression("2 + 2")
	assert.NoError(t, err)

	// Получаем задачу
	task, exprID, found := tm.GetTask()
	assert.True(t, found)
	assert.Equal(t, id, exprID)
	assert.Equal(t, "processing", task.Status)

	// Завершаем задачу
	tm.CompleteTask(id, task.ID, "", 4.0)

	// Пытаемся получить задачу снова (все задачи завершены)
	_, _, found = tm.GetTask()
	assert.False(t, found)
}

// TestCompleteTask проверяет завершение задачи.
func TestCompleteTask(t *testing.T) {
	tm := task_manager.NewTaskManager()

	// Добавляем выражение
	id, err := tm.AddExpression("2 + 2")
	assert.NoError(t, err)

	// Получаем задачу
	task, exprID, found := tm.GetTask()
	assert.True(t, found)
	assert.Equal(t, id, exprID)

	// Завершаем задачу
	success := tm.CompleteTask(id, task.ID, "", 4.0)
	assert.True(t, success)

	// Проверяем, что задача завершена
	expr, found := tm.GetExpression(id)
	assert.True(t, found)
	assert.Equal(t, "completed", expr.Status)
	assert.Equal(t, 4.0, *expr.Result)

	// Пытаемся завершить несуществующую задачу
	success = tm.CompleteTask("invalid-id", "invalid-task-id", "", 0.0)
	assert.False(t, success)
}

// TestImpossibleTask проверяет обработку невозможной задачи.
func TestImpossibleTask(t *testing.T) {
	tm := task_manager.NewTaskManager()

	// Добавляем выражение
	id, err := tm.AddExpression("2 + 2")
	assert.NoError(t, err)

	// Получаем задачу
	task, exprID, found := tm.GetTask()
	assert.True(t, found)
	assert.Equal(t, id, exprID)

	// Помечаем задачу как невозможную
	success := tm.CompleteTask(id, task.ID, "division by zero", 0.0)
	assert.True(t, success)

	// Проверяем, что выражение помечено как "error"
	expr, found := tm.GetExpression(id)
	assert.True(t, found)
	assert.Equal(t, "error", expr.Status)
	assert.Equal(t, "division by zero", expr.Error)
}

// TestAreDependenciesCompleted проверяет проверку завершенности зависимостей.
func TestAreDependenciesCompleted(t *testing.T) {
	tm := task_manager.NewTaskManager()

	// Создаем задачи с зависимостями
	tasks := []models.Task{
		{ID: "task1", Status: "completed"},
		{ID: "task2", Status: "pending"},
	}

	// Проверяем завершенность зависимостей
	completed := tm.AreDependenciesCompleted(tasks, []string{"task1"})
	assert.True(t, completed)

	completed = tm.AreDependenciesCompleted(tasks, []string{"task2"})
	assert.False(t, completed)

	completed = tm.AreDependenciesCompleted(tasks, []string{"task3"})
	assert.False(t, completed)
}
