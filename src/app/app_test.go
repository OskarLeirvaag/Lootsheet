package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/repo"
)

func TestRunInitCreatesSQLiteDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var stdout bytes.Buffer

	if err := Run(context.Background(), []string{"init"}, &stdout); err != nil {
		t.Fatalf("run app: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "LootSheet initialized") {
		t.Fatalf("init output missing readiness line: %q", output)
	}

	if !strings.Contains(output, configPath) {
		t.Fatalf("init output missing config path: %q", output)
	}

	databasePath := filepath.Join(dataDir, "ledger.db")
	if !strings.Contains(output, databasePath) {
		t.Fatalf("init output missing database path: %q", output)
	}

	if !strings.Contains(output, "Seeded accounts: 16") {
		t.Fatalf("init output missing seed count: %q", output)
	}
}

func TestRunAccountListReadsFromSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"account", "list"}, &stdout); err != nil {
		t.Fatalf("run account list: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "CODE  TYPE       ACTIVE  NAME") {
		t.Fatalf("account list missing header: %q", output)
	}

	if !strings.Contains(output, "1000  asset      yes     Party Cash") {
		t.Fatalf("account list missing Party Cash: %q", output)
	}
}

func TestRunDatabaseStatusBeforeInit(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"db", "status"}, &stdout); err != nil {
		t.Fatalf("run db status: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Exists: no") {
		t.Fatalf("db status missing Exists=no: %q", output)
	}

	if !strings.Contains(output, "State: uninitialized") {
		t.Fatalf("db status missing uninitialized state: %q", output)
	}

	if !strings.Contains(output, "Target schema version: 2") {
		t.Fatalf("db status missing target schema version: %q", output)
	}

	if !strings.Contains(output, "Applied migrations: 0") {
		t.Fatalf("db status missing migration count: %q", output)
	}

	if !strings.Contains(output, "Pending migrations: 0") {
		t.Fatalf("db status missing pending migration count: %q", output)
	}
}

