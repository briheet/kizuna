package telegram

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	client *tgbotapi.BotAPI
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}
