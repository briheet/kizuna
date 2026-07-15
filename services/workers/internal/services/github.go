package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/briheet/kizuna/workers/internal/domain"
	githubprovider "github.com/briheet/kizuna/workers/internal/providers/github"
	"github.com/briheet/kizuna/workers/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type GithubService struct {
	repo     repository.GraphRepository
	client   *githubprovider.Client
	embedder *EmbedderService
}

type GithubJobPayload struct {
	TopicID    string          `json:"topic_id"`
	SourceType string          `json:"source_type"`
	Name       string          `json:"name"`
	SourceLink string          `json:"source_link"`
	Scope      string          `json:"scope"`
	Config     GithubJobConfig `json:"config"`
}

type GithubJobConfig struct {
	Owner    string `json:"owner"`
	Repo     string `json:"repo"`
	Since    string `json:"since"`
	Limit    int    `json:"limit"`
	PageSize int    `json:"page_size"`
	Page     int    `json:"page"`
}

func NewGithubService(repo repository.GraphRepository, client *githubprovider.Client, embedder *EmbedderService) *GithubService {
	return &GithubService{
		repo:     repo,
		client:   client,
		embedder: embedder,
	}
}

func (s *GithubService) HandleJob(ctx context.Context, jobID uuid.UUID, kind string, payload json.RawMessage) error {
	var p GithubJobPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decode github job payload: %w", err)
	}

	topicID, err := uuid.Parse(p.TopicID)
	if err != nil {
		return fmt.Errorf("parse github topic id: %w", err)
	}

	sourceID, err := s.repo.UpsertDataSource(ctx, repository.DataSourceInput{
		TopicID:    topicID,
		SourceType: "github",
		Name:       p.Name,
		ExternalID: fmt.Sprintf("github:%s/%s", p.Config.Owner, p.Config.Repo),
		SourceLink: p.SourceLink,
		Config:     payload,
	})
	if err != nil {
		return err
	}

	switch domain.JobKind(kind) {
	case domain.JobKindGithubRepository:
		return s.handleRepository(ctx, sourceID, p)
	case domain.JobKindGithubIssues:
		return s.handleIssues(ctx, sourceID, p)
	case domain.JobKindGithubPullRequests:
		return s.handlePullRequests(ctx, sourceID, p)
	case domain.JobKindGithubCommits:
		return s.handleCommits(ctx, sourceID, p)
	case domain.JobKindGithubReleases:
		return s.handleReleases(ctx, sourceID, p)
	default:
		return fmt.Errorf("unsupported github job kind: %s", kind)
	}
}

func (s *GithubService) handleRepository(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload) error {
	repo, _, err := s.client.GetRepository(ctx, githubprovider.RepoRequest{Owner: p.Config.Owner, Repo: p.Config.Repo})
	if err != nil {
		return err
	}

	props, err := json.Marshal(repo)
	if err != nil {
		return err
	}
	return s.saveGraph(ctx, sourceID, repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{{
			Node: repository.GraphNodeInput{
				NodeType:   "github_repository",
				ExternalID: fmt.Sprintf("github:%s/%s", p.Config.Owner, p.Config.Repo),
				SourceLink: repo.GetHTMLURL(),
				Title:      repo.GetFullName(),
				Path:       repo.GetFullName(),
				Properties: props,
			},
			Chunks: []repository.ChunkInput{{Index: 0, Content: repo.GetDescription()}},
		}},
	})
}

