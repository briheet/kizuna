package repository

import (
	"context"

	"github.com/briheet/kizuna/backend/internal/types"
)

type IngestionRepository interface {
	CreateJobs(ctx context.Context, req *types.CreateJobsRequest) error
}
