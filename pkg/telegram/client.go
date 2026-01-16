package telegram

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	bot *tgbotapi.BotAPI
}

func NewClient(tg_token string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(tg_token)
	if err != nil {
		return nil, err
	}
	slog.Info("bot authorized", "account", bot.Self.UserName)
	return &Client{bot: bot}, nil
}

func (c *Client) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := c.bot.Send(msg)
	return err
}

func (c *Client) SendMessageWithMarkup(chatID int64, msg tgbotapi.MessageConfig) {
	if _, err := c.bot.Send(msg); err != nil {
		slog.Error("error sending message with buttons", "error", err)
	}
}

func (c *Client) GetBotAPI() *tgbotapi.BotAPI {
	return c.bot
}
