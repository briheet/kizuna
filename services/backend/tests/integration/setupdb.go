package integration

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/briheet/kizuna/backend/internal/cmd"
	"github.com/briheet/kizuna/backend/internal/config"
	"github.com/briheet/kizuna/backend/internal/db"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

type BaseDB struct {
	Db           *cockroachdb.CockroachDBContainer
	Client       *db.Client
	PGXURL       string
	MigrationURL string
}

func setupDB(t *testing.T) *BaseDB {
	t.Helper()
	ctx := t.Context()

	cockroachdbContainer, err := cockroachdb.Run(
		ctx,
		"cockroachdb/cockroach:latest-v23.1",
		cockroachdb.WithInsecure(),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(cockroachdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	})
	if err != nil {
		log.Printf("failed to start container: %s", err)
		require.NoError(t, err)
	}

	host, err := cockroachdbContainer.Host(ctx)
	require.NoError(t, err)

	port, err := cockroachdbContainer.MappedPort(ctx, "26257/tcp")
	require.NoError(t, err)

	pgxURL := fmt.Sprintf(
		"postgresql://root@%s:%s/defaultdb?sslmode=disable",
		host,
		port.Port(),
	)

	migrationURL := fmt.Sprintf(
		"cockroachdb://root@%s:%s/defaultdb?sslmode=disable",
		host,
		port.Port(),
	)

	command := cmd.MigrateCmd(ctx)
	command.SetArgs([]string{
		"--filepath",
		"file://../../migration",
		"--dburl",
		migrationURL,
	})

	err = command.Execute()
	require.NoError(t, err)

	client, err := db.NewClient(ctx, &config.Config{
		Db: config.DbConfig{
			DatabaseURL: pgxURL,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close(context.Background()))
	})

	return &BaseDB{
		Db:           cockroachdbContainer,
		Client:       client,
		PGXURL:       pgxURL,
		MigrationURL: migrationURL,
	}
}

func (db *BaseDB) Restore(t *testing.T) {
	t.Helper()

	_, err := db.Client.Conn().Exec(t.Context(), `
		TRUNCATE TABLE
			graph_edges,
			chunks,
			graph_nodes,
			jobs,
			data_sources,
			topics,
			teams,
			organisations
		CASCADE;
	`)
	require.NoError(t, err)
}
