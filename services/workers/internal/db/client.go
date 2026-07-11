package db

import (
	"context"

	"github.com/briheet/kizuna/workers/internal/config"
	"github.com/jackc/pgx/v5"
)

type Client struct {
	conn *pgx.Conn
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	dbConfig, err := pgx.ParseConfig(cfg.Db.DatabaseURL)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) Conn() *pgx.Conn {
	return c.conn
}

func (c *Client) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}
