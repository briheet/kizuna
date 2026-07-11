package worker

import (
	"context"
	"time"

	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/providers"
	"go.uber.org/zap"
)

type WorkerCategory string

const (
	WorkerCategoryGithub  WorkerCategory = "github"
	WorkerCategoryDiscord WorkerCategory = "discord"
	WorkerCategorySlack   WorkerCategory = "slack"
)

type WorkerBuilder func(ctx context.Context,
	dbClient *db.Client,
	logger *logger.Logger,
	client *providers.Client) Worker

var WorkerFuncs = map[WorkerCategory]WorkerBuilder{
	WorkerCategoryGithub: NewGithubWorker,
}

type Worker interface {
	Name() string
	Start(ctx context.Context) error
}

// Every job will have a definition to hold which contains how will it executed.
// It will also contain its handler which will be will be built and stored.
// This handler will be involked when we want to run our jobs.
type JobWorker struct {
	// Name of that job.
	WorkerName string

	// Kind of job.
	Kind string

	// Particular client for their execution.
	Client *db.Client

	// Config for that particular jobs.
	Config JobConfig

	// Jobs particular handler which will get executed.
	Handler Handler

	// Logger for logging lmao
	Logger *logger.Logger
}

// Interface for our handler function to implement.
type Handler interface {
	Handle(ctx context.Context, job Job) error
}

// HandlerFunc allows ordinary functions to be registered as handlers.
type HandlerFunc func(ctx context.Context, job Job) error

// HandlerFunc implementation
func (f HandlerFunc) Handle(ctx context.Context, job Job) error {
	return f(ctx, job)
}

// Github worker
func NewGithubWorker(
	ctx context.Context,
	dbClient *db.Client,
	logger *logger.Logger,
	client *providers.Client,
) Worker {
	return &JobWorker{
		WorkerName: "github-ingestion-worker",
		Kind:       "github.ingest",
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
			return nil
		}),
	}
}

func (j *JobWorker) Name() string {
	return j.WorkerName
}

func (j *JobWorker) Start(ctx context.Context) error {
	j.Logger.Info("worker started", zap.String("worker", j.Name()))

	// Create a new ticker for polling
	ticker := time.NewTicker(j.Config.MinimumPollInterval)
	defer ticker.Stop()

	for {
		select {
		// Context closes
		case <-ctx.Done():
			return ctx.Err()

		// Polling case
		case <-ticker.C:
			jobs, err := j.claimJobs(ctx)
			if err != nil {
				// Handle cases
				return nil
			}

			for _, job := range jobs {
				go j.runJob(ctx, job)
			}

		}

	}
}

func (j *JobWorker) runJob(ctx context.Context, job Job) {

}

func (j *JobWorker) claimJobs(ctx context.Context) ([]Job, error) {
	return []Job{}, nil
}
