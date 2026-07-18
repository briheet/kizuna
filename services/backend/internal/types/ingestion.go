package types

import (
	"encoding/json"
)

// This is a generic pattern all requests will follow
// It will then be validated and switched accordingly
type CreateIngestionRequest struct {
	TopicID    string          `json:"topic_id,omitempty" validate:"omitempty,uuid4"`
	SourceType string          `json:"source_type" validate:"required,oneof=github slack discord jira confluence"`
	Name       string          `json:"name" validate:"required"`
	SourceLink string          `json:"source_link" validate:"omitempty,url"`
	Scope      []string        `json:"scope" validate:"required,min=1,dive,required"`
	Config     json.RawMessage `json:"config" validate:"required"`
}

type CreateIngestionResponse struct {
	TopicID     string `json:"topic_id"`
	SourceType  string `json:"source_type"`
	JobsCreated int    `json:"jobs_created"`
	State       string `json:"state"`
}

// Particular config of the repo
type CreateGithubJobsConfig struct {
	Owner    string `json:"owner" validate:"required"`
	Repo     string `json:"repo" validate:"required"`
	Since    string `json:"since" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Limit    int    `json:"limit" validate:"required,min=1,max=10000"`
	PageSize int    `json:"page_size" validate:"required,min=1,max=100"`
	Page     int    `json:"page" validate:"required,min=1"`
}

type GithubIngestionJobPayload struct {
	TopicID    string                 `json:"topic_id"`
	SourceType string                 `json:"source_type"`
	Name       string                 `json:"name"`
	SourceLink string                 `json:"source_link"`
	Scope      string                 `json:"scope"`
	Config     CreateGithubJobsConfig `json:"config"`
}

type CreateSlackJobsConfig struct {
	TeamID    string `json:"team_id" validate:"required"`
	ChannelID string `json:"channel_id" validate:"omitempty"`
	Since     string `json:"since" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Limit     int    `json:"limit" validate:"required,min=1,max=10000"`
	PageSize  int    `json:"page_size" validate:"required,min=1,max=100"`
	Cursor    string `json:"cursor" validate:"omitempty"`
	Remaining int    `json:"remaining" validate:"omitempty,min=0,max=10000"`
}

type SlackIngestionJobPayload struct {
	TopicID    string                `json:"topic_id"`
	SourceType string                `json:"source_type"`
	Name       string                `json:"name"`
	SourceLink string                `json:"source_link"`
	Scope      string                `json:"scope"`
	Config     CreateSlackJobsConfig `json:"config"`
}

type CreateDiscordJobsConfig struct {
	GuildID         string `json:"guild_id" validate:"required"`
	ChannelID       string `json:"channel_id" validate:"omitempty"`
	Since           string `json:"since" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Limit           int    `json:"limit" validate:"required,min=1,max=10000"`
	PageSize        int    `json:"page_size" validate:"required,min=1,max=100"`
	BeforeMessageID string `json:"before_message_id" validate:"omitempty"`
	Remaining       int    `json:"remaining" validate:"omitempty,min=0,max=10000"`
}

type DiscordIngestionJobPayload struct {
	TopicID    string                  `json:"topic_id"`
	SourceType string                  `json:"source_type"`
	Name       string                  `json:"name"`
	SourceLink string                  `json:"source_link"`
	Scope      string                  `json:"scope"`
	Config     CreateDiscordJobsConfig `json:"config"`
}

type CreateJiraJobsConfig struct {
	ProjectKey string `json:"project_key" validate:"required"`
	Since      string `json:"since" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type JiraIngestionJobPayload struct {
	TopicID    string               `json:"topic_id"`
	SourceType string               `json:"source_type"`
	Name       string               `json:"name"`
	SourceLink string               `json:"source_link"`
	Scope      string               `json:"scope"`
	Config     CreateJiraJobsConfig `json:"config"`
}

type CreateConfluenceJobsConfig struct {
	SpaceID string `json:"space_id" validate:"required"`
	Since   string `json:"since" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type ConfluenceIngestionJobPayload struct {
	TopicID    string                     `json:"topic_id"`
	SourceType string                     `json:"source_type"`
	Name       string                     `json:"name"`
	SourceLink string                     `json:"source_link"`
	Scope      string                     `json:"scope"`
	Config     CreateConfluenceJobsConfig `json:"config"`
}
