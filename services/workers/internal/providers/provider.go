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

var ActiveProviders = []string{"github", "discord", "slack", "confluence", "jira"}

type Client struct {
	cfg        *config.Config
	confluence *confluence.Client
	github     *github.Client
	discord    *discord.Client
	slack      *slack.Client
	jira       *jira.Client
}

func NewClientProvider(ctx context.Context, cfg *config.Config) (*Client, error) {
	confluenceClient, err := confluence.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

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

	jiraClient, err := jira.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		cfg:        cfg,
		confluence: confluenceClient,
		github:     githubClient,
		discord:    discordClient,
		slack:      slackClient,
		jira:       jiraClient,
	}, nil
}

func (c *Client) Github() *github.Client { return c.github }

func (c *Client) Discord() *discord.Client { return c.discord }

func (c *Client) Slack() *slack.Client { return c.slack }

func (c *Client) Confluence() *confluence.Client { return c.confluence }

func (c *Client) Jira() *jira.Client { return c.jira }
