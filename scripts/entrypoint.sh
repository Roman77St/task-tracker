#!/bin/sh

# Останавливает скрипт при любой ошибке
set -e

echo "Запуск миграций..."
./migrate -path ./migrations -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up
echo "Миграции завершены успешно. Запуск бота..."
exec ./bot

# Использование exec в конце позволяет процессу бота стать "владельцем" контейнера (PID 1).
# Это нужно для того, чтобы Docker мог корректно передавать боту сигналы завершения
# (например, при команде docker stop).