func (s *GithubService) handleIssues(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload) error {
	if p.Config.PageSize <= 0 {
		return fmt.Errorf("github page_size is required")
	}
	if p.Config.Limit <= 0 {
		return fmt.Errorf("github limit is required")
	}

	var since time.Time
	if p.Config.Since != "" {
		parsed, err := time.Parse(time.RFC3339, p.Config.Since)
		if err != nil {
			return err
		}
		since = parsed
	}

	page := p.Config.Page
	if page == 0 {
		page = 1
	}
	remaining := p.Config.Limit

	for remaining > 0 {
		pageSize := p.Config.PageSize
		if remaining < pageSize {
			pageSize = remaining
		}

		issues, resp, err := s.client.ListIssues(ctx, githubprovider.ListIssuesRequest{
			RepoRequest: githubprovider.RepoRequest{Owner: p.Config.Owner, Repo: p.Config.Repo},
			State:       "all",
			Since:       since,
			Page:        page,
			PerPage:     pageSize,
		})
		if err != nil {
			return err
		}
		if len(issues) == 0 {
			return nil
		}

		processed := 0
		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(10)

		// Fetch and build each issue graph concurrently, but persist each issue as
		// its own transaction so the job can fail/retry without one huge DB write.
		for _, issue := range issues {
			if issue.IsPullRequest() {
				continue
			}
			processed++

			issue := issue
			g.Go(func() error {
				return s.handleIssue(ctx, sourceID, p, issue)
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}

		remaining -= processed
		if resp == nil || resp.NextPage == 0 || len(issues) < pageSize {
			return nil
		}
		page = resp.NextPage
	}

	return nil
}

func (s *GithubService) handleIssue(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload, issue *githubprovider.Issue) error {
	repoExternalID := fmt.Sprintf("github:%s/%s", p.Config.Owner, p.Config.Repo)
	issueExternalID := fmt.Sprintf("%s/issues/%d", repoExternalID, issue.GetNumber())
	issueProps, err := json.Marshal(issue)
	if err != nil {
		return err
	}

	// Build the deterministic subgraph for one issue: repo -> issue -> comments.
	// The repo node is included so edge endpoints are resolved from this graph input.
	graph := repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_repository",
					ExternalID: repoExternalID,
					SourceLink: p.SourceLink,
					Title:      p.Name,
					Path:       fmt.Sprintf("%s/%s", p.Config.Owner, p.Config.Repo),
				},
			},
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_issue",
					ExternalID: issueExternalID,
					SourceLink: issue.GetHTMLURL(),
					Title:      issue.GetTitle(),
					Path:       fmt.Sprintf("issues/%d", issue.GetNumber()),
					Properties: issueProps,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: issue.GetTitle() + "\n\n" + issue.GetBody()}},
			},
		},
		Edges: []repository.GraphEdgeInput{{
			FromExternalID: repoExternalID,
			ToExternalID:   issueExternalID,
			EdgeType:       "has_issue",
			EdgeScope:      "github",
			Confidence:     1,
		}},
	}

	page := 1
	for {
		comments, resp, err := s.client.ListIssueComments(ctx, githubprovider.IssueRequest{
			RepoRequest: githubprovider.RepoRequest{Owner: p.Config.Owner, Repo: p.Config.Repo},
			Number:      issue.GetNumber(),
			Page:        page,
			PerPage:     p.Config.PageSize,
		})
		if err != nil {
			return err
		}
		for _, comment := range comments {
			commentExternalID := fmt.Sprintf("%s/comments/%d", issueExternalID, comment.GetID())
			props, err := json.Marshal(comment)
			if err != nil {
				return err
			}
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
				Node: repository.GraphNodeInput{
					NodeType:   "github_issue_comment",
					ExternalID: commentExternalID,
					SourceLink: comment.GetHTMLURL(),
					Title:      fmt.Sprintf("Issue #%d comment", issue.GetNumber()),
					Path:       fmt.Sprintf("issues/%d/comments/%d", issue.GetNumber(), comment.GetID()),
					Properties: props,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: comment.GetBody()}},
			})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{
				FromExternalID: issueExternalID,
				ToExternalID:   commentExternalID,
				EdgeType:       "has_comment",
				EdgeScope:      "github",
				Confidence:     1,
			})
		}
		if resp == nil || resp.NextPage == 0 || len(comments) < p.Config.PageSize {
			break
		}
		page = resp.NextPage
	}

	return s.saveGraph(ctx, sourceID, graph)
}

