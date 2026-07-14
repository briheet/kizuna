package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/briheet/kizuna/workers/internal/domain"
	"github.com/briheet/kizuna/workers/internal/providers/github"
	"github.com/briheet/kizuna/workers/internal/repository"
	"github.com/google/uuid"
)

type GithubService struct {
	repo   repository.GithubRepository
	client *github.Client
}

func NewGithubService(repo repository.GithubRepository, client *github.Client) *GithubService {
	return &GithubService{
		repo:   repo,
		client: client,
	}
}

func (s *GithubService) HandleJob(ctx context.Context, jobID uuid.UUID, kind string, payload json.RawMessage) error {

	switch domain.JobKind(kind) {
	case domain.JobKindGithubRepository:
		if err := s.HandleJobKindGithubRepository(); err != nil {
			return err
		}
	case domain.JobKindGithubIssues:
	case domain.JobKindGithubPullRequests:
	case domain.JobKindGithubCommits:
	case domain.JobKindGithubReleases:
	default:
		return fmt.Errorf("unsupported github job kind: %s", kind)
	}

	return s.repo.HandleJob(ctx, jobID, kind, payload)
}

func (s *GithubService) HandleJobKindGithubRepository() error {
	return nil
}
