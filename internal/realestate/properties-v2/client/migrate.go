package client

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

// schema embeds all runtime migration SQL files.
// 000_rtree_stubs.sql is sqlc-only and excluded here.
//
//go:embed sql/schema/001_properties.sql sql/schema/002_listings.sql sql/schema/003_rentals.sql sql/schema/004_media.sql sql/schema/005_spatial.sql sql/schema/006_history_triggers.sql
var schema embed.FS

// migration represents a single schema migration identified by a unique name.
// Once applied, its name is recorded in the schema_migrations table and never
// re-applied — even if the SQL uses CREATE TABLE IF NOT EXISTS.
type migration struct {
	name string
	sql  string
}

var migrations []migration

func init() {
	files := []struct{ name, embed string }{
		{"001_properties", mustRead("sql/schema/001_properties.sql")},
		{"002_listings", mustRead("sql/schema/002_listings.sql")},
		{"003_rentals", mustRead("sql/schema/003_rentals.sql")},
		{"004_media", mustRead("sql/schema/004_media.sql")},
		{"005_spatial", mustRead("sql/schema/005_spatial.sql")},
		{"006_history_triggers", mustRead("sql/schema/006_history_triggers.sql")},
	}
	for _, f := range files {
		migrations = append(migrations, migration{name: f.name, sql: f.embed})
	}
}

func mustRead(path string) string {
	b, err := schema.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("migrate: missing embedded file %s: %v", path, err))
	}
	return string(b)
}

// Migrate creates the schema_migrations tracking table then applies any
// migrations that have not yet run. Safe to call on every startup.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       TEXT PRIMARY KEY NOT NULL,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var exists int
		err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, m.name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if exists > 0 {
			continue // already applied
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", m.name, err)
		}

		if _, err := tx.ExecContext(ctx, m.sql); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", m.name, err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(name) VALUES (?)`, m.name,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.name, err)
		}
	}
	return nil
}

// CreateTable runs all migrations. Panics on error (startup path).
func CreateTable(db *sql.DB) {
	CreateTableContext(context.Background(), db)
}

// CreateTableContext runs all migrations. Panics on error (startup path).
func CreateTableContext(ctx context.Context, db *sql.DB) {
	if err := Migrate(ctx, db); err != nil {
		panic(fmt.Sprintf("migrate: %v", err))
	}
}