func TestRunJournalPostCreatesPostedEntry(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var stdout bytes.Buffer
	err := Run(context.Background(), []string{
		"journal",
		"post",
		"--date", "2026-03-08",
		"--description", "Restock arrows",
		"--debit", "5100:25:Quiver refill",
		"--credit", "1000:25",
	}, &stdout)
	if err != nil {
		t.Fatalf("run journal post: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Posted journal entry #1") {
		t.Fatalf("journal post output missing entry number: %q", output)
	}

	if !strings.Contains(output, "Debits: 25") || !strings.Contains(output, "Credits: 25") {
		t.Fatalf("journal post output missing totals: %q", output)
	}
}

func TestRunJournalPostRejectsUnbalancedEntry(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	err := Run(context.Background(), []string{
		"journal",
		"post",
		"--date", "2026-03-08",
		"--description", "Broken entry",
		"--debit", "5100:25",
		"--credit", "1000:20",
	}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected journal post to fail")
	}

	if !strings.Contains(err.Error(), "journal entry is not balanced") {
		t.Fatalf("error = %q, want balance error", err)
	}
}

func TestRunJournalReverseCreatesReversalEntry(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	// Post an entry first.
	var postStdout bytes.Buffer
	err := Run(context.Background(), []string{
		"journal",
		"post",
		"--date", "2026-03-08",
		"--description", "Restock arrows",
		"--debit", "5100:25:Quiver refill",
		"--credit", "1000:25",
	}, &postStdout)
	if err != nil {
		t.Fatalf("run journal post: %v", err)
	}

	// Get the entry ID from the database.
	entryID := getFirstJournalEntryID(t, databasePath)

	// Reverse it.
	var reverseStdout bytes.Buffer
	err = Run(context.Background(), []string{
		"journal",
		"reverse",
		"--entry-id", entryID,
		"--date", "2026-03-09",
	}, &reverseStdout)
	if err != nil {
		t.Fatalf("run journal reverse: %v", err)
	}

	output := reverseStdout.String()
	if !strings.Contains(output, "Reversed journal entry as #2") {
		t.Fatalf("reverse output missing entry number: %q", output)
	}

	if !strings.Contains(output, "Debits: 25") || !strings.Contains(output, "Credits: 25") {
		t.Fatalf("reverse output missing totals: %q", output)
	}

	if !strings.Contains(output, "Reversal of entry #1") {
		t.Fatalf("reverse output missing default description: %q", output)
	}
}

func TestRunJournalReverseWithCustomDescription(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var postStdout bytes.Buffer
	err := Run(context.Background(), []string{
		"journal",
		"post",
		"--date", "2026-03-08",
		"--description", "Wrong entry",
		"--debit", "5100:50",
		"--credit", "1000:50",
	}, &postStdout)
	if err != nil {
		t.Fatalf("run journal post: %v", err)
	}

	entryID := getFirstJournalEntryID(t, databasePath)

	var reverseStdout bytes.Buffer
	err = Run(context.Background(), []string{
		"journal",
		"reverse",
		"--entry-id", entryID,
		"--date", "2026-03-09",
		"--description", "Correcting duplicate purchase",
	}, &reverseStdout)
	if err != nil {
		t.Fatalf("run journal reverse: %v", err)
	}

	if !strings.Contains(reverseStdout.String(), "Correcting duplicate purchase") {
		t.Fatalf("reverse output missing custom description: %q", reverseStdout.String())
	}
}

func getFirstJournalEntryID(t *testing.T, databasePath string) string {
	t.Helper()

	db, err := repo.OpenDBForTest(databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	var entryID string
	if err := db.QueryRow("SELECT id FROM journal_entries ORDER BY entry_number LIMIT 1").Scan(&entryID); err != nil {
		t.Fatalf("query first journal entry ID: %v", err)
	}

	return entryID
}

func TestRunAccountCreateAddsNewAccount(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var stdout bytes.Buffer
	err := Run(context.Background(), []string{
		"account", "create",
		"--code", "5600",
		"--name", "Tavern Reparations",
		"--type", "expense",
	}, &stdout)
	if err != nil {
		t.Fatalf("run account create: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Created account") {
		t.Fatalf("account create output missing confirmation: %q", output)
	}

	if !strings.Contains(output, "Code: 5600") {
		t.Fatalf("account create output missing code: %q", output)
	}

	if !strings.Contains(output, "Tavern Reparations") {
		t.Fatalf("account create output missing name: %q", output)
	}

	// Verify it appears in account list
	var listStdout bytes.Buffer
	if err := Run(context.Background(), []string{"account", "list"}, &listStdout); err != nil {
		t.Fatalf("run account list: %v", err)
	}

	if !strings.Contains(listStdout.String(), "5600") {
		t.Fatalf("account list missing new account: %q", listStdout.String())
	}
}

func TestRunAccountCreateRejectsDuplicateCode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	err := Run(context.Background(), []string{
		"account", "create",
		"--code", "1000",
		"--name", "Duplicate Cash",
		"--type", "asset",
	}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected account create with duplicate code to fail")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %q, want duplicate code error", err)
	}
}

func TestRunAccountRenameChangesName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var stdout bytes.Buffer
	err := Run(context.Background(), []string{
		"account", "rename",
		"--code", "1000",
		"--name", "Gold Hoard",
	}, &stdout)
	if err != nil {
		t.Fatalf("run account rename: %v", err)
	}

	if !strings.Contains(stdout.String(), "Renamed account 1000") {
		t.Fatalf("rename output missing confirmation: %q", stdout.String())
	}

	var listStdout bytes.Buffer
	if err := Run(context.Background(), []string{"account", "list"}, &listStdout); err != nil {
		t.Fatalf("run account list: %v", err)
	}

	if !strings.Contains(listStdout.String(), "Gold Hoard") {
		t.Fatalf("account list missing renamed account: %q", listStdout.String())
	}
}

