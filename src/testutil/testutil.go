// Package testutil provides cross-package test helpers for initializing
// temporary databases, running ad-hoc queries, and loading SQL fixtures.
// It is intentionally excluded from deadcode analysis because its functions
// are only reachable from _test.go files.
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// InitTestDB creates a temporary SQLite database initialized with the full
// LootSheet schema and seed data. It returns the path to the database file.
// The database is automatically cleaned up when the test finishes.
func InitTestDB(t testing.TB) string {
	t.Helper()

	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := ledger.EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	return databasePath
}

// RunSQLiteQueryForTest opens a database, runs a query that returns a single
// text column per row, and returns all rows joined by newlines.
func RunSQLiteQueryForTest(t testing.TB, databasePath string, query string) string {
	t.Helper()

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		t.Fatalf("run test query: %v", err)
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			t.Fatalf("scan test row: %v", err)
		}
		lines = append(lines, value)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate test rows: %v", err)
	}

	return strings.Join(lines, "\n")
}

// RunSQLiteScriptForTest opens a database and executes an arbitrary SQL script.
func RunSQLiteScriptForTest(t testing.TB, databasePath string, sqlScript string) {
	t.Helper()

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), sqlScript); err != nil {
		t.Fatalf("run test script: %v: %s", err, fmt.Sprintf("%.200s", sqlScript))
	}
}

// LoadFixtureForTest reads a checked-in SQL fixture from the repository's
// fixtures directory.
func LoadFixtureForTest(t testing.TB, fixtureName string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve testutil caller path")
	}

	// fixtures/ is at repo root; this file is at src/testutil/testutil.go
	fixturePath := filepath.Join(filepath.Dir(currentFile), "..", "..", "fixtures", fixtureName)
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture %q: %v", fixtureName, err)
	}

	return string(content)
}

// ApplyFixtureForTest loads and executes a checked-in SQL fixture against an
// initialized test database.
func ApplyFixtureForTest(t testing.TB, databasePath string, fixtureName string) {
	t.Helper()

	RunSQLiteScriptForTest(t, databasePath, LoadFixtureForTest(t, fixtureName))
}

// LoadMigrationAssetsForTest returns the full and legacy (v1-only) init assets
// for migration testing.
func LoadMigrationAssetsForTest(t testing.TB) (config.InitAssets, config.InitAssets) {
	t.Helper()

	fullAssets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	legacyAssets := fullAssets
	legacyAssets.Migrations = append([]config.InitMigration(nil), fullAssets.Migrations[:1]...)
	legacyAssets.SchemaVersion = legacyAssets.Migrations[len(legacyAssets.Migrations)-1].Version

	return fullAssets, legacyAssets
}

// ExecSQLiteForTest opens a database and executes a parameterized SQL statement.
func ExecSQLiteForTest(t testing.TB, databasePath string, query string, args ...any) {
	t.Helper()

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("exec test query: %v", err)
	}
}

// AcceptQuest transitions a quest to 'accepted' status via direct SQL.
// Use this only for test setup — it skips domain validation.
func AcceptQuest(t testing.TB, databasePath string, questID string, acceptedDate string) {
	t.Helper()

	ExecSQLiteForTest(t, databasePath,
		`UPDATE quests SET status = 'accepted', accepted_on = ? WHERE id = ?`,
		acceptedDate, questID)
}

// CompleteQuest transitions a quest to 'completed' status via direct SQL.
// Use this only for test setup — it skips domain validation.
func CompleteQuest(t testing.TB, databasePath string, questID string, completedDate string) {
	t.Helper()

	ExecSQLiteForTest(t, databasePath,
		`UPDATE quests SET status = 'completed', completed_on = ? WHERE id = ?`,
		completedDate, questID)
}
