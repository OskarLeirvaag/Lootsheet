package repo

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

func initTestDatabase(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	return databasePath
}

func TestGetTrialBalanceWithPostedEntries(t *testing.T) {
	databasePath := initTestDatabase(t)
	ctx := context.Background()

	// Post an entry: Dr Adventuring Supplies 5000:50, Cr Party Cash 1000:50
	_, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Buy supplies",
		Lines: []service.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 50, Memo: "Rations"},
			{AccountCode: "1000", CreditAmount: 50},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	// Post another entry: Dr Party Cash 1000:200, Cr Quest Income 4000:200
	_, err = PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-02",
		Description: "Quest reward",
		Lines: []service.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 200},
			{AccountCode: "4000", CreditAmount: 200},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	if !report.Balanced {
		t.Fatalf("expected trial balance to be balanced, got debits=%d credits=%d", report.TotalDebits, report.TotalCredits)
	}

	if report.TotalDebits != 250 || report.TotalCredits != 250 {
		t.Fatalf("total debits=%d credits=%d, want 250/250", report.TotalDebits, report.TotalCredits)
	}

	if len(report.Accounts) != 3 {
		t.Fatalf("account count = %d, want 3", len(report.Accounts))
	}

	// Verify per-account totals (ordered by code).
	accountsByCode := map[string]TrialBalanceRow{}
	for _, row := range report.Accounts {
		accountsByCode[row.AccountCode] = row
	}

	// Party Cash (asset): debits=200, credits=50, balance=150
	cash := accountsByCode["1000"]
	if cash.TotalDebits != 200 || cash.TotalCredits != 50 || cash.Balance != 150 {
		t.Fatalf("Party Cash = debits:%d credits:%d balance:%d, want 200/50/150", cash.TotalDebits, cash.TotalCredits, cash.Balance)
	}
	if cash.AccountType != service.AccountTypeAsset {
		t.Fatalf("Party Cash type = %q, want asset", cash.AccountType)
	}

	// Quest Income (income): debits=0, credits=200, balance=200
	income := accountsByCode["4000"]
	if income.TotalDebits != 0 || income.TotalCredits != 200 || income.Balance != 200 {
		t.Fatalf("Quest Income = debits:%d credits:%d balance:%d, want 0/200/200", income.TotalDebits, income.TotalCredits, income.Balance)
	}
	if income.AccountType != service.AccountTypeIncome {
		t.Fatalf("Quest Income type = %q, want income", income.AccountType)
	}

	// Adventuring Supplies (expense): debits=50, credits=0, balance=50
	supplies := accountsByCode["5000"]
	if supplies.TotalDebits != 50 || supplies.TotalCredits != 0 || supplies.Balance != 50 {
		t.Fatalf("Adventuring Supplies = debits:%d credits:%d balance:%d, want 50/0/50", supplies.TotalDebits, supplies.TotalCredits, supplies.Balance)
	}
	if supplies.AccountType != service.AccountTypeExpense {
		t.Fatalf("Adventuring Supplies type = %q, want expense", supplies.AccountType)
	}
}

func TestGetTrialBalanceEmptyLedger(t *testing.T) {
	databasePath := initTestDatabase(t)
	ctx := context.Background()

	report, err := GetTrialBalance(ctx, databasePath)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	if len(report.Accounts) != 0 {
		t.Fatalf("account count = %d, want 0 for empty ledger", len(report.Accounts))
	}

	if report.TotalDebits != 0 || report.TotalCredits != 0 {
		t.Fatalf("totals = %d/%d, want 0/0 for empty ledger", report.TotalDebits, report.TotalCredits)
	}

	if !report.Balanced {
		t.Fatal("expected empty trial balance to be balanced")
	}
}

func TestGetTrialBalanceExcludesAccountsWithNoPostedTransactions(t *testing.T) {
	databasePath := initTestDatabase(t)
	ctx := context.Background()

	// Post a single entry touching only two accounts.
	_, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Simple entry",
		Lines: []service.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 100},
			{AccountCode: "1000", CreditAmount: 100},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	// Only the 2 accounts with transactions should appear, not all 16 seed accounts.
	if len(report.Accounts) != 2 {
		t.Fatalf("account count = %d, want 2", len(report.Accounts))
	}
}

