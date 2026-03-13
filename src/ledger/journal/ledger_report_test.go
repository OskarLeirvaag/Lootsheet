package journal

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestGetAccountLedgerWithTransactions(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	if _, err := PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25, Memo: "Quiver refill"},
			{AccountCode: "1000", CreditAmount: 25},
		},
	}); err != nil {
		t.Fatalf("post entry 1: %v", err)
	}

	if _, err := PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Quest reward earned",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 100, Memo: "Goblin bounty"},
			{AccountCode: "4000", CreditAmount: 100},
		},
	}); err != nil {
		t.Fatalf("post entry 2: %v", err)
	}

	report, err := GetAccountLedger(ctx, databasePath, campaignID, "1000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if report.AccountCode != "1000" {
		t.Fatalf("account code = %q, want 1000", report.AccountCode)
	}

	if report.AccountName != "Party Cash" {
		t.Fatalf("account name = %q, want Party Cash", report.AccountName)
	}

	if report.AccountType != ledger.AccountTypeAsset {
		t.Fatalf("account type = %q, want asset", report.AccountType)
	}

	if len(report.Entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(report.Entries))
	}

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
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	posted, err := PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post entry: %v", err)
	}

	if _, err := ReverseJournalEntry(ctx, databasePath, campaignID, posted.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("reverse entry: %v", err)
	}

	report, err := GetAccountLedger(ctx, databasePath, campaignID, "1000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if len(report.Entries) != 1 {
		t.Fatalf("entry count = %d, want 1 (only reversal)", len(report.Entries))
	}

	e := report.Entries[0]
	if e.EntryNumber != 2 {
		t.Fatalf("reversal entry number = %d, want 2", e.EntryNumber)
	}
	if e.DebitAmount != 25 || e.CreditAmount != 0 {
		t.Fatalf("reversal amounts = debit:%d credit:%d, want debit:25 credit:0", e.DebitAmount, e.CreditAmount)
	}

	if report.Balance != 25 {
		t.Fatalf("final balance = %d, want 25", report.Balance)
	}
}

func TestGetAccountLedgerIncomeAccountCreditNormal(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	if _, err := PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Quest reward",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 100},
			{AccountCode: "4000", CreditAmount: 100},
		},
	}); err != nil {
		t.Fatalf("post entry: %v", err)
	}

	report, err := GetAccountLedger(ctx, databasePath, campaignID, "4000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if report.AccountType != ledger.AccountTypeIncome {
		t.Fatalf("account type = %q, want income", report.AccountType)
	}

	if report.Balance != 100 {
		t.Fatalf("balance = %d, want 100 (credit-normal)", report.Balance)
	}
}

func TestGetAccountLedgerEmptyAccount(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	report, err := GetAccountLedger(ctx, databasePath, campaignID, "1000")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if len(report.Entries) != 0 {
		t.Fatalf("entry count = %d, want 0", len(report.Entries))
	}

	if report.Balance != 0 {
		t.Fatalf("balance = %d, want 0", report.Balance)
	}
}

func TestGetAccountLedgerNonexistentAccount(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	_, err := GetAccountLedger(context.Background(), databasePath, campaignID, "9999")
	if err == nil {
		t.Fatal("expected error for nonexistent account")
	}
}

func TestGetAccountLedgerExpenseAccountDebitNormal(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	if _, err := PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Buy arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 50},
			{AccountCode: "1000", CreditAmount: 50},
		},
	}); err != nil {
		t.Fatalf("post entry: %v", err)
	}

	report, err := GetAccountLedger(ctx, databasePath, campaignID, "5100")
	if err != nil {
		t.Fatalf("get account ledger: %v", err)
	}

	if report.AccountType != ledger.AccountTypeExpense {
		t.Fatalf("account type = %q, want expense", report.AccountType)
	}

	if report.Balance != 50 {
		t.Fatalf("balance = %d, want 50 (debit-normal)", report.Balance)
	}
}
