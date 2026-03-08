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
	buffer := NewBuffer(80, 18, theme.Base)

	rows := make([]string, 0, 12)
	for index := 0; index < 12; index++ {
		rows = append(rows, fmt.Sprintf("%04d asset active Account %02d", 1000+index, index))
	}

	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts: ListScreenData{
			HeaderLines:  []string{"Chart of accounts from smoke.db.", "Read-only list view."},
			SummaryLines: []string{"Accounts: 12 total", "Active: 12  Inactive: 0"},
			RowLines:     rows,
		},
	}
	shell := NewShell(&data)

	shell.HandleAction(ActionShowAccounts)
	shell.HandleAction(ActionScrollDown)
	shell.HandleAction(ActionScrollDown)
	shell.Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	if !strings.Contains(output, "Section: Accounts") {
		t.Fatalf("accounts screen missing section header:\n%s", output)
	}
	if !strings.Contains(output, "Accounts 3-") || !strings.Contains(output, "/12") {
		t.Fatalf("accounts screen missing scroll title:\n%s", output)
	}
	if strings.Contains(output, "Account 00") {
		t.Fatalf("accounts screen did not scroll:\n%s", output)
	}
	if !strings.Contains(output, "Account 02") {
		t.Fatalf("accounts screen missing expected visible row:\n%s", output)
	}
	if !strings.Contains(output, "↑↓ scroll") {
		t.Fatalf("accounts screen missing scroll help:\n%s", output)
	}
}
