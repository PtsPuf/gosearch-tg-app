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

# Копируем ТОЛЬКО файл управления зависимостями
COPY backend/go.mod ./

# Скачиваем зависимости (если они появятся) или создаем go.sum
# Этот слой будет кэшироваться, если go.mod не менялся
RUN go mod download

# Копируем исходный код бэкенда
COPY backend/ .

# Собираем приложение Go. Флаги для статической сборки и удаления отладочной информации.
# Выходной файл будет /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server main.go

# Этап 2: Создание минимального финального образа
# Используем базовый образ Alpine Linux той же версии, что и builder (примерно)
FROM alpine:3.19 AS final

# Копируем корневые сертификаты из сборщика
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем бинарный файл gosearch из сборщика
COPY --from=builder /go/bin/gosearch /usr/local/bin/gosearch

# Копируем скомпилированное приложение из сборщика
COPY --from=builder /app/server /app/server

# Добавляем права на выполнение для бинарных файлов
RUN chmod +x /app/server /usr/local/bin/gosearch

# Устанавливаем рабочую директорию
WORKDIR /app

# Диагностика: Проверяем наличие и права файлов перед запуском
RUN ls -l /app
RUN ls -l /usr/local/bin/gosearch

# Открываем порт, который слушает наше приложение
EXPOSE 8080

# Команда для запуска приложения при старте контейнера (относительно WORKDIR)
CMD ["./server"] 