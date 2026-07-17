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

func (s *IngestionService) JobsStatus(ctx context.Context, req types.JobsStatusRequest) (*types.JobsStatusResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}
	return s.repo.JobsStatus(ctx, req)
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

	return s.repo.CreateJobs(ctx, req.TopicID, req.Name, jobs)
}

func (s *IngestionService) CreateSlackJobs(ctx context.Context, req *types.CreateIngestionRequest, cfg *types.CreateSlackJobsConfig) error {
	jobs := make([]types.Job, 0, len(req.Scope))

	for _, scope := range req.Scope {
		var kind string
		switch domain.JobScopeSlack(scope) {
		case domain.JobScopeSlackWorkspace:
			kind = "slack.workspace.ingest"
		case domain.JobScopeSlackChannels:
			kind = "slack.channels.ingest"
		case domain.JobScopeSlackMessages:
			kind = "slack.messages.ingest"
		case domain.JobScopeSlackThreads:
			kind = "slack.threads.ingest"
		case domain.JobScopeSlackUsers:
			kind = "slack.users.ingest"
		case domain.JobScopeSlackFiles:
			kind = "slack.files.ingest"
		case domain.JobScopeSlackReactions:
			kind = "slack.reactions.ingest"
		default:
			return fmt.Errorf("unsupported slack scope: %s", scope)
		}

		payload, err := json.Marshal(types.SlackIngestionJobPayload{
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
			Queue:   "slack",
			Payload: payload,
		})
	}

	return s.repo.CreateJobs(ctx, req.TopicID, req.Name, jobs)
}

func (s *IngestionService) CreateDiscordJobs(ctx context.Context, req *types.CreateIngestionRequest, cfg *types.CreateDiscordJobsConfig) error {
	jobs := make([]types.Job, 0, len(req.Scope))

	for _, scope := range req.Scope {
		var kind string
		jobCfg := *cfg
		switch domain.JobScopeDiscord(scope) {
		case domain.JobScopeDiscordGuild:
			kind = "discord.guild.ingest"
		case domain.JobScopeDiscordChannels:
			kind = "discord.channels.ingest"
		case domain.JobScopeDiscordMessages:
			kind = "discord.messages.ingest"
		case domain.JobScopeDiscordThreads:
			kind = "discord.threads.ingest"
		case domain.JobScopeDiscordMembers:
			kind = "discord.members.ingest"
		case domain.JobScopeDiscordReactions:
			kind = "discord.reactions.ingest"
		default:
			return fmt.Errorf("unsupported discord scope: %s", scope)
		}

		payload, err := json.Marshal(types.DiscordIngestionJobPayload{
			TopicID:    req.TopicID,
			SourceType: req.SourceType,
			Name:       req.Name,
			SourceLink: req.SourceLink,
			Scope:      scope,
			Config:     jobCfg,
		})
		if err != nil {
			return err
		}

		jobs = append(jobs, types.Job{
			ID:      uuid.New(),
			Kind:    kind,
			Queue:   "discord",
			Payload: payload,
		})
	}

	return s.repo.CreateJobs(ctx, req.TopicID, req.Name, jobs)
}

func (s *IngestionService) CreateJiraJobs(ctx context.Context, req *types.CreateIngestionRequest, cfg *types.CreateJiraJobsConfig) error {
	jobs := make([]types.Job, 0, len(req.Scope))

	for _, scope := range req.Scope {
		var kind string
		switch domain.JobScopeJira(scope) {
		case domain.JobScopeJiraProjects:
			kind = "jira.projects.ingest"
		case domain.JobScopeJiraIssues:
			kind = "jira.issues.ingest"
		case domain.JobScopeJiraComments:
			kind = "jira.comments.ingest"
		case domain.JobScopeJiraIssueLinks:
			kind = "jira.issue_links.ingest"
		case domain.JobScopeJiraAttachments:
			kind = "jira.attachments.ingest"
		default:
			return fmt.Errorf("unsupported jira scope: %s", scope)
		}

		payload, err := json.Marshal(types.JiraIngestionJobPayload{
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
			Queue:   "jira",
			Payload: payload,
		})
	}

	return s.repo.CreateJobs(ctx, req.TopicID, req.Name, jobs)
}

func (s *IngestionService) CreateConfluenceJobs(ctx context.Context, req *types.CreateIngestionRequest, cfg *types.CreateConfluenceJobsConfig) error {
	jobs := make([]types.Job, 0, len(req.Scope))

	for _, scope := range req.Scope {
		var kind string
		switch domain.JobScopeConfluence(scope) {
		case domain.JobScopeConfluenceSpaces:
			kind = "confluence.spaces.ingest"
		case domain.JobScopeConfluencePages:
			kind = "confluence.pages.ingest"
		case domain.JobScopeConfluenceComments:
			kind = "confluence.comments.ingest"
		case domain.JobScopeConfluenceAttachments:
			kind = "confluence.attachments.ingest"
		default:
			return fmt.Errorf("unsupported confluence scope: %s", scope)
		}

		payload, err := json.Marshal(types.ConfluenceIngestionJobPayload{
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
			Queue:   "confluence",
			Payload: payload,
		})
	}

	return s.repo.CreateJobs(ctx, req.TopicID, req.Name, jobs)
}