func (s *GithubService) handlePullRequests(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload) error {
	if p.Config.PageSize <= 0 {
		return fmt.Errorf("github page_size is required")
	}
	if p.Config.Limit <= 0 {
		return fmt.Errorf("github limit is required")
	}

	page := p.Config.Page
	if page == 0 {
		page = 1
	}
	remaining := p.Config.Limit

	for remaining > 0 {
		pageSize := p.Config.PageSize
		if remaining < pageSize {
			pageSize = remaining
		}

		pulls, resp, err := s.client.ListPullRequests(ctx, githubprovider.ListPullRequestsRequest{
			RepoRequest: githubprovider.RepoRequest{Owner: p.Config.Owner, Repo: p.Config.Repo},
			State:       "all",
			Page:        page,
			PerPage:     pageSize,
		})
		if err != nil {
			return err
		}
		if len(pulls) == 0 {
			return nil
		}

		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(10)

		for _, pr := range pulls {
			pr := pr
			g.Go(func() error {
				return s.handlePullRequest(ctx, sourceID, p, pr)
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}

		remaining -= len(pulls)
		if resp == nil || resp.NextPage == 0 || len(pulls) < pageSize {
			return nil
		}
		page = resp.NextPage
	}

	return nil
}

func (s *GithubService) handlePullRequest(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload, pr *githubprovider.PullRequest) error {
	number := pr.GetNumber()
	req := githubprovider.PullRequestRequest{
		RepoRequest: githubprovider.RepoRequest{Owner: p.Config.Owner, Repo: p.Config.Repo},
		Number:      number,
		Page:        1,
		PerPage:     p.Config.PageSize,
	}

	repoExternalID := fmt.Sprintf("github:%s/%s", p.Config.Owner, p.Config.Repo)
	prExternalID := fmt.Sprintf("%s/pulls/%d", repoExternalID, number)
	props, err := json.Marshal(pr)
	if err != nil {
		return err
	}
	graph := repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_repository",
					ExternalID: repoExternalID,
					SourceLink: p.SourceLink,
					Title:      p.Name,
					Path:       fmt.Sprintf("%s/%s", p.Config.Owner, p.Config.Repo),
				},
			},
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_pull_request",
					ExternalID: prExternalID,
					SourceLink: pr.GetHTMLURL(),
					Title:      pr.GetTitle(),
					Path:       fmt.Sprintf("pulls/%d", number),
					Properties: props,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: pr.GetTitle() + "\n\n" + pr.GetBody()}},
			},
		},
		Edges: []repository.GraphEdgeInput{{
			FromExternalID: repoExternalID,
			ToExternalID:   prExternalID,
			EdgeType:       "has_pull_request",
			EdgeScope:      "github",
			Confidence:     1,
		}},
	}

	for {
		comments, resp, err := s.client.ListPullRequestComments(ctx, req)
		if err != nil {
			return err
		}
		for _, comment := range comments {
			props, err := json.Marshal(comment)
			if err != nil {
				return err
			}
			commentExternalID := fmt.Sprintf("%s/comments/%d", prExternalID, comment.GetID())
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
				Node: repository.GraphNodeInput{
					NodeType:   "github_pull_request_comment",
					ExternalID: commentExternalID,
					SourceLink: comment.GetHTMLURL(),
					Title:      fmt.Sprintf("PR #%d comment", number),
					Path:       fmt.Sprintf("pulls/%d/comments/%d", number, comment.GetID()),
					Properties: props,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: comment.GetBody()}},
			})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{
				FromExternalID: prExternalID,
				ToExternalID:   commentExternalID,
				EdgeType:       "has_comment",
				EdgeScope:      "github",
				Confidence:     1,
			})
		}
		if resp == nil || resp.NextPage == 0 || len(comments) < p.Config.PageSize {
			break
		}
		req.Page = resp.NextPage
	}

	req.Page = 1
	for {
		reviews, resp, err := s.client.ListPullRequestReviews(ctx, req)
		if err != nil {
			return err
		}
		for _, review := range reviews {
			props, err := json.Marshal(review)
			if err != nil {
				return err
			}
			reviewExternalID := fmt.Sprintf("%s/reviews/%d", prExternalID, review.GetID())
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
				Node: repository.GraphNodeInput{
					NodeType:   "github_pull_request_review",
					ExternalID: reviewExternalID,
					SourceLink: review.GetHTMLURL(),
					Title:      fmt.Sprintf("PR #%d review", number),
					Path:       fmt.Sprintf("pulls/%d/reviews/%d", number, review.GetID()),
					Properties: props,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: review.GetBody()}},
			})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{
				FromExternalID: prExternalID,
				ToExternalID:   reviewExternalID,
				EdgeType:       "has_review",
				EdgeScope:      "github",
				Confidence:     1,
			})
		}
		if resp == nil || resp.NextPage == 0 || len(reviews) < p.Config.PageSize {
			break
		}
		req.Page = resp.NextPage
	}

	req.Page = 1
	for {
		commits, resp, err := s.client.ListPullRequestCommits(ctx, req)
		if err != nil {
			return err
		}
		for _, commit := range commits {
			message := ""
			if commit.Commit != nil {
				message = commit.Commit.GetMessage()
			}
			props, err := json.Marshal(commit)
			if err != nil {
				return err
			}
			commitExternalID := fmt.Sprintf("%s/commits/%s", prExternalID, commit.GetSHA())
			graph.Nodes = append(graph.Nodes, repository.GraphNodeWithChunks{
				Node: repository.GraphNodeInput{
					NodeType:   "github_pull_request_commit",
					ExternalID: commitExternalID,
					SourceLink: commit.GetHTMLURL(),
					Title:      commit.GetSHA(),
					Path:       fmt.Sprintf("pulls/%d/commits/%s", number, commit.GetSHA()),
					Properties: props,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: message}},
			})
			graph.Edges = append(graph.Edges, repository.GraphEdgeInput{
				FromExternalID: prExternalID,
				ToExternalID:   commitExternalID,
				EdgeType:       "has_commit",
				EdgeScope:      "github",
				Confidence:     1,
			})
		}
		if resp == nil || resp.NextPage == 0 || len(commits) < p.Config.PageSize {
			break
		}
		req.Page = resp.NextPage
	}

	return s.saveGraph(ctx, sourceID, graph)
}

