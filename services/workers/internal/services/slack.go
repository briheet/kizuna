package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/briheet/kizuna/workers/internal/domain"
	slackprovider "github.com/briheet/kizuna/workers/internal/providers/slack"
	"github.com/briheet/kizuna/workers/internal/repository"
	"github.com/briheet/kizuna/workers/internal/types"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type SlackService struct {
	repo     repository.GraphRepository
	client   *slackprovider.Client
	embedder *EmbedderService
}

func NewSlackService(repo repository.GraphRepository, client *slackprovider.Client, embedder *EmbedderService) *SlackService {
	return &SlackService{repo: repo, client: client, embedder: embedder}
}

func (s *SlackService) HandleJob(ctx context.Context, jobID uuid.UUID, kind string, payload json.RawMessage) error {
	var p types.SlackJobPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	topicID, err := uuid.Parse(p.TopicID)
	if err != nil {
		return err
	}

	sourceID, err := s.repo.UpsertDataSource(ctx, repository.DataSourceInput{
		TopicID:    topicID,
		SourceType: "slack",
		Name:       p.Name,
		ExternalID: fmt.Sprintf("slack:%s", p.Config.TeamID),
		SourceLink: p.SourceLink,
		Config:     payload,
	})
	if err != nil {
		return err
	}

	// Each job scope writes the smallest useful graph for that Slack resource.
	switch domain.JobKind(kind) {
	case domain.JobKindSlackWorkspace:
		return s.handleWorkspace(ctx, sourceID, p)
	case domain.JobKindSlackChannels:
		return s.handleChannels(ctx, sourceID, p)
	case domain.JobKindSlackMessages, domain.JobKindSlackThreads:
		return s.handleMessages(ctx, sourceID, p, domain.JobKind(kind) == domain.JobKindSlackThreads)
	case domain.JobKindSlackUsers:
		return s.handleUsers(ctx, sourceID, p)
	case domain.JobKindSlackFiles:
		return s.handleFiles(ctx, sourceID, p)
	default:
		return fmt.Errorf("unsupported slack job kind: %s", kind)
	}
}

func (s *SlackService) handleWorkspace(ctx context.Context, sourceID uuid.UUID, p types.SlackJobPayload) error {
	// Workspace scope stores only the root Slack workspace node.
	return s.repo.SaveGraph(ctx, sourceID, repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: repository.GraphNodeInput{
		NodeType:   "slack_workspace",
		ExternalID: fmt.Sprintf("slack:%s", p.Config.TeamID),
		SourceLink: p.SourceLink,
		Title:      p.Name,
		Path:       p.Config.TeamID,
	}}}})
}

func (s *SlackService) handleChannels(ctx context.Context, sourceID uuid.UUID, p types.SlackJobPayload) error {
	// Channels scope stores channel metadata and workspace -> channel edges.
	pageSize := p.Config.PageSize
	remaining := p.Config.Limit
	cursor := p.Config.Cursor

	for remaining > 0 {
		currentPageSize := pageSize
		if remaining < currentPageSize {
			currentPageSize = remaining
		}

		channels, nextCursor, err := s.client.ListConversations(ctx, slackprovider.ListConversationsRequest{Limit: currentPageSize, Cursor: cursor})
		if err != nil {
			return err
		}
		if len(channels) == 0 {
			return nil
		}

		graph := repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: s.workspaceNode(p)}}}
		for _, channel := range channels {
			node, err := s.slackChannelNode(p, channel)
			if err != nil {
				return err
			}
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{Node: node})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{
				FromExternalID: s.workspaceExternalID(p),
				ToExternalID:   fmt.Sprintf("%s/channels/%s", s.workspaceExternalID(p), channel.ID),
				EdgeType:       "has_channel",
				EdgeScope:      "slack",
				Confidence:     1,
			})
		}
		if err := s.repo.SaveGraph(ctx, sourceID, graph); err != nil {
			return err
		}

		remaining -= len(channels)
		if nextCursor == "" || len(channels) < currentPageSize {
			return nil
		}
		cursor = nextCursor
	}

	return nil
}

func (s *SlackService) handleMessages(ctx context.Context, sourceID uuid.UUID, p types.SlackJobPayload, includeThreads bool) error {
	// Messages scope fans out per channel; threads scope also includes replies.
	channels, err := s.slackChannels(ctx, p)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	for _, channel := range channels {
		channel := channel
		g.Go(func() error {
			return s.handleChannelMessages(ctx, sourceID, p, channel, includeThreads)
		})
	}
	return g.Wait()
}

