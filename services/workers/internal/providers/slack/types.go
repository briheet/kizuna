package slack

import slacksdk "github.com/slack-go/slack"

type ListConversationsRequest = slacksdk.GetConversationsParameters
type ConversationRequest = slacksdk.GetConversationInfoInput
type ConversationHistoryRequest = slacksdk.GetConversationHistoryParameters
type ConversationRepliesRequest = slacksdk.GetConversationRepliesParameters
type ConversationMembersRequest = slacksdk.GetUsersInConversationParameters
type ListFilesRequest = slacksdk.ListFilesParameters
type ListReactionsRequest = slacksdk.ListReactionsParameters

type ListUsersRequest struct {
	Cursor   string
	Limit    int
	Presence bool
	TeamID   string
}

type FileRequest struct {
	FileID string
	Count  int
	Page   int
}

type ReactionsRequest struct {
	Item   slacksdk.ItemRef
	Params slacksdk.GetReactionsParameters
}

type Channel = slacksdk.Channel
type Message = slacksdk.Message
type User = slacksdk.User
type File = slacksdk.File
type Comment = slacksdk.Comment
type Paging = slacksdk.Paging
type ReactedItem = slacksdk.ReactedItem
type ConversationHistoryResponse = slacksdk.GetConversationHistoryResponse
