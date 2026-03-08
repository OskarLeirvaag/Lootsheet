package repo

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

func initLedgerTestDatabase(t *testing.T) string {
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

func TestGetAccountLedgerWithTransactions(t *testing.T) {
	databasePath := initLedgerTestDatabase(t)
	ctx := context.Background()

	// Post two entries touching account 1000 (Party Cash, asset).
	// Entry 1: Dr Arrows & Ammunition 5100:25, Cr Party Cash 1000:25
	if _, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []service.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25, Memo: "Quiver refill"},
			{AccountCode: "1000", CreditAmount: 25},
		},
	}); err != nil {
		t.Fatalf("post entry 1: %v", err)
	}

	// Entry 2: Dr Party Cash 1000:100, Cr Quest Income 4000:100
	if _, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Quest reward earned",
		Lines: []service.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 100, Memo: "Goblin bounty"},
			{AccountCode: "4000", CreditAmount: 100},
		},
	}); err != nil {
		t.Fatalf("post entry 2: %v", err)
	}

	report, err := GetAccountLedger(ctx, databasePath, "1000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if report.AccountCode != "1000" {
		t.Fatalf("account code = %q, want 1000", report.AccountCode)
	}

	if report.AccountName != "Party Cash" {
		t.Fatalf("account name = %q, want Party Cash", report.AccountName)
	}

	if report.AccountType != service.AccountTypeAsset {
		t.Fatalf("account type = %q, want asset", report.AccountType)
	}

	if len(report.Entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(report.Entries))
	}

	// First entry: credit 25 to asset account -> balance -25
	e1 := report.Entries[0]
	if e1.EntryNumber != 1 {
		t.Fatalf("entry 1 number = %d, want 1", e1.EntryNumber)
	}
	if e1.CreditAmount != 25 || e1.DebitAmount != 0 {
		t.Fatalf("entry 1 amounts = debit:%d credit:%d, want debit:0 credit:25", e1.DebitAmount, e1.CreditAmount)
	}
	if e1.RunningBalance != -25 {
		t.Fatalf("entry 1 running balance = %d, want -25", e1.RunningBalance)
	}
	if e1.Description != "Restock arrows" {
		t.Fatalf("entry 1 description = %q, want Restock arrows", e1.Description)
	}

	// Second entry: debit 100 to asset account -> balance 75
	e2 := report.Entries[1]
	if e2.EntryNumber != 2 {
		t.Fatalf("entry 2 number = %d, want 2", e2.EntryNumber)
	}
	if e2.DebitAmount != 100 || e2.CreditAmount != 0 {
		t.Fatalf("entry 2 amounts = debit:%d credit:%d, want debit:100 credit:0", e2.DebitAmount, e2.CreditAmount)
	}
	if e2.RunningBalance != 75 {
		t.Fatalf("entry 2 running balance = %d, want 75", e2.RunningBalance)
	}
	if e2.Memo != "Goblin bounty" {
		t.Fatalf("entry 2 memo = %q, want Goblin bounty", e2.Memo)
	}

	if report.Balance != 75 {
		t.Fatalf("final balance = %d, want 75", report.Balance)
	}
}

func TestGetAccountLedgerReversedEntriesExcluded(t *testing.T) {
	databasePath := initLedgerTestDatabase(t)
	ctx := context.Background()

	// Post an entry.
	posted, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []service.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post entry: %v", err)
	}

	// Reverse it.
	if _, err := ReverseJournalEntry(ctx, databasePath, posted.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("reverse entry: %v", err)
	}

	// The ledger should show 2 entries: the reversal entry (posted) but not
	// the original (now status=reversed). The reversal entry has swapped amounts.
	report, err := GetAccountLedger(ctx, databasePath, "1000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	// Only the reversal entry should appear (the original is status=reversed).
	if len(report.Entries) != 1 {
		t.Fatalf("entry count = %d, want 1 (only reversal)", len(report.Entries))
	}

	// The reversal entry swaps amounts: original credit 25 becomes debit 25.
	e := report.Entries[0]
	if e.EntryNumber != 2 {
		t.Fatalf("reversal entry number = %d, want 2", e.EntryNumber)
	}
	if e.DebitAmount != 25 || e.CreditAmount != 0 {
		t.Fatalf("reversal amounts = debit:%d credit:%d, want debit:25 credit:0", e.DebitAmount, e.CreditAmount)
	}

	// Asset account: debit increases balance.
	if report.Balance != 25 {
		t.Fatalf("final balance = %d, want 25", report.Balance)
	}
}

func TestGetAccountLedgerIncomeAccountCreditNormal(t *testing.T) {
	databasePath := initLedgerTestDatabase(t)
	ctx := context.Background()

	// Post an entry: Dr Party Cash 1000:100, Cr Quest Income 4000:100
	if _, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Quest reward",
		Lines: []service.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 100},
			{AccountCode: "4000", CreditAmount: 100},
		},
	}); err != nil {
		t.Fatalf("post entry: %v", err)
	}

	report, err := GetAccountLedger(ctx, databasePath, "4000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if report.AccountType != service.AccountTypeIncome {
		t.Fatalf("account type = %q, want income", report.AccountType)
	}

	if len(report.Entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(report.Entries))
	}

	// Income is credit-normal: credit 100 -> balance +100
	if report.Balance != 100 {
		t.Fatalf("balance = %d, want 100 (credit-normal)", report.Balance)
	}

	if report.Entries[0].RunningBalance != 100 {
		t.Fatalf("running balance = %d, want 100", report.Entries[0].RunningBalance)
	}
}

func TestGetAccountLedgerEmptyAccount(t *testing.T) {
	databasePath := initLedgerTestDatabase(t)
	ctx := context.Background()

	report, err := GetAccountLedger(ctx, databasePath, "1000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if report.AccountCode != "1000" {
		t.Fatalf("account code = %q, want 1000", report.AccountCode)
	}

	if len(report.Entries) != 0 {
		t.Fatalf("entry count = %d, want 0", len(report.Entries))
	}

	if report.Balance != 0 {
		t.Fatalf("balance = %d, want 0", report.Balance)
	}
}

func TestGetAccountLedgerNonexistentAccount(t *testing.T) {
	databasePath := initLedgerTestDatabase(t)
	ctx := context.Background()

	_, err := GetAccountLedger(ctx, databasePath, "9999")
	if err == nil {
		t.Fatal("expected error for nonexistent account")
	}
}

func TestGetAccountLedgerExpenseAccountDebitNormal(t *testing.T) {
	databasePath := initLedgerTestDatabase(t)
	ctx := context.Background()

	// Dr Arrows & Ammunition 5100:50, Cr Party Cash 1000:50
	if _, err := PostJournalEntry(ctx, databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Buy arrows",
		Lines: []service.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 50},
			{AccountCode: "1000", CreditAmount: 50},
		},
	}); err != nil {
		t.Fatalf("post entry: %v", err)
	}

	report, err := GetAccountLedger(ctx, databasePath, "5100")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if report.AccountType != service.AccountTypeExpense {
		t.Fatalf("account type = %q, want expense", report.AccountType)
	}

	// Expense is debit-normal: debit 50 -> balance +50
	if report.Balance != 50 {
		t.Fatalf("balance = %d, want 50 (debit-normal)", report.Balance)
	}
}
