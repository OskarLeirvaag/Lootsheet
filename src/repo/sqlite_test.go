package repo

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

func TestEnsureSQLiteInitializedCreatesSchemaAndSeeds(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	result, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	if !result.Initialized {
		t.Fatal("expected database to be initialized")
	}

	accountCount := strings.TrimSpace(runSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM accounts;"))
	if accountCount != "16" {
		t.Fatalf("account count = %q, want 16", accountCount)
	}

	schemaVersion := strings.TrimSpace(runSQLiteQueryForTest(t, databasePath, "SELECT value FROM settings WHERE key = 'schema_version';"))
	if schemaVersion != "2" {
		t.Fatalf("schema version = %q, want 2", schemaVersion)
	}

	migrationRows := strings.TrimSpace(runSQLiteQueryForTest(
		t,
		databasePath,
		"SELECT version || '\t' || name FROM schema_migrations ORDER BY version;",
	))
	if migrationRows != "1\t001_init.sql\n2\t002_add_journal_entry_reversal_tracking.sql" {
		t.Fatalf("migration rows = %q, want init migration records", migrationRows)
	}
}

func TestEnsureSQLiteInitializedDoesNotReseedExistingDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	runSQLiteScriptForTest(
		t,
		databasePath,
		"INSERT INTO accounts (id, code, name, type, active) VALUES ('custom_tavern_reparations', '5600', 'Tavern Reparations', 'expense', 1);",
	)

	result, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("re-run sqlite initialization: %v", err)
	}

	if result.Initialized {
		t.Fatal("expected existing database to remain untouched")
	}

	accountCount := strings.TrimSpace(runSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM accounts;"))
	if accountCount != "17" {
		t.Fatalf("account count = %q, want 17", accountCount)
	}
}

func TestListAccountsReturnsSeededAccounts(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	accounts, err := ListAccounts(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	if len(accounts) != 16 {
		t.Fatalf("account count = %d, want 16", len(accounts))
	}

	if accounts[0].Code != "1000" || accounts[0].Name != "Party Cash" {
		t.Fatalf("first account = %+v, want Party Cash at code 1000", accounts[0])
	}
}

func TestPostJournalEntryCreatesPostedEntryAndLines(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	result, err := PostJournalEntry(context.Background(), databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []service.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25, Memo: "Quiver refill"},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	if result.EntryNumber != 1 {
		t.Fatalf("entry number = %d, want 1", result.EntryNumber)
	}

	if result.LineCount != 2 || result.DebitTotal != 25 || result.CreditTotal != 25 {
		t.Fatalf("result = %+v, want 2 lines and 25/25 totals", result)
	}

	entryRow := strings.TrimSpace(runSQLiteQueryForTest(
		t,
		databasePath,
		"SELECT status || '\t' || entry_date || '\t' || description || '\t' || posted_at FROM journal_entries;",
	))
	fields := strings.Split(entryRow, "\t")
	if len(fields) != 4 {
		t.Fatalf("entry row columns = %d, want 4", len(fields))
	}

	if fields[0] != "posted" || fields[1] != "2026-03-08" || fields[2] != "Restock arrows" {
		t.Fatalf("entry row = %q, want posted journal entry", entryRow)
	}

	if fields[3] == "" {
		t.Fatalf("posted_at is empty in row %q", entryRow)
	}

	lineCount := strings.TrimSpace(runSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_lines;"))
	if lineCount != "2" {
		t.Fatalf("journal line count = %q, want 2", lineCount)
	}
}

func TestPostJournalEntryRejectsUnbalancedInput(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	_, err = PostJournalEntry(context.Background(), databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Broken entry",
		Lines: []service.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 20},
		},
	})
	if err == nil {
		t.Fatal("expected post journal entry to fail")
	}

	if !strings.Contains(err.Error(), "journal entry is not balanced") {
		t.Fatalf("error = %q, want balance error", err)
	}

	entryCount := strings.TrimSpace(runSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_entries;"))
	if entryCount != "0" {
		t.Fatalf("journal entry count = %q, want 0", entryCount)
	}
}

func TestGetDatabaseStatusReturnsUninitializedForMissingDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "missing.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.Exists {
		t.Fatal("expected missing database to report Exists=false")
	}

	if status.Initialized {
		t.Fatal("expected missing database to report Initialized=false")
	}

	if status.State != DatabaseStateUninitialized {
		t.Fatalf("state = %q, want %q", status.State, DatabaseStateUninitialized)
	}

	if status.SchemaVersion != "" {
		t.Fatalf("schema version = %q, want empty", status.SchemaVersion)
	}

	if status.TargetSchemaVersion != "2" {
		t.Fatalf("target schema version = %q, want 2", status.TargetSchemaVersion)
	}

	if len(status.AppliedMigrations) != 0 {
		t.Fatalf("applied migrations = %d, want 0", len(status.AppliedMigrations))
	}

	if len(status.PendingMigrations) != 0 {
		t.Fatalf("pending migrations = %d, want 0", len(status.PendingMigrations))
	}
}

