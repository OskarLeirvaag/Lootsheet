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

	// Write metadata lines first (plain text).
	for _, line := range metaLines {
		if y >= content.Y+content.H {
			return
		}
		buffer.WriteString(content.X, y, theme.Text, clipText(line, content.W))
		y++
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
