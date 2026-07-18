package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

func MigrateCmd(ctx context.Context) *cobra.Command {

	var filepath string
	var dburl string
	migrateCmd := &cobra.Command{
		Use:   "migrate up",
		Short: "Used for applying migration",
		RunE: func(cmd *cobra.Command, args []string) error {

			m, err := migrate.New(
				filepath,
				dburl,
			)
			if err != nil {
				return err
			}

			migrationErr := m.Up()
			sourceErr, databaseErr := m.Close()
			if migrationErr != nil && !errors.Is(migrationErr, migrate.ErrNoChange) {
				return migrationErr
			}
			if sourceErr != nil {
				return fmt.Errorf("close migration source: %w", sourceErr)
			}
			if databaseErr != nil {
				return fmt.Errorf("close migration database: %w", databaseErr)
			}

			return nil
		},
	}

	migrateCmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", filepath, "Pass at start to give filepath of migration files.")
	migrateCmd.PersistentFlags().StringVarP(&dburl, "dburl", "u", dburl, "Pass at start to give dburl.")

	return migrateCmd
}
