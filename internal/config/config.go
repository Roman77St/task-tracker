package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	TelegramToken string
	LogLevel      string
}

func New() (*Config, error) {
	tg_token := os.Getenv("TELEGRAM_TOKEN")
	if tg_token == "" {
		return nil, fmt.Errorf("environment variable \"TELEGRAM_TOKEN\" is missing")
	}
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel == "" {
		loglevel = "info"
	}
	return &Config{
		TelegramToken: tg_token,
		LogLevel:      loglevel,
	}, nil
}

func SetConfigLevel(conf Config, programmLevel *slog.LevelVar) {
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