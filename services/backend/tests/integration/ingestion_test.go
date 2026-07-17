package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/briheet/kizuna/backend/internal/api"
	"github.com/briheet/kizuna/backend/internal/config"
	"github.com/briheet/kizuna/backend/internal/logger"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestCreateIngestionJobs(t *testing.T) {
	db := setupDB(t)
	require.NotNil(t, db)
	require.NotNil(t, db.Db)
	require.NotNil(t, db.Client)

	api := api.NewApi(t.Context(), &config.Config{}, logger.NewNopLogger(), db.Client)

	tests := []struct {
		name         string
		sourceType   string
		sourceName   string
		sourceLink   string
		scopes       []string
		config       map[string]any
		expectedKind []string
		expectedQ    string
	}{
		{
			name:       "github",
			sourceType: "github",
			sourceName: "golang/go",
			sourceLink: "https://github.com/golang/go",
			scopes:     []string{"issues", "pull_requests"},
			config: map[string]any{
				"owner":     "golang",
				"repo":      "go",
				"limit":     100,
				"page_size": 50,
				"page":      1,
			},
			expectedKind: []string{"github.issues.ingest", "github.pull_requests.ingest"},
			expectedQ:    "github",
		},
		{
			name:       "slack",
			sourceType: "slack",
			sourceName: "engineering",
			sourceLink: "https://example.slack.com/archives/C123",
			scopes:     []string{"channels", "messages"},
			config: map[string]any{
				"team_id":    "T123",
				"channel_id": "C123",
				"limit":      100,
				"page_size":  50,
			},
			expectedKind: []string{"slack.channels.ingest", "slack.messages.ingest"},
			expectedQ:    "slack",
		},
		{
			name:       "discord",
			sourceType: "discord",
			sourceName: "go-community",
			sourceLink: "https://discord.com/channels/123/456",
			scopes:     []string{"channels", "messages"},
			config: map[string]any{
				"guild_id":   "123",
				"channel_id": "456",
				"limit":      100,
				"page_size":  50,
			},
			expectedKind: []string{"discord.channels.ingest", "discord.messages.ingest"},
			expectedQ:    "discord",
		},
		{
			name:       "jira",
			sourceType: "jira",
			sourceName: "ENG",
			sourceLink: "https://example.atlassian.net/jira/software/projects/ENG",
			scopes:     []string{"issues", "comments"},
			config: map[string]any{
				"project_key": "ENG",
			},
			expectedKind: []string{"jira.issues.ingest", "jira.comments.ingest"},
			expectedQ:    "jira",
		},
		{
			name:       "confluence",
			sourceType: "confluence",
			sourceName: "Engineering Space",
			sourceLink: "https://example.atlassian.net/wiki/spaces/ENG",
			scopes:     []string{"spaces", "pages"},
			config: map[string]any{
				"space_id": "ENG",
			},
			expectedKind: []string{"confluence.spaces.ingest", "confluence.pages.ingest"},
			expectedQ:    "confluence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.Restore(t)

			body, err := json.Marshal(map[string]any{
				"source_type": tt.sourceType,
				"name":        tt.sourceName,
				"source_link": tt.sourceLink,
				"scope":       tt.scopes,
				"config":      tt.config,
			})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/createJobs", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			api.Routes().ServeHTTP(rec, req)

			require.Equal(t, http.StatusAccepted, rec.Code)

			var created types.CreateIngestionResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&created))
			require.NotEmpty(t, created.TopicID)
			require.Equal(t, tt.sourceType, created.SourceType)
			require.Equal(t, len(tt.scopes), created.JobsCreated)
			require.Equal(t, "available", created.State)

			rows, err := db.Client.Conn().Query(t.Context(), `
				SELECT kind, queue, payload
				FROM jobs
				ORDER BY kind;
			`)
			require.NoError(t, err)
			defer rows.Close()

			jobs := make(map[string]struct {
				Queue   string
				Payload map[string]any
			})

			for rows.Next() {
				var kind string
				var queue string
				var payload []byte

				require.NoError(t, rows.Scan(&kind, &queue, &payload))

				var decoded map[string]any
				require.NoError(t, json.Unmarshal(payload, &decoded))

				jobs[kind] = struct {
					Queue   string
					Payload map[string]any
				}{
					Queue:   queue,
					Payload: decoded,
				}
			}
			require.NoError(t, rows.Err())

			require.Len(t, jobs, len(tt.expectedKind))

			expectedScopes := make(map[string]bool, len(tt.scopes))
			for _, scope := range tt.scopes {
				expectedScopes[scope] = true
			}

			for _, kind := range tt.expectedKind {
				job, ok := jobs[kind]
				require.True(t, ok, "missing job kind %s", kind)
				require.Equal(t, tt.expectedQ, job.Queue)
				require.Equal(t, created.TopicID, job.Payload["topic_id"])
				require.Equal(t, tt.sourceType, job.Payload["source_type"])
				require.Equal(t, tt.sourceName, job.Payload["name"])
				require.Equal(t, tt.sourceLink, job.Payload["source_link"])

				scope, ok := job.Payload["scope"].(string)
				require.True(t, ok)
				require.True(t, expectedScopes[scope], "unexpected scope %s", scope)
			}
		})
	}
}

func TestJobsStatus(t *testing.T) {
	db := setupDB(t)
	db.Restore(t)

	app := api.NewApi(t.Context(), &config.Config{}, logger.NewNopLogger(), db.Client)
	body, err := json.Marshal(map[string]any{
		"source_type": "github",
		"name":        "golang/go",
		"source_link": "https://github.com/golang/go",
		"scope":       []string{"issues", "pull_requests"},
		"config": map[string]any{
			"owner":     "golang",
			"repo":      "go",
			"limit":     100,
			"page_size": 50,
			"page":      1,
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/createJobs", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())

	var created types.CreateIngestionResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&created))
	require.NotEmpty(t, created.TopicID)

	err = db.Client.ExecuteTx(t.Context(), func(tx pgx.Tx) error {
		_, err := tx.Exec(t.Context(), `
			update jobs
			set state = 'discarded', attempt = max_attempt, attempted_at = now(), updated_at = now()
			where kind = 'github.issues.ingest';
		`)
		return err
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/jobsStatus?topic_id="+created.TopicID+"&source_type=github", nil)
	rec = httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp types.JobsStatusResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	counts := map[string]int{}
	for _, count := range resp.Counts {
		counts[count.State] = count.Count
	}

	require.Equal(t, 1, counts["available"])
	require.Equal(t, 1, counts["discarded"])
	require.Len(t, resp.RecentFailures, 1)
	require.Equal(t, "github.issues.ingest", resp.RecentFailures[0].Kind)
	require.Equal(t, "discarded", resp.RecentFailures[0].State)
	require.Equal(t, "github", resp.RecentFailures[0].Queue)
}
