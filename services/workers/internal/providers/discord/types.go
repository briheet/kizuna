package discord

import (
	"time"

	discordsdk "github.com/bwmarrin/discordgo"
)

type GuildRequest struct {
	GuildID string
}

type ChannelRequest struct {
	ChannelID string
}

type MessageRequest struct {
	ChannelID string
	MessageID string
}

type ListMessagesRequest struct {
	ChannelID string
	Limit     int
	BeforeID  string
	AfterID   string
	AroundID  string
}

type ListMembersRequest struct {
	GuildID string
	AfterID string
	Limit   int
}

type ListThreadsRequest struct {
	ChannelID string
	Before    *time.Time
	Limit     int
}

type ListThreadMembersRequest struct {
	ThreadID   string
	AfterID    string
	Limit      int
	WithMember bool
}

type ListReactionsRequest struct {
	ChannelID string
	MessageID string
	EmojiID   string
	Limit     int
	BeforeID  string
	AfterID   string
}

type Guild = discordsdk.Guild
type Channel = discordsdk.Channel
type Message = discordsdk.Message
type User = discordsdk.User
type Member = discordsdk.Member
type ThreadMember = discordsdk.ThreadMember
type ThreadsList = discordsdk.ThreadsList
