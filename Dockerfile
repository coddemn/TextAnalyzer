# === Этап 1: сборка ===
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Кэшируем зависимости: копируем только mod-файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Устанавливаем swag и генерируем docs.go ПЕРЕД сборкой
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/app/main.go -o docs

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/analyzer \
    ./cmd/app

# === Этап 2: финальный образ ===
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/analyzer /app/analyzer

RUN adduser -D -u 1001 appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/analyzer"]