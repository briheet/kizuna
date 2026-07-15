package worker

import (
	"context"
	"time"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/providers"
	"github.com/briheet/kizuna/workers/internal/repository/cockroachdb"
	"github.com/briheet/kizuna/workers/internal/repository/embedder"
	"github.com/briheet/kizuna/workers/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Slack worker
func NewSlackWorker(
	config *config.Config,
	dbClient *db.Client,
	logger *logger.Logger,
	client *providers.Client,
) Worker {
	graphRepo := cockroachdb.NewGraphRepository(dbClient)
	embedderRepo := embedder.NewNomicRepository(config.Embedder.BaseURL)
	embedderService := services.NewEmbedderService(embedderRepo)
	slackService := services.NewSlackService(graphRepo, client.Slack(), embedderService)

	return &JobWorker{
		ID:         uuid.New(),
		WorkerName: "slack-ingestion-worker",
		Kind:       "",
		Queue:      string(WorkerCategorySlack),
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
			logger.Info("handling slack job", zap.String("job_id", job.ID.String()))
			return slackService.HandleJob(ctx, job.ID, job.Kind, job.Payload)
		}),
	}
}
