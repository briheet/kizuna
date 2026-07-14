package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type GithubRepository interface {
	HandleJob(ctx context.Context, jobID uuid.UUID, kind string, payload json.RawMessage) error
}