func TestGetTrialBalanceReversedEntriesDoNotDoubleCount(t *testing.T) {
	databasePath := initTestDatabase(t)
	ctx := context.Background()

	// Post an entry: Dr 5000:75, Cr 1000:75
	posted, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Buy supplies",
		Lines: []service.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 75},
			{AccountCode: "1000", CreditAmount: 75},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	// Also post a second entry that stays posted: Dr 1000:200, Cr 4000:200
	_, err = PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Quest reward",
		Lines: []service.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 200},
			{AccountCode: "4000", CreditAmount: 200},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	// Reverse the first entry. This creates a reversal (posted) and marks
	// the original as 'reversed'. The reversed original is excluded from
	// the trial balance; only the reversal entry (posted) is included.
	_, err = ReverseJournalEntry(ctx, databasePath, posted.ID, "2026-03-02", "")
	if err != nil {
		t.Fatalf("reverse journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	if !report.Balanced {
		t.Fatalf("expected balanced trial balance, got debits=%d credits=%d", report.TotalDebits, report.TotalCredits)
	}

	// The trial balance should include lines from:
	//   - Entry 2 (posted): Dr 1000:200, Cr 4000:200
	//   - Reversal (posted): Dr 1000:75, Cr 5000:75
	// But NOT the original reversed entry.
	// Total debits = 200 + 75 = 275, total credits = 200 + 75 = 275
	if report.TotalDebits != 275 || report.TotalCredits != 275 {
		t.Fatalf("totals = debits:%d credits:%d, want 275/275", report.TotalDebits, report.TotalCredits)
	}

	accountsByCode := map[string]TrialBalanceRow{}
	for _, row := range report.Accounts {
		accountsByCode[row.AccountCode] = row
	}

	// Party Cash 1000 (asset): debits=200+75=275, credits=0
	cash := accountsByCode["1000"]
	if cash.TotalDebits != 275 || cash.TotalCredits != 0 {
		t.Fatalf("Party Cash = debits:%d credits:%d, want 275/0", cash.TotalDebits, cash.TotalCredits)
	}

	// Quest Income 4000 (income): debits=0, credits=200
	income := accountsByCode["4000"]
	if income.TotalDebits != 0 || income.TotalCredits != 200 {
		t.Fatalf("Quest Income = debits:%d credits:%d, want 0/200", income.TotalDebits, income.TotalCredits)
	}

	// Adventuring Supplies 5000 (expense): debits=0, credits=75 (from reversal only)
	supplies := accountsByCode["5000"]
	if supplies.TotalDebits != 0 || supplies.TotalCredits != 75 {
		t.Fatalf("Adventuring Supplies = debits:%d credits:%d, want 0/75", supplies.TotalDebits, supplies.TotalCredits)
	}
}

func TestGetTrialBalanceNormalBalanceDirections(t *testing.T) {
	databasePath := initTestDatabase(t)
	ctx := context.Background()

	// Create a liability account for testing.
	_, err := CreateAccount(ctx, databasePath, "2100", "Test Liability", service.AccountTypeLiability)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	// Create an equity account for testing.
	_, err = CreateAccount(ctx, databasePath, "3100", "Test Equity", service.AccountTypeEquity)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	// Dr Party Cash 1000:500, Cr Test Liability 2100:300, Cr Test Equity 3100:200
	_, err = PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Test normal balances",
		Lines: []service.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 500},
			{AccountCode: "2100", CreditAmount: 300},
			{AccountCode: "3100", CreditAmount: 200},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	accountsByCode := map[string]TrialBalanceRow{}
	for _, row := range report.Accounts {
		accountsByCode[row.AccountCode] = row
	}

	// Asset: normal debit balance = debits - credits
	cash := accountsByCode["1000"]
	if cash.Balance != 500 {
		t.Fatalf("Party Cash balance = %d, want 500", cash.Balance)
	}

	// Liability: normal credit balance = credits - debits
	liability := accountsByCode["2100"]
	if liability.Balance != 300 {
		t.Fatalf("Test Liability balance = %d, want 300", liability.Balance)
	}

	// Equity: normal credit balance = credits - debits
	equity := accountsByCode["3100"]
	if equity.Balance != 200 {
		t.Fatalf("Test Equity balance = %d, want 200", equity.Balance)
	}
}
