package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"task-traker/internal/config"
	"task-traker/internal/repository"
	"task-traker/pkg/telegram"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.TODO()
	// Инициализируем начальный логгер
	programLevel := &slog.LevelVar{}
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(logHandler))

	// Загружаем переменные среды из файла .env
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .enf file")
		os.Exit(1)
	}

	// Загружаем конфигурацию
	conf, err := config.New()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	// Устанавливаем уровень логгирования из конфига
	config.SetConfigLevel(*conf, programLevel)

	// Инициализируем телеграм бот
	bot, err := telegram.NewClient(conf.TelegramToken)
	if err != nil {
		slog.Error("bot authorization error", "error", err)
	}

	//Инициализируем базу данных
	rep, err := repository.InitDB(ctx)
	if err != nil {
		slog.Error("Database connection error", "error", err)
		os.Exit(1)
	}

	_ = rep.DB

	// Просто тест
	id, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		slog.Error("Невозможно конвертировать число")
	}
	err = bot.SendMessage(id, "Hello World")
	if err != nil {
		slog.Warn("message not sent", "warning", err)
	}
}
