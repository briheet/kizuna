package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	aiclient "github.com/briheet/kizuna/backend/internal/ai"
	"github.com/briheet/kizuna/backend/internal/config"
	"github.com/briheet/kizuna/backend/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestOpenAIRepositorySummarize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/responses", r.URL.Path)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		var request responseRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "gpt-5.6-luna", request.Model)
		require.Equal(t, 700, request.MaxOutputTokens)
		require.False(t, request.Store)
		require.Contains(t, request.Instructions, "using only the retrieved sources")
		require.Contains(t, request.Input, "Where is the deployment guide?")
		require.Contains(t, request.Input, "github_pull_request")

		require.NoError(t, json.NewEncoder(w).Encode(responsePayload{
			Status: "completed",
			Output: []struct {
				Type    string `json:"type"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			}{
				{
					Type: "message",
					Content: []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{{Type: "output_text", Text: "Use the production runbook [1]."}},
				},
			},
		}))
	}))
	defer server.Close()

	client := aiclient.NewClient(&config.Config{AI: config.AIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	}})
	repo := NewOpenAIRepository(client, "gpt-5.6-luna", 700)

	answer, err := repo.Summarize(t.Context(), "Where is the deployment guide?", []repository.AnswerSource{
		{
			Reference:  1,
			Title:      "Production runbook",
			SourceType: "github",
			NodeType:   "github_pull_request",
			Content:    "Deploy with the production workflow.",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Use the production runbook [1].", answer)
}

func TestOpenAIRepositoryRejectsMissingText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(responsePayload{Status: "completed"}))
	}))
	defer server.Close()

	client := aiclient.NewClient(&config.Config{AI: config.AIConfig{APIKey: "test-key", BaseURL: server.URL}})
	repo := NewOpenAIRepository(client, "gpt-5.6-luna", 700)

	_, err := repo.Summarize(t.Context(), "question", nil)
	require.ErrorContains(t, err, "response contained no text")
}

func TestBuildAnswerInputLimitsContext(t *testing.T) {
	input, err := buildAnswerInput("question", []repository.AnswerSource{
		{Reference: 1, Content: strings.Repeat("a", maxSourceCharacters+100)},
	})
	require.NoError(t, err)
	require.Less(t, len([]rune(input)), maxSourceCharacters+500)
}