func (s *SlackService) handleChannelMessages(ctx context.Context, sourceID uuid.UUID, p types.SlackJobPayload, channel slackprovider.Channel, includeThreads bool) error {
	pageSize := p.Config.PageSize
	remaining := p.Config.Limit
	if p.Config.Remaining > 0 {
		remaining = p.Config.Remaining
	}
	if pageSize <= 0 {
		return fmt.Errorf("slack page_size is required")
	}
	if remaining <= 0 {
		return fmt.Errorf("slack limit is required")
	}

	oldest := ""
	if p.Config.Since != "" {
		since, err := time.Parse(time.RFC3339, p.Config.Since)
		if err != nil {
			return err
		}
		oldest = fmt.Sprintf("%d", since.Unix())
	}

	channelExternalID := fmt.Sprintf("%s/channels/%s", s.workspaceExternalID(p), channel.ID)
	channelNode, err := s.slackChannelNode(p, channel)
	if err != nil {
		return err
	}

	cursor := p.Config.Cursor
	for remaining > 0 {
		currentPageSize := pageSize
		if remaining < currentPageSize {
			currentPageSize = remaining
		}

		history, err := s.client.ListConversationHistory(ctx, slackprovider.ConversationHistoryRequest{ChannelID: channel.ID, Cursor: cursor, Limit: currentPageSize, Oldest: oldest})
		if err != nil {
			return err
		}
		if len(history.Messages) == 0 {
			return nil
		}

		// One page graph keeps the channel, messages, optional replies, chunks, and edges together.
		graph := repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: s.workspaceNode(p)}, {Node: channelNode}}}
		graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: s.workspaceExternalID(p), ToExternalID: channelExternalID, EdgeType: "has_channel", EdgeScope: "slack", Confidence: 1})

		for _, message := range history.Messages {
			messageExternalID := fmt.Sprintf("%s/messages/%s", channelExternalID, message.Timestamp)
			props, err := json.Marshal(message)
			if err != nil {
				return err
			}
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
				Node:   repository.GraphNodeInput{NodeType: "slack_message", ExternalID: messageExternalID, SourceLink: p.SourceLink, Title: message.Timestamp, Path: fmt.Sprintf("channels/%s/messages/%s", channel.ID, message.Timestamp), Properties: props},
				Chunks: []repository.ChunkInput{{Index: 0, Content: message.Text}},
			})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: channelExternalID, ToExternalID: messageExternalID, EdgeType: "has_message", EdgeScope: "slack", Confidence: 1})

			if includeThreads && message.ThreadTimestamp != "" && message.ThreadTimestamp == message.Timestamp {
				replyCursor := ""
				for {
					replies, _, nextCursor, err := s.client.ListConversationReplies(ctx, slackprovider.ConversationRepliesRequest{ChannelID: channel.ID, Timestamp: message.Timestamp, Cursor: replyCursor, Limit: pageSize})
					if err != nil {
						return err
					}
					for _, reply := range replies {
						replyExternalID := fmt.Sprintf("%s/replies/%s", messageExternalID, reply.Timestamp)
						props, err := json.Marshal(reply)
						if err != nil {
							return err
						}
						graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
							Node:   repository.GraphNodeInput{NodeType: "slack_thread_reply", ExternalID: replyExternalID, SourceLink: p.SourceLink, Title: reply.Timestamp, Path: fmt.Sprintf("channels/%s/messages/%s/replies/%s", channel.ID, message.Timestamp, reply.Timestamp), Properties: props},
							Chunks: []repository.ChunkInput{{Index: 0, Content: reply.Text}},
						})
						graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: messageExternalID, ToExternalID: replyExternalID, EdgeType: "has_reply", EdgeScope: "slack", Confidence: 1})
					}
					if nextCursor == "" || len(replies) < pageSize {
						break
					}
					replyCursor = nextCursor
				}
			}
		}

		if err := s.saveGraph(ctx, sourceID, graph); err != nil {
			return err
		}

		remaining -= len(history.Messages)
		cursor = history.ResponseMetaData.NextCursor
		if cursor == "" || len(history.Messages) < currentPageSize {
			return nil
		}
	}

	return nil
}

func (s *SlackService) handleUsers(ctx context.Context, sourceID uuid.UUID, p types.SlackJobPayload) error {
	// Users scope stores Slack users as nodes connected to the workspace.
	users, err := s.client.ListUsers(ctx, slackprovider.ListUsersRequest{Limit: 100, TeamID: p.Config.TeamID})
	if err != nil {
		return err
	}

	graph := repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: s.workspaceNode(p)}}}
	for _, user := range users {
		props, err := json.Marshal(user)
		if err != nil {
			return err
		}
		externalID := fmt.Sprintf("%s/users/%s", s.workspaceExternalID(p), user.ID)
		graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{Node: repository.GraphNodeInput{NodeType: "slack_user", ExternalID: externalID, Title: user.Name, Path: fmt.Sprintf("users/%s", user.ID), Properties: props}})
		graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: s.workspaceExternalID(p), ToExternalID: externalID, EdgeType: "has_user", EdgeScope: "slack", Confidence: 1})
	}
	return s.repo.SaveGraph(ctx, sourceID, graph)
}

