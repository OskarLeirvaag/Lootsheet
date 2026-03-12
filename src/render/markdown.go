package render

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// styledSpan is a fragment of text with a single style.
type styledSpan struct {
	Text  string
	Style tcell.Style
}

// styledLine is a sequence of styled spans forming one visual line.
type styledLine struct {
	Spans []styledSpan
}

// parseMarkdownLines converts a body string into styled lines ready for rendering.
// It handles block-level formatting (headings, lists, blockquotes, code fences)
// and inline formatting (bold, italic, inline code, @references).
func parseMarkdownLines(body string, width int, theme *Theme) []styledLine {
	if width <= 0 || theme == nil {
		return nil
	}

	rawLines := strings.Split(body, "\n")
	var result []styledLine
	inCodeFence := false

	for _, raw := range rawLines {
		if strings.HasPrefix(raw, "```") {
			inCodeFence = !inCodeFence
			result = append(result, styledLine{Spans: []styledSpan{
				{Text: raw, Style: theme.MarkdownCode},
			}})
			continue
		}

		if inCodeFence {
			result = append(result, styledLine{Spans: []styledSpan{
				{Text: raw, Style: theme.MarkdownCode},
			}})
			continue
		}

		spans, indent := parseBlockLine(raw, theme)
		wrapped := wrapSpans(spans, width, indent)
		result = append(result, wrapped...)
	}

	return result
}

// parseBlockLine determines the block type by prefix and returns inline-parsed spans
// plus a continuation indent width.
func parseBlockLine(line string, theme *Theme) ([]styledSpan, int) {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)

	// Headings.
	if text, ok := strings.CutPrefix(trimmed, "### "); ok {
		return parseInlineSpans(text, theme.MarkdownBold), 0
	}
	if text, ok := strings.CutPrefix(trimmed, "## "); ok {
		return parseInlineSpans(text, theme.MarkdownBold), 0
	}
	if text, ok := strings.CutPrefix(trimmed, "# "); ok {
		return parseInlineSpans(text, theme.MarkdownHeading), 0
	}

	// List items.
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		bullet := string(trimmed[0])
		text := trimmed[2:]
		spans := append([]styledSpan{{Text: bullet + " ", Style: theme.MarkdownHeading}}, parseInlineSpans(text, theme.Text)...) //nolint:gocritic // appendAssign: intentional inline append
		return spans, 2
	}

	// Numbered list.
	if idx := strings.Index(trimmed, ". "); idx > 0 && idx <= 3 {
		prefix := trimmed[:idx+2]
		allDigits := true
		for _, r := range trimmed[:idx] {
			if !unicode.IsDigit(r) {
				allDigits = false
				break
			}
		}
		if allDigits {
			text := trimmed[idx+2:]
			spans := append([]styledSpan{{Text: prefix, Style: theme.MarkdownHeading}}, parseInlineSpans(text, theme.Text)...) //nolint:gocritic // appendAssign: intentional inline append
			return spans, len([]rune(prefix))
		}
	}

	// Blockquote.
	if text, ok := strings.CutPrefix(trimmed, "> "); ok {
		spans := append([]styledSpan{{Text: "> ", Style: theme.MarkdownBlockquote}}, parseInlineSpans(text, theme.MarkdownBlockquote)...) //nolint:gocritic // appendAssign: intentional inline append
		return spans, 2
	}

	// Normal paragraph.
	if trimmed == "" {
		return []styledSpan{{Text: "", Style: theme.Text}}, 0
	}
	return parseInlineSpans(trimmed, theme.Text), 0
}

