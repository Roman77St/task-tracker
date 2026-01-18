FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -ldflags="-s -w"
# Эти флаги удаляют таблицу символов и отладочную информацию из бинарного файла.
# Это уменьшает размер исполняемого файла на 20-30% без потери производительности.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o bot ./cmd/server/main.go

# ca-certificates - обязательны для работы по сети, tzdata - для работы с часовым поясом
# по умолчанию могут отсутствовать в alpine
FROM alpine:3.23
RUN apk --no-cache add ca-certificates tzdata

# Создаем не-root пользователя для безопасности
RUN adduser -D -u 10001 appuser
WORKDIR /home/appuser

COPY --from=builder /app/bot .
COPY --from=builder /go/bin/migrate ./migrate
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/scripts/entrypoint.sh ./entrypoint.sh

# Не для продакшн! лучше передавать переменные окружения через Docker Compose
# или параметры запуска (docker run --env-file .env my-app-image)
# COPY --from=builder /app/.env .

RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]