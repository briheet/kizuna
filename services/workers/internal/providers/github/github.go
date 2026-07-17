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

func (c *Client) GetRepository(ctx context.Context, req RepoRequest) (*Repository, *Response, error) {
	return c.client.Repositories.Get(ctx, req.Owner, req.Repo)
}

func (c *Client) GetReadme(ctx context.Context, req RepoRequest) (*RepositoryContent, *Response, error) {
	return c.client.Repositories.GetReadme(ctx, req.Owner, req.Repo, nil)
}

func (c *Client) ListIssues(ctx context.Context, req ListIssuesRequest) ([]*Issue, *Response, error) {
	return c.client.Issues.ListByRepo(ctx, req.Owner, req.Repo, &githubsdk.IssueListByRepoOptions{
		State:     req.State,
		Sort:      req.Sort,
		Direction: req.Direction,
		Since:     req.Since,
		ListOptions: githubsdk.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
}

func (c *Client) ListIssueComments(ctx context.Context, req IssueRequest) ([]*IssueComment, *Response, error) {
	return c.client.Issues.ListComments(ctx, req.Owner, req.Repo, req.Number, &githubsdk.IssueListCommentsOptions{
		ListOptions: githubsdk.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
}

func (c *Client) ListPullRequests(ctx context.Context, req ListPullRequestsRequest) ([]*PullRequest, *Response, error) {
	return c.client.PullRequests.List(ctx, req.Owner, req.Repo, &githubsdk.PullRequestListOptions{
		State:     req.State,
		Sort:      req.Sort,
		Direction: req.Direction,
		ListOptions: githubsdk.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
}

func (c *Client) ListPullRequestComments(ctx context.Context, req PullRequestRequest) ([]*PullRequestComment, *Response, error) {
	return c.client.PullRequests.ListComments(ctx, req.Owner, req.Repo, req.Number, &githubsdk.PullRequestListCommentsOptions{
		ListOptions: githubsdk.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
}

func (c *Client) ListPullRequestReviews(ctx context.Context, req PullRequestRequest) ([]*PullRequestReview, *Response, error) {
	return c.client.PullRequests.ListReviews(ctx, req.Owner, req.Repo, req.Number, &githubsdk.ListOptions{
		Page:    req.Page,
		PerPage: req.PerPage,
	})
}

func (c *Client) ListPullRequestCommits(ctx context.Context, req PullRequestRequest) ([]*RepositoryCommit, *Response, error) {
	return c.client.PullRequests.ListCommits(ctx, req.Owner, req.Repo, req.Number, &githubsdk.ListOptions{
		Page:    req.Page,
		PerPage: req.PerPage,
	})
}

func (c *Client) ListCommits(ctx context.Context, req ListCommitsRequest) ([]*RepositoryCommit, *Response, error) {
	return c.client.Repositories.ListCommits(ctx, req.Owner, req.Repo, &githubsdk.CommitsListOptions{
		SHA:   req.SHA,
		Since: req.Since,
		Until: req.Until,
		ListOptions: githubsdk.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
}

func (c *Client) ListLabels(ctx context.Context, req ListLabelsRequest) ([]*Label, *Response, error) {
	return c.client.Issues.ListLabels(ctx, req.Owner, req.Repo, &githubsdk.ListOptions{
		Page:    req.Page,
		PerPage: req.PerPage,
	})
}

func (c *Client) ListIssueLabels(ctx context.Context, req IssueRequest) ([]*Label, *Response, error) {
	return c.client.Issues.ListLabelsByIssue(ctx, req.Owner, req.Repo, req.Number, &githubsdk.ListOptions{
		Page:    req.Page,
		PerPage: req.PerPage,
	})
}

func (c *Client) ListMilestones(ctx context.Context, req ListMilestonesRequest) ([]*Milestone, *Response, error) {
	return c.client.Issues.ListMilestones(ctx, req.Owner, req.Repo, &githubsdk.MilestoneListOptions{
		State:     req.State,
		Sort:      req.Sort,
		Direction: req.Direction,
		ListOptions: githubsdk.ListOptions{
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
}

func (c *Client) ListReleases(ctx context.Context, req ListReleasesRequest) ([]*RepositoryRelease, *Response, error) {
	return c.client.Repositories.ListReleases(ctx, req.Owner, req.Repo, &githubsdk.ListOptions{
		Page:    req.Page,
		PerPage: req.PerPage,
	})
}
