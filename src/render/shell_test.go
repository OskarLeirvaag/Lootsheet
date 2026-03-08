package render

import (
	"fmt"
	"strings"
	"testing"
)

func TestShellRenderShowsTabsAndFooterHelp(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(100, 28, theme.Base)

	data := DefaultShellData()
	NewShell(&data).Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	for _, token := range []string{
		"LootSheet TUI",
		"Section: Dashboard",
		"Sections: [Dashboard]  Accounts  Journal  Quests  Loot",
		"1-5 jump",
		"q quit",
		"Ctrl+L refresh",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("shell output missing %q:\n%s", token, output)
		}
	}
}

func TestShellRenderShowsScrollableAccountsScreen(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(96, 28, theme.Base)

	items := make([]ListItemData, 0, 12)
	for index := 0; index < 12; index++ {
		row := fmt.Sprintf("%04d asset active Account %02d", 1000+index, index)
		items = append(items, ListItemData{
			Key:         fmt.Sprintf("%04d", 1000+index),
			Row:         row,
			DetailTitle: fmt.Sprintf("Account %02d", index),
			DetailLines: []string{fmt.Sprintf("Detail for account %02d", index)},
		})
	}

	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts: ListScreenData{
			HeaderLines:  []string{"Chart of accounts from smoke.db.", "Read-only list view."},
			SummaryLines: []string{"Accounts: 12 total", "Active: 12  Inactive: 0"},
			Items:        items,
		},
	}
	shell := NewShell(&data)

	shell.HandleAction(ActionShowAccounts)
	for index := 0; index < 6; index++ {
		shell.HandleAction(ActionMoveDown)
	}
	shell.Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	if !strings.Contains(output, "Section: Accounts") {
		t.Fatalf("accounts screen missing section header:\n%s", output)
	}
	if !strings.Contains(output, "Accounts ") || !strings.Contains(output, "/12") {
		t.Fatalf("accounts screen missing scroll title:\n%s", output)
	}
	if strings.Contains(output, "Account 00") {
		t.Fatalf("accounts screen did not scroll:\n%s", output)
	}
	if !strings.Contains(output, "Account 06") {
		t.Fatalf("accounts screen missing expected visible row:\n%s", output)
	}
	if !strings.Contains(output, "↑↓ select") {
		t.Fatalf("accounts screen missing selection help:\n%s", output)
	}
	if !strings.Contains(output, "Detail for account 06") {
		t.Fatalf("accounts screen missing detail pane content:\n%s", output)
	}
}

func TestShellRenderKeepsDetailVisibleOnStandardTerminal(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(80, 24, theme.Base)

	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts: ListScreenData{
			HeaderLines:  []string{"Chart of accounts from smoke.db.", "Interactive list view."},
			SummaryLines: []string{"Accounts: 2 total", "Active: 2  Inactive: 0"},
			Items: []ListItemData{
				{
					Key:         "1000",
					Row:         "1000 asset active Party Cash",
					DetailTitle: "Account 1000",
					DetailLines: []string{"Name: Party Cash", "Status: active"},
				},
				{
					Key:         "1100",
					Row:         "1100 asset active Quest Receivable",
					DetailTitle: "Account 1100",
					DetailLines: []string{"Name: Quest Receivable", "Status: active"},
				},
			},
		},
	}

	shell := NewShell(&data)
	shell.HandleAction(ActionShowAccounts)
	shell.Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	for _, token := range []string{"Account 1000", "Name: Party Cash"} {
		if !strings.Contains(output, token) {
			t.Fatalf("standard terminal output missing %q:\n%s", token, output)
		}
	}
}

func TestShellPrimaryActionOpensConfirmAndEmitsCommand(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "1000",
					Row:         "1000 asset active Party Cash",
					DetailTitle: "Account 1000",
					DetailLines: []string{"Name: Party Cash"},
					PrimaryAction: &ItemActionData{
						ID:           "account.deactivate",
						Label:        "t deactivate",
						ConfirmTitle: "Deactivate account 1000?",
						ConfirmLines: []string{"Party Cash"},
					},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowAccounts)

	result := shell.HandleAction(ActionPrimary)
	if !result.Redraw {
		t.Fatal("expected primary action to trigger redraw")
	}
	if shell.confirm == nil {
		t.Fatal("expected confirm modal to open")
	}

	cancel := shell.HandleAction(ActionQuit)
	if !cancel.Redraw {
		t.Fatal("expected quit inside confirm modal to cancel it")
	}
	if shell.confirm != nil {
		t.Fatal("expected confirm modal to close on quit")
	}

	shell.HandleAction(ActionPrimary)
	result = shell.HandleAction(ActionConfirm)
	if result.Command == nil {
		t.Fatal("expected confirm action to emit command")
	}
	if result.Command.ID != "account.deactivate" {
		t.Fatalf("command id = %q, want account.deactivate", result.Command.ID)
	}
	if result.Command.ItemKey != "1000" {
		t.Fatalf("command item key = %q, want 1000", result.Command.ItemKey)
	}
}

func TestShellReloadPreservesSelectionByKey(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts: ListScreenData{
			Items: []ListItemData{
				{Key: "1000", Row: "1000 asset active Party Cash"},
				{Key: "1100", Row: "1100 asset active Quest Receivable"},
				{Key: "1200", Row: "1200 asset active Loot Inventory"},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowAccounts)
	shell.HandleAction(ActionMoveDown)

	if item := shell.currentSelectedItem(SectionAccounts); item == nil || item.Key != "1100" {
		t.Fatalf("selected item before reload = %#v, want key 1100", item)
	}

	reloaded := ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts: ListScreenData{
			Items: []ListItemData{
				{Key: "0900", Row: "0900 asset active Treasury Box"},
				{Key: "1100", Row: "1100 asset active Quest Receivable"},
				{Key: "1200", Row: "1200 asset active Loot Inventory"},
			},
		},
	}
	shell.Reload(&reloaded)

	if item := shell.currentSelectedItem(SectionAccounts); item == nil || item.Key != "1100" {
		t.Fatalf("selected item after reload = %#v, want key 1100", item)
	}
}
