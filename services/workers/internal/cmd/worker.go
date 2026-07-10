package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func WorkerCmd(ctx context.Context) *cobra.Command {
	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "Worker entrypoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return workerCmd
}
