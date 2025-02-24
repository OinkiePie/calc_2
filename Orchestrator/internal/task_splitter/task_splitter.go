package task_splitter

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator/internal/models"
	"github.com/OinkiePie/calc_2/pkg/logger"
	"github.com/OinkiePie/calc_2/pkg/operators"
	"github.com/google/uuid"
)

var (
	ErrEmptyInput        = errors.New("empty input")
	ErrUnopenedParen     = errors.New("unopened parenthesis")
	ErrUnclosedParen     = errors.New("unclosed parenthesis")
	ErrInvalidSyntax     = errors.New("invalid syntax")
	ErrNotEnoughOperands = errors.New("not enough operands")
	// ErrTooManyOperators  = errors.New("too many operators") - эквивалентно ErrNotEnoughOperands
	ErrUnaryMinus = errors.New("not enough operands for the unary minus")
	ErrRPN        = errors.New("error during converting to RPN")
)

// ParseExpression разбирает математическое выражение, представленное в виде строки, и преобразует его в набор задач для выполнения.
//
// Args:
//
//	id: string - Уникальный идентификатор для связывания задач с выражением.
//	expression: string - Математическое выражение, которое необходимо разобрать.
//
// Returns:
//
//	[]models.Task - Срез задач, представляющих операции, необходимые для вычисления выражения.
//	error - Ошибка, если выражение не может быть разобрано или содержит неверные элементы.
func ParseExpression(id, expression string) ([]models.Task, error) {
	// expression = strings.TrimSpace(expression) не требуется
	// т.к. была передана уже обрезанная строка

	// Удаляем все пробелы внутри строки
	expression = strings.ReplaceAll(expression, " ", "")

	rpn, err := infixToRPN(expression)
	if err != nil {
		return nil, err
	}

	var tasks []models.Task
	// fmt.Println("RPN:", strings.Join(rpn, " "))
	tasks, err = rpnToTasks(id, rpn)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// precedence определяет приоритет оператора для правильной вложенности при разбиении на задачи.
// Более высокий приоритет означает, что оператор должен быть выполнен раньше.
//
// Args:
//
//	op: string - Строка, представляющая оператор (+, -, *, /, ^, u-).
//
// Returns:
//
//	int - Целое число, представляющее приоритет оператора. Чем больше число, тем выше приоритет.
//	     Возвращает 0 для неопознанных операторов.
func precedence(op string) int {
	switch op {
	case operators.OpAdd, operators.OpSubtract:
		return 1
	case operators.OpMultiply, operators.OpDivide:
		return 2
	case operators.OpPower:
		return 3
	case operators.OpUnaryMinus:
		return 4
	default:
		return 0
	}
}

// isOperator проверяет, является ли токен строкой, представляющей математический оператор.
//
// Args:
//
//	token: string - Строка, которую необходимо проверить.
//
// Returns:
//
//	bool - true, если токен является одним из допустимых операторов (+, -, *, /, ^, u-), иначе false.
func isOperator(token string) bool {
	switch token {
	case operators.OpAdd, operators.OpSubtract, operators.OpMultiply, operators.OpDivide, operators.OpPower:
		return true
	default:
		return false
	}
}

// isUnaryMinus определяет, следует ли обрабатывать знак минус как унарный (например, "-5") или бинарный (например, "3 - 5").
//
// Args:
//
//	tokens: []string - Срез строк, представляющий токены выражения.
//	i: int - Индекс текущего токена в срезе.
//
// Returns:
//
//	bool - true, если минус должен быть обработан как унарный, иначе false.
func isUnaryMinus(tokens []string, i int) bool {
	if i == 0 {
		return true // Минус в начале выражения - унарный
	}
	prevToken := tokens[i-1]
	return prevToken == operators.ParenLeft || isOperator(prevToken)
}

// infixToRPN преобразует математическое выражение в инфиксной нотации (обычная запись) в обратную польскую нотацию (RPN).
// RPN упрощает вычисление выражений с помощью стека.
//
// Args:
//
//	expression: string - Математическое выражение в инфиксной нотации.
//
// Returns:
//
//	[]string - Срез строк, представляющий выражение в обратной польской нотации (RPN).
//	error - Ошибка, если выражение не может быть преобразовано.
func infixToRPN(expression string) ([]string, error) {

	tokens := tokenize(expression) // Сначала разбиваем на токены
	output := []string{}           // Выходная очередь
	stack := []string{}            // Стек операторов

	for i, token := range tokens {
		switch {
		case isNumber(token): // Если число, добавляем в выходную очередь
			output = append(output, token)
		case token == operators.ParenLeft: // Если открывающая скобка, помещаем в стек
			stack = append(stack, token)
		case token == operators.ParenRight: // Если закрывающая скобка
			for len(stack) > 0 && stack[len(stack)-1] != operators.ParenLeft {
				// Переносим операторы из стека в выходную очередь,
				// пока он не опустеет, или мы не встретим открывающую скобку
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				// Если в стеке не осталось открывающей скобки, это означает, что у нас была
				// закрывающая скобка, но не было соответствующей открывающей скобки в выражении
				return nil, ErrUnopenedParen
			}
			stack = stack[:len(stack)-1] // Удаляем открывающую скобку из стека
		case isOperator(token): // Если оператор
			if token == "-" && isUnaryMinus(tokens, i) {
				token = operators.OpUnaryMinus // Помечаем как унарный минус
			}
			for len(stack) > 0 && precedence(token) <= precedence(stack[len(stack)-1]) {
				// Переносим операторы из стека в выходную очередь, пока приоритет текущего оператора
				// меньше или равен приоритету оператора на вершине стека
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token) // Помещаем текущий оператор в стек
		default:
			return nil, ErrInvalidSyntax
		}
	}

	// Переносим все оставшиеся операторы из стека в выходную очередь
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		if top == operators.ParenLeft || top == operators.ParenRight {
			return nil, ErrUnclosedParen
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

// tokenize разбивает входную строку математического выражения на отдельные токены (числа, операторы, скобки).
// Токены используются для дальнейшей обработки выражения.
//
// Args:
//
//	expression: string - Строка, содержащая математическое выражение.
//
// Returns:
//
//	[]string - Срез строк, представляющий токены выражения.
func tokenize(expression string) []string {
	var tokens []string
	var currentNumber string

	for _, r := range expression {
		s := string(r)
		if unicode.IsDigit(r) || s == operators.Point {
			currentNumber += s
		} else {
			if currentNumber != "" {
				tokens = append(tokens, currentNumber)
				currentNumber = ""
			}
			if s != "" {
				tokens = append(tokens, s)
			}
		}
	}

	if currentNumber != "" {
		tokens = append(tokens, currentNumber)
	}
	return tokens
}

// isNumber проверяет, является ли переданный токен числом.
//
// Args:
//
//	token: string - Строка, которую необходимо проверить.
//
// Returns:
//
//	bool - true, если токен может быть преобразован в число, иначе false.
func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

// opTime возвращает время операции
//
// Args:
//
//	operator: string - математический оператор.
//
// Returns:
//
//	int - длительность операции в миллисекундах, если значение отсутствует - 0.
func opTime(operator string) int {
	// Время не вынесено в отдельную переменную т.к. при этом конфиг не успевает инициализироваться
	duration, ok := map[string]int{
		operators.OpAdd:      config.Cfg.Math.TIME_ADDITION_MS,
		operators.OpSubtract: config.Cfg.Math.TIME_SUBTRACTION_MS,
		operators.OpMultiply: config.Cfg.Math.TIME_MULTIPLICATION_MS,
		operators.OpDivide:   config.Cfg.Math.TIME_DIVISION_MS,
		operators.OpPower:    config.Cfg.Math.TIME_POWER_MS,
	}[operator]

	if !ok {
		logger.Log.Warnf("opTime: оператор %s не найден", operator)
		return 0
	}
	return duration
}

// rpnToTasks преобразует выражение, представленное в обратной польской нотации (RPN), в набор задач (models.Task) с учетом зависимостей между ними.
// Каждая задача представляет собой операцию, которую необходимо выполнить для вычисления части выражения.
//
// Args:
//
//	expression: string - Исходное математическое выражение.
//	rpn: []string - Срез строк, представляющий выражение в обратной польской нотации (RPN).
//
// Returns:
//
//	[]models.Task - Срез задач, представляющих операции для вычисления выражения, с установленными зависимостями.
//	error: Ошибка, если RPN выражение не может быть преобразовано в задачи или содержит неверные элементы.
func rpnToTasks(expression string, rpn []string) ([]models.Task, error) {
	var tasks []models.Task
	var stack []string // Стек для чисел и ID задач

	for _, token := range rpn {
		switch token {
		case operators.OpAdd, operators.OpSubtract, operators.OpMultiply, operators.OpDivide, operators.OpPower:
			if len(stack) < 2 {
				return nil, ErrNotEnoughOperands
			}

			operand2Str := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			operand1Str := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			id := uuid.New().String()
			task := models.Task{
				ID:             id,
				Args:           make([]*float64, 2),
				Operation:      token,
				Operation_time: opTime(token),
				Status:         "pending",
				Expression:     expression,
				Dependencies:   make([]string, 2),
			}

			// Проверяем, являются ли операнды числами или ID задач
			if num1, err := strconv.ParseFloat(operand1Str, 64); err == nil {
				task.Args[0] = &num1 // Arg1 - число
			} else {
				// Arg1 - nil (зависимость)
				task.Dependencies[0] = operand1Str
			}

			if num2, err := strconv.ParseFloat(operand2Str, 64); err == nil {
				task.Args[1] = &num2 // Arg2 - число
			} else {
				// Arg2 - nil (зависимость)
				task.Dependencies[1] = operand2Str
			}

			tasks = append(tasks, task)
			stack = append(stack, id) // Результат этой задачи будет использован далее
		case operators.OpUnaryMinus:
			// Унарный минус
			// Если унарному минусу одиноко, он начинает грустить
			if len(stack) < 1 {
				return nil, ErrUnaryMinus
			}
			operandStr := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			id := uuid.New().String()
			task := models.Task{
				ID:             id,
				Operation:      operators.OpUnaryMinus,
				Operation_time: config.Cfg.Math.TIME_UNARY_MINUS_MS,
				Status:         "pending",
				Expression:     expression,
			}

			if num, err := strconv.ParseFloat(operandStr, 64); err == nil {
				task.Args[0] = &num // Arg1 - число
			} else {
				// Arg1 - nil (зависимость)
				task.Dependencies[0] = operandStr
			}
			tasks = append(tasks, task)
			stack = append(stack, id)

		default:
			// Число
			if _, err := strconv.ParseFloat(token, 64); err != nil {
				// В результате всех прошлых операций и проверок можно сделать вывод, что
				// суда ничего не может попасть (не может быть инородным символом или оператором
				// в результате проверки в infixToRPN и прошлых case соответственно).
				return nil, ErrRPN
			}
			stack = append(stack, token) // Просто добавляем число в стек
		}
	}

	if len(stack) != 1 {
		return nil, ErrRPN
	}

	// Все задачи созданы.
	return tasks, nil
}
