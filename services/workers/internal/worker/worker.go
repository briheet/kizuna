package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/google/uuid"
)

// Main shell. Holds state to all workers.
type Orchestrator struct {
	// Main process context
	ctx context.Context

	// Controlling workers
	cancel context.CancelFunc

	// Main workers: github, slack, discord, etc
	workers []Worker

	// Workers needs to be synced
	wg sync.WaitGroup

	// Config for checking in on workers
	config OrchestratorConfig
}

type OrchestratorConfig struct {
	// Time to check in on workers.
	RescueInterval time.Duration

	// If not responding then.
	ShutdownTimeout time.Duration
}

type Worker interface {
	Name() string
	Start(ctx context.Context) error
}

// type for job state
type JobState string

// Enum on jobs
const (
	JobStateScheduled JobState = "scheduled"
	JobStateAvailable JobState = "available"
	JobStateRunning   JobState = "running"
	JobStateRetryable JobState = "retryable"
	JobStateCompleted JobState = "completed"
	JobStateDiscarded JobState = "discarded"
	JobStateCancelled JobState = "cancelled"
)

// This is response of table and particular jobs info.
type Job struct {
	// Id for that particular job.
	ID uuid.UUID

	// Kind of job like github, slack, discord, etc.
	Kind string

	// Payload for that particular job.
	Payload json.RawMessage

	// Current state of the job.
	State JobState

	// Priority Of execution. Not sure to keep this for now.
	Priority int

	// Number of Attempts for that job.
	Attempt int

	// Max Attempts for that particular job.
	MaxAttempts int

	// Worker ID for that job.
	WorkerID *uuid.UUID

	// If worker goes down, his jobs should be able to be reclaimed by orther workers
	LeaseExpiresAt *time.Time

	// When was that job scheduled.
	ScheduledAt time.Time

	// When was the last time the job was attempted. Nullable at start.
	AttemptedAt *time.Time

	// Was the job completed ? If so then when. Nullable at start.
	CompletedAt *time.Time

	// When was the job first time created at.
	CreatedAt time.Time

	// Was the job updated of its status ? If so then when.
	UpdatedAt time.Time
}

// Jobs config on how they will behave.
type JobConfig struct {
	// Minimum Polling for particular jobs.
	MinimumPollInterval time.Duration

	// Maximum polling for particular jobs.
	MaximumPollInterval time.Duration

	// pooling time + backoff for ease and removing thundering.
	PollBackoff float64

	// Max number of jobs run for a particular work at a time.
	MaxConcurrency int

	// Single Db query for that particular job.
	ClaimBatchSize int

	// Time limit for that particular job.
	JobTimeout time.Duration

	// Time for which that particular job can be leased out.
	LeaseDuration time.Duration

	// checks for that particular jobs every particular time interval.
	HeartBeatInterval time.Duration

	// If a job fails, worker polls and check and requeues.
	RescueInterval time.Duration

	// Time limit for the job shutdown if exceeds timelimit.
	ShutdownTimeout time.Duration
}

// Every job will have a definition to hold which contains how will it executed.
// It will also contain its handler which will be will be built and stored.
// This handler will be involked when we want to run our jobs.
type JobWorker struct {
	// Name of that job.
	Name string

	// Kind of job.
	Kind string

	// Particular client for their execution.
	client *db.Client

	// Config for that particular jobs.
	Config JobConfig

	// Jobs particular handler which will get executed.
	Handler Handler
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

// Inits a new worker
func NewOrchestrator(ctx context.Context, config *config.Config) *Orchestrator {
	ctx, cancel := context.WithCancel(ctx)

	// Build empty queue for now
	return &Orchestrator{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (o *Orchestrator) Start() error {
	return nil
}
