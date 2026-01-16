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

type State int

const (
	StateIdle State = iota
	StateWaitTaskTitle
	StateWaitTaskDeadline
)

type UserSession struct {
	State State
	Title string
}

type Handler struct {
	Bot         *telegram.Client
	TaskService *service.TaskService
	Sessions map[int64]*UserSession
}

func (h Handler) Start(ctx context.Context) error {
	if h.Sessions == nil {
		h.Sessions = make(map[int64]*UserSession)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	botAPI := h.Bot.GetBotAPI()
	updates := botAPI.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if update.Message == nil {
				continue
			}

			slog.Info("Новое сообщение", "от", update.Message.From.UserName, "текст", update.Message.Text)
			userID := update.Message.Chat.ID
			requestCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			session, ok := h.Sessions[userID]
			if !ok {
				session = &UserSession{State: StateIdle}
				h.Sessions[userID] = session
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					h.handleStartCommand(requestCtx, update.Message)
				case "add":
					h.handleAddCommand(requestCtx, update.Message)
					session.State = StateWaitTaskTitle
				case "list":
					h.handleListCommand(requestCtx, update.Message)
				default:
					h.Bot.SendMessage(userID, "Неизвестная команда")
				}
				cancel()
				continue
			}

			switch session.State {
			case StateWaitTaskTitle:
				h.handleAddTitleTask(requestCtx, update.Message, session)
			case StateWaitTaskDeadline:
				h.handleAddDeadlineTask(requestCtx, update.Message, session)
			case StateIdle:
				switch update.Message.Text {
				case "➕ Добавить задачу":
					h.handleAddCommand(requestCtx, update.Message)
					session.State = StateWaitTaskTitle
				case "📋 Все задачи":
					h.handleListCommand(requestCtx, update.Message)
					// Добавить логику
				default:
					h.Bot.SendMessage(userID, "Используйте кнопки меню или команды.")
				}
			}

			cancel()
		}
	}
}

func (h Handler) handleStartCommand(ctx context.Context, m *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(m.From.ID,
		"Привет! Я запоминаю задачи и присылаю уведомления о дедлайне.")
	msg.ReplyMarkup = mainMenuKeyboard()
	h.Bot.SendMessageWithMarkup(m.From.ID, msg)
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

func (h Handler) handleAddCommand(ctx context.Context, m *tgbotapi.Message) {
	h.Bot.SendMessage(m.From.ID, "Напишите текс задачи")
}

func (h Handler) handleAddTitleTask(ctx context.Context, m *tgbotapi.Message, session *UserSession){
	session.Title = m.Text
	session.State = StateWaitTaskDeadline
	h.Bot.SendMessage(m.Chat.ID, "Теперь введите дату и время (ДД.ММ.ГГГГ ЧЧ:ММ):")
}

func (h Handler) handleAddDeadlineTask(ctx context.Context, m *tgbotapi.Message, session *UserSession){
	err := h.TaskService.CreateTask(ctx, m.From.ID, session.Title, m.Text)
	if err != nil {
		slog.Error("Task creation filed", "error", err)
		h.Bot.SendMessage(m.From.ID, "Ошибка формата даты. Попробуйте еще раз.")
		return
	}
	h.Bot.SendMessage(m.From.ID, "✅ Задача сохранена!")
	session.State = StateIdle
	session.Title = ""
}