// parseInlineSpans processes inline markdown formatting within a line.
func parseInlineSpans(text string, baseStyle tcell.Style) []styledSpan {
	if text == "" {
		return nil
	}

	var spans []styledSpan
	runes := []rune(text)
	i := 0

	var current []rune

	flushCurrent := func() {
		if len(current) > 0 {
			spans = append(spans, styledSpan{Text: string(current), Style: baseStyle})
			current = nil
		}
	}

	for i < len(runes) {
		// Bold: **text**
		if i+1 < len(runes) && runes[i] == '*' && runes[i+1] == '*' {
			end := findClosing(runes, i+2, "**")
			if end >= 0 {
				flushCurrent()
				inner := string(runes[i+2 : end])
				boldStyle := baseStyle.Bold(true)
				spans = append(spans, styledSpan{Text: inner, Style: boldStyle})
				i = end + 2
				continue
			}
		}

		// Italic: *text*
		if runes[i] == '*' && (i+1 >= len(runes) || runes[i+1] != '*') {
			end := findClosingSingle(runes, i+1, '*')
			if end >= 0 {
				flushCurrent()
				inner := string(runes[i+1 : end])
				italicStyle := baseStyle.Italic(true)
				spans = append(spans, styledSpan{Text: inner, Style: italicStyle})
				i = end + 1
				continue
			}
		}

		// Inline code: `text`
		if runes[i] == '`' {
			end := findClosingSingle(runes, i+1, '`')
			if end >= 0 {
				flushCurrent()
				inner := string(runes[i+1 : end])
				spans = append(spans, styledSpan{Text: inner, Style: baseStyle.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)})
				i = end + 1
				continue
			}
		}

		// Reference: @type/name
		if runes[i] == '@' && i+1 < len(runes) {
			end := i + 1
			for end < len(runes) && !isRefTerminatorRune(runes[end]) {
				end++
			}
			ref := string(runes[i:end])
			if strings.Contains(ref, "/") && len(ref) > 2 {
				flushCurrent()
				// Use MarkdownReference style from theme via a small hack:
				// we pass it through as underlined base
				refStyle := baseStyle.Foreground(tcell.NewRGBColor(140, 190, 160))
				spans = append(spans, styledSpan{Text: ref, Style: refStyle})
				i = end
				continue
			}
		}

		current = append(current, runes[i])
		i++
	}
	flushCurrent()

	return spans
}

func isRefTerminatorRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '.' || r == ')' || r == ']'
}

// findClosing finds the position of a two-character closing marker.
func findClosing(runes []rune, start int, marker string) int {
	mr := []rune(marker)
	for i := start; i+1 < len(runes); i++ {
		if runes[i] == mr[0] && runes[i+1] == mr[1] {
			return i
		}
	}
	return -1
}

// findClosingSingle finds the position of a single-character closing marker.
func findClosingSingle(runes []rune, start int, marker rune) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == marker {
			return i
		}
	}
	return -1
}

// wrapSpans wraps a slice of spans to fit within the given width.
// Continuation lines are indented by contIndent spaces.
func wrapSpans(spans []styledSpan, maxWidth int, contIndent int) []styledLine {
	if maxWidth <= 0 {
		return nil
	}

	// Flatten to measure.
	var fullText strings.Builder
	for _, sp := range spans {
		fullText.WriteString(sp.Text)
	}
	if fullText.Len() == 0 {
		return []styledLine{{Spans: spans}}
	}

	// Simple character-count wrapping.
	totalRunes := []rune(fullText.String())
	if len(totalRunes) <= maxWidth {
		return []styledLine{{Spans: spans}}
	}

	// Need to wrap. Build lines character by character.
	var lines []styledLine
	col := 0
	lineNum := 0
	var currentSpans []styledSpan
	var currentRunes []rune
	var currentStyle tcell.Style
	styleSet := false

	// Build a flat list of (rune, style) pairs.
	type styledRune struct {
		R rune
		S tcell.Style
	}
	var flat []styledRune
	for _, sp := range spans {
		for _, r := range sp.Text {
			flat = append(flat, styledRune{R: r, S: sp.Style})
		}
	}

	flushRunes := func() {
		if len(currentRunes) > 0 {
			currentSpans = append(currentSpans, styledSpan{Text: string(currentRunes), Style: currentStyle})
			currentRunes = nil
		}
	}

	lineWidth := maxWidth
	for _, sr := range flat {
		if col >= lineWidth {
			flushRunes()
			lines = append(lines, styledLine{Spans: currentSpans})
			currentSpans = nil
			styleSet = false
			col = 0
			lineNum++
			lineWidth = maxWidth - contIndent
			if contIndent > 0 {
				currentSpans = append(currentSpans, styledSpan{
					Text:  strings.Repeat(" ", contIndent),
					Style: sr.S,
				})
				col = contIndent
			}
		}

		if !styleSet || sr.S != currentStyle {
			flushRunes()
			currentStyle = sr.S
			styleSet = true
		}
		currentRunes = append(currentRunes, sr.R)
		col++
	}
	flushRunes()
	if len(currentSpans) > 0 {
		lines = append(lines, styledLine{Spans: currentSpans})
	}

	if len(lines) == 0 {
		return []styledLine{{Spans: spans}}
	}
	return lines
}

// WriteStyledSpans writes styled spans to the buffer at the given position.
func WriteStyledSpans(buffer *Buffer, x, y int, spans []styledSpan, maxWidth int) {
	if buffer == nil {
		return
	}
	col := x
	for _, sp := range spans {
		for _, r := range sp.Text {
			if col-x >= maxWidth {
				return
			}
			buffer.Set(col, y, r, sp.Style)
			col++
		}
	}
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
		WriteStyledSpans(buffer, content.X, y, sl.Spans, content.W)
		y++
	}
}