func (s *SlackService) handleFiles(ctx context.Context, sourceID uuid.UUID, p types.SlackJobPayload) error {
	// Files scope stores file nodes and embeds their searchable text.
	pageSize := p.Config.PageSize
	remaining := p.Config.Limit
	cursor := p.Config.Cursor

	for remaining > 0 {
		currentPageSize := pageSize
		if remaining < currentPageSize {
			currentPageSize = remaining
		}

		files, next, err := s.client.ListFiles(ctx, slackprovider.ListFilesRequest{Channel: p.Config.ChannelID, Limit: currentPageSize, Cursor: cursor})
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return nil
		}

		graph := repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: s.workspaceNode(p)}}}
		for _, file := range files {
			props, err := json.Marshal(file)
			if err != nil {
				return err
			}
			externalID := fmt.Sprintf("%s/files/%s", s.workspaceExternalID(p), file.ID)
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
				Node:   repository.GraphNodeInput{NodeType: "slack_file", ExternalID: externalID, SourceLink: file.URLPrivate, Title: file.Title, Path: fmt.Sprintf("files/%s", file.ID), Properties: props},
				Chunks: []repository.ChunkInput{{Index: 0, Content: file.Title}},
			})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: s.workspaceExternalID(p), ToExternalID: externalID, EdgeType: "has_file", EdgeScope: "slack", Confidence: 1})
		}
		if err := s.saveGraph(ctx, sourceID, graph); err != nil {
			return err
		}

		remaining -= len(files)
		if next == nil || next.Cursor == "" || len(files) < currentPageSize {
			return nil
		}
		cursor = next.Cursor
	}

	return nil
}

func (s *SlackService) slackChannels(ctx context.Context, p types.SlackJobPayload) ([]slackprovider.Channel, error) {
	// A channel id limits ingestion to one channel; otherwise use visible workspace channels.
	if p.Config.ChannelID != "" {
		channel, err := s.client.GetConversation(ctx, slackprovider.ConversationRequest{ChannelID: p.Config.ChannelID})
		if err != nil {
			return nil, err
		}
		return []slackprovider.Channel{*channel}, nil
	}
	pageSize := p.Config.PageSize
	remaining := p.Config.Limit
	cursor := p.Config.Cursor
	var channels []slackprovider.Channel

	for remaining > 0 {
		currentPageSize := pageSize
		if remaining < currentPageSize {
			currentPageSize = remaining
		}

		page, nextCursor, err := s.client.ListConversations(ctx, slackprovider.ListConversationsRequest{Limit: currentPageSize, Cursor: cursor})
		if err != nil {
			return nil, err
		}
		channels = append(channels, page...)
		remaining -= len(page)
		if nextCursor == "" || len(page) < currentPageSize {
			return channels, nil
		}
		cursor = nextCursor
	}

	return channels, nil
}

func (s *SlackService) saveGraph(ctx context.Context, sourceID uuid.UUID, graph repository.GraphInput) error {
	// Only chunks get embeddings; metadata-only nodes are saved without vectors.
	texts := make([]string, 0)
	for _, node := range graph.Nodes {
		for _, chunk := range node.Chunks {
			if chunk.Content != "" {
				texts = append(texts, chunk.Content)
			}
		}
	}
	if len(texts) == 0 {
		return s.repo.SaveGraph(ctx, sourceID, graph)
	}

	embeddings, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	index := 0
	for nodeIndex := range graph.Nodes {
		for chunkIndex := range graph.Nodes[nodeIndex].Chunks {
			if graph.Nodes[nodeIndex].Chunks[chunkIndex].Content == "" {
				continue
			}
			graph.Nodes[nodeIndex].Chunks[chunkIndex].Embedding = embeddings[index]
			index++
		}
	}
	return s.repo.SaveGraph(ctx, sourceID, graph)
}

func (s *SlackService) workspaceExternalID(p types.SlackJobPayload) string {
	return fmt.Sprintf("slack:%s", p.Config.TeamID)
}

func (s *SlackService) workspaceNode(p types.SlackJobPayload) repository.GraphNodeInput {
	return repository.GraphNodeInput{NodeType: "slack_workspace", ExternalID: s.workspaceExternalID(p), SourceLink: p.SourceLink, Title: p.Name, Path: p.Config.TeamID}
}

func (s *SlackService) slackChannelNode(p types.SlackJobPayload, channel slackprovider.Channel) (repository.GraphNodeInput, error) {
	props, err := json.Marshal(channel)
	if err != nil {
		return repository.GraphNodeInput{}, err
	}
	return repository.GraphNodeInput{NodeType: "slack_channel", ExternalID: fmt.Sprintf("%s/channels/%s", s.workspaceExternalID(p), channel.ID), SourceLink: p.SourceLink, Title: channel.Name, Path: fmt.Sprintf("channels/%s", channel.ID), Properties: props}, nil
}
