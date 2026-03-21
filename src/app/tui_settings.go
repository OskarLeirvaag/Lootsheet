package app

import (
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

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

func buildSettingsCampaignItems(campaigns []campaign.Record, activeCampaignID string) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(campaigns))

	for _, c := range campaigns {
		active := ""
		if c.ID == activeCampaignID {
			active = " (active)"
		}

		renameAction := render.ItemActionData{
			Trigger:     render.ActionEdit,
			ID:          tuiCommandCampaignRename,
			Label:       "u rename",
			Mode:        render.ItemActionModeInput,
			InputTitle:  fmt.Sprintf("Rename campaign %q", c.Name),
			InputPrompt: "New name",
			Placeholder: c.Name,
			InputHelp: []string{
				"ID: " + c.ID,
				"Current name: " + c.Name,
			},
		}

		actions := []render.ItemActionData{renameAction}

		if c.ID != activeCampaignID {
			deleteAction := render.ItemActionData{
				Trigger:      render.ActionDelete,
				ID:           tuiCommandCampaignDelete,
				Label:        "d remove",
				ConfirmTitle: fmt.Sprintf("Delete campaign %q?", c.Name),
				ConfirmLines: []string{
					c.Name,
					"This will permanently delete the campaign and all its data.",
					"The active campaign cannot be deleted.",
				},
			}
			actions = append(actions, deleteAction)
		}

		items = append(items, render.ListItemData{
			Key:         c.ID,
			Row:         fmt.Sprintf("%-40s%s", c.Name, active),
			DetailTitle: "Campaign: " + c.Name,
			DetailLines: []string{
				"ID: " + c.ID,
				"Name: " + c.Name,
				"Status: " + active,
			},
			Actions: actions,
		})
	}

	return items
}

func summarizeSettingsCampaigns(campaigns []campaign.Record, activeCampaignID string) []string {
	activeName := ""
	for _, c := range campaigns {
		if c.ID == activeCampaignID {
			activeName = c.Name
			break
		}
	}
	return []string{
		fmt.Sprintf("Campaigns: %d  Active: %s", len(campaigns), activeName),
	}
}
