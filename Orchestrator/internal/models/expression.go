package models

// Expression - структура для представления арифметического выражения
type Expression struct {
	ID               string
	Status           string   // "pending", "processing", "completed"
	Result           *float64 // Указатель, чтобы отличать `null` от `0`
	Tasks            []Task
	ExpressionString string // Исходное выражение
}

type ExpressionResponse struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Result *float64 `json:"result,omitempty"` //omitempty - если result nil, то не выводить его
}

type ExpressionAdd struct {
	Expression string `json:"expression"`
}
