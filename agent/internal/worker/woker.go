package worker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/OinkiePie/calc_2/agent/internal/client"
	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/models"
	"github.com/OinkiePie/calc_2/pkg/operators"
)

var (
	errDivisionByZero = errors.New("division by zero not allowed")
	errFirstNil       = errors.New("first operator cannot be nil")
)

// Worker представляет собой рабочего, выполняющего задачи.
type Worker struct {
	errChan   chan error        // Канал для отправки ошибок, возникающих при выполнении задач.
	workerID  int               // Уникальный идентификатор рабочего.
	apiClient *client.APIClient // API-клиент для получения и отправки задач.
	wg        *sync.WaitGroup   // WaitGroup для сигнализации о завершении работы.
}

// NewWorker создает новый экземпляр рабочего.
//
// Args:
//
//	workerID: int - Уникальный идентификатор рабочего.
//	apiClient: *client.APIClient - API-клиент для получения и отправки задач.
//	wg: *sync.WaitGroup - WaitGroup для сигнализации о завершении работы.
//	errChan: chan error - Канал для отправки ошибок, возникающих при выполнении задач.
//
// Returns:
//
//	*Worker: Указатель на новый экземпляр структуры Worker.
func NewWorker(workerID int, apiClient *client.APIClient, wg *sync.WaitGroup, errChan chan error) *Worker {
	return &Worker{
		workerID:  workerID,
		apiClient: apiClient,
		wg:        wg,
		errChan:   errChan,
	}
}

// Start запускает рабочего для выполнения задач, получаемых от API.
//
// Args:
//
//	ctx: context.Context - Контекст, используемый для отмены работы воркера.
//
// Описание:
//
//	Воркер постоянно пытается получить задачу от API. Если задача получена,
//	воркер запускает вычисление в отдельной горутине с таймаутом, указанным
//	в задаче. Результат вычисления (или ошибка) отправляется обратно в API.
//	Воркер завершает работу при отмене контекста.
//
// Обработка ошибок:
//   - Если нет задач для выполнения, воркер ждет 2 секунды и повторяет попытку.
//   - Если не удается получить задачу, воркер ждет 5 секунд и повторяет попытку.
//   - Если полученная задача невыполнима, в API отправляется сообщение об ошибке.
//   - Если во время вычисления происходит паника, она перехватывается, логируется и отправляется в канал ошибок.
//   - Если при отправке результата возникает ошибка, воркер ждет 5 секунд и повторяет попытку.
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()
	prevErr := errors.New("")
	waiting := true

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debugf("Рабочий %d отключен", w.workerID)
			return
		default:
			task, err := w.apiClient.GetTask()
			if err != nil {
				// Дабы избежать бесконечно спама в консоль сверяем с предыдущей ошибкой
				if prevErr.Error() != err.Error() {
					logger.Log.Errorf("Рабочий %d: Ошибка при получении задачи: %v. Повторные запросы каждые %d мс.",
						w.workerID, err, config.Cfg.Server.Agent.AGENT_REPEAT_ERR)
				}
				prevErr = err
				time.Sleep(time.Duration(config.Cfg.Server.Agent.AGENT_REPEAT_ERR) * time.Millisecond)
				continue
			}

			if task == nil {
				// Дабы избежать бесконечно спама в консоль проверяем был ли уже лог о ожидании
				if waiting {
					logger.Log.Debugf("Рабочий %d: Нет доступных задач. Повторные запросы каждые %d мс.",
						w.workerID, config.Cfg.Server.Agent.AGENT_REPEAT)
					waiting = false
				}
				time.Sleep(time.Duration(config.Cfg.Server.Agent.AGENT_REPEAT) * time.Millisecond)
				continue
			}

			logger.Log.Debugf("Рабочий %d: Получена задача %s", w.workerID, task.ID)
			waiting = true
			//  Создаем контекст с таймаутом
			taskCtx, cancel := context.WithTimeout(context.Background(), time.Duration(task.Operation_time)*time.Millisecond)
			defer cancel()

			//  Запускаем вычисление в горутине
			resultChan := make(chan float64, 1)
			errorChan := make(chan error, 1)

			go func(t *models.TaskResponse) {
				defer func() {
					if r := recover(); r != nil {
						close(resultChan)
						close(errorChan)
						w.errChan <- fmt.Errorf("ошибка во время вычисления: %v", r)
					}
				}()

				result, err := Calculate(t)
				if err != nil {
					errorChan <- err
					return
				}
				resultChan <- result
			}(task)

			// Ожидаем результат и таймаут
			var result float64
			select {
			case result = <-resultChan:
				<-taskCtx.Done()
				logger.Log.Debugf("Рабочий %d: Задача %s успешно выполнена", w.workerID, task.ID)
			case err = <-errorChan:
				logger.Log.Debugf("Рабочий %d: Задача %s невыполнима: %v", w.workerID, task.ID, err)
				// Перезаписываем поле Error чтобы обработчик понял что выражение невыполнимо
				task.Error = fmt.Sprintf("IMPOSSIBLE: %v", err)
				result = 0
			}

			if math.IsInf(result, 1) {
				result = 0
				task.Error = "result is +Inf"
			}

			if math.IsInf(result, -1) {
				result = 0
				task.Error = "result is -Inf"
			}

			// Отправляем результат (даже если был таймаут)
			completedTask := models.TaskCompleted{
				Expression: task.Expression,
				ID:         task.ID,
				Result:     result,
				Error:      task.Error,
			}

			err = w.apiClient.CompleteTask(completedTask)
			if err != nil {
				logger.Log.Errorf("Рабочий %d: Ошибка при отправлении задачи %s: %v", w.workerID, task.ID, err)
				time.Sleep(5 * time.Second)
			} else {
				logger.Log.Debugf("Рабочий %d: Задача %s успешно отправлена", w.workerID, task.ID)
			}
		}
	}
}

// Calculate выполняет математическую операцию над двумя аргументами, указанными в задаче.
// Поддерживаемые операции: сложение, вычитание, умножение, деление и возведение в степень.
// Если второй аргумент равен nil, выполняется унарный минус для первого аргумента.
//
// Args:
//
//	task: (*models.TaskResponse) - Задача, содержащая аргументы и операцию.
//
// Returns:
//
//	float64 - Результат выполнения операции.
//	error - Ошибка, если операция не может быть выполнена (например, деление на ноль или неизвестная операция).
func Calculate(task *models.TaskResponse) (float64, error) {
	var arg1, arg2 float64

	// Nil попадает в операндом только если оператором является унарный минус.
	// При этом число которое необходимо обратить всегда первый операнд.
	if task.Args[0] == nil {
		// Перовый оператор никогода не может быть nil
		return 0, errFirstNil
	}
	arg1 = *task.Args[0]
	// Если 2й операнд - nil, то операнд всегда унарный минус
	if task.Args[1] == nil {
		return -*task.Args[0], nil
	}
	arg2 = *task.Args[1]

	switch task.Operation {

	case operators.OpAdd:
		return arg1 + arg2, nil

	case operators.OpSubtract:
		return arg1 - arg2, nil

	case operators.OpMultiply:
		return arg1 * arg2, nil

	case operators.OpDivide:
		if arg2 == 0 {
			return 0, errDivisionByZero
		}
		return arg1 / arg2, nil

	case operators.OpPower:
		return math.Pow(arg1, arg2), nil
	}

	return 0, fmt.Errorf("unknown operator: %s", task.Operation)
}
