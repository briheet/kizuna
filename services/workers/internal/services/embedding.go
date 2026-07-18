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
	return s.repo.EmbedDocuments(ctx, texts)
}

func (s *EmbedderService) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return s.repo.EmbedQuery(ctx, text)
}

func (s *EmbedderService) Dimensions() int {
	return s.repo.Dimensions()
}
