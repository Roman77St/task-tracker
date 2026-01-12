package main

import (
	"log/slog"
	"os"
	"strings"
	"task-traker/internal/config"
	"task-traker/pkg/telegram"

	"github.com/joho/godotenv"
)

func main() {
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
	config, err := config.New()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	// Устанавливаем уровень логгирования из конфига
	setConfigLevel(*config, programLevel)

	// Инициализируем телеграм бот
	bot, err := telegram.NewClient(config.TelegramToken)
	if err != nil {
		slog.Error("bot authorization error", "error", err)
	}

	// Просто тест
	err = bot.SendMessage(1178229376, "Hello World")
	if err != nil {
		slog.Warn("message not sent", "warning", err)
	}
}

func setConfigLevel(conf config.Config, programmLevel *slog.LevelVar) {
	level := strings.ToUpper(conf.LogLevel)
	if level != "INFO" {
		switch level {
		case "DEBUG":
			programmLevel.Set(slog.LevelDebug)
		case "ERROR":
			programmLevel.Set(slog.LevelError)
		case "WARN", "WARNING":
			programmLevel.Set(slog.LevelWarn)
		default:
			programmLevel.Set(slog.LevelInfo)
		}
	}
}
