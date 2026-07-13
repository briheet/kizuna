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
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGithubIngestion(t *testing.T) {

	db := setupDB(t)
	require.NotNil(t, db)
	require.NotNil(t, db.Db)
	require.NotNil(t, db.Client)
	db.Restore(t)

	topicID := uuid.New().String()
	body := []byte(`{
		"topic_id": "` + topicID + `",
		"source_type": "github",
		"name": "golang/go",
		"source_link": "https://github.com/golang/go",
		"scope": ["issues", "pull_requests"],
		"config": {
			"owner": "golang",
			"repo": "go"
		}
	}`)

	api := api.NewApi(t.Context(), &config.Config{}, logger.NewNopLogger(), db.Client)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/createJobs", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	api.Routes().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	rows, err := db.Client.Conn().Query(t.Context(), `
		SELECT kind, queue, payload
		FROM jobs
		ORDER BY kind;
	`)
	require.NoError(t, err)
	defer rows.Close()

	jobs := make(map[string]struct {
		Kind    string
		Queue   string
		Payload types.GithubIngestionJobPayload
	})

	for rows.Next() {
		var job struct {
			Kind    string
			Queue   string
			Payload types.GithubIngestionJobPayload
		}
		var payload []byte

		require.NoError(t, rows.Scan(&job.Kind, &job.Queue, &payload))
		require.NoError(t, json.Unmarshal(payload, &job.Payload))

		jobs[job.Kind] = job
	}
	require.NoError(t, rows.Err())

	require.Len(t, jobs, 2)

	issuesJob := jobs["github.issues.ingest"]
	require.Equal(t, "github", issuesJob.Queue)
	require.Equal(t, "issues", issuesJob.Payload.Scope)
	require.Equal(t, topicID, issuesJob.Payload.TopicID)
	require.Equal(t, "golang", issuesJob.Payload.Config.Owner)
	require.Equal(t, "go", issuesJob.Payload.Config.Repo)

	pullRequestsJob := jobs["github.pull_requests.ingest"]
	require.Equal(t, "github", pullRequestsJob.Queue)
	require.Equal(t, "pull_requests", pullRequestsJob.Payload.Scope)
	require.Equal(t, topicID, pullRequestsJob.Payload.TopicID)
	require.Equal(t, "golang", pullRequestsJob.Payload.Config.Owner)
	require.Equal(t, "go", pullRequestsJob.Payload.Config.Repo)
}
