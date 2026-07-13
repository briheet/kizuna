package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrationUp(t *testing.T) {

	db := setupDB(t)
	require.NotNil(t, db)
	require.NotNil(t, db.Db)
	require.NotNil(t, db.Client)

	var exists bool

	err := db.Client.Conn().QueryRow(t.Context(), `
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
