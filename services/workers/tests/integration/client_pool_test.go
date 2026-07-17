package integration

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestClientSupportsConcurrentTransactions(t *testing.T) {
	db := setupDB(t)
	db.Restore(t)

	const transactionCount = 16
	errors := make(chan error, transactionCount)
	var workers sync.WaitGroup

	for index := range transactionCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			err := db.Client.ExecuteTx(t.Context(), func(tx pgx.Tx) error {
				_, err := tx.Exec(t.Context(), `
					insert into organisations (id, name, created_at)
					values ($1, $2, now());
				`, uuid.New(), fmt.Sprintf("organisation-%d", index))
				return err
			})
			errors <- err
		}()
	}

	workers.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}

	var count int
	err := db.Client.Conn().QueryRow(t.Context(), `select count(*) from organisations;`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, transactionCount, count)
}
