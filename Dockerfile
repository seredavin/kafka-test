# Multi-stage build для минимального размера образа

# Стадия сборки
FROM golang:1.23-alpine AS builder

# Установка необходимых инструментов
RUN apk add --no-cache git ca-certificates tzdata

# Рабочая директория
WORKDIR /build

# Копируем go mod файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download
RUN go mod verify

# Копируем исходный код
COPY . .

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-s -w -extldflags "-static"' \
    -o kafka-producer-ui .

# Финальная стадия
FROM scratch

# Копируем CA сертификаты для HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем информацию о временных зонах
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копируем бинарный файл
COPY --from=builder /build/kafka-producer-ui /kafka-producer-ui

# Метаданные
LABEL maintainer="your-email@example.com"
LABEL description="Kafka Producer UI - TUI for sending messages to Apache Kafka"
LABEL version="1.0.0"

# Точка входа
ENTRYPOINT ["/kafka-producer-ui"]

# Примечание: Поскольку это TUI приложение, запускать нужно с флагами -it:
# docker run -it --rm kafka-producer-ui

