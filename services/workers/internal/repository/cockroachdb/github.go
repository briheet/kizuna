package cockroachdb

import (
	"context"
	"encoding/json"

	"github.com/briheet/kizuna/workers/internal/db"
	"github.com/google/uuid"
)

type GithubRepository struct {
	db *db.Client
}

func NewGithubRepository(db *db.Client) *GithubRepository {
	return &GithubRepository{db: db}
}

func (r *GithubRepository) HandleJob(ctx context.Context, jobID uuid.UUID, kind string, payload json.RawMessage) error {
	return nil
}
