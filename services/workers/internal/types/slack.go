package types

type SlackJobPayload struct {
	TopicID    string         `json:"topic_id"`
	SourceType string         `json:"source_type"`
	Name       string         `json:"name"`
	SourceLink string         `json:"source_link"`
	Scope      string         `json:"scope"`
	Config     SlackJobConfig `json:"config"`
}

type SlackJobConfig struct {
	TeamID    string `json:"team_id"`
	ChannelID string `json:"channel_id"`
	Since     string `json:"since"`
	Limit     int    `json:"limit"`
	PageSize  int    `json:"page_size"`
	Cursor    string `json:"cursor"`
	Remaining int    `json:"remaining"`
}
