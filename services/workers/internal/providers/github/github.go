package github

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	githubsdk "github.com/google/go-github/v89/github"
)

type Client struct {
	client *githubsdk.Client
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	client, err := githubsdk.NewClient(githubsdk.WithAuthToken(cfg.Github.Token))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}
