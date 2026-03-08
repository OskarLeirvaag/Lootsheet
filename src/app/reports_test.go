package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
)

func setupReportTestEnv(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	databasePath := filepath.Join(dataDir, "ledger.db")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := ledger.EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize database: %v", err)
	}

	return databasePath
}

func TestRunTrialBalanceShowsBalancedOutput(t *testing.T) {
	databasePath := setupReportTestEnv(t)
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

	var stdout bytes.Buffer
	if err := Run(ctx, []string{"report", "trial-balance"}, &stdout); err != nil {
		t.Fatalf("run trial balance: %v", err)
	}

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
	_ = setupReportTestEnv(t)
	ctx := context.Background()

	var stdout bytes.Buffer
	if err := Run(ctx, []string{"report", "trial-balance"}, &stdout); err != nil {
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
	databasePath := setupReportTestEnv(t)
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

	var stdout bytes.Buffer
	if err := Run(ctx, []string{"report", "trial-balance"}, &stdout); err != nil {
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

func TestRunReportMissingSubcommand(t *testing.T) {
	_ = setupReportTestEnv(t)
	ctx := context.Background()

	var stdout bytes.Buffer
	err := Run(ctx, []string{"report"}, &stdout)
	if err != nil {
		t.Fatalf("expected no error for report with no subcommand (shows help): %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "report") {
		t.Fatalf("expected report help output, got: %q", output)
	}
}

func TestRunReportUnknownSubcommand(t *testing.T) {
	_ = setupReportTestEnv(t)
	ctx := context.Background()

	var stdout bytes.Buffer
	err := Run(ctx, []string{"report", "bogus"}, &stdout)

	// Cobra surfaces an "unknown command" error for unrecognized subcommands.
	if err == nil {
		// If Cobra falls through to the parent RunE (shows help), that is
		// acceptable; verify the help text was printed instead.
		output := stdout.String()
		if !strings.Contains(output, "report") {
			t.Fatalf("expected either error or help output, got: %q", output)
		}
	}
}

func TestRunReportRoutesToTrialBalance(t *testing.T) {
	_ = setupReportTestEnv(t)
	ctx := context.Background()

	var stdout bytes.Buffer
	if err := Run(ctx, []string{"report", "trial-balance"}, &stdout); err != nil {
		t.Fatalf("run report trial-balance: %v", err)
	}

	if !strings.Contains(stdout.String(), "Trial Balance") {
		t.Fatalf("output missing Trial Balance header: %q", stdout.String())
	}
}

func TestRunPromisedQuestsReport(t *testing.T) {
	databasePath := setupReportTestEnv(t)
	ctx := context.Background()

	if _, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Pending Promise",
		Patron:             "Lady Mirelle",
		PromisedBaseReward: 400,
		PartialAdvance:     50,
		BonusConditions:    "Extra if the prisoners survive",
		Status:             "offered",
	}); err != nil {
		t.Fatalf("create quest: %v", err)
	}

	var stdout bytes.Buffer
	if err := Run(ctx, []string{"report", "promised-quests"}, &stdout); err != nil {
		t.Fatalf("run promised quests report: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Promised But Unearned Quests") {
		t.Fatalf("output missing title: %q", output)
	}
	if !strings.Contains(output, "Pending Promise") {
		t.Fatalf("output missing quest title: %q", output)
	}
	if !strings.Contains(output, "Lady Mirelle") {
		t.Fatalf("output missing patron: %q", output)
	}
	if !strings.Contains(output, "4 GP") {
		t.Fatalf("output missing promised reward: %q", output)
	}
	if !strings.Contains(output, "5 SP") {
		t.Fatalf("output missing advance: %q", output)
	}
}

func TestRunWriteOffCandidatesReport(t *testing.T) {
	databasePath := setupReportTestEnv(t)
	ctx := context.Background()

	createdQuest, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Stale Receivable",
		Patron:             "Harbormaster Tov",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-01-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := quest.CompleteQuest(ctx, databasePath, createdQuest.ID, "2026-01-02"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	if _, err := quest.CollectQuestPayment(ctx, databasePath, quest.CollectQuestPaymentInput{
		QuestID:     createdQuest.ID,
		Amount:      200,
		Date:        "2026-01-05",
		Description: "Harbormaster paid part in advance of winter taxes",
	}); err != nil {
		t.Fatalf("collect quest payment: %v", err)
	}

	var stdout bytes.Buffer
	if err := Run(ctx, []string{"report", "writeoff-candidates", "--as-of", "2026-03-15", "--min-age-days", "30"}, &stdout); err != nil {
		t.Fatalf("run write-off candidates report: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Write-Off Candidates") {
		t.Fatalf("output missing title: %q", output)
	}
	if !strings.Contains(output, "Stale Receivable") {
		t.Fatalf("output missing quest title: %q", output)
	}
	if !strings.Contains(output, "Harbormaster To") {
		t.Fatalf("output missing patron: %q", output)
	}
	if !strings.Contains(output, "3 GP") {
		t.Fatalf("output missing outstanding amount: %q", output)
	}
	if !strings.Contains(output, "72") {
		t.Fatalf("output missing age days: %q", output)
	}
}
