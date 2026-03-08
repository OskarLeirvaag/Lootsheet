package app

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

func setupReportTestApp(t *testing.T) (*Application, string, *bytes.Buffer) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	var stdout bytes.Buffer
	application, err := New(cfg, &stdout, io.Discard)
	if err != nil {
		t.Fatalf("create application: %v", err)
	}

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := ledger.EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize database: %v", err)
	}

	return application, databasePath, &stdout
}

func TestRunTrialBalanceShowsBalancedOutput(t *testing.T) {
	application, databasePath, stdout := setupReportTestApp(t)
	ctx := context.Background()

	_, err := journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Buy supplies",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 50, Memo: "Rations"},
			{AccountCode: "1000", CreditAmount: 50},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	_, err = journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-02",
		Description: "Quest reward",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 200},
			{AccountCode: "4000", CreditAmount: 200},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	hctx := application.handlerContext()
	if err := application.runReport(ctx, []string{"trial-balance"}); err != nil {
		t.Fatalf("run trial balance: %v", err)
	}
	_ = hctx // used implicitly via runReport

	output := stdout.String()

	if !strings.Contains(output, "Trial Balance") {
		t.Fatalf("output missing header: %q", output)
	}

	if !strings.Contains(output, "CODE") || !strings.Contains(output, "ACCOUNT") {
		t.Fatalf("output missing column headers: %q", output)
	}

	if !strings.Contains(output, "1000") {
		t.Fatalf("output missing Party Cash account: %q", output)
	}

	if !strings.Contains(output, "4000") {
		t.Fatalf("output missing Quest Income account: %q", output)
	}

	if !strings.Contains(output, "5000") {
		t.Fatalf("output missing Adventuring Supplies account: %q", output)
	}

	if !strings.Contains(output, "BALANCED") {
		t.Fatalf("output missing BALANCED label: %q", output)
	}

	if strings.Contains(output, "UNBALANCED") {
		t.Fatalf("output should not contain UNBALANCED: %q", output)
	}
}

func TestRunTrialBalanceEmptyLedger(t *testing.T) {
	application, _, stdout := setupReportTestApp(t)
	ctx := context.Background()

	if err := application.runReport(ctx, []string{"trial-balance"}); err != nil {
		t.Fatalf("run trial balance: %v", err)
	}

	output := stdout.String()

	if !strings.Contains(output, "Trial Balance") {
		t.Fatalf("output missing header: %q", output)
	}

	if !strings.Contains(output, "BALANCED") {
		t.Fatalf("output missing BALANCED label for empty ledger: %q", output)
	}
}

func TestRunTrialBalanceAfterReversal(t *testing.T) {
	application, databasePath, stdout := setupReportTestApp(t)
	ctx := context.Background()

	posted, err := journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Buy supplies",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 100},
			{AccountCode: "1000", CreditAmount: 100},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	_, err = journal.ReverseJournalEntry(ctx, databasePath, posted.ID, "2026-03-02", "")
	if err != nil {
		t.Fatalf("reverse journal entry: %v", err)
	}

	if err := application.runReport(ctx, []string{"trial-balance"}); err != nil {
		t.Fatalf("run trial balance: %v", err)
	}

	output := stdout.String()

	if !strings.Contains(output, "BALANCED") {
		t.Fatalf("output missing BALANCED after reversal: %q", output)
	}

	if strings.Contains(output, "UNBALANCED") {
		t.Fatalf("output should not contain UNBALANCED after reversal: %q", output)
	}
}

func TestRunReportDispatcherMissingSubcommand(t *testing.T) {
	application, _, _ := setupReportTestApp(t)
	ctx := context.Background()

	err := application.runReport(ctx, []string{})
	if err == nil {
		t.Fatal("expected error for missing report subcommand")
	}

	if !strings.Contains(err.Error(), "missing report subcommand") {
		t.Fatalf("error = %q, want missing subcommand error", err)
	}
}

func TestRunReportDispatcherUnknownSubcommand(t *testing.T) {
	application, _, _ := setupReportTestApp(t)
	ctx := context.Background()

	err := application.runReport(ctx, []string{"bogus"})
	if err == nil {
		t.Fatal("expected error for unknown report subcommand")
	}

	if !strings.Contains(err.Error(), "unknown report subcommand") {
		t.Fatalf("error = %q, want unknown subcommand error", err)
	}
}

func TestRunReportDispatcherRoutesToTrialBalance(t *testing.T) {
	application, _, stdout := setupReportTestApp(t)
	ctx := context.Background()

	if err := application.runReport(ctx, []string{"trial-balance"}); err != nil {
		t.Fatalf("run report trial-balance: %v", err)
	}

	if !strings.Contains(stdout.String(), "Trial Balance") {
		t.Fatalf("output missing Trial Balance header: %q", stdout.String())
	}
}
