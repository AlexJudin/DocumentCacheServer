# Фаза сборки
FROM golang:1.21.0 AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Создаем рабочую директорию
WORKDIR /usr/src/app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /alexeyjudin ./cmd/main.go

CMD ["/alexeyjudin"]
