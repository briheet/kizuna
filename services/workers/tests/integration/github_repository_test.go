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

func TestGraphRepositorySaveGithubGraph(t *testing.T) {
	db := setupDB(t)
	db.Restore(t)

	ctx := t.Context()
	orgID := uuid.New()
	teamID := uuid.New()
	topicID := uuid.New()

	err := db.Client.ExecuteTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			insert into organisations (id, name, created_at) values ($1, 'org', now());
		`, orgID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			insert into teams (id, organisation_id, name, created_at) values ($1, $2, 'team', now());
		`, teamID, orgID); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			insert into topics (id, team_id, name, created_at) values ($1, $2, 'topic', now());
		`, topicID, teamID)
		return err
	})
	require.NoError(t, err)

	repo := cockroachdb.NewGraphRepository(db.Client)
	config, err := json.Marshal(map[string]string{"owner": "golang", "repo": "go"})
	require.NoError(t, err)

	dataSourceID, err := repo.UpsertDataSource(ctx, repository.DataSourceInput{
		TopicID:    topicID,
		SourceType: "github",
		Name:       "golang/go",
		ExternalID: "github:golang/go",
		SourceLink: "https://github.com/golang/go",
		Config:     config,
	})
	require.NoError(t, err)

	sameDataSourceID, err := repo.UpsertDataSource(ctx, repository.DataSourceInput{
		TopicID:    topicID,
		SourceType: "github",
		Name:       "golang/go",
		ExternalID: "github:golang/go",
		SourceLink: "https://github.com/golang/go",
		Config:     config,
	})
	require.NoError(t, err)
	require.Equal(t, dataSourceID, sameDataSourceID)

	props, err := json.Marshal(map[string]any{"number": 42})
	require.NoError(t, err)
	embedding := make([]float32, 768)
	embedding[0] = 0.1

	err = repo.SaveGraph(ctx, dataSourceID, repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{
		Node: repository.GraphNodeInput{
			NodeType:   "github_issue",
			ExternalID: "github:golang/go/issues/42",
			SourceLink: "https://github.com/golang/go/issues/42",
			Title:      "issue title",
			Path:       "issues/42",
			Properties: props,
		},
		Chunks: []repository.ChunkInput{
			{Index: 0, Content: "issue body"},
			{Index: 1, Content: "comment body"},
		},
	}}})
	require.NoError(t, err)

	err = repo.SaveGraph(ctx, dataSourceID, repository.GraphInput{Nodes: []repository.GraphNodeWithChunks{{
		Node: repository.GraphNodeInput{
			NodeType:   "github_issue",
			ExternalID: "github:golang/go/issues/42",
			SourceLink: "https://github.com/golang/go/issues/42",
			Title:      "updated issue title",
			Path:       "issues/42",
			Properties: props,
		},
		Chunks: []repository.ChunkInput{
			{Index: 0, Content: "updated issue body", Embedding: embedding},
		},
	}}})
	require.NoError(t, err)

	var nodeCount int
	err = db.Client.Conn().QueryRow(ctx, `
		select count(*) from graph_nodes where data_source_id = $1;
	`, dataSourceID).Scan(&nodeCount)
	require.NoError(t, err)
	require.Equal(t, 1, nodeCount)

	var title string
	var chunkCount int
	var content string
	var hasEmbedding bool
	err = db.Client.Conn().QueryRow(ctx, `
		select gn.title, count(c.id), min(c.content), bool_and(c.embedding is not null)
		from graph_nodes gn
		join chunks c on c.graph_node_id = gn.id
		where gn.data_source_id = $1
		group by gn.title;
	`, dataSourceID).Scan(&title, &chunkCount, &content, &hasEmbedding)
	require.NoError(t, err)
	require.Equal(t, "updated issue title", title)
	require.Equal(t, 2, chunkCount)
	require.Equal(t, "comment body", content)
	require.False(t, hasEmbedding)

	err = db.Client.Conn().QueryRow(ctx, `
		select content, embedding is not null
		from chunks
		where graph_node_id = (
			select id from graph_nodes where data_source_id = $1 and external_id = 'github:golang/go/issues/42'
		)
		  and chunk_index = 0;
	`, dataSourceID).Scan(&content, &hasEmbedding)
	require.NoError(t, err)
	require.Equal(t, "updated issue body", content)
	require.True(t, hasEmbedding)

	err = repo.SaveGraph(ctx, dataSourceID, repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_repository",
					ExternalID: "github:golang/go",
					SourceLink: "https://github.com/golang/go",
					Title:      "golang/go",
					Path:       "golang/go",
				},
			},
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_issue_comment",
					ExternalID: "github:golang/go/issues/42/comments/1",
					SourceLink: "https://github.com/golang/go/issues/42#comment-1",
					Title:      "Issue #42 comment",
					Path:       "issues/42/comments/1",
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: "comment body", Embedding: embedding}},
			},
		},
		Edges: []repository.GraphEdgeInput{{
			FromExternalID: "github:golang/go",
			ToExternalID:   "github:golang/go/issues/42/comments/1",
			EdgeType:       "has_comment",
			EdgeScope:      "github",
			Confidence:     1,
		}},
	})
	require.NoError(t, err)

	var edgeCount int
	err = db.Client.Conn().QueryRow(ctx, `
		select count(*) from graph_edges where root_data_source_id = $1 and edge_type = 'has_comment';
	`, dataSourceID).Scan(&edgeCount)
	require.NoError(t, err)
	require.Equal(t, 1, edgeCount)
}