func TestGetDatabaseStatusReturnsAppliedMigrationsAfterInit(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	status, err := GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if !status.Exists || !status.Initialized {
		t.Fatalf("status = %+v, want existing initialized database", status)
	}

	if status.State != DatabaseStateCurrent {
		t.Fatalf("state = %q, want %q", status.State, DatabaseStateCurrent)
	}

	if status.SchemaVersion != "2" {
		t.Fatalf("schema version = %q, want 2", status.SchemaVersion)
	}

	if status.TargetSchemaVersion != "2" {
		t.Fatalf("target schema version = %q, want 2", status.TargetSchemaVersion)
	}

	if len(status.AppliedMigrations) != 2 {
		t.Fatalf("applied migrations = %d, want 2", len(status.AppliedMigrations))
	}

	if len(status.PendingMigrations) != 0 {
		t.Fatalf("pending migrations = %d, want 0", len(status.PendingMigrations))
	}

	if status.AppliedMigrations[1].Name != "002_add_journal_entry_reversal_tracking.sql" {
		t.Fatalf("second migration = %+v, want 002_add_journal_entry_reversal_tracking.sql", status.AppliedMigrations[1])
	}
}

func TestGetDatabaseStatusFallsBackToLegacySettingsVersion(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "legacy.db")

	runSQLiteScriptForTest(t, databasePath, `
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO settings (key, value) VALUES ('schema_version', '1');
`)

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if !status.Exists || !status.Initialized {
		t.Fatalf("status = %+v, want existing initialized legacy database", status)
	}

	if status.State != DatabaseStateUpgradeable {
		t.Fatalf("state = %q, want %q", status.State, DatabaseStateUpgradeable)
	}

	if status.SchemaVersion != "1" {
		t.Fatalf("schema version = %q, want 1", status.SchemaVersion)
	}

	if len(status.AppliedMigrations) != 0 {
		t.Fatalf("applied migrations = %d, want 0 for legacy fallback", len(status.AppliedMigrations))
	}

	if len(status.PendingMigrations) != 1 || status.PendingMigrations[0].Version != "2" {
		t.Fatalf("pending migrations = %+v, want version 2 pending migration", status.PendingMigrations)
	}
}

func TestMigrateSQLiteDatabaseAppliesPendingMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	fullAssets, legacyAssets := loadMigrationAssetsForTest(t)

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
		t.Fatalf("initialize legacy sqlite database: %v", err)
	}

	result, err := MigrateSQLiteDatabase(context.Background(), databasePath, fullAssets)
	if err != nil {
		t.Fatalf("migrate sqlite database: %v", err)
	}

	if !result.Migrated {
		t.Fatal("expected database migration to apply pending migrations")
	}

	if result.MetadataRepaired {
		t.Fatal("expected metadata_repaired=false for normal migration")
	}

	if result.FromSchemaVersion != "1" || result.ToSchemaVersion != "2" {
		t.Fatalf("result versions = %+v, want 1 -> 2", result)
	}

	if len(result.AppliedMigrations) != 1 || result.AppliedMigrations[0].Version != "2" {
		t.Fatalf("applied migrations = %+v, want version 2", result.AppliedMigrations)
	}

	schemaVersion := strings.TrimSpace(runSQLiteQueryForTest(t, databasePath, "SELECT value FROM settings WHERE key = 'schema_version';"))
	if schemaVersion != "2" {
		t.Fatalf("schema version = %q, want 2", schemaVersion)
	}

	reversedAtColumn := strings.TrimSpace(runSQLiteQueryForTest(
		t,
		databasePath,
		"SELECT COUNT(*) FROM pragma_table_info('journal_entries') WHERE name = 'reversed_at';",
	))
	if reversedAtColumn != "1" {
		t.Fatalf("reversed_at column count = %q, want 1", reversedAtColumn)
	}
}

func TestMigrateSQLiteDatabaseBackfillsLegacyMetadataBeforeApplyingMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "legacy.db")

	fullAssets, legacyAssets := loadMigrationAssetsForTest(t)

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
		t.Fatalf("initialize legacy sqlite database: %v", err)
	}

	runSQLiteScriptForTest(t, databasePath, "DROP TABLE schema_migrations;")

	result, err := MigrateSQLiteDatabase(context.Background(), databasePath, fullAssets)
	if err != nil {
		t.Fatalf("migrate sqlite database: %v", err)
	}

	if !result.Migrated {
		t.Fatal("expected legacy database migration to apply pending migrations")
	}

	if !result.MetadataRepaired {
		t.Fatal("expected metadata_repaired=true for legacy metadata fallback")
	}

	migrationRows := strings.TrimSpace(runSQLiteQueryForTest(
		t,
		databasePath,
		"SELECT version || '\t' || name FROM schema_migrations ORDER BY version;",
	))
	if migrationRows != "1\t001_init.sql\n2\t002_add_journal_entry_reversal_tracking.sql" {
		t.Fatalf("migration rows = %q, want backfilled and applied migration records", migrationRows)
	}
}

func loadMigrationAssetsForTest(t *testing.T) (config.InitAssets, config.InitAssets) {
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

func runSQLiteQueryForTest(t *testing.T, databasePath string, query string) string {
	t.Helper()

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(query)
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

func runSQLiteScriptForTest(t *testing.T, databasePath string, sqlScript string) {
	t.Helper()

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(sqlScript); err != nil {
		t.Fatalf("run test script: %v: %s", err, fmt.Sprintf("%.200s", sqlScript))
	}
}
