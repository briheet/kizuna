package repository

import "context"

type EmbedderRepository interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
}
