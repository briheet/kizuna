package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/briheet/kizuna/backend/internal/api"
	"github.com/briheet/kizuna/backend/internal/config"
	"github.com/briheet/kizuna/backend/internal/logger"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	db := setupDB(t)
	db.Restore(t)

	vector := func(first float32) string {
		values := make([]string, 768)
		values[0] = strconv.FormatFloat(float64(first), 'f', -1, 32)
		for i := 1; i < len(values); i++ {
			values[i] = "0"
		}
		return "[" + strings.Join(values, ",") + "]"
	}

	nomic := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/embed", r.URL.Path)
		var request types.NomicEmbeddingRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "nomic-embed-text:v1.5", request.Model)
		require.Equal(t, []string{"search_query: find nearest"}, request.Input)

		embedding := make([]float32, 768)
		embedding[0] = 1
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embeddings": [][]float32{embedding},
		})
	}))
	defer nomic.Close()

	openAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/responses", r.URL.Path)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"status": "completed",
			"output": []map[string]any{
				{
					"type": "message",
					"content": []map[string]any{
						{"type": "output_text", "text": "The nearest result is supported by the retrieved records [1]."},
					},
				},
			},
		}))
	}))
	defer openAI.Close()

	app := api.NewApi(t.Context(), &config.Config{
		AI: config.AIConfig{
			APIKey:          "test-key",
			BaseURL:         openAI.URL,
			Model:           "gpt-5.6-luna",
			MaxOutputTokens: 700,
		},
		Embedder: config.EmbedderConfig{
			BaseURL: nomic.URL,
			Model:   "nomic-embed-text:v1.5",
		},
	}, logger.NewNopLogger(), db.Client)

	orgID := uuid.New()
	teamID := uuid.New()
	topicID := uuid.New()
	secondTopicID := uuid.New()
	sourceID := uuid.New()
	secondSourceID := uuid.New()
	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()

	err := db.Client.ExecuteTx(t.Context(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(t.Context(), `insert into organisations (id, name, created_at) values ($1, 'org', now());`, orgID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into teams (id, organisation_id, name, created_at) values ($1, $2, 'team', now());`, teamID, orgID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into topics (id, team_id, name, created_at) values ($1, $2, 'topic', now());`, topicID, teamID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into topics (id, team_id, name, created_at) values ($1, $2, 'second topic', now());`, secondTopicID, teamID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into data_sources (id, topic_id, source_type, name, created_at, updated_at) values ($1, $2, 'github', 'repo', now(), now());`, sourceID, topicID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into data_sources (id, topic_id, source_type, name, created_at, updated_at) values ($1, $2, 'slack', 'workspace', now(), now());`, secondSourceID, secondTopicID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into graph_nodes (id, data_source_id, node_type, external_id, title, source_link, created_at, updated_at) values ($1, $2, 'github_issue', 'a', 'nearest', 'https://example.com/a', now(), now()), ($3, $2, 'github_issue', 'b', 'far', 'https://example.com/b', now(), now());`, nodeA, sourceID, nodeB); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into graph_nodes (id, data_source_id, node_type, external_id, title, source_link, created_at, updated_at) values ($1, $2, 'slack_message', 'c', 'cross-topic', 'https://example.com/c', now(), now());`, nodeC, secondSourceID); err != nil {
			return err
		}
		if _, err := tx.Exec(t.Context(), `insert into graph_edges (id, root_data_source_id, from_node_id, to_node_id, edge_type, edge_scope, confidence, created_at) values ($1, $2, $3, $4, 'related_to', 'github', 1, now());`, uuid.New(), sourceID, nodeA, nodeB); err != nil {
			return err
		}
		_, err := tx.Exec(t.Context(), `insert into chunks (id, graph_node_id, chunk_index, content, embedding, created_at) values ($1, $2, 0, 'nearest chunk', $3::VECTOR, now()), ($4, $5, 0, 'far chunk', $6::VECTOR, now()), ($7, $8, 0, 'cross-topic chunk', $9::VECTOR, now());`, uuid.New(), nodeA, vector(1), uuid.New(), nodeB, vector(-1), uuid.New(), nodeC, vector(0.5))
		return err
	})
	require.NoError(t, err)

	body, err := json.Marshal(types.SearchRequest{
		Query: "find nearest",
		Limit: 3,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	app.Routes().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotEmpty(t, rec.Body.String())

	var resp types.SearchResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Results, 3)
	require.Equal(t, "The nearest result is supported by the retrieved records [1].", resp.Summary)
	require.Equal(t, "nearest chunk", resp.Results[0].Content)
	require.Equal(t, "cross-topic chunk", resp.Results[1].Content)
	require.Equal(t, "nearest", resp.Results[0].Title)
	require.Equal(t, "https://example.com/a", resp.Results[0].SourceLink)
	require.Equal(t, "github", resp.Results[0].SourceType)
	require.Equal(t, "github_issue", resp.Results[0].NodeType)
	require.Len(t, resp.Edges, 1)
	require.Equal(t, "related_to", resp.Edges[0].EdgeType)
	require.Len(t, resp.RelatedNodes, 3)

	links := map[string]bool{}
	for _, node := range resp.RelatedNodes {
		links[node.SourceLink] = true
	}
	require.True(t, links["https://example.com/a"])
	require.True(t, links["https://example.com/b"])
}
