FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем файлы управления зависимостями
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код приложения
COPY .env ./.env
COPY pkg ./pkg
COPY web ./web
COPY config ./config

# Собираем приложение
RUN go build -o /app/web-app ./web/cmd/main.go

FROM alpine:latest

WORKDIR /app

# Копируем исполняемый файл из билдера
COPY --from=builder /app/web-app ./

# Копируем .env файл
COPY --from=builder /app/.env ./.env

# Копируем config и static
COPY --from=builder /app/config ./config
COPY --from=builder /app/web/static ./web/static

# Устанавливаем права на выполнение
RUN chmod +x ./web-app

# Запускаем приложение
ARG DEFAULT_PORT=8081
ENV PORT_WEB=$DEFAULT_PORT

EXPOSE $PORT_WEB
CMD ["./web-app"]