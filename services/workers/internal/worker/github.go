package worker

import (
	"context"
	"time"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/providers"
	"github.com/briheet/kizuna/workers/internal/repository"
	"github.com/briheet/kizuna/workers/internal/repository/cockroachdb"
	"github.com/briheet/kizuna/workers/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Github worker
func NewGithubWorker(
	config *config.Config,
	dbClient *db.Client,
	logger *logger.Logger,
	client *providers.Client,
	embedderRepository repository.EmbedderRepository,
) Worker {
	githubRepo := cockroachdb.NewGraphRepository(dbClient)
	embedderService := services.NewEmbedderService(embedderRepository)
	githubService := services.NewGithubService(githubRepo, client.Github(), embedderService)

	return &JobWorker{
		ID:         uuid.New(),
		WorkerName: "github-ingestion-worker",
		Kind:       "",
		Queue:      string(WorkerCategoryGithub),
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
			return githubService.HandleJob(ctx, job.ID, job.Kind, job.Payload)
		}),
	}
}