func TestRunAccountDeactivateAndActivate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	// Deactivate
	var deactivateStdout bytes.Buffer
	err := Run(context.Background(), []string{
		"account", "deactivate",
		"--code", "1000",
	}, &deactivateStdout)
	if err != nil {
		t.Fatalf("run account deactivate: %v", err)
	}

	if !strings.Contains(deactivateStdout.String(), "Deactivated account 1000") {
		t.Fatalf("deactivate output missing confirmation: %q", deactivateStdout.String())
	}

	// Verify it shows as inactive in list
	var listStdout bytes.Buffer
	if err := Run(context.Background(), []string{"account", "list"}, &listStdout); err != nil {
		t.Fatalf("run account list: %v", err)
	}

	if !strings.Contains(listStdout.String(), "1000  asset      no") {
		t.Fatalf("account list missing inactive account: %q", listStdout.String())
	}

	// Reactivate
	var activateStdout bytes.Buffer
	err = Run(context.Background(), []string{
		"account", "activate",
		"--code", "1000",
	}, &activateStdout)
	if err != nil {
		t.Fatalf("run account activate: %v", err)
	}

	if !strings.Contains(activateStdout.String(), "Activated account 1000") {
		t.Fatalf("activate output missing confirmation: %q", activateStdout.String())
	}

	// Verify it shows as active again
	var listStdout2 bytes.Buffer
	if err := Run(context.Background(), []string{"account", "list"}, &listStdout2); err != nil {
		t.Fatalf("run account list: %v", err)
	}

	if !strings.Contains(listStdout2.String(), "1000  asset      yes") {
		t.Fatalf("account list missing reactivated account: %q", listStdout2.String())
	}
}

func TestRunDatabaseStatusAfterInitShowsAppliedMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"db", "status"}, &stdout); err != nil {
		t.Fatalf("run db status: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Exists: yes") {
		t.Fatalf("db status missing Exists=yes: %q", output)
	}

	if !strings.Contains(output, "State: current") {
		t.Fatalf("db status missing current state: %q", output)
	}

	if !strings.Contains(output, "Schema version: 2") {
		t.Fatalf("db status missing schema version: %q", output)
	}

	if !strings.Contains(output, "Target schema version: 2") {
		t.Fatalf("db status missing target schema version: %q", output)
	}

	if !strings.Contains(output, "Applied migrations: 2") {
		t.Fatalf("db status missing migration count: %q", output)
	}

	if !strings.Contains(output, "Pending migrations: 0") {
		t.Fatalf("db status missing pending migration count: %q", output)
	}

	if !strings.Contains(output, "2  002_add_journal_entry_reversal_tracking.sql") {
		t.Fatalf("db status missing second migration row: %q", output)
	}
}

func TestRunDatabaseStatusShowsUpgradeableDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	fullAssets, legacyAssets := loadMigrationAssetsForAppTest(t)
	if _, err := repo.EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
		t.Fatalf("initialize legacy db: %v", err)
	}

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"db", "status"}, &stdout); err != nil {
		t.Fatalf("run db status: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "State: upgradeable") {
		t.Fatalf("db status missing upgradeable state: %q", output)
	}

	if !strings.Contains(output, "Schema version: 1") {
		t.Fatalf("db status missing schema version 1: %q", output)
	}

	if !strings.Contains(output, "Target schema version: "+fullAssets.SchemaVersion) {
		t.Fatalf("db status missing target schema version: %q", output)
	}

	if !strings.Contains(output, "Pending migrations: 1") {
		t.Fatalf("db status missing pending migration count: %q", output)
	}

	if !strings.Contains(output, "2  002_add_journal_entry_reversal_tracking.sql") {
		t.Fatalf("db status missing pending migration row: %q", output)
	}
}

func TestRunDatabaseMigrateAppliesPendingMigration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	_, legacyAssets := loadMigrationAssetsForAppTest(t)
	if _, err := repo.EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
		t.Fatalf("initialize legacy db: %v", err)
	}

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"db", "migrate"}, &stdout); err != nil {
		t.Fatalf("run db migrate: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "State: migrated") {
		t.Fatalf("db migrate missing migrated state: %q", output)
	}

	if !strings.Contains(output, "From schema version: 1") {
		t.Fatalf("db migrate missing from schema version: %q", output)
	}

	if !strings.Contains(output, "To schema version: 2") {
		t.Fatalf("db migrate missing to schema version: %q", output)
	}

	if !strings.Contains(output, "2  002_add_journal_entry_reversal_tracking.sql") {
		t.Fatalf("db migrate missing applied migration row: %q", output)
	}
}

func loadMigrationAssetsForAppTest(t *testing.T) (config.InitAssets, config.InitAssets) {
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
