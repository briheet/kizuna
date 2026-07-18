package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/briheet/kizuna/workers/internal/domain"
	discordprovider "github.com/briheet/kizuna/workers/internal/providers/discord"
	"github.com/briheet/kizuna/workers/internal/repository"
	"github.com/briheet/kizuna/workers/internal/types"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type DiscordService struct {
	repo     repository.GraphRepository
	client   *discordprovider.Client
	embedder *EmbedderService
}

func NewDiscordService(repo repository.GraphRepository, client *discordprovider.Client, embedder *EmbedderService) *DiscordService {
	return &DiscordService{repo: repo, client: client, embedder: embedder}
}

func (s *DiscordService) HandleJob(ctx context.Context, jobID uuid.UUID, kind string, payload json.RawMessage) error {
	var p types.DiscordJobPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	topicID, err := uuid.Parse(p.TopicID)
	if err != nil {
		return err
	}

	sourceID, err := s.repo.UpsertDataSource(ctx, repository.DataSourceInput{
		TopicID:    topicID,
		SourceType: "discord",
		Name:       p.Name,
		ExternalID: fmt.Sprintf("discord:%s", p.Config.GuildID),
		SourceLink: p.SourceLink,
		Config:     payload,
	})
	if err != nil {
		return err
	}

	// Each job scope writes the smallest useful graph for that Discord resource.
	switch domain.JobKind(kind) {
	case domain.JobKindDiscordGuild:
		return s.handleGuild(ctx, sourceID, p)
	case domain.JobKindDiscordChannels:
		return s.handleChannels(ctx, sourceID, p)
	case domain.JobKindDiscordMessages, domain.JobKindDiscordThreads:
		return s.handleMessages(ctx, sourceID, p)
	case domain.JobKindDiscordMembers:
		return s.handleMembers(ctx, sourceID, p)
	default:
		return fmt.Errorf("unsupported discord job kind: %s", kind)
	}
}

func (s *DiscordService) handleGuild(ctx context.Context, sourceID uuid.UUID, p types.DiscordJobPayload) error {
	// Guild scope stores only the root Discord guild node.
	guild, err := s.client.GetGuild(ctx, discordprovider.GuildRequest{GuildID: p.Config.GuildID})
	if err != nil {
		return err
	}
	props, err := json.Marshal(guild)
	if err != nil {
		return err
	}
	guildExternalID := fmt.Sprintf("discord:%s", p.Config.GuildID)
	return s.repo.SaveGraph(ctx, sourceID, repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: repository.GraphNodeInput{NodeType: "discord_guild", ExternalID: guildExternalID, SourceLink: p.SourceLink, Title: guild.Name, Path: p.Config.GuildID, Properties: props}}}})
}

func (s *DiscordService) handleChannels(ctx context.Context, sourceID uuid.UUID, p types.DiscordJobPayload) error {
	// Channels scope stores channel metadata and guild -> channel edges.
	channels, err := s.discordChannels(ctx, p)
	if err != nil {
		return err
	}

	guildExternalID := fmt.Sprintf("discord:%s", p.Config.GuildID)
	graph := repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: repository.GraphNodeInput{NodeType: "discord_guild", ExternalID: guildExternalID, SourceLink: p.SourceLink, Title: p.Name, Path: p.Config.GuildID}}}}
	for _, channel := range channels {
		props, err := json.Marshal(channel)
		if err != nil {
			return err
		}
		channelExternalID := fmt.Sprintf("%s/channels/%s", guildExternalID, channel.ID)
		graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{Node: repository.GraphNodeInput{NodeType: "discord_channel", ExternalID: channelExternalID, SourceLink: fmt.Sprintf("https://discord.com/channels/%s/%s", p.Config.GuildID, channel.ID), Title: channel.Name, Path: fmt.Sprintf("channels/%s", channel.ID), Properties: props}})
		graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: guildExternalID, ToExternalID: channelExternalID, EdgeType: "has_channel", EdgeScope: "discord", Confidence: 1})
	}
	return s.repo.SaveGraph(ctx, sourceID, graph)
}

func (s *DiscordService) handleMessages(ctx context.Context, sourceID uuid.UUID, p types.DiscordJobPayload) error {
	// Messages scope fans out per channel and stores message chunks for search.
	channels, err := s.discordChannels(ctx, p)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	for _, channel := range channels {
		g.Go(func() error {
			return s.handleChannelMessages(ctx, sourceID, p, channel)
		})
	}
	return g.Wait()
}

