package models

// Expression представляет структуру арифметического выражения.
type Expression struct {
	// ID - Уникальный идентификатор выражения.
	ID string
	// Status - Статус выражения ("pending", "processing", "completed").
	Status string
	// Result - Указатель на результат вычисления выражения. Может быть nil, если вычисление ещё не завершено.
	Result *float64
	// Tasks - Список задач, составляющих выражение.
	Tasks []Task
	// ExpressionString - Исходное выражение в виде строки.
	ExpressionString string
}

// ExpressionResponse представляет структуру для отправки информации о выражении в HTTP-ответе.
type ExpressionResponse struct {
	// ID - Уникальный идентификатор выражения.
	ID string `json:"id"`
	// Status - Статус выражения.
	Status string `json:"status"`
	// Result - Указатель на результат вычисления выражения. Если nil, то поле не включается в JSON-ответ (omitempty).
	Result *float64 `json:"result,omitempty"` //omitempty - если result nil, то не выводить его
}

// ExpressionAdd представляет структуру для получения математического выражения из HTTP-запроса.
// Используется для декодирования вырожения из тела запроса.
type ExpressionAdd struct {
	// Expression - Математическое выражение в виде строки.
	Expression string `json:"expression"`
}
