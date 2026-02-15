package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tursodatabase/go-libsql"
)

type LibSQLConfig struct {
	DataDir  string
	URL      string
	Token    string
	Interval time.Duration
}

type LibSQLConnection struct {
	Connector *libsql.Connector
	DB        *sql.DB
}

func NewLibSQLConnection(cfg *LibSQLConfig) (*LibSQLConnection, error) {
	if cfg.Interval == 0 {
		cfg.Interval = time.Minute * 5
	}

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	localDBPath := filepath.Join(cfg.DataDir, "data.db")

	opts := []libsql.Option{
		libsql.WithSyncInterval(cfg.Interval),
	}

	if cfg.Token != "" {
		opts = append(opts, libsql.WithAuthToken(cfg.Token))
	}

	connector, err := libsql.NewEmbeddedReplicaConnector(localDBPath, cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s %s: %w", localDBPath, cfg.URL, err)
	}

	return &LibSQLConnection{
		Connector: connector,
		DB:        sql.OpenDB(connector),
	}, nil
}

func (c *LibSQLConnection) Sync() error {
	if c.Connector != nil {
		_, err := c.Connector.Sync()
		return err
	}
	return nil
}

func (c *LibSQLConnection) Close() error {
	if c.Connector != nil {
		return c.Connector.Close()
	}
	return nil
}
