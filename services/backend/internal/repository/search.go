package repository

import (
	"context"

	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/google/uuid"
)

type SearchRepository interface {
	SearchChunks(ctx context.Context, embedding []float32, limit int) ([]types.SearchResult, error)
	GetRelatedGraph(ctx context.Context, nodeIDs []uuid.UUID) ([]types.RelatedNode, []types.SearchEdge, error)
}

type EmbedderRepository interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
