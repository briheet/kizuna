package services

import (
	"context"

	"github.com/briheet/kizuna/backend/internal/repository"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/google/uuid"
)

type SearchService struct {
	repo     repository.SearchRepository
	embedder repository.EmbedderRepository
	answerer repository.AnswerRepository
}

func NewSearchService(
	repo repository.SearchRepository,
	embedder repository.EmbedderRepository,
	answerer repository.AnswerRepository,
) *SearchService {
	return &SearchService{repo: repo, embedder: embedder, answerer: answerer}
}

func (s *SearchService) Search(ctx context.Context, req types.SearchRequest) (*types.SearchResponse, error) {
	vectors, err := s.embedder.Embed(ctx, []string{"search_query: " + req.Query})
	if err != nil {
		return nil, err
	}

	results, err := s.repo.SearchChunks(ctx, vectors[0], req.Limit)
	if err != nil {
		return nil, err
	}

	nodeIDs := make([]uuid.UUID, 0, len(results))
	for _, result := range results {
		id, err := uuid.Parse(result.GraphNodeID)
		if err != nil {
			return nil, err
		}
		nodeIDs = append(nodeIDs, id)
	}

	nodes, edges, err := s.repo.GetRelatedGraph(ctx, nodeIDs)
	if err != nil {
		return nil, err
	}

	answerSources := make([]repository.AnswerSource, len(results))
	for index, result := range results {
		answerSources[index] = repository.AnswerSource{
			Reference:  index + 1,
			Title:      result.Title,
			SourceType: result.SourceType,
			NodeType:   result.NodeType,
			Content:    result.Content,
		}
	}

	summary, err := s.answerer.Summarize(ctx, req.Query, answerSources)
	if err != nil {
		return nil, err
	}

	return &types.SearchResponse{
		Summary:      summary,
		Results:      results,
		RelatedNodes: nodes,
		Edges:        edges,
	}, nil
}
