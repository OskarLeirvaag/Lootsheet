package render

import (
	"strings"
	"testing"
)

func TestDashboardRenderShowsPanelsAndFooterHelp(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(80, 24, theme.Base)

	Dashboard{}.Render(buffer, &theme, keymap)

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

	Dashboard{}.Render(buffer, &theme, keymap)

	output := buffer.PlainText()
	if !strings.Contains(output, "Terminal too small") {
		t.Fatalf("compact output missing resize message:\n%s", output)
	}
}
