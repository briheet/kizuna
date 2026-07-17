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

const (
	defaultOrganisationID = "00000000-0000-4000-8000-000000000001"
	defaultTeamID         = "00000000-0000-4000-8000-000000000002"
)

func (r *CockroachDbIngestionRepository) CreateJobs(ctx context.Context, topicID string, topicName string, jobs []types.Job) error {
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
		if _, err := tx.Exec(ctx, `
			insert into organisations (id, name, created_at)
			values ($1, 'Kizuna', now())
			on conflict (id) do nothing;
		`, defaultOrganisationID); err != nil {
			return fmt.Errorf("create ingestion organisation: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			insert into teams (id, organisation_id, name, created_at)
			values ($1, $2, 'Development', now())
			on conflict (id) do nothing;
		`, defaultTeamID, defaultOrganisationID); err != nil {
			return fmt.Errorf("create ingestion team: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			insert into topics (id, team_id, name, created_at)
			values ($1, $2, $3, now())
			on conflict (id) do nothing;
		`, topicID, defaultTeamID, topicName); err != nil {
			return fmt.Errorf("create ingestion topic: %w", err)
		}

		for _, job := range jobs {
			if _, err := tx.Exec(
				ctx,
				query,
				job.ID,
				job.Kind,
				job.Queue,
				job.Payload,
			); err != nil {
				return fmt.Errorf("insert ingestion job %s: %w", job.Kind, err)
			}
		}
		return nil
	})
}

func (r *CockroachDbIngestionRepository) JobsStatus(ctx context.Context, req types.JobsStatusRequest) (*types.JobsStatusResponse, error) {
	resp := &types.JobsStatusResponse{}

	err := r.db.ExecuteTx(ctx, func(tx pgx.Tx) error {
		countRows, err := tx.Query(ctx, `
			select state, count(*)
			from jobs
			where ($1 = '' or payload->>'topic_id' = $1)
			  and ($2 = '' or payload->>'source_type' = $2)
			  and ($3 = '' or state = $3)
			group by state
			order by state;
		`, req.TopicID, req.SourceType, req.State)
		if err != nil {
			return err
		}
		defer countRows.Close()

		for countRows.Next() {
			var count types.JobStateCount
			if err := countRows.Scan(&count.State, &count.Count); err != nil {
				return err
			}
			resp.Counts = append(resp.Counts, count)
		}
		if err := countRows.Err(); err != nil {
			return err
		}

		failureRows, err := tx.Query(ctx, `
			select
				id::string,
				kind,
				queue,
				state,
				attempt,
				max_attempt,
				coalesce(last_error, ''),
				payload,
				coalesce(completed_at, attempted_at, updated_at, created_at, now()) as recent_at
			from jobs
			where ($1 = '' or payload->>'topic_id' = $1)
			  and ($2 = '' or payload->>'source_type' = $2)
			  and ($3 = '' or state = $3)
			  and (state in ('failed', 'discarded') or last_error is not null)
			order by recent_at desc
			limit $4;
		`, req.TopicID, req.SourceType, req.State, req.Limit)
		if err != nil {
			return err
		}
		defer failureRows.Close()

		for failureRows.Next() {
			var job types.FailedJob
			if err := failureRows.Scan(
				&job.ID,
				&job.Kind,
				&job.Queue,
				&job.State,
				&job.Attempt,
				&job.MaxAttempt,
				&job.LastError,
				&job.Payload,
				&job.RecentAt,
			); err != nil {
				return err
			}
			resp.RecentFailures = append(resp.RecentFailures, job)
		}

		return failureRows.Err()
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
