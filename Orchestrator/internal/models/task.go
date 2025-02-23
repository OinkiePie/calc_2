package models

// Expression - структура для части арифметического выражения
type Task struct {
	ID             string
	Args           []*float64
	Operation      string
	Operation_time int
	Dependencies   []string // ID задачи, результат которой нужен (если есть зависимость)
	Status         string   // "pending", "processing", "completed"
	Result         *float64 // Указатель на результат (мб nil)
	Expression     string
}

type TaskResponse struct {
	ID             string     `json:"id"`
	Args           []*float64 `json:"args"`
	Operation      string     `json:"operation"`
	Operation_time int        `json:"operation_time"`
	Expression     string     `json:"expression"`
}

type TaskCompleted struct {
	Expression string  `json:"expression"` // ID корневого выражения
	ID         string  `json:"id"`         //ID таска
	Result     float64 `json:"result"`     //Результат вычислений
}
