package types

import (
	"encoding/json"
	"time"
)

type JobsStatusRequest struct {
	TopicID    string `validate:"omitempty,uuid4"`
	SourceType string `validate:"omitempty,oneof=github slack discord jira confluence"`
	State      string `validate:"omitempty,oneof=available running completed discarded cancelled failed"`
	Limit      int    `validate:"omitempty,min=1,max=100"`
}

type JobsStatusResponse struct {
	Counts         []JobStateCount `json:"counts"`
	RecentFailures []FailedJob     `json:"recent_failures"`
}

type JobStateCount struct {
	State string `json:"state"`
	Count int    `json:"count"`
}

type FailedJob struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Queue      string          `json:"queue"`
	State      string          `json:"state"`
	Attempt    int             `json:"attempt"`
	MaxAttempt int             `json:"max_attempt"`
	LastError  string          `json:"last_error"`
	Payload    json.RawMessage `json:"payload"`
	RecentAt   time.Time       `json:"recent_at"`
}
