# Этап 1: Сборка приложения
# Используем официальный образ Go нужной версии (1.24.1 или новее)
FROM golang:1.24.1-alpine AS builder

# Устанавливаем Git, т.к. go install может его использовать для скачивания
# и ca-certificates для HTTPS
RUN apk add --no-cache git ca-certificates

# Устанавливаем gosearch
# Выполняем это до копирования исходников, чтобы кэшировать этот слой
RUN go install github.com/ibnaleem/gosearch@latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы управления зависимостями
COPY backend/go.mod backend/go.sum ./

# Скачиваем зависимости (если они появятся)
# Этот слой будет кэшироваться, если go.mod/go.sum не менялись
RUN go mod download

# Копируем исходный код бэкенда
COPY backend/ .

# Собираем приложение Go. Флаги для статической сборки и удаления отладочной информации.
# Выходной файл будет /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server main.go

# Этап 2: Создание минимального финального образа
# Используем базовый образ Alpine Linux (очень маленький)
FROM alpine:latest

# Копируем корневые сертификаты из сборщика
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем бинарный файл gosearch из сборщика
# Путь может отличаться, если `go install` устанавливает в другое место
# Обычно это /go/bin/gosearch внутри стандартного golang образа
COPY --from=builder /go/bin/gosearch /usr/local/bin/gosearch

# Копируем скомпилированное приложение из сборщика
COPY --from=builder /app/server /app/server

# Устанавливаем рабочую директорию
WORKDIR /app

# Открываем порт, который слушает наше приложение
EXPOSE 8080

# Команда для запуска приложения при старте контейнера
CMD ["/app/server"] 