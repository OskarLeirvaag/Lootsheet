package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
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

	if !strings.Contains(output, "Detail: -") {
		t.Fatalf("db status missing blank detail: %q", output)
	}

	if !strings.Contains(output, "Target schema version: 6") {
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

	if !strings.Contains(output, "Debits: 2 SP 5 CP") || !strings.Contains(output, "Credits: 2 SP 5 CP") {
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

	if !strings.Contains(output, "Debits: 2 SP 5 CP") || !strings.Contains(output, "Credits: 2 SP 5 CP") {
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

func TestRunEntryExpenseUsesDefaultDateAndFundingAccount(t *testing.T) {
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

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{
		"entry", "expense",
		"--account", "5100",
		"--amount", "25",
		"--description", "Restock arrows",
		"--memo", "Quiver refill",
	}, &stdout); err != nil {
		t.Fatalf("run entry expense: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Recorded expense as journal entry #1") {
		t.Fatalf("expense output missing entry number: %q", output)
	}
	if !strings.Contains(output, "Amount: 2 SP 5 CP") {
		t.Fatalf("expense output missing amount: %q", output)
	}

	databasePath := filepath.Join(dataDir, "ledger.db")
	entryRow := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, `
		SELECT entry_date || '|' || description
		FROM journal_entries
		ORDER BY entry_number
	`))
	if entryRow != "2026-03-10|Restock arrows" {
		t.Fatalf("expense entry row = %q, want default dated entry", entryRow)
	}
	lineRows := ledger.RunSQLiteQueryForTest(t, databasePath, `
		SELECT a.code || '|' || debit_amount || '|' || credit_amount || '|' || COALESCE(jl.memo, '')
		FROM journal_lines jl
		JOIN accounts a ON a.id = jl.account_id
		ORDER BY line_number
	`)
	if !strings.Contains(lineRows, "5100|25|0|Quiver refill") || !strings.Contains(lineRows, "1000|0|25|") {
		t.Fatalf("expense lines = %q, want expense debit and cash credit", lineRows)
	}
}

func TestRunEntryIncomeUsesDefaultDateAndDepositAccount(t *testing.T) {
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

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{
		"entry", "income",
		"--account", "4000",
		"--amount", "1gp",
		"--description", "Goblin bounty",
		"--memo", "Mayor payout",
	}, &stdout); err != nil {
		t.Fatalf("run entry income: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Recorded income as journal entry #1") {
		t.Fatalf("income output missing entry number: %q", output)
	}

	databasePath := filepath.Join(dataDir, "ledger.db")
	lineRows := ledger.RunSQLiteQueryForTest(t, databasePath, `
		SELECT a.code || '|' || debit_amount || '|' || credit_amount || '|' || COALESCE(jl.memo, '')
		FROM journal_lines jl
		JOIN accounts a ON a.id = jl.account_id
		ORDER BY line_number
	`)
	if !strings.Contains(lineRows, "1000|100|0|") || !strings.Contains(lineRows, "4000|0|100|Mayor payout") {
		t.Fatalf("income lines = %q, want cash debit and income credit", lineRows)
	}
}

func TestRunEntryCustomUsesDefaultDate(t *testing.T) {
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

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{
		"entry", "custom",
		"--description", "Gear transfer",
		"--debit", "1300:500",
		"--credit", "1000:500",
	}, &stdout); err != nil {
		t.Fatalf("run entry custom: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Recorded custom entry as journal entry #1") {
		t.Fatalf("custom output missing entry number: %q", output)
	}

	databasePath := filepath.Join(dataDir, "ledger.db")
	entryRow := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, `
		SELECT entry_date || '|' || description
		FROM journal_entries
		ORDER BY entry_number
	`))
	if entryRow != "2026-03-10|Gear transfer" {
		t.Fatalf("custom entry row = %q, want default dated custom entry", entryRow)
	}
}

