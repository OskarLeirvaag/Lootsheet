// Package export provides CSV, Excel, and PDF export functions for ledger reports.
package export

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
)

// accountTypeOrder defines the canonical display ordering for account groups.
var accountTypeOrder = []ledger.AccountType{
	ledger.AccountTypeAsset,
	ledger.AccountTypeLiability,
	ledger.AccountTypeEquity,
	ledger.AccountTypeIncome,
	ledger.AccountTypeExpense,
}

// accountTypeLabel returns the human-readable group label for an account type.
func accountTypeLabel(t ledger.AccountType) string {
	switch t {
	case ledger.AccountTypeAsset:
		return "Assets"
	case ledger.AccountTypeLiability:
		return "Liabilities"
	case ledger.AccountTypeEquity:
		return "Equity"
	case ledger.AccountTypeIncome:
		return "Income"
	case ledger.AccountTypeExpense:
		return "Expenses"
	default:
		return string(t)
	}
}

// groupAccountsByType partitions trial balance rows by account type,
// preserving original order within each group.
func groupAccountsByType(accounts []report.TrialBalanceRow) map[ledger.AccountType][]report.TrialBalanceRow {
	groups := make(map[ledger.AccountType][]report.TrialBalanceRow)
	for _, a := range accounts {
		groups[a.AccountType] = append(groups[a.AccountType], a)
	}
	return groups
}

// WriteTrialBalanceCSV writes the trial balance report as a machine-readable CSV.
// Amounts are raw copper-piece int64 values. Accounts are grouped by type with
// subtotals per group and a grand total at the bottom.
func WriteTrialBalanceCSV(w io.Writer, tb report.TrialBalanceReport, _ map[string]journal.AccountLedgerReport) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Header row.
	if err := cw.Write([]string{"Code", "Name", "Type", "Debits (CP)", "Credits (CP)", "Balance (CP)"}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	groups := groupAccountsByType(tb.Accounts)

	for _, acctType := range accountTypeOrder {
		rows, ok := groups[acctType]
		if !ok {
			continue
		}

		label := accountTypeLabel(acctType)

		// Blank separator row.
		if err := cw.Write([]string{"", "", "", "", "", ""}); err != nil {
			return fmt.Errorf("write CSV blank row: %w", err)
		}

		// Group label row.
		if err := cw.Write([]string{"", fmt.Sprintf("--- %s ---", label), "", "", "", ""}); err != nil {
			return fmt.Errorf("write CSV group label: %w", err)
		}

		var groupDebits, groupCredits, groupBalance int64

		for _, row := range rows {
			if err := cw.Write([]string{
				row.AccountCode,
				row.AccountName,
				string(row.AccountType),
				i64(row.TotalDebits),
				i64(row.TotalCredits),
				i64(row.Balance),
			}); err != nil {
				return fmt.Errorf("write CSV account row: %w", err)
			}
			groupDebits += row.TotalDebits
			groupCredits += row.TotalCredits
			groupBalance += row.Balance
		}

		// Subtotal row.
		if err := cw.Write([]string{
			"", fmt.Sprintf("Subtotal %s", label), "",
			i64(groupDebits), i64(groupCredits), i64(groupBalance),
		}); err != nil {
			return fmt.Errorf("write CSV subtotal row: %w", err)
		}
	}

	// Blank row before grand total.
	if err := cw.Write([]string{"", "", "", "", "", ""}); err != nil {
		return fmt.Errorf("write CSV blank row: %w", err)
	}

	// Grand total row.
	if err := cw.Write([]string{
		"", "Grand Total", "",
		i64(tb.TotalDebits), i64(tb.TotalCredits), "",
	}); err != nil {
		return fmt.Errorf("write CSV grand total: %w", err)
	}

	// Balanced status row.
	if err := cw.Write([]string{
		"", fmt.Sprintf("Balanced: %t", tb.Balanced), "", "", "", "",
	}); err != nil {
		return fmt.Errorf("write CSV balanced row: %w", err)
	}

	cw.Flush()
	return cw.Error()
}

// WriteAccountLedgerCSV writes an individual account's ledger entries as CSV.
// Amounts are raw copper-piece int64 values.
func WriteAccountLedgerCSV(w io.Writer, rpt *journal.AccountLedgerReport) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Header row.
	if err := cw.Write([]string{
		"Entry#", "Date", "Description", "Memo",
		"Debit (CP)", "Credit (CP)", "Running Balance (CP)",
	}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, e := range rpt.Entries {
		if err := cw.Write([]string{
			fmt.Sprintf("%d", e.EntryNumber),
			e.EntryDate,
			e.Description,
			e.Memo,
			i64(e.DebitAmount),
			i64(e.CreditAmount),
			i64(e.RunningBalance),
		}); err != nil {
			return fmt.Errorf("write CSV entry row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}

// i64 formats an int64 as a decimal string.
func i64(v int64) string {
	return fmt.Sprintf("%d", v)
}
