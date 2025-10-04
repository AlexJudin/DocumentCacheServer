# Фаза сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Аргументы для сборки
ARG APP_NAME=doc_serv
ARG VERSION=1.0.0
ARG BUILD_TIME=unknown

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o /app/${APP_NAME} ./cmd/

# Фаза запуска
FROM scratch AS runtime

# Копируем сертификаты и временную зону из builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копируем собранное приложение
COPY --from=builder /app/${APP_NAME} /app/${APP_NAME}

# Устанавливаем рабочую директорию
WORKDIR /app

# Экспортируем порт
EXPOSE 7540

# Команда для запуска
CMD ["/app/doc_serv"]