func (s *GithubService) handleCommits(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload) error {
	if p.Config.PageSize <= 0 {
		return fmt.Errorf("github page_size is required")
	}
	if p.Config.Limit <= 0 {
		return fmt.Errorf("github limit is required")
	}

	var since time.Time
	if p.Config.Since != "" {
		parsed, err := time.Parse(time.RFC3339, p.Config.Since)
		if err != nil {
			return err
		}
		since = parsed
	}

	page := p.Config.Page
	if page == 0 {
		page = 1
	}
	remaining := p.Config.Limit

	for remaining > 0 {
		pageSize := p.Config.PageSize
		if remaining < pageSize {
			pageSize = remaining
		}

		commits, resp, err := s.client.ListCommits(ctx, githubprovider.ListCommitsRequest{
			RepoRequest: githubprovider.RepoRequest{Owner: p.Config.Owner, Repo: p.Config.Repo},
			Since:       since,
			Page:        page,
			PerPage:     pageSize,
		})
		if err != nil {
			return err
		}
		if len(commits) == 0 {
			return nil
		}

		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(10)

		for _, commit := range commits {
			commit := commit
			g.Go(func() error {
				return s.handleCommit(ctx, sourceID, p, commit)
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}

		remaining -= len(commits)
		if resp == nil || resp.NextPage == 0 || len(commits) < pageSize {
			return nil
		}
		page = resp.NextPage
	}

	return nil
}

func (s *GithubService) handleCommit(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload, commit *githubprovider.RepositoryCommit) error {
	message := ""
	if commit.Commit != nil {
		message = commit.Commit.GetMessage()
	}

	repoExternalID := fmt.Sprintf("github:%s/%s", p.Config.Owner, p.Config.Repo)
	commitExternalID := fmt.Sprintf("%s/commits/%s", repoExternalID, commit.GetSHA())
	props, err := json.Marshal(commit)
	if err != nil {
		return err
	}

	return s.saveGraph(ctx, sourceID, repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_repository",
					ExternalID: repoExternalID,
					SourceLink: p.SourceLink,
					Title:      p.Name,
					Path:       fmt.Sprintf("%s/%s", p.Config.Owner, p.Config.Repo),
				},
			},
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_commit",
					ExternalID: commitExternalID,
					SourceLink: commit.GetHTMLURL(),
					Title:      commit.GetSHA(),
					Path:       fmt.Sprintf("commits/%s", commit.GetSHA()),
					Properties: props,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: message}},
			},
		},
		Edges: []repository.GraphEdgeInput{{
			FromExternalID: repoExternalID,
			ToExternalID:   commitExternalID,
			EdgeType:       "has_commit",
			EdgeScope:      "github",
			Confidence:     1,
		}},
	})
}

