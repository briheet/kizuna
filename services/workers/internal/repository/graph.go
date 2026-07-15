package repository

import (
	"context"

	"github.com/google/uuid"
)

type GraphRepository interface {
	UpsertDataSource(ctx context.Context, input DataSourceInput) (uuid.UUID, error)
	SaveGraph(ctx context.Context, dataSourceID uuid.UUID, graph GraphInput) error
}

type DataSourceInput struct {
	TopicID    uuid.UUID
	SourceType string
	Name       string
	ExternalID string
	SourceLink string
	Config     []byte
}

type GraphInput struct {
	Nodes []GraphNodeWithChunks
	Edges []GraphEdgeInput
}

type GraphNodeWithChunks struct {
	Node   GraphNodeInput
	Chunks []ChunkInput
}

type GraphNodeInput struct {
	NodeType   string
	ExternalID string
	SourceLink string
	Title      string
	Path       string
	Properties []byte
}

type ChunkInput struct {
	Index     int
	Content   string
	Embedding []float32
}

type GraphEdgeInput struct {
	FromExternalID string
	ToExternalID   string
	EdgeType       string
	EdgeScope      string
	Confidence     float64
	Properties     []byte
}
