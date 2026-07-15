package repository

import (
	"context"

	"github.com/google/uuid"
)

type GithubRepository interface {
	UpsertDataSource(ctx context.Context, input GithubDataSourceInput) (uuid.UUID, error)
	SaveNodeWithChunks(ctx context.Context, dataSourceID uuid.UUID, node GithubGraphNodeInput, chunks []GithubChunkInput) error
	SaveGithubGraph(ctx context.Context, dataSourceID uuid.UUID, graph GithubGraphInput) error
}

type GithubDataSourceInput struct {
	TopicID    uuid.UUID
	Name       string
	ExternalID string
	SourceLink string
	Config     []byte
}

type GithubGraphNodeInput struct {
	NodeType   string
	ExternalID string
	SourceLink string
	Title      string
	Path       string
	Properties []byte
}

type GithubChunkInput struct {
	Index     int
	Content   string
	Embedding []float32
}

type GithubGraphInput struct {
	Nodes []GithubGraphNodeWithChunks
	Edges []GithubGraphEdgeInput
}

type GithubGraphNodeWithChunks struct {
	Node   GithubGraphNodeInput
	Chunks []GithubChunkInput
}

type GithubGraphEdgeInput struct {
	FromExternalID string
	ToExternalID   string
	EdgeType       string
	EdgeScope      string
	Confidence     float64
	Properties     []byte
}
