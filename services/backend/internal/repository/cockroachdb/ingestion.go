package cockroachdb

import (
	"context"
	"fmt"

	"github.com/briheet/kizuna/backend/internal/db"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/jackc/pgx/v5"
)

type CockroachDbIngestionRepository struct {
	db *db.Client
}

func NewCockroachDbIngestionRepository(db *db.Client) *CockroachDbIngestionRepository {
	return &CockroachDbIngestionRepository{db: db}
}

func (r *CockroachDbIngestionRepository) CreateJobs(ctx context.Context, jobs []types.Job) error {
	query := `
  			insert into jobs (
  				id,
  				kind,
  				queue,
  				payload,
  				state,
  				priority,
  				attempt,
  				max_attempt,
  				scheduled_at,
  				created_at,
  				updated_at
  			) values (
  				$1,
  				$2,
  				$3,
  				$4,
  				'available',
  				0,
  				0,
  				10,
  				now(),
  				now(),
  				now()
  			);
  		`

	return r.db.ExecuteTx(ctx, func(tx pgx.Tx) error {
		for _, job := range jobs {
			if _, err := tx.Exec(
				ctx,
				query,
				job.ID,
				job.Kind,
				job.Queue,
				job.Payload,
			); err != nil {
				return fmt.Errorf("insert github ingestion job %s: %w", job.Kind, err)
			}
		}
		return nil
	})
}
