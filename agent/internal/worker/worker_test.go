package worker_test

import (
	"context"
	"sync"
	"testing"

	"github.com/OinkiePie/calc_2/agent/internal/client"
	"github.com/OinkiePie/calc_2/agent/internal/worker"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/models"
	"github.com/OinkiePie/calc_2/pkg/operators"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Отключаем логирование в тестах, чтобы не засорять вывод.
	logger.InitLogger(logger.Options{Level: 6})
}

func TestCalculate_Addition(t *testing.T) {
	arg1 := 5.0
	arg2 := 3.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, &arg2},
		Operation: operators.OpAdd,
	}

	result, err := worker.Calculate(task)

	assert.NoError(t, err)
	assert.Equal(t, 8.0, result)
}

func TestCalculate_Subtraction(t *testing.T) {
	arg1 := 5.0
	arg2 := 3.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, &arg2},
		Operation: operators.OpSubtract,
	}

	result, err := worker.Calculate(task)

	assert.NoError(t, err)
	assert.Equal(t, 2.0, result)
}

func TestCalculate_Multiplication(t *testing.T) {
	arg1 := 5.0
	arg2 := 3.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, &arg2},
		Operation: operators.OpMultiply,
	}

	result, err := worker.Calculate(task)

	assert.NoError(t, err)
	assert.Equal(t, 15.0, result)
}

func TestCalculate_Division(t *testing.T) {
	arg1 := 15.0
	arg2 := 3.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, &arg2},
		Operation: operators.OpDivide,
	}

	result, err := worker.Calculate(task)

	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)
}

func TestCalculate_DivisionByZero(t *testing.T) {
	arg1 := 15.0
	arg2 := 0.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, &arg2},
		Operation: operators.OpDivide,
	}

	_, err := worker.Calculate(task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero not allowed")
}

func TestCalculate_Power(t *testing.T) {
	arg1 := 2.0
	arg2 := 3.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, &arg2},
		Operation: operators.OpPower,
	}

	result, err := worker.Calculate(task)

	assert.NoError(t, err)
	assert.Equal(t, 8.0, result)
}

func TestCalculate_UnaryMinus(t *testing.T) {
	arg1 := 5.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, nil},
		Operation: operators.OpUnaryMinus, // Операция не важна для унарного минуса.
	}

	result, err := worker.Calculate(task)

	assert.NoError(t, err)
	assert.Equal(t, -5.0, result)
}

func TestCalculate_FirstArgumentNil(t *testing.T) {
	task := &models.TaskResponse{
		Args:      []*float64{nil, new(float64)},
		Operation: operators.OpAdd,
	}

	_, err := worker.Calculate(task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "first operator cannot be nil")
}

func TestCalculate_UnknownOperator(t *testing.T) {
	arg1 := 5.0
	arg2 := 3.0
	task := &models.TaskResponse{
		Args:      []*float64{&arg1, &arg2},
		Operation: "unknown",
	}

	_, err := worker.Calculate(task)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown operator")
}

func TestWorker_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	apiClient := &client.APIClient{}
	wg := &sync.WaitGroup{}
	errChan := make(chan error, 1)
	worker := worker.NewWorker(1, apiClient, wg, errChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Start(ctx)
	}()

	cancel()
	wg.Wait()
}
