package domain

type JobScopeGithub string
type JobScopeSlack string
type JobScopeDiscord string
type JobScopeJira string
type JobScopeConfluence string

const (
	SourceTypeGithub     string = "github"
	SourceTypeSlack      string = "slack"
	SourceTypeDiscord    string = "discord"
	SourceTypeJira       string = "jira"
	SourceTypeConfluence string = "confluence"

	JobScopeGithubRepository   JobScopeGithub = "repository"
	JobScopeGithubIssues       JobScopeGithub = "issues"
	JobScopeGithubPullRequests JobScopeGithub = "pull_requests"
	JobScopeGithubCommits      JobScopeGithub = "commits"
	JobScopeGithubReleases     JobScopeGithub = "releases"

	JobScopeSlackWorkspace JobScopeSlack = "workspace"
	JobScopeSlackChannels  JobScopeSlack = "channels"
	JobScopeSlackMessages  JobScopeSlack = "messages"
	JobScopeSlackThreads   JobScopeSlack = "threads"
	JobScopeSlackUsers     JobScopeSlack = "users"
	JobScopeSlackFiles     JobScopeSlack = "files"
	JobScopeSlackReactions JobScopeSlack = "reactions"

	JobScopeDiscordGuild     JobScopeDiscord = "guild"
	JobScopeDiscordChannels  JobScopeDiscord = "channels"
	JobScopeDiscordMessages  JobScopeDiscord = "messages"
	JobScopeDiscordThreads   JobScopeDiscord = "threads"
	JobScopeDiscordMembers   JobScopeDiscord = "members"
	JobScopeDiscordReactions JobScopeDiscord = "reactions"

	JobScopeJiraProjects    JobScopeJira = "projects"
	JobScopeJiraIssues      JobScopeJira = "issues"
	JobScopeJiraComments    JobScopeJira = "comments"
	JobScopeJiraIssueLinks  JobScopeJira = "issue_links"
	JobScopeJiraAttachments JobScopeJira = "attachments"

	JobScopeConfluenceSpaces      JobScopeConfluence = "spaces"
	JobScopeConfluencePages       JobScopeConfluence = "pages"
	JobScopeConfluenceComments    JobScopeConfluence = "comments"
	JobScopeConfluenceAttachments JobScopeConfluence = "attachments"
)
