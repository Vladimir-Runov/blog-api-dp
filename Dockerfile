# Этап сборки
#FROM golang:1.24-alpine AS builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем модули
COPY go.mod go.sum ./

RUN go version
RUN go mod download
COPY . .

# Собираем из папки cmd/api
RUN go build -o main ./cmd/api

# Минимальный образ для запуска
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Копируем бинарник из сборочного этапа
COPY --from=builder /app/main .

# Копируем миграции
COPY --from=builder /app/migrations ./migrations

# Права на выполнение
RUN chmod +x main

# Запуск
EXPOSE 8080
CMD ["./main"]
