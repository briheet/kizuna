package worker

import (
	"context"
	"time"

	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/providers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type WorkerCategory string

const (
	WorkerCategoryGithub     WorkerCategory = "github"
	WorkerCategoryDiscord    WorkerCategory = "discord"
	WorkerCategorySlack      WorkerCategory = "slack"
	WorkerCategoryConfluence WorkerCategory = "confluence"
	WorkerCategoryJira       WorkerCategory = "jira"
)

type WorkerBuilder func(
	dbClient *db.Client,
	logger *logger.Logger,
	client *providers.Client) Worker

var WorkerFuncs = map[WorkerCategory]WorkerBuilder{
	WorkerCategoryGithub:     NewGithubWorker,
	WorkerCategoryDiscord:    NewDiscordWorker,
	WorkerCategorySlack:      NewSlackWorker,
	WorkerCategoryConfluence: NewConfluenceWorker,
	WorkerCategoryJira:       NewJiraWorker,
}

type Worker interface {
	Name() string
	Start(ctx context.Context) error
}

// Every job will have a definition to hold which contains how will it executed.
// It will also contain its handler which will be will be built and stored.
// This handler will be involked when we want to run our jobs.
type JobWorker struct {
	// Worker id
	ID uuid.UUID

	// Name of that job.
	WorkerName string

	// Kind of job.
	Kind string

	// Queue. Which queue the job belongs to
	Queue string

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
	jobCtx, cancel := context.WithTimeout(ctx, j.Config.JobTimeout)
	defer cancel()

	if err := j.Handler.Handle(jobCtx, job); err != nil {
		_, updateErr := j.Client.Conn().Exec(ctx, `
			update jobs
			set
			state = case
			when attempt >= max_attempt then 'discarded'
			else 'available'
			end,
			scheduled_at = case
			when attempt >= max_attempt then scheduled_at
			else now() + interval '1 minute'
			end,
			updated_at = now()
			where id = $1;
		`, job.ID)
		if updateErr != nil {
			j.Logger.Error("failed to mark job failed", zap.String("job_id", job.ID.String()), zap.Error(updateErr))
		}

		j.Logger.Error("job failed", zap.String("job_id", job.ID.String()), zap.Error(err))
		return
	}

	if _, err := j.Client.Conn().Exec(ctx, `
		update jobs
		set
			state = 'completed',
			completed_at = now(),
			updated_at = now()
		where id = $1;
	`, job.ID); err != nil {
		j.Logger.Error("failed to mark job completed", zap.String("job_id", job.ID.String()), zap.Error(err))
	}

}

func (j *JobWorker) claimJobs(ctx context.Context) ([]Job, error) {

	var jobs []Job

	// Build and Query the db for that particular worker type jobs
	// Read query for jobs
	readQuery := `
		select id from jobs where
		queue = $1 and
		kind = $2 and
		state = 'available' and
		scheduled_at <= now()
		order by priority desc,
		scheduled_at asc,
		created_at asc
		limit $3
		for update skip locked;`

	updateQuery := `
		update jobs set
		state = 'running',
		worker_id = $1,
		attempt = attempt + 1,
		attempted_at = now(),
		updated_at = now()
		where id = ANY($2::uuid[]) returning
		id,
  		kind,
  		queue,
  		payload,
  		state,
  		priority,
  		attempt,
  		max_attempt,
  		worker_id,
  		scheduled_at,
  		attempted_at,
  		completed_at,
  		created_at,
  		updated_at;`

	executorFunc := func(tx pgx.Tx) error {
		pgRows, err := tx.Query(ctx, readQuery, j.Queue, j.Kind, j.Config.ClaimBatchSize)
		if err != nil {
			return err
		}
		defer pgRows.Close()

		var jobIDs []uuid.UUID
		for pgRows.Next() {
			var id uuid.UUID
			if err := pgRows.Scan(&id); err != nil {
				return err
			}

			jobIDs = append(jobIDs, id)
		}

		if err := pgRows.Err(); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, updateQuery, j.ID, jobIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var job Job

			if err := rows.Scan(
				&job.ID,
				&job.Kind,
				&job.Queue,
				&job.Payload,
				&job.State,
				&job.Priority,
				&job.Attempt,
				&job.MaxAttempts,
				&job.WorkerID,
				&job.ScheduledAt,
				&job.AttemptedAt,
				&job.CompletedAt,
				&job.CreatedAt,
				&job.UpdatedAt,
			); err != nil {
				return err
			}

			jobs = append(jobs, job)
		}

		return rows.Err()

	}

	if err := j.Client.ExecuteTx(ctx, executorFunc); err != nil {
		return nil, err
	}

	return jobs, nil
}
