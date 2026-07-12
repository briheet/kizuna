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

func (c *Client) GetGuild(ctx context.Context, req GuildRequest) (*Guild, error) {
	return c.client.Guild(req.GuildID)
}

func (c *Client) ListGuildChannels(ctx context.Context, req GuildRequest) ([]*Channel, error) {
	return c.client.GuildChannels(req.GuildID)
}

func (c *Client) GetChannel(ctx context.Context, req ChannelRequest) (*Channel, error) {
	return c.client.Channel(req.ChannelID)
}

func (c *Client) ListMessages(ctx context.Context, req ListMessagesRequest) ([]*Message, error) {
	return c.client.ChannelMessages(req.ChannelID, req.Limit, req.BeforeID, req.AfterID, req.AroundID)
}

func (c *Client) GetMessage(ctx context.Context, req MessageRequest) (*Message, error) {
	return c.client.ChannelMessage(req.ChannelID, req.MessageID)
}

func (c *Client) ListGuildMembers(ctx context.Context, req ListMembersRequest) ([]*Member, error) {
	return c.client.GuildMembers(req.GuildID, req.AfterID, req.Limit)
}

func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	return c.client.User(userID)
}

func (c *Client) ListActiveThreads(ctx context.Context, req ChannelRequest) (*ThreadsList, error) {
	return c.client.ThreadsActive(req.ChannelID)
}

func (c *Client) ListArchivedThreads(ctx context.Context, req ListThreadsRequest) (*ThreadsList, error) {
	return c.client.ThreadsArchived(req.ChannelID, req.Before, req.Limit)
}

func (c *Client) ListPrivateArchivedThreads(ctx context.Context, req ListThreadsRequest) (*ThreadsList, error) {
	return c.client.ThreadsPrivateArchived(req.ChannelID, req.Before, req.Limit)
}

func (c *Client) ListJoinedPrivateArchivedThreads(ctx context.Context, req ListThreadsRequest) (*ThreadsList, error) {
	return c.client.ThreadsPrivateJoinedArchived(req.ChannelID, req.Before, req.Limit)
}

func (c *Client) ListThreadMembers(ctx context.Context, req ListThreadMembersRequest) ([]*ThreadMember, error) {
	return c.client.ThreadMembers(req.ThreadID, req.Limit, req.WithMember, req.AfterID)
}

func (c *Client) ListReactions(ctx context.Context, req ListReactionsRequest) ([]*User, error) {
	return c.client.MessageReactions(req.ChannelID, req.MessageID, req.EmojiID, req.Limit, req.BeforeID, req.AfterID)
}
