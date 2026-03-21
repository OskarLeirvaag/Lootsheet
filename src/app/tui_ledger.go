package app

import (
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

const (
	ledgerMaxEntries = 50
	ledgerDescWidth  = 20
)

func summarizeLedgerSection(tb report.TrialBalanceReport, accountCount int) []string {
	status := "Balanced"
	if !tb.Balanced {
		status = "UNBALANCED"
	}
	return []string{
		fmt.Sprintf("Accounts: %d with postings / %d total", len(tb.Accounts), accountCount),
		fmt.Sprintf("Total debits: %s  Total credits: %s", currency.FormatAmount(tb.TotalDebits), currency.FormatAmount(tb.TotalCredits)),
		fmt.Sprintf("Status: %s", status),
	}
}

func buildLedgerItems(accounts []ledger.AccountRecord, ledgers map[string]journal.AccountLedgerReport, balances map[string]report.TrialBalanceRow) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(accounts))
	for _, record := range accounts {
		bal := balances[record.Code]

		detailLines := []string{
			"Account: " + record.Code + " — " + record.Name,
			"Type: " + string(record.Type),
			"",
			fmt.Sprintf("Total debits:  %s", currency.FormatAmount(bal.TotalDebits)),
			fmt.Sprintf("Total credits: %s", currency.FormatAmount(bal.TotalCredits)),
			fmt.Sprintf("Net balance:   %s", currency.FormatAmount(bal.Balance)),
		}

		if rpt, ok := ledgers[record.Code]; ok && len(rpt.Entries) > 0 {
			detailLines = append(detailLines, "", "\u2500\u2500 Postings \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500")
			entries := rpt.Entries
			if len(entries) > ledgerMaxEntries {
				entries = entries[len(entries)-ledgerMaxEntries:]
			}
			for _, e := range entries {
				amount := formatLedgerAmount(e.DebitAmount, e.CreditAmount)
				detailLines = append(detailLines,
					fmt.Sprintf("#%-3d  %-10s  %-20s  %8s  %s",
						e.EntryNumber, e.EntryDate, truncate(e.Description, ledgerDescWidth), amount, currency.FormatAmount(e.RunningBalance)))
			}
		}

		items = append(items, render.ListItemData{
			Key:         record.Code,
			Row:         fmt.Sprintf("%-4s %-9s %10s %10s %10s  %s", record.Code, string(record.Type), currency.FormatAmount(bal.TotalDebits), currency.FormatAmount(bal.TotalCredits), currency.FormatAmount(bal.Balance), record.Name),
			DetailTitle: "Account " + record.Code,
			DetailLines: detailLines,
		})
	}

	return items
}

// buildSettingsAccountItems builds the CRUD-enabled account list for the Settings tab.
func buildSettingsAccountItems(accounts []ledger.AccountRecord) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(accounts))
	for _, record := range accounts {
		status := "inactive"
		toggleAction := render.ItemActionData{
			Trigger:      render.ActionToggle,
			ID:           tuiCommandAccountActivate,
			Label:        "t activate",
			ConfirmTitle: fmt.Sprintf("Activate account %s?", record.Code),
			ConfirmLines: []string{
				record.Name,
				"This account will become available for new journal entries again.",
			},
		}
		if record.Active {
			status = "active"
			toggleAction = render.ItemActionData{
				Trigger:      render.ActionToggle,
				ID:           tuiCommandAccountDeactivate,
				Label:        "t deactivate",
				ConfirmTitle: fmt.Sprintf("Deactivate account %s?", record.Code),
				ConfirmLines: []string{
					record.Name,
					"Inactive accounts stay in history but cannot be used in new journal entries.",
				},
			}
		}

		renameAction := render.ItemActionData{
			Trigger:     render.ActionEdit,
			ID:          tuiCommandAccountRename,
			Label:       "u rename",
			Mode:        render.ItemActionModeInput,
			InputTitle:  fmt.Sprintf("Rename account %s", record.Code),
			InputPrompt: "New name",
			Placeholder: record.Name,
			InputHelp: []string{
				"Code: " + record.Code + " (immutable)",
				"Type: " + string(record.Type) + " (immutable)",
				"Current name: " + record.Name,
			},
		}

		deleteAction := render.ItemActionData{
			Trigger:      render.ActionDelete,
			ID:           tuiCommandAccountDelete,
			Label:        "d remove",
			ConfirmTitle: fmt.Sprintf("Remove account %s?", record.Code),
			ConfirmLines: []string{
				record.Name,
				"Accounts with postings cannot be removed. Unused accounts will be deleted immediately.",
			},
		}

		items = append(items, render.ListItemData{
			Key:         record.Code,
			Row:         fmt.Sprintf("%-4s %-9s %-8s %s", record.Code, string(record.Type), status, record.Name),
			DetailTitle: "Account " + record.Code,
			DetailLines: []string{
				"Name: " + record.Name,
				"Type: " + string(record.Type),
				"Status: " + status,
				"Code: " + record.Code + " (immutable)",
			},
			Actions: []render.ItemActionData{renameAction, deleteAction, toggleAction},
		})
	}

	return items
}

func formatLedgerAmount(debit, credit int64) string {
	if debit > 0 {
		return "+" + currency.FormatAmount(debit)
	}
	return "-" + currency.FormatAmount(credit)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "\u2026"
}

func buildLedgerViewData(tb report.TrialBalanceReport, accountLedgers map[string]journal.AccountLedgerReport) render.LedgerViewData {
	rows := make([]render.LedgerViewRow, 0, len(tb.Accounts))
	for _, row := range tb.Accounts {
		rows = append(rows, render.LedgerViewRow{
			AccountCode:  row.AccountCode,
			AccountName:  row.AccountName,
			AccountType:  string(row.AccountType),
			TotalDebits:  currency.FormatAmount(row.TotalDebits),
			TotalCredits: currency.FormatAmount(row.TotalCredits),
			Balance:      currency.FormatAmount(row.Balance),
		})
	}

	details := make(map[string]render.LedgerAccountDetail, len(accountLedgers))
	for code, rpt := range accountLedgers {
		entries := make([]render.LedgerDetailEntry, 0, len(rpt.Entries))
		for _, e := range rpt.Entries {
			debit := ""
			credit := ""
			if e.DebitAmount > 0 {
				debit = currency.FormatAmount(e.DebitAmount)
			}
			if e.CreditAmount > 0 {
				credit = currency.FormatAmount(e.CreditAmount)
			}
			entries = append(entries, render.LedgerDetailEntry{
				EntryNumber:    e.EntryNumber,
				Date:           e.EntryDate,
				Description:    e.Description,
				Debit:          debit,
				Credit:         credit,
				RunningBalance: currency.FormatAmount(e.RunningBalance),
			})
		}
		var totalDebits, totalCredits int64
		for _, e := range rpt.Entries {
			totalDebits += e.DebitAmount
			totalCredits += e.CreditAmount
		}
		details[code] = render.LedgerAccountDetail{
			AccountCode:  rpt.AccountCode,
			AccountName:  rpt.AccountName,
			AccountType:  string(rpt.AccountType),
			Entries:      entries,
			TotalDebits:  currency.FormatAmount(totalDebits),
			TotalCredits: currency.FormatAmount(totalCredits),
			Balance:      currency.FormatAmount(rpt.Balance),
		}
	}

	return render.LedgerViewData{
		Rows:          rows,
		AccountDetail: details,
		TotalDebits:   currency.FormatAmount(tb.TotalDebits),
		TotalCredits:  currency.FormatAmount(tb.TotalCredits),
		Balanced:      tb.Balanced,
		Available:     true,
	}
}
