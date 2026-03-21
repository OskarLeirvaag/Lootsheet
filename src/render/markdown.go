package render

import (
	"github.com/OskarLeirvaag/Lootsheet/src/render/markdown"
	"github.com/gdamore/tcell/v2"
)

// Type aliases re-export markdown types so render-internal code
// can continue using unqualified names.
type styledLine = markdown.StyledLine

// markdownStyles builds a MarkdownStyles from a Theme.
func markdownStyles(theme *Theme) markdown.MarkdownStyles {
	return markdown.MarkdownStyles{
		Text:       theme.Text,
		Muted:      theme.Muted,
		Heading:    theme.MarkdownHeading,
		Bold:       theme.MarkdownBold,
		Code:       theme.MarkdownCode,
		Blockquote: theme.MarkdownBlockquote,
		Reference:  theme.MarkdownReference,
	}
}

// parseMarkdownLines delegates to the markdown sub-package.
func parseMarkdownLines(body string, width int, theme *Theme) []styledLine {
	if theme == nil {
		return nil
	}
	ms := markdownStyles(theme)
	return markdown.ParseMarkdownLines(body, width, &ms)
}

// wrapPlainText wraps a single line of plain text to fit within width.
// Returns one or more lines. Empty input returns a single empty string.
func wrapPlainText(text string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	runes := []rune(text)
	if len(runes) <= width {
		return []string{text}
	}
	var lines []string
	for len(runes) > width {
		lines = append(lines, string(runes[:width]))
		runes = runes[width:]
	}
	lines = append(lines, string(runes))
	return lines
}

// DrawStyledPanel renders a panel with styled line content instead of plain strings.
func DrawStyledPanel(buffer *Buffer, rect Rect, theme *Theme, title string, metaLines []string, styledLines []styledLine, borderStyle, titleStyle tcell.Style) {
	if buffer == nil || rect.Empty() {
		return
	}

	// Draw the panel chrome (border + title).
	DrawPanel(buffer, rect, theme, Panel{
		Title:       title,
		BorderStyle: &borderStyle,
		TitleStyle:  &titleStyle,
		Texture:     PanelTextureNone,
	})

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	y := content.Y

	// Write metadata lines with wrapping.
	for _, line := range metaLines {
		wrapped := wrapPlainText(line, content.W)
		for _, wl := range wrapped {
			if y >= content.Y+content.H {
				return
			}
			buffer.WriteString(content.X, y, theme.Text, wl)
			y++
		}
	}

	// Blank separator after metadata.
	if len(metaLines) > 0 && len(styledLines) > 0 {
		y++
	}

	// Write styled markdown lines.
	for _, sl := range styledLines {
		if y >= content.Y+content.H {
			return
		}
		markdown.WriteStyledSpans(buffer, content.X, y, sl.Spans, content.W)
		y++
	}
}
