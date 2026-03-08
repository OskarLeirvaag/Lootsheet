package render

import (
	"strings"
	"testing"
)

func TestDashboardRenderShowsPanelsAndFooterHelp(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(80, 24, theme.Base)

	(&Dashboard{}).Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	for _, token := range []string{
		"LootSheet Dashboard",
		"Accounts",
		"Journal",
		"Ledger Snapshot",
		"Quest Register",
		"Loot Register",
		"q quit",
		"Esc quit",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("dashboard output missing %q:\n%s", token, output)
		}
	}
}

func TestDashboardRenderFallsBackForSmallTerminals(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(30, 8, theme.Base)

	(&Dashboard{}).Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	if !strings.Contains(output, "Terminal too small") {
		t.Fatalf("compact output missing resize message:\n%s", output)
	}
}

func TestDashboardRenderUsesProvidedData(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(80, 24, theme.Base)

	(&Dashboard{
		Data: DashboardData{
			HeaderLines:   []string{"Live dashboard summary.", "Ctrl+L refreshes the snapshot."},
			AccountsLines: []string{"Accounts: 16 total", "Active: 16  Inactive: 0"},
			JournalLines:  []string{"Entries: 2 total", "Latest: #2 2026-03-09"},
			LedgerLines:   []string{"Status: BALANCED"},
			QuestLines:    []string{"Receivables: 1"},
			LootLines:     []string{"Tracked items: 1"},
		},
	}).Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	for _, token := range []string{
		"Live dashboard summary.",
		"Accounts: 16 total",
		"Entries: 2 total",
		"Status: BALANCED",
		"Receivables: 1",
		"Tracked items: 1",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("dashboard output missing %q:\n%s", token, output)
		}
	}
}
