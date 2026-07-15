package services

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/repository"
)

type EmbedderService struct {
	repo repository.EmbedderRepository
}

func NewEmbedderService(repo repository.EmbedderRepository) *EmbedderService {
	return &EmbedderService{repo: repo}
}

func (s *EmbedderService) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return s.repo.Embed(ctx, texts)
}

func (s *EmbedderService) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	vectors, err := s.repo.Embed(ctx, []string{text})
	if err != nil || len(vectors) == 0 {
		return nil, err
	}

	return vectors[0], nil
}

func (s *EmbedderService) Dimensions() int {
	return s.repo.Dimensions()
}
