package journal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// AccountLedgerReport contains the full ledger view for a single account,
// including transaction history with running balance.
type AccountLedgerReport struct {
	AccountCode string
	AccountName string
	AccountType ledger.AccountType
	Entries     []LedgerEntry
	Balance     int64
}

// LedgerEntry represents a single line in an account's ledger view.
type LedgerEntry struct {
	EntryNumber    int
	EntryDate      string
	Description    string
	Memo           string
	DebitAmount    int64
	CreditAmount   int64
	RunningBalance int64
}

// GetAccountLedger returns the transaction history for a single account
// with running balance. Only posted entries are included; reversed entries
// are excluded (their reversal entries appear separately).
//
// Running balance follows normal accounting conventions:
//   - Asset and expense accounts: debits increase, credits decrease
//   - Liability, equity, and income accounts: credits increase, debits decrease
func GetAccountLedger(ctx context.Context, databasePath string, accountCode string) (AccountLedgerReport, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (AccountLedgerReport, error) {
		// Resolve the account code to account info.
		var accountID, accountName, accountTypeStr string
		if err := db.QueryRowContext(ctx,
			"SELECT id, name, type FROM accounts WHERE code = ?", accountCode,
		).Scan(&accountID, &accountName, &accountTypeStr); err != nil {
			return AccountLedgerReport{}, fmt.Errorf("account code %q does not exist", accountCode)
		}

		accountType := ledger.AccountType(accountTypeStr)
		if !accountType.Valid() {
			return AccountLedgerReport{}, fmt.Errorf("account %q has invalid type %q", accountCode, accountTypeStr)
		}

		// Query all journal lines for this account from posted entries only.
		rows, err := db.QueryContext(ctx, `
			SELECT
				je.entry_number,
				je.entry_date,
				je.description,
				jl.memo,
				jl.debit_amount,
				jl.credit_amount
			FROM journal_lines jl
			JOIN journal_entries je ON je.id = jl.journal_entry_id
			WHERE jl.account_id = ?
			  AND je.status = 'posted'
			ORDER BY je.entry_date, je.entry_number, jl.line_number
		`, accountID)
		if err != nil {
			return AccountLedgerReport{}, fmt.Errorf("query account ledger: %w", err)
		}
		defer rows.Close()

		// Determine balance direction: asset/expense are debit-normal,
		// liability/equity/income are credit-normal.
		debitNormal := accountType == ledger.AccountTypeAsset || accountType == ledger.AccountTypeExpense

		var entries []LedgerEntry
		var runningBalance int64

		for rows.Next() {
			var e LedgerEntry
			if err := rows.Scan(&e.EntryNumber, &e.EntryDate, &e.Description, &e.Memo, &e.DebitAmount, &e.CreditAmount); err != nil {
				return AccountLedgerReport{}, fmt.Errorf("scan ledger entry: %w", err)
			}

			if debitNormal {
				runningBalance += e.DebitAmount - e.CreditAmount
			} else {
				runningBalance += e.CreditAmount - e.DebitAmount
			}
			e.RunningBalance = runningBalance

			entries = append(entries, e)
		}

		if err := rows.Err(); err != nil {
			return AccountLedgerReport{}, fmt.Errorf("iterate ledger entries: %w", err)
		}

		return AccountLedgerReport{
			AccountCode: accountCode,
			AccountName: accountName,
			AccountType: accountType,
			Entries:     entries,
			Balance:     runningBalance,
		}, nil
	})
}
