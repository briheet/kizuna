package jira

import models "github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"

type SearchProjectsRequest struct {
	Options    *models.ProjectSearchOptionsScheme
	StartAt    int
	MaxResults int
}

type ProjectRequest struct {
	ProjectKeyOrID string
	Expand         []string
}

type SearchIssuesRequest struct {
	JQL        string
	Fields     []string
	Expand     []string
	StartAt    int
	MaxResults int
	Validate   string
}

type SearchIssuesJQLRequest struct {
	JQL           string
	Fields        []string
	Expand        []string
	MaxResults    int
	NextPageToken string
}

type IssueRequest struct {
	IssueKeyOrID string
	Fields       []string
	Expand       []string
}

type ListIssueCommentsRequest struct {
	IssueKeyOrID string
	OrderBy      string
	Expand       []string
	StartAt      int
	MaxResults   int
}

type IssueCommentRequest struct {
	IssueKeyOrID string
	CommentID    string
}

type ListLabelsRequest struct {
	StartAt    int
	MaxResults int
}

type SearchProjectVersionsRequest struct {
	ProjectKeyOrID string
	Options        *models.VersionGetsOptions
	StartAt        int
	MaxResults     int
}

type UserRequest struct {
	AccountID string
	Expand    []string
}

type ListUsersRequest struct {
	StartAt    int
	MaxResults int
}

type Response = models.ResponseScheme
type Project = models.ProjectScheme
type ProjectSearch = models.ProjectSearchScheme
type Issue = models.IssueScheme
type IssueSearch = models.IssueSearchScheme
type IssueSearchJQL = models.IssueSearchJQLScheme
type IssueComment = models.IssueCommentScheme
type IssueCommentPage = models.IssueCommentPageScheme
type IssueLink = models.IssueLinkScheme
type IssueLinkPage = models.IssueLinkPageScheme
type IssueLabels = models.IssueLabelsScheme
type Component = models.ComponentScheme
type Version = models.VersionScheme
type VersionPage = models.VersionPageScheme
type IssueAttachmentMetadata = models.IssueAttachmentMetadataScheme
type User = models.UserScheme
