package domain

type JobKind string

const (
	JobKindGithubRepository   JobKind = "github.repository.ingest"
	JobKindGithubIssues       JobKind = "github.issues.ingest"
	JobKindGithubPullRequests JobKind = "github.pull_requests.ingest"
	JobKindGithubCommits      JobKind = "github.commits.ingest"
	JobKindGithubReleases     JobKind = "github.releases.ingest"

	JobKindSlackWorkspace JobKind = "slack.workspace.ingest"
	JobKindSlackChannels  JobKind = "slack.channels.ingest"
	JobKindSlackMessages  JobKind = "slack.messages.ingest"
	JobKindSlackThreads   JobKind = "slack.threads.ingest"
	JobKindSlackUsers     JobKind = "slack.users.ingest"
	JobKindSlackFiles     JobKind = "slack.files.ingest"
	JobKindSlackReactions JobKind = "slack.reactions.ingest"

	JobKindDiscordGuild     JobKind = "discord.guild.ingest"
	JobKindDiscordChannels  JobKind = "discord.channels.ingest"
	JobKindDiscordMessages  JobKind = "discord.messages.ingest"
	JobKindDiscordThreads   JobKind = "discord.threads.ingest"
	JobKindDiscordMembers   JobKind = "discord.members.ingest"
	JobKindDiscordReactions JobKind = "discord.reactions.ingest"
)
