package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

type BaseDB struct {
	Db     *cockroachdb.CockroachDBContainer
	Client *db.Client
	PGXURL string
}

func setupDB(t *testing.T) *BaseDB {
	t.Helper()
	ctx := t.Context()

	container, err := cockroachdb.Run(
		ctx,
		"cockroachdb/cockroach:latest-v26.2",
		cockroachdb.WithInsecure(),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "26257/tcp")
	require.NoError(t, err)

	pgxURL := fmt.Sprintf(
		"postgresql://root@%s:%s/defaultdb?sslmode=disable",
		host,
		port.Port(),
	)

	client, err := db.NewClient(ctx, &config.Config{
		Db: config.DbConfig{
			DatabaseURL: pgxURL,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close(context.Background()))
	})

	migration, err := os.ReadFile("../../../backend/migration/002_combined.up.sql")
	require.NoError(t, err)

	err = client.ExecuteTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, string(migration))
		return err
	})
	require.NoError(t, err)

	return &BaseDB{
		Db:     container,
		Client: client,
		PGXURL: pgxURL,
	}
}

func (db *BaseDB) Restore(t *testing.T) {
	t.Helper()

	err := db.Client.ExecuteTx(t.Context(), func(tx pgx.Tx) error {
		_, err := tx.Exec(t.Context(), `
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
		return err
	})
	require.NoError(t, err)
}
