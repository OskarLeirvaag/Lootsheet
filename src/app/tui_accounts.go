package app

import (
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func summarizeAccounts(accounts []ledger.AccountRecord) []string {
	counts := map[ledger.AccountType]int{
		ledger.AccountTypeAsset:     0,
		ledger.AccountTypeLiability: 0,
		ledger.AccountTypeEquity:    0,
		ledger.AccountTypeIncome:    0,
		ledger.AccountTypeExpense:   0,
	}

	active := 0
	for _, record := range accounts {
		counts[record.Type]++
		if record.Active {
			active++
		}
	}

	return []string{
		fmt.Sprintf("Accounts: %d total", len(accounts)),
		fmt.Sprintf("Active: %d  Inactive: %d", active, len(accounts)-active),
		fmt.Sprintf("Assets: %d  Liabilities: %d", counts[ledger.AccountTypeAsset], counts[ledger.AccountTypeLiability]),
		fmt.Sprintf("Equity: %d  Income: %d  Expenses: %d", counts[ledger.AccountTypeEquity], counts[ledger.AccountTypeIncome], counts[ledger.AccountTypeExpense]),
	}
}

func buildAccountItems(accounts []ledger.AccountRecord) []render.ListItemData {
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
				"Used accounts may be marked inactive. Accounts with postings cannot be deleted.",
			},
			Actions: []render.ItemActionData{renameAction, deleteAction, toggleAction},
		})
	}

	return items
}