func getFirstJournalEntryID(t *testing.T, databasePath string) string {
	t.Helper()

	db, err := ledger.OpenDBForTest(databasePath)
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

	if !strings.Contains(output, "Detail: -") {
		t.Fatalf("db status missing blank detail: %q", output)
	}

	if !strings.Contains(output, "Schema version: 6") {
		t.Fatalf("db status missing schema version: %q", output)
	}

	if !strings.Contains(output, "Target schema version: 6") {
		t.Fatalf("db status missing target schema version: %q", output)
	}

	if !strings.Contains(output, "Applied migrations: 6") {
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

	if !strings.Contains(output, "Pending migrations: 5") {
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

	ledger.RunSQLiteScriptForTest(t, databasePath, `CREATE TABLE outsiders (id TEXT PRIMARY KEY);`)

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

	if !strings.Contains(output, "To schema version: 6") {
		t.Fatalf("db migrate missing to schema version: %q", output)
	}

	if !strings.Contains(output, "2  002_add_journal_entry_reversal_tracking.sql") {
		t.Fatalf("db migrate missing applied migration row: %q", output)
	}

	if !strings.Contains(output, "3  003_add_loot_item_type.sql") {
		t.Fatalf("db migrate missing third applied migration row: %q", output)
	}
}

func TestRunQuestCreateListAcceptCompleteCollect(t *testing.T) {
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

	// Create a quest.
	var createStdout bytes.Buffer
	err := Run(context.Background(), []string{
		"quest", "create",
		"--title", "Clear the Goblin Cave",
		"--patron", "Mayor Thornton",
		"--reward", "500",
		"--status", "offered",
	}, &createStdout)
	if err != nil {
		t.Fatalf("run quest create: %v", err)
	}

	createOutput := createStdout.String()
	if !strings.Contains(createOutput, "Created quest") {
		t.Fatalf("quest create output missing confirmation: %q", createOutput)
	}
	if !strings.Contains(createOutput, "Clear the Goblin Cave") {
		t.Fatalf("quest create output missing title: %q", createOutput)
	}
	if !strings.Contains(createOutput, "Reward: 5 GP") {
		t.Fatalf("quest create output missing reward: %q", createOutput)
	}

	// List quests.
	var listStdout bytes.Buffer
	if err := Run(context.Background(), []string{"quest", "list"}, &listStdout); err != nil {
		t.Fatalf("run quest list: %v", err)
	}

	listOutput := listStdout.String()
	if !strings.Contains(listOutput, "Clear the Goblin Cave") {
		t.Fatalf("quest list missing quest: %q", listOutput)
	}
	if !strings.Contains(listOutput, "offered") {
		t.Fatalf("quest list missing status: %q", listOutput)
	}

	// Get quest ID from database.
	questID := getFirstQuestID(t, databasePath)

	// Accept the quest.
	var acceptStdout bytes.Buffer
	err = Run(context.Background(), []string{
		"quest", "accept",
		"--id", questID,
		"--date", "2026-03-05",
	}, &acceptStdout)
	if err != nil {
		t.Fatalf("run quest accept: %v", err)
	}

	if !strings.Contains(acceptStdout.String(), "Accepted quest") {
		t.Fatalf("accept output missing confirmation: %q", acceptStdout.String())
	}

	// Complete the quest.
	var completeStdout bytes.Buffer
	err = Run(context.Background(), []string{
		"quest", "complete",
		"--id", questID,
		"--date", "2026-03-10",
	}, &completeStdout)
	if err != nil {
		t.Fatalf("run quest complete: %v", err)
	}

	if !strings.Contains(completeStdout.String(), "Completed quest") {
		t.Fatalf("complete output missing confirmation: %q", completeStdout.String())
	}

	// Collect full payment.
	var collectStdout bytes.Buffer
	err = Run(context.Background(), []string{
		"quest", "collect",
		"--id", questID,
		"--amount", "500",
		"--date", "2026-03-12",
	}, &collectStdout)
	if err != nil {
		t.Fatalf("run quest collect: %v", err)
	}

	collectOutput := collectStdout.String()
	if !strings.Contains(collectOutput, "Collected quest payment as journal entry #1") {
		t.Fatalf("collect output missing entry number: %q", collectOutput)
	}
	if !strings.Contains(collectOutput, "Amount: 5 GP") {
		t.Fatalf("collect output missing amount: %q", collectOutput)
	}
	if !strings.Contains(collectOutput, "Debits: 5 GP") || !strings.Contains(collectOutput, "Credits: 5 GP") {
		t.Fatalf("collect output missing totals: %q", collectOutput)
	}

	// Verify quest is now paid in the list.
	var listStdout2 bytes.Buffer
	if err := Run(context.Background(), []string{"quest", "list"}, &listStdout2); err != nil {
		t.Fatalf("run quest list: %v", err)
	}

	if !strings.Contains(listStdout2.String(), "paid") {
		t.Fatalf("quest list missing paid status: %q", listStdout2.String())
	}
}

func TestRunQuestCreateWithAcceptedStatus(t *testing.T) {
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
		"quest", "create",
		"--title", "Escort Mission",
		"--reward", "200",
		"--status", "accepted",
		"--accepted-on", "2026-03-01",
	}, &stdout)
	if err != nil {
		t.Fatalf("run quest create accepted: %v", err)
	}

	if !strings.Contains(stdout.String(), "Status: accepted") {
		t.Fatalf("quest create output missing accepted status: %q", stdout.String())
	}
}

func TestRunQuestCollectPartialPayment(t *testing.T) {
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

	// Create an accepted quest, complete it, then partial collect.
	var createStdout bytes.Buffer
	err := Run(context.Background(), []string{
		"quest", "create",
		"--title", "Long Quest",
		"--reward", "1000",
		"--status", "accepted",
		"--accepted-on", "2026-03-01",
	}, &createStdout)
	if err != nil {
		t.Fatalf("run quest create: %v", err)
	}

	questID := getFirstQuestID(t, databasePath)

	err = Run(context.Background(), []string{
		"quest", "complete",
		"--id", questID,
		"--date", "2026-03-10",
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run quest complete: %v", err)
	}

	// Partial payment.
	var collectStdout bytes.Buffer
	err = Run(context.Background(), []string{
		"quest", "collect",
		"--id", questID,
		"--amount", "400",
		"--date", "2026-03-12",
	}, &collectStdout)
	if err != nil {
		t.Fatalf("run quest collect partial: %v", err)
	}

	// Verify quest is partially_paid.
	var listStdout bytes.Buffer
	if err := Run(context.Background(), []string{"quest", "list"}, &listStdout); err != nil {
		t.Fatalf("run quest list: %v", err)
	}

	if !strings.Contains(listStdout.String(), "partially_paid") {
		t.Fatalf("quest list missing partially_paid status: %q", listStdout.String())
	}
}

func getFirstQuestID(t *testing.T, databasePath string) string {
	t.Helper()

	db, err := ledger.OpenDBForTest(databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	var questID string
	if err := db.QueryRow("SELECT id FROM quests ORDER BY created_at LIMIT 1").Scan(&questID); err != nil {
		t.Fatalf("query first quest ID: %v", err)
	}

	return questID
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
