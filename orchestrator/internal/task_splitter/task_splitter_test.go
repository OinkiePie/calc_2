package task_splitter_test

import (
	"errors"
	"testing"

	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/orchestrator/internal/task_splitter"
	"github.com/OinkiePie/calc_2/pkg/models"
)

type test struct {
	name       string
	id         string
	expression string
	want       []models.Task
	wantErr    error
}

func TestParseExpression(t *testing.T) {
	config.InitConfig() // Не обращаем внимание на ошибку т.к. это не имеет смысла в тесте

	tests := []test{
		{
			name:       "Simple addition",
			expression: "2 + 3",
			want: []models.Task{
				{
					Operation: "+",
					Args:      []*float64{floatPtr(2), floatPtr(3)},
				},
			},
			wantErr: nil,
		},

		{
			name:       "Multiplication and subtraction",
			expression: "5 * 4 - 1",
			want: []models.Task{
				{
					Operation: "*",
					Args:      []*float64{floatPtr(5), floatPtr(4)},
				},
				{
					Operation: "-",
					Args:      []*float64{nil, floatPtr(1)},
				},
			},
			wantErr: nil,
		},

		{
			name:       "Parentheses and power",
			expression: "(2 + 3) ^ 2",
			want: []models.Task{
				{
					Operation: "+",
					Args:      []*float64{floatPtr(2), floatPtr(3)},
				},
				{
					Operation: "^",
					Args:      []*float64{nil, floatPtr(2)},
				},
			},
			wantErr: nil,
		},

		{
			name:       "Unary minus",
			expression: "-5 + 3",
			want: []models.Task{
				{
					Operation: "-u",
					Args:      []*float64{floatPtr(5), nil},
				},
				{
					Operation: "+",
					Args:      []*float64{nil, floatPtr(3)},
				},
			},
			wantErr: nil,
		},

		{
			name:       "Unopened Parenthesis",
			expression: "2 + 3)",
			want:       nil,
			wantErr:    task_splitter.ErrUnopenedParen,
		},
		{
			name:       "Unclosed Parenthesis",
			expression: "(2 + 3",
			want:       nil,
			wantErr:    task_splitter.ErrUnclosedParen,
		},

		{
			name:       "Invalid syntax",
			expression: "2 + a",
			want:       nil,
			wantErr:    task_splitter.ErrInvalidSyntax,
		},

		{
			name:       "Bad unary minus",
			expression: "-+1",
			want:       nil,
			wantErr:    task_splitter.ErrUnaryMinus,
		},

		{
			name:       "Not enough operands",
			expression: "3+",
			want:       nil,
			wantErr:    task_splitter.ErrNotEnoughOperands,
		},

		{
			name:       "One operand",
			expression: "42",
			want:       nil,
			wantErr:    task_splitter.ErrOneOperand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := task_splitter.ParseExpression("ABOBA42", tt.expression) // фиксированный id для прохождения компиляции

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseExpression() len = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {

				if tt.want[i].Operation != got[i].Operation {
					t.Errorf("Operation got = %s, want %s", tt.want[i].Operation, got[i].Operation)
					return
				}

				for j := range tt.want[i].Args {

					if !(nil == tt.want[i].Args[j] && nil == got[i].Args[j]) {

					}
					if *tt.want[i].Args[j] != *got[i].Args[j] {
						return
					}

				}
			}

		})
	}
}

// Helper function для создания указателя на float64
func floatPtr(f float64) *float64 {
	return &f
}
