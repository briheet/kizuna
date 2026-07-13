package github

import (
	"time"

	githubsdk "github.com/google/go-github/v89/github"
)

type RepoRequest struct {
	Owner string
	Repo  string
}

type ListIssuesRequest struct {
	RepoRequest
	State     string
	Sort      string
	Direction string
	Since     time.Time
	Page      int
	PerPage   int
}

type ListPullRequestsRequest struct {
	RepoRequest
	State     string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

type IssueRequest struct {
	RepoRequest
	Number  int
	Page    int
	PerPage int
}

type PullRequestRequest struct {
	RepoRequest
	Number  int
	Page    int
	PerPage int
}

type ListCommitsRequest struct {
	RepoRequest
	SHA     string
	Since   time.Time
	Until   time.Time
	Page    int
	PerPage int
}

type ListMilestonesRequest struct {
	RepoRequest
	State     string
	Sort      string
	Direction string
	Page      int
	PerPage   int
}

type ListLabelsRequest struct {
	RepoRequest
	Page    int
	PerPage int
}

type ListReleasesRequest struct {
	RepoRequest
	Page    int
	PerPage int
}

type Repository = githubsdk.Repository
type Issue = githubsdk.Issue
type IssueComment = githubsdk.IssueComment
type PullRequest = githubsdk.PullRequest
type PullRequestComment = githubsdk.PullRequestComment
type PullRequestReview = githubsdk.PullRequestReview
type RepositoryCommit = githubsdk.RepositoryCommit
type Label = githubsdk.Label
type Milestone = githubsdk.Milestone
type RepositoryRelease = githubsdk.RepositoryRelease
type Response = githubsdk.Response