func (s *GithubService) handleReleases(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload) error {
	if p.Config.PageSize <= 0 {
		return fmt.Errorf("github page_size is required")
	}
	if p.Config.Limit <= 0 {
		return fmt.Errorf("github limit is required")
	}

	page := p.Config.Page
	if page == 0 {
		page = 1
	}
	remaining := p.Config.Limit

	for remaining > 0 {
		pageSize := p.Config.PageSize
		if remaining < pageSize {
			pageSize = remaining
		}

		releases, resp, err := s.client.ListReleases(ctx, githubprovider.ListReleasesRequest{
			RepoRequest: githubprovider.RepoRequest{Owner: p.Config.Owner, Repo: p.Config.Repo},
			Page:        page,
			PerPage:     pageSize,
		})
		if err != nil {
			return err
		}
		if len(releases) == 0 {
			return nil
		}

		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(10)

		for _, release := range releases {
			release := release
			g.Go(func() error {
				return s.handleRelease(ctx, sourceID, p, release)
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}

		remaining -= len(releases)
		if resp == nil || resp.NextPage == 0 || len(releases) < pageSize {
			return nil
		}
		page = resp.NextPage
	}

	return nil
}

func (s *GithubService) handleRelease(ctx context.Context, sourceID uuid.UUID, p GithubJobPayload, release *githubprovider.RepositoryRelease) error {
	repoExternalID := fmt.Sprintf("github:%s/%s", p.Config.Owner, p.Config.Repo)
	releaseExternalID := fmt.Sprintf("%s/releases/%d", repoExternalID, release.GetID())
	props, err := json.Marshal(release)
	if err != nil {
		return err
	}

	return s.saveGraph(ctx, sourceID, repository.GraphInput{
		Nodes: []repository.GraphNodeWithChunks{
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_repository",
					ExternalID: repoExternalID,
					SourceLink: p.SourceLink,
					Title:      p.Name,
					Path:       fmt.Sprintf("%s/%s", p.Config.Owner, p.Config.Repo),
				},
			},
			{
				Node: repository.GraphNodeInput{
					NodeType:   "github_release",
					ExternalID: releaseExternalID,
					SourceLink: release.GetHTMLURL(),
					Title:      release.GetName(),
					Path:       fmt.Sprintf("releases/%d", release.GetID()),
					Properties: props,
				},
				Chunks: []repository.ChunkInput{{Index: 0, Content: release.GetName() + "\n\n" + release.GetBody()}},
			},
		},
		Edges: []repository.GraphEdgeInput{{
			FromExternalID: repoExternalID,
			ToExternalID:   releaseExternalID,
			EdgeType:       "has_release",
			EdgeScope:      "github",
			Confidence:     1,
		}},
	})
}

func (s *GithubService) saveGraph(ctx context.Context, sourceID uuid.UUID, graph repository.GraphInput) error {
	// Embed all chunk text after the graph is built so chunk indexes and
	// embeddings stay aligned before the single repo transaction.
	texts := make([]string, 0)
	for _, node := range graph.Nodes {
		for _, chunk := range node.Chunks {
			if chunk.Content != "" {
				texts = append(texts, chunk.Content)
			}
		}
	}
	if len(texts) == 0 {
		return s.repo.SaveGraph(ctx, sourceID, graph)
	}

	embeddings, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	embeddingIndex := 0
	for nodeIndex := range graph.Nodes {
		for chunkIndex := range graph.Nodes[nodeIndex].Chunks {
			if graph.Nodes[nodeIndex].Chunks[chunkIndex].Content == "" {
				continue
			}
			graph.Nodes[nodeIndex].Chunks[chunkIndex].Embedding = embeddings[embeddingIndex]
			embeddingIndex++
		}
	}

	return s.repo.SaveGraph(ctx, sourceID, graph)
}
