package worker

import (
	"context"
	"time"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/providers"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Github worker
func NewJiraWorker(
	config *config.Config,
	dbClient *db.Client,
	logger *logger.Logger,
	client *providers.Client,
) Worker {
	return &JobWorker{
		ID:         uuid.New(),
		WorkerName: "jira-ingestion-worker",
		Kind:       "jira.ingest",
		Queue:      string(WorkerCategoryJira),
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
			logger.Info("handling jira job", zap.String("job_id", job.ID.String()))
			return nil
		}),
	}
}
