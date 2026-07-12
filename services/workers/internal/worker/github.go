package worker

import (
	"context"
	"time"

	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/providers"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Github worker
func NewGithubWorker(
	ctx context.Context,
	dbClient *db.Client,
	logger *logger.Logger,
	client *providers.Client,
) Worker {
	return &JobWorker{
		ID:         uuid.New(),
		WorkerName: "github-ingestion-worker",
		Kind:       "github.ingest",
		Queue:      "github",
		Client:     dbClient,
		Logger:     logger,
		Config: JobConfig{
			MinimumPollInterval: 2 * time.Second,
			ClaimBatchSize:      10,
			MaxConcurrency:      10,
			JobTimeout:          2 * time.Minute,
			LeaseDuration:       5 * time.Minute,
		},
		Handler: HandlerFunc(func(ctx context.Context, job Job) error {
			logger.Info("handling github job", zap.String("job_id", job.ID.String()))
			return nil
		}),
	}
}
