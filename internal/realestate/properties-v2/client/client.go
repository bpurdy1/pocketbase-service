package client

import (
	"database/sql"

	"github.com/caarlos0/env/v11"

	"sqlite-realestate/client/db"
)

type Config struct {
}

type Client struct {
	db.Querier
	db *sql.DB
}

// RawDB returns the underlying *sql.DB for queries that bypass sqlc (e.g. spatial R*Tree).
func (c *Client) RawDB() *sql.DB { return c.db }

func NewClient(sqlDB *sql.DB) *Client {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return &Client{
		Querier: db.New(sqlDB),
		db:      sqlDB,
	}
}
