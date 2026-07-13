package repository

import (
	"context"

	"github.com/briheet/kizuna/backend/internal/types"
)

type IngestionRepository interface {
	CreateJobs(ctx context.Context, jobs []types.Job) error
}
