package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
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

	if !strings.Contains(output, "Applied migrations: 0") {
		t.Fatalf("db status missing migration count: %q", output)
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

func TestRunDatabaseStatusAfterInitShowsAppliedMigration(t *testing.T) {
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

	if !strings.Contains(output, "State: initialized") {
		t.Fatalf("db status missing initialized state: %q", output)
	}

	if !strings.Contains(output, "Schema version: 1") {
		t.Fatalf("db status missing schema version: %q", output)
	}

	if !strings.Contains(output, "Applied migrations: 1") {
		t.Fatalf("db status missing migration count: %q", output)
	}

	if !strings.Contains(output, "1  001_init.sql") {
		t.Fatalf("db status missing migration row: %q", output)
	}
}
