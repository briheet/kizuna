package integration

import (
	"fmt"
	"log"
	"testing"

	"github.com/briheet/kizuna/backend/internal/cmd"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

type BaseDB struct {
	Db *cockroachdb.CockroachDBContainer
}

func setupDB(t *testing.T) (*BaseDB, error) {
	t.Helper()

	cockroachdbContainer, err := cockroachdb.Run(
		t.Context(),
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
		return nil, err
	}

	return &BaseDB{
		Db: cockroachdbContainer,
	}, nil
}

func TestMigrationUp(t *testing.T) {

	ctx := t.Context()

	db, err := setupDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.Db)

	host, err := db.Db.Host(ctx)
	require.NoError(t, err)

	port, err := db.Db.MappedPort(ctx, "26257/tcp")
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

	client, err := pgxpool.New(ctx, pgxURL)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	var exists bool

	err = client.QueryRow(t.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = 'graph_nodes'
		)
	`).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists)
}
