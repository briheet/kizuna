package worker

import (
	"context"
	"sync"
	"time"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/providers"
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

// Inits a new worker
func NewOrchestrator(ctx context.Context, config *config.Config) (*Orchestrator, error) {
	ctx, cancel := context.WithCancel(ctx)

	// Get clients
	providerClients, err := providers.NewClientProvider(ctx, config)
	if err != nil {
		return nil, err
	}

	// Build workers
	var workers []Worker

	// Get active providers stated in ./internal/providers
	// After getting them, build and store
	for _, category := range providers.ActiveProviders {
		builderFunc := WorkerFuncs[WorkerCategory(category)]
		workers = append(workers, builderFunc(ctx, providerClients))
	}

	// Build empty queue for now
	return &Orchestrator{
		ctx:     ctx,
		cancel:  cancel,
		workers: workers,
		wg:      sync.WaitGroup{},
	}, nil
}

func (o *Orchestrator) Start() error {
	return nil
}
