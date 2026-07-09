package cmd

import (
	"context"

	"github.com/briheet/kizuna/internal/config"
	"github.com/briheet/kizuna/internal/logger"
	"github.com/spf13/cobra"
)

func ApiCmd(ctx context.Context) *cobra.Command {

	var configPath string

	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Base api for Kizuna",
		RunE: func(cmd *cobra.Command, args []string) error {

			// Load config
			_, err := config.LoadConfig(ctx, configPath)
			if err != nil {
				return err
			}

			log, err := logger.NewLogger("api")
			if err != nil {
				return err
			}
			defer func() { _ = log.Sync() }()

			log.Info("Hi")

			return nil
		},
	}

	apiCmd.PersistentFlags().StringVarP(&configPath, "configPath", "c", configPath, "Pass at start to give config path to the application")

	return apiCmd
}
