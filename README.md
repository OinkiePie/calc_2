# Calc Service: Распределенная система для вычисления математических выражений

## Описание

Calc Service - это распределенная система, предназначенная для вычисления сложных математических выражений. Она состоит из следующих компонентов:

*   **Web Service:** Предоставляет API для взаимодействия с пользователем. Позволяет отправлять выражения на вычисление и получать результаты.
*   **Orchestrator:** Координирует процесс вычисления, разбивая сложные выражения на простые задачи и распределяя их между агентами.
*   **Agent:** Выполняет отдельные математические задачи, полученные от оркестратора, и возвращает результаты.

## Архитектура
```
calc_2/
├── cmd/
│   └── main.go                   // Точка входа приложения.
│      
├── pkg/
│   ├── models/
│   │   ├── expression.go     // Структуры данных выражения.
│   │   └── task.go           // Структуры данных задач.
│   ├── logger/
│   │   └── logger.go             // Логирует сообщения.
│   └── operators/
│       └── operators.go          // Символы математических операций.
│
├── config/
│   ├── config.go                 // Загружает и обрабатывает конфигурации.
│   ├── dev.yaml                  // Конфигурация для разработки.
│   └── prod.yaml                 // Конфигурация для production.
│
├── web/
│   ├── web.go                    // Точка входа Веб Сервиса.
│   ├── internal/
│   │   ├── router/
│   │   │    └── router.go        // Маршруты Веб Сервиса
│   │   └── handlers/
│   │       └── handlers.go       // Обработчики запросов Веб Сервису
│   └── static/
│       ├── index.html            // Основной файл интерфейса.
│       ├── favicon.ico           // Иконка, отображаемая в браузере.
│       ├── script.js             // JavaScript код интерфейса.
│       └── style.css             // CSS стили интерфейса.
│
├── agent/
│   └── agent.go                  // Точка входа Агента. 
│
└── orchestrator/
    ├── orchestrator.go           // Точка входа Оркестратора. 
    └── internal/
        ├── handlers/
        │   └── handlers.go       // Обрабатывает запросы Оркестратору
        ├── middlewares/
        │   └── middlewares.go    // Обработчики запросов Окестратору
        ├── router/
        │   └── router.go         // Маршруты Окестратора
        ├── task_manager/
        │   └── task_manager.go   // Управляет задачами
        └── task_splitter/
            └── task_splitter.go  // Разбивает выражение на задачи
```
## Запуск проекта

1.  Клонируйте репозиторий:

    ```bash
    git clone https://github.com/OinkiePie/calc_2.git
    cd calc_service
    ```

2.  Запустите сервисы с помощью единой точки входа команды:

    ```bash
    go run ./cmd/main.go
    ```


## Использование

### 1. Отправка выражения на вычисление

Для отправки математического выражения на вычисление используйте следующий запрос `curl`:

```bash
curl --location 'http://localhost:8082/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'


напиши мне код для агента. 
он должен делать запросы без остановки на localgost:8080/internal/task, если была получения ошибка что задач для решения нет ждет секунду. 
На вход он получает задачу
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
}
в начале выполения он запускает контекст с таймаутом Operation_time , и продолжает выполнение полсе окончания решения задачи (если заняло больше времени чем Operation_time ) или конца Operation_time (если выполнено раньше)

когда задача была решена отправляет её на localgost:8080/internal/task в формате
type TaskCompleted struct {
	// Expression - ID корневого выражения, к которому принадлежит задача.
	Expression string `json:"expression"`
	// ID - Уникальный идентификатор задачи.
	ID string `json:"id"`
	// Result - Результат вычисления задачи.
	Result float64 `json:"result"`
}.

При старте демон запускает несколько горутин, каждая из которых выступает в роли независимого вычислителя. Количество горутин регулируется переменной конфига, можешь получить её по адресу config.Cfg.server.agent.COMPUTING_POWER.
Сделай разбитие на файлы
agent.go - запускает сервис
internal - остальные процессы.
можешь добавить папки по необходимости