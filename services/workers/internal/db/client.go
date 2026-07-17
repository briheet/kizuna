package db

import (
	"context"
	"time"

	"github.com/briheet/kizuna/workers/internal/config"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.Db.DatabaseURL)
	if err != nil {
		return nil, err
	}

	dbConfig.MaxConns = 8
	dbConfig.MinConns = 2
	dbConfig.MaxConnLifetime = 5 * time.Minute
	dbConfig.MaxConnLifetimeJitter = 30 * time.Second
	dbConfig.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &Client{
		pool: pool,
	}, nil
}

func (c *Client) Conn() *pgxpool.Pool {
	return c.pool
}

func (c *Client) ExecuteTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return crdbpgx.ExecuteTx(ctx, c.pool, pgx.TxOptions{}, fn)
}

func (c *Client) Close(_ context.Context) error {
	c.pool.Close()
	return nil
}
