package providers

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/providers/github"
)

var ActiveProviders = []string{"github"}

type Client struct {
	github *github.Client
}

func NewClientProvider(ctx context.Context, cfg *config.Config) (*Client, error) {
	githubClient, err := github.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		github: githubClient,
	}, nil
}

func (c *Client) Github() *github.Client { return c.github }
