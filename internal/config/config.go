package config

import (
	"fmt"
	"os"
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
