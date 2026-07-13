package types

import (
	"encoding/json"
)

// This is a generic pattern all requests will follow
// It will then be validated and switched accordingly
type CreateIngestionRequest struct {
	TopicID    string          `json:"topic_id" validate:"required,uuid4"`
	SourceType string          `json:"source_type" validate:"required,oneof=github slack discord jira confluence"`
	Name       string          `json:"name" validate:"required"`
	SourceLink string          `json:"source_link" validate:"omitempty,url"`
	Scope      []string        `json:"scope" validate:"required,min=1,dive,required"`
	Config     json.RawMessage `json:"config" validate:"required"`
}

// Particular config of the repo
type CreateGithubJobsConfig struct {
	Owner string `json:"owner" validate:"required"`
	Repo  string `json:"repo" validate:"required"`
	Since string `json:"since" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type GithubIngestionJobPayload struct {
	TopicID    string                 `json:"topic_id"`
	SourceType string                 `json:"source_type"`
	Name       string                 `json:"name"`
	SourceLink string                 `json:"source_link"`
	Scope      string                 `json:"scope"`
	Config     CreateGithubJobsConfig `json:"config"`
}
