package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"task-traker/internal/config"
	"task-traker/internal/delivery/telegramHandler"
	"task-traker/internal/repository"
	"task-traker/internal/service"
	"task-traker/pkg/telegram"

	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// Инициализируем начальный логгер
	programLevel := &slog.LevelVar{}
	// для продакшн
	// logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	// для разработки
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})

	slog.SetDefault(slog.New(logHandler))

	// Загружаем переменные среды из файла .envh.Bot.SendMessage(m.Chat.ID, "Пример: /add Задача, 20.01.2026 15:00")
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
	db, err := repository.InitDB(ctx)
	if err != nil {
		slog.Error("Database connection error", "error", err)
		os.Exit(1)
	}
	defer db.DB.Close()

	// Создаем сервис и передаём ему репозиторий.
	taskService := service.TaskService{
		Repo: db,
	}

	// Запуск воркера уведомлений
	go func() {
		slog.Info("Starting a background notification worker")
		taskService.StartNotificationWorker(ctx, bot)
	}()

	handler := telegramHandler.Handler{
		Bot: bot,
		TaskService: &taskService,
	}

	err = handler.Start(ctx)
	if err != nil {
		slog.Error("Telegram answer", "error", err)
	}
}
