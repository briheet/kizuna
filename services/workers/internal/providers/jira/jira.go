package jira

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	jirasdk "github.com/ctreminiom/go-atlassian/v2/jira/v3"
)

type Client struct {
	client *jirasdk.Client
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	instance, err := jirasdk.New(nil, cfg.Jira.Host)
	if err != nil {
		return nil, err
	}

	instance.Auth.SetBasicAuth(cfg.Jira.Mail, cfg.Jira.Token)
	instance.Auth.SetUserAgent("curl/7.54.0")

	return &Client{
		client: instance,
	}, nil
}

func (c *Client) SearchProjects(ctx context.Context, req SearchProjectsRequest) (*ProjectSearch, *Response, error) {
	return c.client.Project.Search(ctx, req.Options, req.StartAt, req.MaxResults)
}

func (c *Client) GetProject(ctx context.Context, req ProjectRequest) (*Project, *Response, error) {
	return c.client.Project.Get(ctx, req.ProjectKeyOrID, req.Expand)
}

func (c *Client) SearchIssues(ctx context.Context, req SearchIssuesRequest) (*IssueSearch, *Response, error) {
	return c.client.Issue.Search.Get(ctx, req.JQL, req.Fields, req.Expand, req.StartAt, req.MaxResults, req.Validate)
}

func (c *Client) SearchIssuesJQL(ctx context.Context, req SearchIssuesJQLRequest) (*IssueSearchJQL, *Response, error) {
	return c.client.Issue.Search.SearchJQL(ctx, req.JQL, req.Fields, req.Expand, req.MaxResults, req.NextPageToken)
}

func (c *Client) GetIssue(ctx context.Context, req IssueRequest) (*Issue, *Response, error) {
	return c.client.Issue.Get(ctx, req.IssueKeyOrID, req.Fields, req.Expand)
}

func (c *Client) ListIssueComments(ctx context.Context, req ListIssueCommentsRequest) (*IssueCommentPage, *Response, error) {
	return c.client.Issue.Comment.Gets(ctx, req.IssueKeyOrID, req.OrderBy, req.Expand, req.StartAt, req.MaxResults)
}

func (c *Client) GetIssueComment(ctx context.Context, req IssueCommentRequest) (*IssueComment, *Response, error) {
	return c.client.Issue.Comment.Get(ctx, req.IssueKeyOrID, req.CommentID)
}

func (c *Client) ListIssueLinks(ctx context.Context, issueKeyOrID string) (*IssueLinkPage, *Response, error) {
	return c.client.Issue.Link.Gets(ctx, issueKeyOrID)
}

func (c *Client) GetIssueLink(ctx context.Context, linkID string) (*IssueLink, *Response, error) {
	return c.client.Issue.Link.Get(ctx, linkID)
}

func (c *Client) ListLabels(ctx context.Context, req ListLabelsRequest) (*IssueLabels, *Response, error) {
	return c.client.Issue.Label.Gets(ctx, req.StartAt, req.MaxResults)
}

func (c *Client) ListProjectComponents(ctx context.Context, projectKeyOrID string) ([]*Component, *Response, error) {
	return c.client.Project.Component.Gets(ctx, projectKeyOrID)
}

func (c *Client) ListProjectVersions(ctx context.Context, projectKeyOrID string) ([]*Version, *Response, error) {
	return c.client.Project.Version.Gets(ctx, projectKeyOrID)
}

func (c *Client) SearchProjectVersions(ctx context.Context, req SearchProjectVersionsRequest) (*VersionPage, *Response, error) {
	return c.client.Project.Version.Search(ctx, req.ProjectKeyOrID, req.Options, req.StartAt, req.MaxResults)
}

func (c *Client) GetAttachment(ctx context.Context, attachmentID string) (*IssueAttachmentMetadata, *Response, error) {
	return c.client.Issue.Attachment.Metadata(ctx, attachmentID)
}

func (c *Client) GetUser(ctx context.Context, req UserRequest) (*User, *Response, error) {
	return c.client.User.Get(ctx, req.AccountID, req.Expand)
}

func (c *Client) ListUsers(ctx context.Context, req ListUsersRequest) ([]*User, *Response, error) {
	return c.client.User.Gets(ctx, req.StartAt, req.MaxResults)
}
