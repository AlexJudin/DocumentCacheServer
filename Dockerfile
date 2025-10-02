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

# Аргументы для сборки
ARG APP_NAME=app
ARG VERSION=1.0.0

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /alexeyjudin ./cmd/main.go

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /build_app \
    -ldflags="-w -s -X main.version=${VERSION} -X" \
    -o /app/${APP_NAME} ./cmd/app

CMD ["/build_app"]

# Фаза запуска
FROM scratch AS runtime
