package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/briheet/kizuna/backend/internal/domain"
	"github.com/briheet/kizuna/backend/internal/repository"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/google/uuid"
)

type IngestionService struct {
	repo repository.IngestionRepository
}

func NewIngestionService(repo repository.IngestionRepository) *IngestionService {
	return &IngestionService{repo: repo}
}

// Here need to birfurcate into different different resources
// These would include issues, pull requests, commits, releases
func (s *IngestionService) CreateGithubJobs(ctx context.Context, req *types.CreateIngestionRequest, cfg *types.CreateGithubJobsConfig) error {
	jobs := make([]types.Job, 0, len(req.Scope))

	for _, scope := range req.Scope {
		var kind string
		switch domain.JobScopeGithub(scope) {
		case domain.JobScopeGithubIssues:
			kind = "github.issues.ingest"
		case domain.JobScopeGithubCommits:
			kind = "github.commits.ingest"
		case domain.JobScopeGithubPullRequests:
			kind = "github.pull_requests.ingest"
		case domain.JobScopeGithubReleases:
			kind = "github.releases.ingest"
		case domain.JobScopeGithubRepository:
			kind = "github.repository.ingest"
		default:
			return fmt.Errorf("unsupported github scope: %s", scope)
		}

		payload, err := json.Marshal(types.GithubIngestionJobPayload{
			TopicID:    req.TopicID,
			SourceType: req.SourceType,
			Name:       req.Name,
			SourceLink: req.SourceLink,
			Scope:      scope,
			Config:     *cfg,
		})
		if err != nil {
			return err
		}

		jobs = append(jobs, types.Job{
			ID:      uuid.New(),
			Kind:    kind,
			Queue:   "github",
			Payload: payload,
		})
	}

	return s.repo.CreateJobs(ctx, jobs)
}
