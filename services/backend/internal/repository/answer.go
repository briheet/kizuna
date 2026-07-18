package repository

import "context"

type AnswerSource struct {
	Reference  int    `json:"reference"`
	Title      string `json:"title"`
	SourceType string `json:"source_type"`
	NodeType   string `json:"node_type"`
	Content    string `json:"content"`
}

type AnswerRepository interface {
	Summarize(ctx context.Context, question string, sources []AnswerSource) (string, error)
}
