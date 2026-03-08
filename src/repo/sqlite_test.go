package repo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

func TestEnsureSQLiteInitializedCreatesSchemaAndSeeds(t *testing.T) {
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

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
	if schemaVersion != "1" {
		t.Fatalf("schema version = %q, want 1", schemaVersion)
	}

	migrationRow := strings.TrimSpace(runSQLiteQueryForTest(
		t,
		databasePath,
		"SELECT version || '\t' || name FROM schema_migrations ORDER BY version;",
	))
	if migrationRow != "1\t001_init.sql" {
		t.Fatalf("migration row = %q, want first init migration record", migrationRow)
	}
}

func TestEnsureSQLiteInitializedDoesNotReseedExistingDatabase(t *testing.T) {
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

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
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

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
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

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
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

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
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "missing.db")

	status, err := GetDatabaseStatus(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.Exists {
		t.Fatal("expected missing database to report Exists=false")
	}

	if status.Initialized {
		t.Fatal("expected missing database to report Initialized=false")
	}

	if status.SchemaVersion != "" {
		t.Fatalf("schema version = %q, want empty", status.SchemaVersion)
	}

	if len(status.AppliedMigrations) != 0 {
		t.Fatalf("applied migrations = %d, want 0", len(status.AppliedMigrations))
	}
}

func TestGetDatabaseStatusReturnsAppliedMigrationsAfterInit(t *testing.T) {
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	status, err := GetDatabaseStatus(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if !status.Exists || !status.Initialized {
		t.Fatalf("status = %+v, want existing initialized database", status)
	}

	if status.SchemaVersion != "1" {
		t.Fatalf("schema version = %q, want 1", status.SchemaVersion)
	}

	if len(status.AppliedMigrations) != 1 {
		t.Fatalf("applied migrations = %d, want 1", len(status.AppliedMigrations))
	}

	if status.AppliedMigrations[0].Name != "001_init.sql" {
		t.Fatalf("first migration = %+v, want 001_init.sql", status.AppliedMigrations[0])
	}
}

func TestGetDatabaseStatusFallsBackToLegacySettingsVersion(t *testing.T) {
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		t.Skipf("%s not available: %v", sqliteCommand, err)
	}

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

	status, err := GetDatabaseStatus(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if !status.Exists || !status.Initialized {
		t.Fatalf("status = %+v, want existing initialized legacy database", status)
	}

	if status.SchemaVersion != "1" {
		t.Fatalf("schema version = %q, want 1", status.SchemaVersion)
	}

	if len(status.AppliedMigrations) != 0 {
		t.Fatalf("applied migrations = %d, want 0 for legacy fallback", len(status.AppliedMigrations))
	}
}

func runSQLiteQueryForTest(t *testing.T, databasePath string, sql string) string {
	t.Helper()

	command := exec.Command(sqliteCommand, "-batch", "-noheader", databasePath, sql)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run sqlite query: %v: %s", err, strings.TrimSpace(string(output)))
	}

	return string(output)
}

func runSQLiteScriptForTest(t *testing.T, databasePath string, sql string) {
	t.Helper()

	command := exec.Command(sqliteCommand, databasePath)
	command.Stdin = strings.NewReader(sql)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		t.Fatalf("run sqlite script: %v", err)
	}
}
