package worker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/OinkiePie/calc_2/agent/internal/client"
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

// StartWorker запускает вычислителя
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debugf("Рабочий %d отключен", w.workerID)
			return
		default:
			task, err := w.apiClient.GetTask()
			if err != nil {
				logger.Log.Errorf("Рабочий %d: Ошибка при получении задачи: %v", w.workerID, err)
				time.Sleep(10 * time.Second)
				continue
			}

			if task == nil {
				logger.Log.Debugf("Рабочий %d: Нет доступных задач, ожидаю...", w.workerID)
				time.Sleep(2 * time.Second)
				continue
			}

			logger.Log.Debugf("Рабочий %d: Получена задача %s", w.workerID, task.ID)

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

				result, err := calculate(t)
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
				time.Sleep(10 * time.Second)
			} else {
				logger.Log.Debugf("Рабочий %d: Задача %s успешно отправлена", w.workerID, task.ID)
			}
		}
	}
}

func calculate(task *models.TaskResponse) (float64, error) {
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

	return 0, fmt.Errorf("unidentified operator: %s", task.Operation)
}
