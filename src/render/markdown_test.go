package render

import (
	"strings"
	"testing"
)

func TestDrawStyledPanelProducesOutput(t *testing.T) {
	theme := DefaultTheme()
	buffer := NewBuffer(50, 15, theme.Base)

	body := "# Test Heading\n\nSome **bold** text and @[quest/Clear] reference."
	mdLines := parseMarkdownLines(body, 44, &theme)

	rect := Rect{X: 0, Y: 0, W: 50, H: 15}
	DrawStyledPanel(buffer, rect, &theme, "Detail", []string{"Updated: 2026-03-12"}, mdLines, theme.SectionNotes, theme.SectionNotes)

	output := buffer.PlainText()
	for _, token := range []string{"Detail", "Updated: 2026-03-12", "Test Heading", "bold"} {
		if !strings.Contains(output, token) {
			t.Fatalf("styled panel missing %q:\n%s", token, output)
		}
	}
}
