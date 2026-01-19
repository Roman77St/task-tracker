package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"task-traker/internal/config"
	httpHandler "task-traker/internal/delivery/http"
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

	// Загружаем переменные среды из файла .env
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "text", "Direct loading of environment variables is possible.")
		// при загрузке в docker (docker run --env-file .env my-app-image) os.Exit не нужен!
		// os.Exit(1)
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

	telegramHandler := telegramHandler.Handler{
		Bot:         bot,
		TaskService: &taskService,
		Sessions: make(map[int64]*telegramHandler.UserSession),
	}
	// Запуск телеграм бота
    go func() {
		slog.Info("Starting a telegram bot")
		err = telegramHandler.Start(ctx)
		if err != nil {
			slog.Error("Telegram answer", "error", err)
		}
	}()

	httpH := httpHandler.NewHandler(&taskService)
	addr := os.Getenv("HTTP_ADDR")
	srv := &http.Server{
		Addr : addr,
		Handler: httpH.InitRouter(),
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func ()  {
		text := fmt.Sprintf("Запуск HTTP сервера на %s", addr)
		slog.Info(text)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed{
			slog.Error("HTTP server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	srv.Shutdown(shutdownCtx)
	slog.Info("App exited")
}
