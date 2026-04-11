// Package spatial wires R*Tree spatial indexing into PocketBase without
// touching any of PocketBase's own collection tables.
//
// How it works:
//
//  1. On startup, migrate() creates three lightweight objects in the same
//     database:
//     - property_spatial_map   (bridges PocketBase TEXT id → R*Tree integer id)
//     - property_rtree         (SQLite R*Tree virtual table, lat/lng indexed)
//     - Four SQLite triggers on PocketBase's "properties" table
//
//  2. From that point on, every INSERT / UPDATE / DELETE that PocketBase
//     issues against "properties" is intercepted at the SQLite level by the
//     triggers — no Go-level hooks are needed.
//
//  3. registerRoutes() adds a single authenticated endpoint:
//     GET /api/spatial/properties/nearby?lat=&lng=&radius=&org=
//
// Usage in server.New():
//
//	if err := spatial.Wire(app, conn.DB); err != nil {
//	    return nil, err
//	}
package spatial

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// Wire registers spatial migrations and API routes on app.
// Call this in server.New() after creating the LibSQL connection.
//
// Migrations run inside OnServe (after PocketBase has created the "properties"
// table) so the triggers can reference it safely.
func Wire(app core.App, db *sql.DB) error {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := migrate(context.Background(), db); err != nil {
			return fmt.Errorf("spatial.Wire: %w", err)
		}
		return e.Next()
	})
	registerRoutes(app, db)
	return nil
}
