# Используем официальный образ Go для сборки
FROM golang:1.25-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей Go
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Второй этап: создаем легкий образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates

# Создаем пользователя app для безопасности
RUN addgroup -S app && adduser -S app -G app

WORKDIR /root/

# Копируем бинарный файл из стадии сборки
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/views ./internal/views

# Делаем файлы доступными для пользователя app
RUN chown -R app:app /root/

# Переключаемся на пользователя app
USER app

# Экспонируем порт
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]