package repository

import (
	"context"

	"github.com/briheet/kizuna/backend/internal/types"
)

type IngestionRepository interface {
	CreateJobs(ctx context.Context, topicID string, topicName string, jobs []types.Job) error
	JobsStatus(ctx context.Context, req types.JobsStatusRequest) (*types.JobsStatusResponse, error)
}
