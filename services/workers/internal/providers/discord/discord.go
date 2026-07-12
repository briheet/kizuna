package discord

import (
	"context"
	"fmt"

	"github.com/briheet/kizuna/workers/internal/config"
	discordsdk "github.com/bwmarrin/discordgo"
)

type Client struct {
	client *discordsdk.Session
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	client, err := discordsdk.New(fmt.Sprintf("%s %s", cfg.Discord.TokenType, cfg.Discord.Token))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}
