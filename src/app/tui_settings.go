package app

import (
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func buildSettingsAccountItems(accounts []ledger.AccountRecord) []render.ListItemData {
	return buildAccountItems(accounts, nil)
}

func buildSettingsCodexTypeItems(codexTypes []codex.CodexType) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(codexTypes))

	for _, ct := range codexTypes {
		renameAction := render.ItemActionData{
			Trigger:     render.ActionEdit,
			ID:          tuiCommandCodexTypeRename,
			Label:       "u rename",
			Mode:        render.ItemActionModeInput,
			InputTitle:  fmt.Sprintf("Rename codex type %q", ct.ID),
			InputPrompt: "New name",
			Placeholder: ct.Name,
			InputHelp: []string{
				"ID: " + ct.ID + " (immutable)",
				"Form: " + ct.FormID,
				"Current name: " + ct.Name,
			},
		}

		deleteAction := render.ItemActionData{
			Trigger:      render.ActionDelete,
			ID:           tuiCommandCodexTypeDelete,
			Label:        "d remove",
			ConfirmTitle: fmt.Sprintf("Remove codex type %q?", ct.ID),
			ConfirmLines: []string{
				ct.Name,
				"Types with existing codex entries cannot be removed.",
			},
		}

		items = append(items, render.ListItemData{
			Key:         "codex_type:" + ct.ID,
			Row:         fmt.Sprintf("%-12s %-12s %s", ct.ID, ct.FormID, ct.Name),
			DetailTitle: "Codex Type: " + ct.Name,
			DetailLines: []string{
				"ID: " + ct.ID,
				"Name: " + ct.Name,
				"Form template: " + ct.FormID,
			},
			Actions: []render.ItemActionData{renameAction, deleteAction},
		})
	}

	return items
}

func summarizeSettingsAccounts(accounts []ledger.AccountRecord) []string {
	active := 0
	for _, record := range accounts {
		if record.Active {
			active++
		}
	}

	return []string{
		fmt.Sprintf("Accounts: %d total  Active: %d  Inactive: %d", len(accounts), active, len(accounts)-active),
	}
}

func summarizeSettingsCodexTypes(codexTypes []codex.CodexType) []string {
	return []string{
		fmt.Sprintf("Codex types: %d", len(codexTypes)),
	}
}
