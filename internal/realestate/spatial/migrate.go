package spatial

import (
	"context"
	"database/sql"
	"fmt"
)

// migrate creates the R*Tree tables and triggers that live alongside
// PocketBase's "properties" table in the same database.
//
// All statements are idempotent (IF NOT EXISTS), so it is safe to call
// on every startup. The triggers fire on SQLite-level INSERT/UPDATE/DELETE,
// so PocketBase's ORM writes are indexed automatically — no Go hooks needed.
//
// Each migration is a slice of SQL statements. Turso's Hrana protocol
// rejects multi-statement strings, so every statement is executed separately
// within the same transaction.
func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS spatial_schema_migrations (
			name       TEXT PRIMARY KEY NOT NULL,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`); err != nil {
		return fmt.Errorf("spatial migrate: create tracking table: %w", err)
	}

	for _, m := range migrations {
		var exists int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM spatial_schema_migrations WHERE name = ?`, m.name,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if exists > 0 {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx %s: %w", m.name, err)
		}

		for _, stmt := range m.stmts {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("apply migration %s: %w", m.name, err)
			}
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO spatial_schema_migrations(name) VALUES (?)`, m.name,
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

type migration struct {
	name  string
	stmts []string
}

var migrations = []migration{
	{
		name: "001_property_spatial_map",
		stmts: []string{
			`CREATE TABLE IF NOT EXISTS property_spatial_map (
				rid         INTEGER PRIMARY KEY AUTOINCREMENT,
				property_id TEXT NOT NULL UNIQUE REFERENCES properties(id) ON DELETE CASCADE
			)`,
		},
	},
	{
		name: "002_property_rtree",
		stmts: []string{
			`CREATE VIRTUAL TABLE IF NOT EXISTS property_rtree USING rtree(
				id,
				min_lat, max_lat,
				min_lng, max_lng
			)`,
		},
	},
	{
		// Turso rejects multi-statement exec — each trigger is its own statement.
		name: "003_triggers",
		stmts: []string{
			// Insert: populate spatial index when coordinates are non-zero
			`CREATE TRIGGER IF NOT EXISTS trg_property_spatial_insert
			AFTER INSERT ON properties
			FOR EACH ROW
			WHEN NEW.lat IS NOT NULL AND NEW.lat != 0
			 AND NEW.lng IS NOT NULL AND NEW.lng != 0
			BEGIN
			    INSERT OR IGNORE INTO property_spatial_map(property_id) VALUES (NEW.id);
			    INSERT OR REPLACE INTO property_rtree(id, min_lat, max_lat, min_lng, max_lng)
			        SELECT m.rid, NEW.lat, NEW.lat, NEW.lng, NEW.lng
			        FROM property_spatial_map m WHERE m.property_id = NEW.id;
			END`,

			// Update: refresh R*Tree entry when lat/lng changes
			`CREATE TRIGGER IF NOT EXISTS trg_property_spatial_update
			AFTER UPDATE OF lat, lng ON properties
			FOR EACH ROW
			WHEN NEW.lat IS NOT NULL AND NEW.lat != 0
			 AND NEW.lng IS NOT NULL AND NEW.lng != 0
			BEGIN
			    INSERT OR IGNORE INTO property_spatial_map(property_id) VALUES (NEW.id);
			    INSERT OR REPLACE INTO property_rtree(id, min_lat, max_lat, min_lng, max_lng)
			        SELECT m.rid, NEW.lat, NEW.lat, NEW.lng, NEW.lng
			        FROM property_spatial_map m WHERE m.property_id = NEW.id;
			END`,

			// Delete: FK-safe fallback for when FK enforcement is off
			`CREATE TRIGGER IF NOT EXISTS trg_property_rtree_cleanup
			AFTER DELETE ON properties
			FOR EACH ROW
			BEGIN
			    DELETE FROM property_rtree
			    WHERE id IN (
			        SELECT rid FROM property_spatial_map WHERE property_id = OLD.id
			    );
			    DELETE FROM property_spatial_map WHERE property_id = OLD.id;
			END`,

			// spatial_map cascade: keep R*Tree row consistent if map row is removed
			`CREATE TRIGGER IF NOT EXISTS trg_property_rtree_from_map_delete
			AFTER DELETE ON property_spatial_map
			FOR EACH ROW
			BEGIN
			    DELETE FROM property_rtree WHERE id = OLD.rid;
			END`,
		},
	},
	{
		// Backfill existing properties into the spatial index.
		name: "004_backfill",
		stmts: []string{
			`INSERT OR IGNORE INTO property_spatial_map(property_id)
			    SELECT id FROM properties
			    WHERE lat IS NOT NULL AND lat != 0
			      AND lng IS NOT NULL AND lng != 0`,

			`INSERT OR IGNORE INTO property_rtree(id, min_lat, max_lat, min_lng, max_lng)
			    SELECT m.rid, p.lat, p.lat, p.lng, p.lng
			    FROM property_spatial_map m
			    JOIN properties p ON p.id = m.property_id`,
		},
	},
}
