package providers

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/providers/confluence"
	"github.com/briheet/kizuna/workers/internal/providers/discord"
	"github.com/briheet/kizuna/workers/internal/providers/github"
	"github.com/briheet/kizuna/workers/internal/providers/jira"
	"github.com/briheet/kizuna/workers/internal/providers/slack"
)

type Client struct {
	cfg        *config.Config
	confluence *confluence.Client
	github     *github.Client
	discord    *discord.Client
	slack      *slack.Client
	jira       *jira.Client
}

func NewClientProvider(ctx context.Context, cfg *config.Config) (*Client, error) {
	githubClient, err := github.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := &Client{
		cfg:    cfg,
		github: githubClient,
	}

	if cfg.Discord.Token != "" {
		client.discord, err = discord.NewClient(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Slack.Token != "" {
		client.slack, err = slack.NewClient(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Confluence.Token != "" {
		client.confluence, err = confluence.NewClient(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Jira.Token != "" {
		client.jira, err = jira.NewClient(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func EnabledProviders(cfg *config.Config) []string {
	active := []string{"github"}
	if cfg.Discord.Token != "" {
		active = append(active, "discord")
	}
	if cfg.Slack.Token != "" {
		active = append(active, "slack")
	}
	if cfg.Confluence.Token != "" {
		active = append(active, "confluence")
	}
	if cfg.Jira.Token != "" {
		active = append(active, "jira")
	}
	return active
}

func (c *Client) Github() *github.Client { return c.github }

func (c *Client) Discord() *discord.Client { return c.discord }

func (c *Client) Slack() *slack.Client { return c.slack }

func (c *Client) Confluence() *confluence.Client { return c.confluence }

func (c *Client) Jira() *jira.Client { return c.jira }
