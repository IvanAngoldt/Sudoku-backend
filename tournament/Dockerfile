# Используем официальный образ Go
FROM golang:1.24-alpine

# Устанавливаем необходимые зависимости
RUN apk add --no-cache gcc musl-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main .

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"] 