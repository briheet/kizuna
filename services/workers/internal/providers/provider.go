package providers

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/providers/discord"
	"github.com/briheet/kizuna/workers/internal/providers/github"
	"github.com/briheet/kizuna/workers/internal/providers/slack"
	"github.com/briheet/kizuna/workers/internal/providers/telegram"
)

var ActiveProviders = []string{"github", "discord", "slack", "telegram"}

type Client struct {
	cfg      *config.Config
	github   *github.Client
	discord  *discord.Client
	slack    *slack.Client
	telegram *telegram.Client
}

func NewClientProvider(ctx context.Context, cfg *config.Config) (*Client, error) {
	githubClient, err := github.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	discordClient, err := discord.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	slackClient, err := slack.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	telegramClient, err := telegram.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		cfg:      cfg,
		github:   githubClient,
		discord:  discordClient,
		slack:    slackClient,
		telegram: telegramClient,
	}, nil
}

func (c *Client) Github() *github.Client { return c.github }

func (c *Client) Discord() *discord.Client { return c.discord }

func (c *Client) Slack() *slack.Client { return c.slack }

func (c *Client) Telegram() *telegram.Client { return c.telegram }
