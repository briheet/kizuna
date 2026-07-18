package embedder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/briheet/kizuna/workers/internal/config"
	embedderclient "github.com/briheet/kizuna/workers/internal/embedder"
	"github.com/stretchr/testify/require"
)

func newTestRepository(server *httptest.Server) *NomicRepository {
	client := embedderclient.NewClient(&config.Config{
		Embedder: config.EmbedderConfig{BaseURL: server.URL},
	})
	return NewNomicRepository(client, "nomic-embed-text:v1.5")
}

func TestNomicRepositoryEmbedsDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/embed", r.URL.Path)

		var request nomicEmbeddingRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "nomic-embed-text:v1.5", request.Model)
		require.Equal(t, []string{"search_document: first", "search_document: second"}, request.Input)

		vectors := [][]float32{make([]float32, nomicDimensions), make([]float32, nomicDimensions)}
		require.NoError(t, json.NewEncoder(w).Encode(nomicEmbeddingResponse{Embeddings: vectors}))
	}))
	defer server.Close()

	repository := newTestRepository(server)
	vectors, err := repository.EmbedDocuments(t.Context(), []string{"first", "second"})
	require.NoError(t, err)
	require.Len(t, vectors, 2)
	require.Len(t, vectors[0], nomicDimensions)
	require.Equal(t, nomicDimensions, repository.Dimensions())
}

func TestNomicRepositoryEmbedsQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request nomicEmbeddingRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, []string{"search_query: where is the runbook?"}, request.Input)
		require.NoError(t, json.NewEncoder(w).Encode(nomicEmbeddingResponse{
			Embeddings: [][]float32{make([]float32, nomicDimensions)},
		}))
	}))
	defer server.Close()

	repository := newTestRepository(server)
	vector, err := repository.EmbedQuery(t.Context(), "where is the runbook?")
	require.NoError(t, err)
	require.Len(t, vector, nomicDimensions)
}

func TestNomicRepositoryRejectsInvalidDimensions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(nomicEmbeddingResponse{Embeddings: [][]float32{{1, 2, 3}}}))
	}))
	defer server.Close()

	_, err := newTestRepository(server).EmbedDocuments(t.Context(), []string{"text"})
	require.ErrorContains(t, err, "3 dimensions, expected 768")
}

func TestNomicRepositoryRejectsEmptyInput(t *testing.T) {
	repository := NewNomicRepository(nil, "nomic-embed-text:v1.5")

	_, err := repository.EmbedDocuments(t.Context(), []string{" "})
	require.ErrorContains(t, err, "input 0 is empty")

	_, err = repository.EmbedQuery(t.Context(), "")
	require.ErrorContains(t, err, "input 0 is empty")
}

func TestNomicRepositoryReturnsServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model is not loaded", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := newTestRepository(server).EmbedDocuments(t.Context(), []string{"text"})
	require.ErrorContains(t, err, "503 Service Unavailable: model is not loaded")
}
