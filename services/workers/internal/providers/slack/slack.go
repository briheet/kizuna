package slack

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	slacksdk "github.com/slack-go/slack"
)

type Client struct {
	client *slacksdk.Client
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	client := slacksdk.New(cfg.Slack.Token)

	return &Client{
		client: client,
	}, nil
}
