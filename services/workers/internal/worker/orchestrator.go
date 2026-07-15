package worker

import (
	"context"
	"sync"
	"time"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/providers"
	"go.uber.org/zap"
)

// Main shell. Holds state to all workers.
type Orchestrator struct {
	// Main workers: github, slack, discord, etc
	workers []Worker

	// Workers needs to be synced
	wg sync.WaitGroup

	// Config for checking in on workers
	config OrchestratorConfig

	// Base logger
	logger *logger.Logger
}

type OrchestratorConfig struct {
	// Time to check in on workers.
	rescueInterval time.Duration

	// If not responding then.
	shutdownTimeout time.Duration
}

// Inits a new worker
func NewOrchestrator(ctx context.Context, config *config.Config, logger *logger.Logger) (*Orchestrator, error) {
	// Get clients
	providerClients, err := providers.NewClientProvider(ctx, config)
	if err != nil {
		return nil, err
	}

	dbClient, err := db.NewClient(ctx, config)
	if err != nil {
		return nil, err
	}

	// Build workers
	var workers []Worker

	// Get active providers stated in ./internal/providers
	// After getting them, build and store
	for _, category := range providers.ActiveProviders {
		builderFunc := WorkerFuncs[WorkerCategory(category)]
		workers = append(workers, builderFunc(config, dbClient, logger, providerClients))
	}

	// Build empty queue for now
	return &Orchestrator{
		workers: workers,
		logger:  logger,
		wg:      sync.WaitGroup{},
	}, nil
}

func (o *Orchestrator) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Make a errChan and start all workers
	errChan := make(chan error, len(o.workers))

	for _, worker := range o.workers {
		go func(worker Worker) {
			errChan <- worker.Start(ctx)
		}(worker)
	}

	select {
	// Any Particular job fails, Rn just log and exit
	// This should be multiple cases and also depend upon custom errors returned
	case err := <-errChan:
		cancel()
		o.logger.Info("Orchestrator error", zap.String("Err:", err.Error()))
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
