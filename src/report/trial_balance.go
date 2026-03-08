package report

import (
	"context"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// TrialBalanceRow represents a single account row in the trial balance report.
type TrialBalanceRow struct {
	AccountCode  string
	AccountName  string
	AccountType  ledger.AccountType
	TotalDebits  int64
	TotalCredits int64
	Balance      int64
}

// TrialBalanceReport contains the full trial balance with per-account totals
// and grand totals. Balanced is true when total debits equal total credits.
type TrialBalanceReport struct {
	Accounts     []TrialBalanceRow
	TotalDebits  int64
	TotalCredits int64
	Balanced     bool
}

// GetTrialBalance queries all accounts that have at least one journal line from
// a posted journal entry and returns their debit/credit totals. Reversed entries
// are excluded because they have their own reversal entry that cancels them out.
// The balance for each account is computed based on its normal balance:
//   - Asset and Expense accounts: balance = total_debits - total_credits
//   - Liability, Equity, and Income accounts: balance = total_credits - total_debits
func GetTrialBalance(ctx context.Context, databasePath string) (TrialBalanceReport, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return TrialBalanceReport{}, err
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return TrialBalanceReport{}, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT a.code, a.name, a.type,
		       COALESCE(SUM(jl.debit_amount), 0),
		       COALESCE(SUM(jl.credit_amount), 0)
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN accounts a ON a.id = jl.account_id
		WHERE je.status = 'posted'
		GROUP BY a.id
		ORDER BY a.code, a.id
	`)
	if err != nil {
		return TrialBalanceReport{}, fmt.Errorf("query trial balance: %w", err)
	}
	defer rows.Close()

	var report TrialBalanceReport
	for rows.Next() {
		var row TrialBalanceRow
		var accountType string

		if err := rows.Scan(&row.AccountCode, &row.AccountName, &accountType, &row.TotalDebits, &row.TotalCredits); err != nil {
			return TrialBalanceReport{}, fmt.Errorf("scan trial balance row: %w", err)
		}

		row.AccountType = ledger.AccountType(accountType)
		if !row.AccountType.Valid() {
			return TrialBalanceReport{}, fmt.Errorf("scan trial balance row: invalid account type %q", accountType)
		}

		switch row.AccountType {
		case ledger.AccountTypeAsset, ledger.AccountTypeExpense:
			row.Balance = row.TotalDebits - row.TotalCredits
		case ledger.AccountTypeLiability, ledger.AccountTypeEquity, ledger.AccountTypeIncome:
			row.Balance = row.TotalCredits - row.TotalDebits
		}

		report.TotalDebits += row.TotalDebits
		report.TotalCredits += row.TotalCredits
		report.Accounts = append(report.Accounts, row)
	}

	if err := rows.Err(); err != nil {
		return TrialBalanceReport{}, fmt.Errorf("iterate trial balance rows: %w", err)
	}

	report.Balanced = report.TotalDebits == report.TotalCredits

	return report, nil
}
