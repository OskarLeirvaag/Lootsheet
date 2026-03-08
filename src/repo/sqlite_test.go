package repo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
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
