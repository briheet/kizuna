package types

type DiscordJobPayload struct {
	TopicID    string           `json:"topic_id"`
	SourceType string           `json:"source_type"`
	Name       string           `json:"name"`
	SourceLink string           `json:"source_link"`
	Scope      string           `json:"scope"`
	Config     DiscordJobConfig `json:"config"`
}

type DiscordJobConfig struct {
	GuildID         string `json:"guild_id"`
	ChannelID       string `json:"channel_id"`
	Since           string `json:"since"`
	Limit           int    `json:"limit"`
	PageSize        int    `json:"page_size"`
	BeforeMessageID string `json:"before_message_id"`
	Remaining       int    `json:"remaining"`
}
