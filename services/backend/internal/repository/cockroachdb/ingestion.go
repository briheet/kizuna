package cockroachdb

import "github.com/briheet/kizuna/backend/internal/db"

type CockroachDbIngestionRepository struct {
	db *db.Client
}

func NewCockroachDbIngestionRepository(db *db.Client) *CockroachDbIngestionRepository {
	return &CockroachDbIngestionRepository{db: db}
}
