package domain

type JobKind string

const (
	JobKindGithubRepository   JobKind = "github.repository.ingest"
	JobKindGithubIssues       JobKind = "github.issues.ingest"
	JobKindGithubPullRequests JobKind = "github.pull_requests.ingest"
	JobKindGithubCommits      JobKind = "github.commits.ingest"
	JobKindGithubReleases     JobKind = "github.releases.ingest"
)
