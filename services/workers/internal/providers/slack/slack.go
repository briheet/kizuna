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

func (c *Client) ListConversations(ctx context.Context, req ListConversationsRequest) ([]Channel, string, error) {
	return c.client.GetConversationsContext(ctx, &req)
}

func (c *Client) GetConversation(ctx context.Context, req ConversationRequest) (*Channel, error) {
	return c.client.GetConversationInfoContext(ctx, &req)
}

func (c *Client) ListConversationHistory(ctx context.Context, req ConversationHistoryRequest) (*ConversationHistoryResponse, error) {
	return c.client.GetConversationHistoryContext(ctx, &req)
}

func (c *Client) ListConversationReplies(ctx context.Context, req ConversationRepliesRequest) ([]Message, bool, string, error) {
	return c.client.GetConversationRepliesContext(ctx, &req)
}

func (c *Client) ListConversationMembers(ctx context.Context, req ConversationMembersRequest) ([]string, string, error) {
	return c.client.GetUsersInConversationContext(ctx, &req)
}

func (c *Client) ListUsers(ctx context.Context, req ListUsersRequest) ([]User, error) {
	return c.client.GetUsersContext(
		ctx,
		slacksdk.GetUsersOptionCursor(req.Cursor),
		slacksdk.GetUsersOptionLimit(req.Limit),
		slacksdk.GetUsersOptionPresence(req.Presence),
		slacksdk.GetUsersOptionTeamID(req.TeamID),
	)
}

func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	return c.client.GetUserInfoContext(ctx, userID)
}

func (c *Client) ListFiles(ctx context.Context, req ListFilesRequest) ([]File, *ListFilesRequest, error) {
	return c.client.ListFilesContext(ctx, req)
}

func (c *Client) GetFile(ctx context.Context, req FileRequest) (*File, []Comment, *Paging, error) {
	return c.client.GetFileInfoContext(ctx, req.FileID, req.Count, req.Page)
}

func (c *Client) GetReactions(ctx context.Context, req ReactionsRequest) (ReactedItem, error) {
	return c.client.GetReactionsContext(ctx, req.Item, req.Params)
}

func (c *Client) ListReactions(ctx context.Context, req ListReactionsRequest) ([]ReactedItem, string, error) {
	return c.client.ListReactionsContext(ctx, req)
}
