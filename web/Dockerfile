FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o build/web.exe ./web/cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/web ./web

EXPOSE 8081

CMD ["./web"]