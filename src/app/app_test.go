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
