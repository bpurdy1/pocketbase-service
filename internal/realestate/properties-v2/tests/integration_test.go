package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"sqlite-realestate/client"
)

var ctx = context.Background()

const fileDBPath = "./testdata/file_db_test.sqlite"

// fileDB is the shared Client used by all TestFileDB_* tests.
// Opened once at package init, closed in TestMain after all tests run.
var fileDB *client.Client

func newFileTestClient() (*client.Client, func()) {
	if err := os.MkdirAll("./testdata", 0755); err != nil {
		panic("mkdir testdata: " + err.Error())
	}
	sqlDB, err := sql.Open("sqlite", fileDBPath)
	if err != nil {
		panic("open file db: " + err.Error())
	}
	if _, err := sqlDB.Exec(`PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;`); err != nil {
		panic("pragmas: " + err.Error())
	}
	if err := client.Migrate(ctx, sqlDB); err != nil {
		panic("migrate: " + err.Error())
	}
	c := client.NewClient(sqlDB)
	cleanup := func() {
		sqlDB.Close()
		os.Remove(fileDBPath)
		os.Remove(fileDBPath + "-wal")
		os.Remove(fileDBPath + "-shm")
	}
	return c, cleanup
}

func TestMain(m *testing.M) {
	// var cleanup func()
	fileDB, _ = newFileTestClient()
	code := m.Run()
	// cleanup()
	os.Exit(code)
}

// newTestClient returns a Client backed by a fresh in-memory SQLite database.
// Each test gets its own isolated DB — nothing persists between tests.
func newTestClient(t *testing.T) (*client.Client, func()) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := sqlDB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		t.Fatalf("pragmas: %v", err)
	}
	if err := client.Migrate(ctx, sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return client.NewClient(sqlDB), func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	}
}

// futureDate returns an RFC3339 datetime N days from now (for expires_at).
func futureDate(days int) string {
	return time.Now().AddDate(0, 0, days).UTC().Format(time.RFC3339)
}

// pastDate returns an RFC3339 datetime N days ago.
func pastDate(days int) string {
	return time.Now().AddDate(0, 0, -days).UTC().Format(time.RFC3339)
}
