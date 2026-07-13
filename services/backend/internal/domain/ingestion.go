package domain

type JobScopeGithub string

const (
	SourceTypeGithub string = "github"

	JobScopeGithubRepository   JobScopeGithub = "repository"
	JobScopeGithubIssues       JobScopeGithub = "issues"
	JobScopeGithubPullRequests JobScopeGithub = "pull_requests"
	JobScopeGithubCommits      JobScopeGithub = "commits"
	JobScopeGithubReleases     JobScopeGithub = "releases"
)
