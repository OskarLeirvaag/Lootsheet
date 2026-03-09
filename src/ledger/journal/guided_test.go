package journal

import (
	"context"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

func TestPostExpenseEntryCreatesBalancedTwoLineEntry(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	result, err := PostExpenseEntry(context.Background(), databasePath, &ExpenseEntryInput{
		Date:               "2026-03-08",
		Description:        "Restock arrows",
		ExpenseAccountCode: "5100",
		FundingAccountCode: "1000",
		Amount:             25,
		Memo:               "Quiver refill",
	})
	if err != nil {
		t.Fatalf("post expense entry: %v", err)
	}

	if result.EntryNumber != 1 {
		t.Fatalf("entry number = %d, want 1", result.EntryNumber)
	}

	lineRows := strings.Split(strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, `
		SELECT a.code || '|' || debit_amount || '|' || credit_amount || '|' || COALESCE(memo, '')
		FROM journal_lines jl
		JOIN accounts a ON a.id = jl.account_id
		ORDER BY line_number
	`)), "\n")
	if len(lineRows) != 2 {
		t.Fatalf("line count = %d, want 2", len(lineRows))
	}

	if lineRows[0] != "5100|25|0|Quiver refill" {
		t.Fatalf("first line = %q, want expense debit", lineRows[0])
	}
	if lineRows[1] != "1000|0|25|" {
		t.Fatalf("second line = %q, want funding credit", lineRows[1])
	}
}

func TestPostExpenseEntryAllowsLiabilityFunding(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ledger.RunSQLiteScriptForTest(t, databasePath, `
		INSERT INTO accounts (id, code, name, type, active)
		VALUES ('inn_tab', '2100', 'Innkeeper Tab', 'liability', 1);
	`)

	result, err := PostExpenseEntry(context.Background(), databasePath, &ExpenseEntryInput{
		Date:               "2026-03-08",
		Description:        "Charge inn stay",
		ExpenseAccountCode: "5300",
		FundingAccountCode: "2100",
		Amount:             800,
	})
	if err != nil {
		t.Fatalf("post expense entry with liability funding: %v", err)
	}

	if result.EntryNumber != 1 {
		t.Fatalf("entry number = %d, want 1", result.EntryNumber)
	}
}

func TestPostExpenseEntryRejectsWrongAccountTypes(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	_, err := PostExpenseEntry(context.Background(), databasePath, &ExpenseEntryInput{
		Date:               "2026-03-08",
		Description:        "Broken expense",
		ExpenseAccountCode: "1000",
		FundingAccountCode: "4000",
		Amount:             25,
	})
	if err == nil {
		t.Fatal("expected guided expense entry to fail")
	}

	if !strings.Contains(err.Error(), `account code "1000" must be an active expense account`) {
		t.Fatalf("error = %q, want expense account type failure", err)
	}
}

func TestPostExpenseEntryRejectsNonPositiveAmount(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	_, err := PostExpenseEntry(context.Background(), databasePath, &ExpenseEntryInput{
		Date:               "2026-03-08",
		Description:        "Broken expense",
		ExpenseAccountCode: "5100",
		FundingAccountCode: "1000",
		Amount:             0,
	})
	if err == nil {
		t.Fatal("expected non-positive guided expense amount to fail")
	}

	if !strings.Contains(err.Error(), "journal entry amount must be positive") {
		t.Fatalf("error = %q, want positive amount failure", err)
	}
}

func TestPostIncomeEntryCreatesBalancedTwoLineEntry(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	result, err := PostIncomeEntry(context.Background(), databasePath, &IncomeEntryInput{
		Date:               "2026-03-08",
		Description:        "Quest bounty received",
		IncomeAccountCode:  "4000",
		DepositAccountCode: "1000",
		Amount:             100,
		Memo:               "Mayor payout",
	})
	if err != nil {
		t.Fatalf("post income entry: %v", err)
	}

	if result.EntryNumber != 1 {
		t.Fatalf("entry number = %d, want 1", result.EntryNumber)
	}

	lineRows := strings.Split(strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, `
		SELECT a.code || '|' || debit_amount || '|' || credit_amount || '|' || COALESCE(memo, '')
		FROM journal_lines jl
		JOIN accounts a ON a.id = jl.account_id
		ORDER BY line_number
	`)), "\n")
	if len(lineRows) != 2 {
		t.Fatalf("line count = %d, want 2", len(lineRows))
	}

	if lineRows[0] != "1000|100|0|" {
		t.Fatalf("first line = %q, want deposit debit", lineRows[0])
	}
	if lineRows[1] != "4000|0|100|Mayor payout" {
		t.Fatalf("second line = %q, want income credit", lineRows[1])
	}
}

func TestPostIncomeEntryRejectsWrongAccountTypes(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	_, err := PostIncomeEntry(context.Background(), databasePath, &IncomeEntryInput{
		Date:               "2026-03-08",
		Description:        "Broken income",
		IncomeAccountCode:  "5100",
		DepositAccountCode: "1000",
		Amount:             100,
	})
	if err == nil {
		t.Fatal("expected guided income entry to fail")
	}

	if !strings.Contains(err.Error(), `account code "5100" must be an active income account`) {
		t.Fatalf("error = %q, want income account type failure", err)
	}
}

func TestPostIncomeEntryRejectsInactiveAccount(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ledger.RunSQLiteScriptForTest(t, databasePath, `UPDATE accounts SET active = 0 WHERE code = '1000';`)

	_, err := PostIncomeEntry(context.Background(), databasePath, &IncomeEntryInput{
		Date:               "2026-03-08",
		Description:        "Broken income",
		IncomeAccountCode:  "4000",
		DepositAccountCode: "1000",
		Amount:             100,
	})
	if err == nil {
		t.Fatal("expected guided income entry with inactive account to fail")
	}

	if !strings.Contains(err.Error(), `account code "1000" is inactive`) {
		t.Fatalf("error = %q, want inactive account failure", err)
	}
}
