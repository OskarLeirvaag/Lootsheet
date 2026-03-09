package render

import (
	"strings"
	"testing"
)

func TestDrawPanelRendersBoxAndTitle(t *testing.T) {
	buffer := NewBuffer(20, 6, DefaultTheme().Base)
	theme := DefaultTheme()

	DrawPanel(buffer, Rect{W: 20, H: 6}, &theme, Panel{
		Title: "Accounts",
		Lines: []string{"Read-only shell", "Placeholder"},
	})

	output := buffer.PlainText()
	if !strings.Contains(output, "╔ Accounts ") {
		t.Fatalf("panel output missing title:\n%s", output)
	}
	if !strings.Contains(output, "│Read-only shell") {
		t.Fatalf("panel output missing body:\n%s", output)
	}
	if !strings.Contains(output, "╚") {
		t.Fatalf("panel output missing bottom border:\n%s", output)
	}
}
