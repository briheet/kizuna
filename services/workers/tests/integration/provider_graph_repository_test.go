package integration

import (
	"encoding/json"
	"testing"

	"github.com/briheet/kizuna/workers/internal/repository"
	"github.com/briheet/kizuna/workers/internal/repository/cockroachdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestGraphRepositorySaveSlackGraph(t *testing.T) {
	db := setupDB(t)
	db.Restore(t)

	ctx := t.Context()
	topicID := createTopic(t, db)
	repo := cockroachdb.NewGraphRepository(db.Client)

	config, err := json.Marshal(map[string]string{"team_id": "T123", "channel_id": "C123"})
	require.NoError(t, err)

	dataSourceID, err := repo.UpsertDataSource(ctx, repository.DataSourceInput{
		TopicID:    topicID,
		SourceType: "slack",
		Name:       "slack engineering",
		ExternalID: "slack:T123",
		SourceLink: "https://slack.com/app_redirect?channel=C123",
		Config:     config,
	})
	require.NoError(t, err)

	embedding := make([]float32, 768)
	embedding[0] = 0.2

	err = repo.SaveGraph(ctx, dataSourceID, repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{
			{
				Node: repository.GraphNodeInput{
					NodeType:   "slack_workspace",
					ExternalID: "slack:T123",
					SourceLink: "https://slack.com",
					Title:      "slack engineering",
					Path:       "T123",
				},
			},
			{
				Node: repository.GraphNodeInput{
					NodeType:   "slack_message",
					ExternalID: "slack:T123/channels/C123/messages/1710000000.000100",
					SourceLink: "https://slack.com/app_redirect?channel=C123",
					Title:      "1710000000.000100",
					Path:       "channels/C123/messages/1710000000.000100",
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: "deploy failed in staging", Embedding: embedding}},
			},
		},
		Edges: []repository.GraphEdgeInput{{
			FromExternalID: "slack:T123",
			ToExternalID:   "slack:T123/channels/C123/messages/1710000000.000100",
			EdgeType:       "has_message",
			EdgeScope:      "slack",
			Confidence:     1,
		}},
	})
	require.NoError(t, err)

	assertProviderGraph(t, db, dataSourceID, "slack_message", "deploy failed in staging", "has_message")
}

func TestGraphRepositorySaveDiscordGraph(t *testing.T) {
	db := setupDB(t)
	db.Restore(t)

	ctx := t.Context()
	topicID := createTopic(t, db)
	repo := cockroachdb.NewGraphRepository(db.Client)

	config, err := json.Marshal(map[string]string{"guild_id": "G123", "channel_id": "C123"})
	require.NoError(t, err)

	dataSourceID, err := repo.UpsertDataSource(ctx, repository.DataSourceInput{
		TopicID:    topicID,
		SourceType: "discord",
		Name:       "discord engineering",
		ExternalID: "discord:G123",
		SourceLink: "https://discord.com/channels/G123/C123",
		Config:     config,
	})
	require.NoError(t, err)

	embedding := make([]float32, 768)
	embedding[0] = 0.3

	err = repo.SaveGraph(ctx, dataSourceID, repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{
			{
				Node: repository.GraphNodeInput{
					NodeType:   "discord_guild",
					ExternalID: "discord:G123",
					SourceLink: "https://discord.com/channels/G123",
					Title:      "discord engineering",
					Path:       "G123",
				},
			},
			{
				Node: repository.GraphNodeInput{
					NodeType:   "discord_message",
					ExternalID: "discord:G123/channels/C123/messages/M123",
					SourceLink: "https://discord.com/channels/G123/C123/M123",
					Title:      "alice",
					Path:       "channels/C123/messages/M123",
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: "panic in worker logs", Embedding: embedding}},
			},
		},
		Edges: []repository.GraphEdgeInput{{
			FromExternalID: "discord:G123",
			ToExternalID:   "discord:G123/channels/C123/messages/M123",
			EdgeType:       "has_message",
			EdgeScope:      "discord",
			Confidence:     1,
		}},
	})
	require.NoError(t, err)

	assertProviderGraph(t, db, dataSourceID, "discord_message", "panic in worker logs", "has_message")
}

func createTopic(t *testing.T, db *BaseDB) uuid.UUID {
	t.Helper()

	orgID := uuid.New()
	teamID := uuid.New()
	topicID := uuid.New()

	err := db.Client.ExecuteTx(t.Context(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(t.Context(), `
			insert into organisations (id, name, created_at) values ($1, 'org', now());
		`, orgID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `
			insert into teams (id, organisation_id, name, created_at) values ($1, $2, 'team', now());
		`, teamID, orgID); err != nil {
			return err
		}
		_, err := tx.Exec(t.Context(), `
			insert into topics (id, team_id, name, created_at) values ($1, $2, 'topic', now());
		`, topicID, teamID)
		return err
	})
	require.NoError(t, err)

	return topicID
}

func assertProviderGraph(t *testing.T, db *BaseDB, dataSourceID uuid.UUID, nodeType string, content string, edgeType string) {
	t.Helper()

	var nodeCount int
	err := db.Client.Conn().QueryRow(t.Context(), `
		select count(*) from graph_nodes where data_source_id = $1;
	`, dataSourceID).Scan(&nodeCount)
	require.NoError(t, err)
	require.Equal(t, 2, nodeCount)

	var savedContent string
	var hasEmbedding bool
	err = db.Client.Conn().QueryRow(t.Context(), `
		select c.content, c.embedding is not null
		from chunks c
		join graph_nodes gn on gn.id = c.graph_node_id
		where gn.data_source_id = $1 and gn.node_type = $2 and c.chunk_index = 0;
	`, dataSourceID, nodeType).Scan(&savedContent, &hasEmbedding)
	require.NoError(t, err)
	require.Equal(t, content, savedContent)
	require.True(t, hasEmbedding)

	var edgeCount int
	err = db.Client.Conn().QueryRow(t.Context(), `
		select count(*) from graph_edges where root_data_source_id = $1 and edge_type = $2;
	`, dataSourceID, edgeType).Scan(&edgeCount)
	require.NoError(t, err)
	require.Equal(t, 1, edgeCount)
}
