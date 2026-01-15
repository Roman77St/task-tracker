package telegramHandler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"task-traker/internal/service"
	"task-traker/pkg/telegram"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	Bot *telegram.Client
	TaskService *service.TaskService
}

func (h Handler) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	botAPI := h.Bot.GetBotAPI()
	updates := botAPI.GetUpdatesChan(u)

	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case update, ok := <- updates:
			if !ok {
				return nil
			}
			if update.Message == nil {
				continue
			}
			slog.Info("Новое сообщение", "от", update.Message.From.UserName, "текст", update.Message.Text)
			requestCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					h.handleStartCommand(requestCtx, update.Message)
				case "add":
					h.handleAddCommand(requestCtx, update.Message)
				case "list":
					h.handleListCommand(requestCtx, update.Message)
				default:
					h.Bot.SendMessage(update.Message.Chat.ID, "Неизвестная команда")
				}
			}
			cancel()
		}
	}
}

func (h Handler) handleStartCommand(ctx context.Context, m *tgbotapi.Message) {
	h.Bot.SendMessage(m.From.ID, "Привет! Присылайте задачу на контроль!\n Пример: /add Задача, 20.01.2026 15:00")

}

func (h Handler) handleAddCommand(ctx context.Context, m *tgbotapi.Message) {
	userID := m.Chat.ID
	args := m.CommandArguments()
	if args == "" {
		h.Bot.SendMessage(m.Chat.ID, "Пример: /add Задача, 20.01.2026 15:00")
        return
	}
	parts := strings.Split(args, ",")
	if len(parts) != 2 {
		h.Bot.SendMessage(m.Chat.ID, "Не корректно введены данные. Пример:\n/add Задача, 20.01.2026 15:00")
		return
	}
	title := strings.TrimSpace(parts[0])
	deadlineStr := strings.TrimSpace(parts[1])
	err := h.TaskService.CreateTask(ctx, userID, title, deadlineStr)
	if err != nil {
		slog.Error("Task creation filed", "user_id", m.Chat.ID, "error", err)
		h.Bot.SendMessage(m.Chat.ID, "Ошибка создания задачи.")
		return
	}

	h.Bot.SendMessage(m.Chat.ID, "Задача сохранена!")
}

func (h Handler) handleListCommand(ctx context.Context, m *tgbotapi.Message) {
	userID := m.Chat.ID
	tasks, err := h.TaskService.Repo.GetTasksByUserID(ctx, userID)
	if err != nil {
		slog.Error("handleListCommand error", "error", err)
		h.Bot.SendMessage(userID, "Ошибка сервера")
		return
	}
	if len(tasks) == 0 {
		h.Bot.SendMessage(userID, "У вас нет активных задач🎉")
		return
	}
	var msg strings.Builder
	msg.WriteString("📋 Ваши активные задачи:\n\n")
	for i, task := range tasks {
		deadlineStr := task.Deadline.Format("02.01.2006 15:04")
		s := fmt.Sprintf("%d. %s\n ⏰ %s\n\n", i+1, task.Title, deadlineStr)
		msg.WriteString(s)
	}
	h.Bot.SendMessage(userID, msg.String())
}