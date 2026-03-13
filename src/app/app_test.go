package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
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

	if !strings.Contains(output, "Detail: -") {
		t.Fatalf("db status missing blank detail: %q", output)
	}

	if !strings.Contains(output, "Target schema version: "+config.SchemaVersion) {
		t.Fatalf("db status missing target schema version: %q", output)
	}

	if !strings.Contains(output, "Applied migrations: 0") {
		t.Fatalf("db status missing migration count: %q", output)
	}

	if !strings.Contains(output, "Pending migrations: 0") {
		t.Fatalf("db status missing pending migration count: %q", output)
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

	if !strings.Contains(output, "Detail: -") {
		t.Fatalf("db status missing blank detail: %q", output)
	}

	if !strings.Contains(output, "Schema version: "+config.SchemaVersion) {
		t.Fatalf("db status missing schema version: %q", output)
	}

	if !strings.Contains(output, "Target schema version: "+config.SchemaVersion) {
		t.Fatalf("db status missing target schema version: %q", output)
	}

	if !strings.Contains(output, "Applied migrations: "+config.SchemaVersion) {
		t.Fatalf("db status missing migration count: %q", output)
	}

	if !strings.Contains(output, "Pending migrations: 0") {
		t.Fatalf("db status missing pending migration count: %q", output)
	}

	if !strings.Contains(output, "2  002_add_journal_entry_reversal_tracking.sql") {
		t.Fatalf("db status missing second migration row: %q", output)
	}

	if !strings.Contains(output, "3  003_add_loot_item_type.sql") {
		t.Fatalf("db status missing third migration row: %q", output)
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
	if _, err := ledger.EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
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

	if !strings.Contains(output, "Detail: -") {
		t.Fatalf("db status missing blank detail: %q", output)
	}

	if !strings.Contains(output, "Schema version: 1") {
		t.Fatalf("db status missing schema version 1: %q", output)
	}

	if !strings.Contains(output, "Target schema version: "+fullAssets.SchemaVersion) {
		t.Fatalf("db status missing target schema version: %q", output)
	}

	pendingCount := strconv.Itoa(mustAtoi(t, config.SchemaVersion) - 1)
	if !strings.Contains(output, "Pending migrations: "+pendingCount) {
		t.Fatalf("db status missing pending migration count: %q", output)
	}

	if !strings.Contains(output, "2  002_add_journal_entry_reversal_tracking.sql") {
		t.Fatalf("db status missing pending migration row: %q", output)
	}

	if !strings.Contains(output, "3  003_add_loot_item_type.sql") {
		t.Fatalf("db status missing third pending migration row: %q", output)
	}
}

func TestRunDatabaseStatusShowsForeignDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}

	testutil.RunSQLiteScriptForTest(t, databasePath, `CREATE TABLE outsiders (id TEXT PRIMARY KEY);`)

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"db", "status"}, &stdout); err != nil {
		t.Fatalf("run db status: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "State: foreign") {
		t.Fatalf("db status missing foreign state: %q", output)
	}

	if !strings.Contains(output, "missing LootSheet migration metadata") {
		t.Fatalf("db status missing foreign detail: %q", output)
	}
}

func TestRunDatabaseStatusShowsDamagedDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}

	if err := os.WriteFile(databasePath, []byte("not a sqlite database"), 0o600); err != nil {
		t.Fatalf("write damaged db file: %v", err)
	}

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"db", "status"}, &stdout); err != nil {
		t.Fatalf("run db status: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "State: damaged") {
		t.Fatalf("db status missing damaged state: %q", output)
	}

	if !strings.Contains(output, "file is not a valid SQLite database") {
		t.Fatalf("db status missing damaged detail: %q", output)
	}
}

func TestRunDatabaseMigrateAppliesPendingMigration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")
	backupDir := filepath.Join(dataDir, "snapshots")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")
	t.Setenv(config.EnvBackupDir, backupDir)

	_, legacyAssets := loadMigrationAssetsForAppTest(t)
	if _, err := ledger.EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
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

	if !strings.Contains(output, "Backup: "+backupDir) {
		t.Fatalf("db migrate missing backup path: %q", output)
	}

	if !strings.Contains(output, "From schema version: 1") {
		t.Fatalf("db migrate missing from schema version: %q", output)
	}

	if !strings.Contains(output, "To schema version: "+config.SchemaVersion) {
		t.Fatalf("db migrate missing to schema version: %q", output)
	}

	if !strings.Contains(output, "2  002_add_journal_entry_reversal_tracking.sql") {
		t.Fatalf("db migrate missing applied migration row: %q", output)
	}

	if !strings.Contains(output, "3  003_add_loot_item_type.sql") {
		t.Fatalf("db migrate missing third applied migration row: %q", output)
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("mustAtoi(%q): %v", s, err)
	}
	return n
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
