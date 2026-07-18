package services

import (
	"context"
	"testing"

	"github.com/briheet/kizuna/backend/internal/repository"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type searchRepositoryStub struct {
	results []types.SearchResult
}

func (s *searchRepositoryStub) SearchChunks(_ context.Context, _ []float32, _ int) ([]types.SearchResult, error) {
	return s.results, nil
}

func (s *searchRepositoryStub) GetRelatedGraph(_ context.Context, _ []uuid.UUID) ([]types.RelatedNode, []types.SearchEdge, error) {
	return nil, nil, nil
}

type embedderRepositoryStub struct {
	input []string
}

func (s *embedderRepositoryStub) Embed(_ context.Context, texts []string) ([][]float32, error) {
	s.input = texts
	return [][]float32{{1}}, nil
}

type answerRepositoryStub struct {
	question string
	sources  []repository.AnswerSource
}

func (s *answerRepositoryStub) Summarize(_ context.Context, question string, sources []repository.AnswerSource) (string, error) {
	s.question = question
	s.sources = sources
	return "Grounded answer [1].", nil
}

func TestSearchRunsGroundedAnswerFlow(t *testing.T) {
	nodeID := uuid.New().String()
	searchRepo := &searchRepositoryStub{results: []types.SearchResult{
		{
			ChunkID:     uuid.New().String(),
			GraphNodeID: nodeID,
			Content:     "Deploy through the production workflow.",
			Title:       "Deployment guide",
			SourceType:  "github",
			NodeType:    "github_pull_request",
			SourceLink:  "https://github.com/example/repo/pull/1",
		},
	}}
	embedderRepo := &embedderRepositoryStub{}
	answerRepo := &answerRepositoryStub{}
	service := NewSearchService(searchRepo, embedderRepo, answerRepo)

	response, err := service.Search(t.Context(), types.SearchRequest{Query: "How do I deploy?", Limit: 6})
	require.NoError(t, err)
	require.Equal(t, []string{"search_query: How do I deploy?"}, embedderRepo.input)
	require.Equal(t, "How do I deploy?", answerRepo.question)
	require.Equal(t, []repository.AnswerSource{
		{
			Reference:  1,
			Title:      "Deployment guide",
			SourceType: "github",
			NodeType:   "github_pull_request",
			Content:    "Deploy through the production workflow.",
		},
	}, answerRepo.sources)
	require.Equal(t, "Grounded answer [1].", response.Summary)
	require.Equal(t, "https://github.com/example/repo/pull/1", response.Results[0].SourceLink)
}
