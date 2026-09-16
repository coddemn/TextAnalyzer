# === Этап 1: сборка ===
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Кэшируем зависимости: копируем только mod-файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/analyzer \
    ./cmd/app

# === Этап 2: финальный образ ===
FROM alpine:3.20

# ca-certificates нужен, если приложение делает HTTPS-запросы
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем только бинарник из этапа сборки
COPY --from=builder /app/bin/analyzer /app/analyzer

# Непривилегированный пользователь (продакшен-паттерн)
RUN adduser -D -u 1001 appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/analyzer"]