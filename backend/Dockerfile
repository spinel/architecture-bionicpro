# Этап сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые зависимости
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/main .

# Открываем порт
EXPOSE 8000

# Запускаем приложение
CMD ["./main"]

