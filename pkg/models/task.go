package models

// Task представляет структуру для части арифметического выражения, которую нужно вычислить.
type Task struct {
	// ID - Уникальный идентификатор задачи.
	ID string
	// Args - Срез указателей на аргументы задачи. Может быть nil, если аргумент является зависит от другой задачи.
	Args []*float64
	// Operation - Операция, которую необходимо выполнить (+, -, *, /, ^, -u).
	Operation string
	// Operation_time - Время, необходимое для выполнения операции.
	Operation_time int
	// Dependencies - Список ID задач, результаты которых необходимы для выполнения данной задачи.
	Dependencies []string
	// Status - Статус задачи ("pending", "processing", "completed").
	Status string
	// Result - Указатель на результат выполнения задачи. Может быть nil, если задача ещё не выполнена.
	Result *float64
	// Expression - ID выражения, к которому принадлежит данная задача.
	Expression string
}

// TaskResponse представляет структуру для отправки информации о задаче в HTTP-ответе.
type TaskResponse struct {
	// ID - Уникальный идентификатор задачи.
	ID string `json:"id"`
	// Args - Срез указателей на аргументы задачи.
	Args []*float64 `json:"args"`
	// Operation - Операция, которую необходимо выполнить.
	Operation string `json:"operation"`
	// Operation_time - Время, необходимое для выполнения операции.
	Operation_time int `json:"operation_time"`
	// Expression - ID выражения, к которому принадлежит данная задача.
	// Используется для опитмизации возвращения резульата агентом.
	Expression string `json:"expression"`
	// Error - Указывает на невыполниасть задачи
	Error string `json:"error"`
}

// TaskCompleted представляет структуру для получения информации о завершенной задаче из HTTP-запроса.
// Используется для декодирования вырожения из тела запроса.
type TaskCompleted struct {
	// Expression - ID корневого выражения, к которому принадлежит задача.
	Expression string `json:"expression"`
	// ID - Уникальный идентификатор задачи.
	ID string `json:"id"`
	// Result - Результат вычисления задачи.
	Result float64 `json:"result"`
	// Error - Указывает на невыполниасть задачи
	Error string `json:"error"`
}
