package cmd

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/logger"
	"github.com/briheet/kizuna/workers/internal/worker"
	"github.com/spf13/cobra"
)

func WorkerCmd(ctx context.Context) *cobra.Command {
	var configPath string

	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "Worker entrypoint",
		RunE: func(cmd *cobra.Command, args []string) error {

			// Load config
			cfg, err := config.LoadConfig(ctx, configPath)
			if err != nil {
				return err
			}

			// Load logger. curr uses uber zap
			log, err := logger.NewLogger("api")
			if err != nil {
				return err
			}
			defer func() { _ = log.Sync() }()

			// Create a new orchestrator
			// This will manage all life cycle for workers
			// These workers include particular workers for github, slack, discord, etc
			orchestrator, err := worker.NewOrchestrator(ctx, cfg)
			if err != nil {
				return err
			}

			// Start the orchestrator
			return orchestrator.Start()
		},
	}

	return workerCmd
}
