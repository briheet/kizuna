package cmd

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/briheet/kizuna/backend/internal/api"
	"github.com/briheet/kizuna/backend/internal/config"
	"github.com/briheet/kizuna/backend/internal/db"
	"github.com/briheet/kizuna/backend/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func ApiCmd(ctx context.Context) *cobra.Command {

	var configPath string

	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Base api for Kizuna",
		RunE: func(cmd *cobra.Command, args []string) error {

			// Load config
			cfg, err := config.LoadConfig(ctx, configPath)
			if err != nil {
				return err
			}

			log, err := logger.NewLogger("api")
			if err != nil {
				return err
			}
			defer func() { _ = log.Sync() }()

			dbClient, err := db.NewClient(ctx, cfg)
			if err != nil {
				return err
			}
			defer dbClient.Close(ctx)

			api := api.NewApi(ctx, cfg, log, dbClient)
			srv := api.Server(cfg.Api.Port)

			apiErr := make(chan error, 1)
			go func() {
				apiErr <- srv.ListenAndServe()
			}()

			log.Info("started api", zap.Int("port", cfg.Api.Port))

			select {
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}
				return nil
			case err := <-apiErr:
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			}
		},
	}

	apiCmd.PersistentFlags().StringVarP(&configPath, "configPath", "c", configPath, "Pass at start to give config path to the application")

	return apiCmd
}