func (s *DiscordService) handleChannelMessages(ctx context.Context, sourceID uuid.UUID, p types.DiscordJobPayload, channel *discordprovider.Channel) error {
	pageSize := p.Config.PageSize
	remaining := p.Config.Limit
	if p.Config.Remaining > 0 {
		remaining = p.Config.Remaining
	}
	if pageSize <= 0 {
		return fmt.Errorf("discord page_size is required")
	}
	if remaining <= 0 {
		return fmt.Errorf("discord limit is required")
	}

	guildExternalID := fmt.Sprintf("discord:%s", p.Config.GuildID)
	channelExternalID := fmt.Sprintf("%s/channels/%s", guildExternalID, channel.ID)
	channelProps, err := json.Marshal(channel)
	if err != nil {
		return err
	}

	beforeID := p.Config.BeforeMessageID
	for remaining > 0 {
		currentPageSize := pageSize
		if remaining < currentPageSize {
			currentPageSize = remaining
		}

		messages, err := s.client.ListMessages(ctx, discordprovider.ListMessagesRequest{ChannelID: channel.ID, Limit: currentPageSize, BeforeID: beforeID})
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}

		// One page graph keeps the channel, messages, chunks, and edges together.
		graph := repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{
			{Node: repository.GraphNodeInput{NodeType: "discord_guild", ExternalID: guildExternalID, SourceLink: p.SourceLink, Title: p.Name, Path: p.Config.GuildID}},
			{Node: repository.GraphNodeInput{NodeType: "discord_channel", ExternalID: channelExternalID, SourceLink: fmt.Sprintf("https://discord.com/channels/%s/%s", p.Config.GuildID, channel.ID), Title: channel.Name, Path: fmt.Sprintf("channels/%s", channel.ID), Properties: channelProps}},
		}}
		graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: guildExternalID, ToExternalID: channelExternalID, EdgeType: "has_channel", EdgeScope: "discord", Confidence: 1})

		for _, message := range messages {
			messageExternalID := fmt.Sprintf("%s/messages/%s", channelExternalID, message.ID)
			props, err := json.Marshal(message)
			if err != nil {
				return err
			}
			author := ""
			if message.Author != nil {
				author = message.Author.Username
			}
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
				Node:   repository.GraphNodeInput{NodeType: "discord_message", ExternalID: messageExternalID, SourceLink: fmt.Sprintf("https://discord.com/channels/%s/%s/%s", p.Config.GuildID, channel.ID, message.ID), Title: author, Path: fmt.Sprintf("channels/%s/messages/%s", channel.ID, message.ID), Properties: props},
				Chunks: []repository.ChunkInput{{Index: 0, Content: message.Content}},
			})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: channelExternalID, ToExternalID: messageExternalID, EdgeType: "has_message", EdgeScope: "discord", Confidence: 1})
		}

		if err := s.saveGraph(ctx, sourceID, graph); err != nil {
			return err
		}

		remaining -= len(messages)
		if len(messages) < currentPageSize {
			return nil
		}
		beforeID = messages[len(messages)-1].ID
	}

	return nil
}

func (s *DiscordService) handleMembers(ctx context.Context, sourceID uuid.UUID, p types.DiscordJobPayload) error {
	// Members scope stores guild members as nodes connected to the guild.
	members, err := s.client.ListGuildMembers(ctx, discordprovider.ListMembersRequest{GuildID: p.Config.GuildID, Limit: 100})
	if err != nil {
		return err
	}

	guildExternalID := fmt.Sprintf("discord:%s", p.Config.GuildID)
	graph := repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{Node: repository.GraphNodeInput{NodeType: "discord_guild", ExternalID: guildExternalID, SourceLink: p.SourceLink, Title: p.Name, Path: p.Config.GuildID}}}}
	for _, member := range members {
		if member.User == nil {
			continue
		}
		props, err := json.Marshal(member)
		if err != nil {
			return err
		}
		externalID := fmt.Sprintf("%s/members/%s", guildExternalID, member.User.ID)
		graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{Node: repository.GraphNodeInput{NodeType: "discord_member", ExternalID: externalID, Title: member.User.Username, Path: fmt.Sprintf("members/%s", member.User.ID), Properties: props}})
		graph.Edges = append(graph.Edges, repository.GraphEdgeInput{FromExternalID: guildExternalID, ToExternalID: externalID, EdgeType: "has_member", EdgeScope: "discord", Confidence: 1})
	}
	return s.repo.SaveGraph(ctx, sourceID, graph)
}

func (s *DiscordService) discordChannels(ctx context.Context, p types.DiscordJobPayload) ([]*discordprovider.Channel, error) {
	// A channel id limits ingestion to one channel; otherwise use visible guild channels.
	if p.Config.ChannelID != "" {
		channel, err := s.client.GetChannel(ctx, discordprovider.ChannelRequest{ChannelID: p.Config.ChannelID})
		if err != nil {
			return nil, err
		}
		return []*discordprovider.Channel{channel}, nil
	}
	return s.client.ListGuildChannels(ctx, discordprovider.GuildRequest{GuildID: p.Config.GuildID})
}

func (s *DiscordService) saveGraph(ctx context.Context, sourceID uuid.UUID, graph repository.GraphInput) error {
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
