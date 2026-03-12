package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

func newLootTestEnv(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	// Initialize the database first.
	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	return databasePath
}

func runLootCmd(t *testing.T, args []string) string {
	t.Helper()
	var stdout bytes.Buffer
	if err := Run(context.Background(), args, &stdout); err != nil {
		t.Fatalf("Run(%v): %v", args, err)
	}
	return stdout.String()
}

func runLootCmdExpectError(t *testing.T, args []string) error {
	t.Helper()
	var stdout bytes.Buffer
	return Run(context.Background(), args, &stdout)
}

func TestRunLootCreateAppraiseRecognizeSell(t *testing.T) {
	databasePath := newLootTestEnv(t)

	// Create a loot item.
	createOutput := runLootCmd(t, []string{
		"loot", "create", "--name", "Ruby Gemstone", "--source", "Dragon Hoard", "--quantity", "1",
	})

	if !strings.Contains(createOutput, "Created loot item") {
		t.Fatalf("loot create output missing confirmation: %q", createOutput)
	}
	if !strings.Contains(createOutput, "Ruby Gemstone") {
		t.Fatalf("loot create output missing name: %q", createOutput)
	}
	if !strings.Contains(createOutput, "Status: held") {
		t.Fatalf("loot create output missing status: %q", createOutput)
	}

	// List loot items.
	listOutput := runLootCmd(t, []string{"loot", "list"})
	if !strings.Contains(listOutput, "Ruby Gemstone") {
		t.Fatalf("loot list missing item: %q", listOutput)
	}
	if !strings.Contains(listOutput, "held") {
		t.Fatalf("loot list missing status: %q", listOutput)
	}

	// Get IDs from database.
	lootItemID := getFirstLootItemID(t, databasePath)

	// Appraise.
	appraiseOutput := runLootCmd(t, []string{
		"loot", "appraise", "--id", lootItemID, "--value", "500", "--date", "2026-03-08", "--appraiser", "Jeweler",
	})
	if !strings.Contains(appraiseOutput, "Appraised loot item") {
		t.Fatalf("appraise output missing confirmation: %q", appraiseOutput)
	}
	if !strings.Contains(appraiseOutput, "Value: 5 GP") {
		t.Fatalf("appraise output missing value: %q", appraiseOutput)
	}

	appraisalID := getFirstLootAppraisalID(t, databasePath)

	// Recognize.
	recognizeOutput := runLootCmd(t, []string{
		"loot", "recognize", "--appraisal-id", appraisalID, "--date", "2026-03-09", "--description", "Recognize ruby gemstone",
	})
	if !strings.Contains(recognizeOutput, "Recognized loot appraisal as journal entry #1") {
		t.Fatalf("recognize output missing entry number: %q", recognizeOutput)
	}
	if !strings.Contains(recognizeOutput, "Debits: 5 GP") || !strings.Contains(recognizeOutput, "Credits: 5 GP") {
		t.Fatalf("recognize output missing totals: %q", recognizeOutput)
	}

	// Verify recognized status.
	listOutput2 := runLootCmd(t, []string{"loot", "list"})
	if !strings.Contains(listOutput2, "recognized") {
		t.Fatalf("loot list missing recognized status: %q", listOutput2)
	}

	// Sell above appraisal.
	sellOutput := runLootCmd(t, []string{
		"loot", "sell", "--id", lootItemID, "--amount", "600", "--date", "2026-03-10", "--description", "Sold ruby to merchant",
	})
	if !strings.Contains(sellOutput, "Sold loot item as journal entry #2") {
		t.Fatalf("sell output missing entry number: %q", sellOutput)
	}
	if !strings.Contains(sellOutput, "Amount: 6 GP") {
		t.Fatalf("sell output missing amount: %q", sellOutput)
	}

	// Verify sold status.
	listOutput3 := runLootCmd(t, []string{"loot", "list"})
	if !strings.Contains(listOutput3, "sold") {
		t.Fatalf("loot list missing sold status: %q", listOutput3)
	}
}

func TestRunLootSellBelowAppraisal(t *testing.T) {
	databasePath := newLootTestEnv(t)

	runLootCmd(t, []string{"loot", "create", "--name", "Chipped Diamond", "--source", "Ruins"})

	lootItemID := getFirstLootItemID(t, databasePath)

	runLootCmd(t, []string{"loot", "appraise", "--id", lootItemID, "--value", "500", "--date", "2026-03-08"})

	appraisalID := getFirstLootAppraisalID(t, databasePath)

	runLootCmd(t, []string{"loot", "recognize", "--appraisal-id", appraisalID, "--date", "2026-03-09"})

	sellOutput := runLootCmd(t, []string{
		"loot", "sell", "--id", lootItemID, "--amount", "300", "--date", "2026-03-10",
	})

	// Debits: 300 (cash) + 200 (loss) = 500; Credits: 500 (inventory).
	if !strings.Contains(sellOutput, "Debits: 5 GP") || !strings.Contains(sellOutput, "Credits: 5 GP") {
		t.Fatalf("sell output missing balanced totals: %q", sellOutput)
	}
}

func TestRunLootCreateMissingName(t *testing.T) {
	_ = newLootTestEnv(t)

	err := runLootCmdExpectError(t, []string{"loot", "create"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}

	if !strings.Contains(err.Error(), "--name is required") {
		t.Fatalf("error = %q, want --name required", err)
	}
}

func getFirstLootItemID(t *testing.T, databasePath string) string {
	t.Helper()

	db, err := ledger.OpenDBForTest(databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	var id string
	if err := db.QueryRowContext(context.Background(), "SELECT id FROM loot_items ORDER BY created_at LIMIT 1").Scan(&id); err != nil {
		t.Fatalf("query first loot item ID: %v", err)
	}

	return id
}

func getFirstLootAppraisalID(t *testing.T, databasePath string) string {
	t.Helper()

	db, err := ledger.OpenDBForTest(databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	var id string
	if err := db.QueryRowContext(context.Background(), "SELECT id FROM loot_appraisals ORDER BY created_at LIMIT 1").Scan(&id); err != nil {
		t.Fatalf("query first loot appraisal ID: %v", err)
	}

	return id
}
