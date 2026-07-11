package worker

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/providers"
)

type WorkerCategory string

const (
	WorkerCategoryGithub  WorkerCategory = "github"
	WorkerCategoryDiscord WorkerCategory = "discord"
	WorkerCategorySlack   WorkerCategory = "slack"
)

type WorkerBuilder func(ctx context.Context, client *providers.Client) Worker

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

// Github worker
func NewGithubWorker(ctx context.Context, client *providers.Client) Worker {
	return &JobWorker{}
}

func (j *JobWorker) Name() string {
	return j.WorkerName
}

func (j *JobWorker) Start(ctx context.Context) error {
	return nil
}
