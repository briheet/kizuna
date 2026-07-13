package services

import "github.com/briheet/kizuna/backend/internal/repository"

type IngestionService struct {
	repo repository.IngestionRepository
}

func NewIngestionService(repo repository.IngestionRepository) *IngestionService {
	return &IngestionService{repo: repo}
}
