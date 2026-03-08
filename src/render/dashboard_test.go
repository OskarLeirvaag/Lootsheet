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
		"Party Hoard",
		"Ledger Snapshot",
		"Quest Register",
		"e  I have an expense",
		"/________\\",
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
			HeaderLines:     []string{"Live dashboard summary.", "Ctrl+L refreshes the snapshot."},
			AccountsLines:   []string{"Accounts: 16 total", "Active: 16  Inactive: 0"},
			JournalLines:    []string{"Entries: 2 total", "Latest: #2 2026-03-09"},
			HoardLines:      []string{"To share now: 15 GP", "Unsold loot: 8 GP"},
			QuickEntryLines: []string{"e  I have an expense", "i  I have income", "a  Add custom entry"},
			LedgerLines:     []string{"Status: BALANCED"},
			QuestLines:      []string{"Receivables: 1"},
			LootLines:       []string{"Tracked items: 1"},
		},
	}).Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	for _, token := range []string{
		"Live dashboard summary.",
		"Accounts: 16 total",
		"Entries: 2 total",
		"To share now: 15 GP",
		"e  I have an expense",
		"Status: BALANCED",
		"Receivables: 1",
		"Unsold loot: 8 GP",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("dashboard output missing %q:\n%s", token, output)
		}
	}
}
