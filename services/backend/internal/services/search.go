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
}

func NewSearchService(repo repository.SearchRepository, embedder repository.EmbedderRepository) *SearchService {
	return &SearchService{repo: repo, embedder: embedder}
}

func (s *SearchService) Search(ctx context.Context, req types.SearchRequest) (*types.SearchResponse, error) {
	topicID, err := uuid.Parse(req.TopicID)
	if err != nil {
		return nil, err
	}

	vectors, err := s.embedder.Embed(ctx, []string{req.Query})
	if err != nil {
		return nil, err
	}

	results, err := s.repo.SearchChunks(ctx, topicID, vectors[0], req.Limit)
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

	return &types.SearchResponse{Results: results, RelatedNodes: nodes, Edges: edges}, nil
}
