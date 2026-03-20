package markdown

import (
	"strings"
	"unicode"

	"github.com/OskarLeirvaag/Lootsheet/src/render/canvas"
	"github.com/gdamore/tcell/v2"
)

// MarkdownStyles carries the resolved styles needed by the markdown parser,
// decoupled from the render.Theme type to avoid an import cycle.
type MarkdownStyles struct {
	Text       tcell.Style
	Muted      tcell.Style // unused currently, but may be useful
	Heading    tcell.Style
	Bold       tcell.Style
	Code       tcell.Style
	Blockquote tcell.Style
	Reference  tcell.Style
}

// StyledSpan is a fragment of text with a single style.
type StyledSpan struct {
	Text  string
	Style tcell.Style
}

// StyledLine is a sequence of styled spans forming one visual line.
type StyledLine struct {
	Spans []StyledSpan
}

// ParseMarkdownLines converts a body string into styled lines ready for rendering.
// It handles block-level formatting (headings, lists, blockquotes, code fences)
// and inline formatting (bold, italic, inline code, @references).
func ParseMarkdownLines(body string, width int, styles *MarkdownStyles) []StyledLine {
	if width <= 0 {
		return nil
	}

	rawLines := strings.Split(body, "\n")
	var result []StyledLine
	inCodeFence := false

	for _, raw := range rawLines {
		if strings.HasPrefix(raw, "```") {
			inCodeFence = !inCodeFence
			result = append(result, StyledLine{Spans: []StyledSpan{
				{Text: raw, Style: styles.Code},
			}})
			continue
		}

		if inCodeFence {
			result = append(result, StyledLine{Spans: []StyledSpan{
				{Text: raw, Style: styles.Code},
			}})
			continue
		}

		spans, indent := parseBlockLine(raw, styles)
		wrapped := wrapSpans(spans, width, indent)
		result = append(result, wrapped...)
	}

	return result
}

// parseBlockLine determines the block type by prefix and returns inline-parsed spans
// plus a continuation indent width.
func parseBlockLine(line string, styles *MarkdownStyles) ([]StyledSpan, int) {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)

	// Headings.
	if text, ok := strings.CutPrefix(trimmed, "### "); ok {
		return parseInlineSpans(text, styles.Bold, styles.Reference), 0
	}
	if text, ok := strings.CutPrefix(trimmed, "## "); ok {
		return parseInlineSpans(text, styles.Bold, styles.Reference), 0
	}
	if text, ok := strings.CutPrefix(trimmed, "# "); ok {
		return parseInlineSpans(text, styles.Heading, styles.Reference), 0
	}

	// List items.
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		bullet := string(trimmed[0])
		text := trimmed[2:]
		spans := append([]StyledSpan{{Text: bullet + " ", Style: styles.Heading}}, parseInlineSpans(text, styles.Text, styles.Reference)...) //nolint:gocritic // appendAssign: intentional inline append
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
			spans := append([]StyledSpan{{Text: prefix, Style: styles.Heading}}, parseInlineSpans(text, styles.Text, styles.Reference)...) //nolint:gocritic // appendAssign: intentional inline append
			return spans, len([]rune(prefix))
		}
	}

	// Blockquote.
	if text, ok := strings.CutPrefix(trimmed, "> "); ok {
		spans := append([]StyledSpan{{Text: "> ", Style: styles.Blockquote}}, parseInlineSpans(text, styles.Blockquote, styles.Reference)...) //nolint:gocritic // appendAssign: intentional inline append
		return spans, 2
	}

	// Normal paragraph.
	if trimmed == "" {
		return []StyledSpan{{Text: "", Style: styles.Text}}, 0
	}
	return parseInlineSpans(trimmed, styles.Text, styles.Reference), 0
}

// parseInlineSpans processes inline markdown formatting within a line.
func parseInlineSpans(text string, baseStyle tcell.Style, opts ...tcell.Style) []StyledSpan {
	if text == "" {
		return nil
	}

	var spans []StyledSpan
	runes := []rune(text)
	i := 0

	var current []rune

	flushCurrent := func() {
		if len(current) > 0 {
			spans = append(spans, StyledSpan{Text: string(current), Style: baseStyle})
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
				spans = append(spans, StyledSpan{Text: inner, Style: boldStyle})
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
				spans = append(spans, StyledSpan{Text: inner, Style: italicStyle})
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
				spans = append(spans, StyledSpan{Text: inner, Style: baseStyle.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)})
				i = end + 1
				continue
			}
		}

		// Reference: @type/name
		if runes[i] == '@' && i+1 < len(runes) {
			end := i + 1
			for end < len(runes) && !IsRefTerminatorRune(runes[end]) {
				end++
			}
			ref := string(runes[i:end])
			if strings.Contains(ref, "/") && len(ref) > 2 {
				flushCurrent()
				rs := baseStyle.Foreground(tcell.NewRGBColor(180, 160, 220))
				if len(opts) > 0 {
					rs = opts[0]
				}
				spans = append(spans, StyledSpan{Text: ref, Style: rs})
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

// IsRefTerminatorRune returns true for runes that terminate an @reference.
func IsRefTerminatorRune(r rune) bool {
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
func wrapSpans(spans []StyledSpan, maxWidth int, contIndent int) []StyledLine {
	if maxWidth <= 0 {
		return nil
	}

	// Flatten to measure.
	var fullText strings.Builder
	for _, sp := range spans {
		fullText.WriteString(sp.Text)
	}
	if fullText.Len() == 0 {
		return []StyledLine{{Spans: spans}}
	}

	// Simple character-count wrapping.
	totalRunes := []rune(fullText.String())
	if len(totalRunes) <= maxWidth {
		return []StyledLine{{Spans: spans}}
	}

	// Need to wrap. Build lines character by character.
	var lines []StyledLine
	col := 0
	lineNum := 0
	var currentSpans []StyledSpan
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
			currentSpans = append(currentSpans, StyledSpan{Text: string(currentRunes), Style: currentStyle})
			currentRunes = nil
		}
	}

	lineWidth := maxWidth
	for _, sr := range flat {
		if col >= lineWidth {
			flushRunes()
			lines = append(lines, StyledLine{Spans: currentSpans})
			currentSpans = nil
			styleSet = false
			col = 0
			lineNum++
			lineWidth = maxWidth - contIndent
			if contIndent > 0 {
				currentSpans = append(currentSpans, StyledSpan{
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
		lines = append(lines, StyledLine{Spans: currentSpans})
	}

	if len(lines) == 0 {
		return []StyledLine{{Spans: spans}}
	}
	return lines
}

// WriteStyledSpans writes styled spans to the buffer at the given position.
func WriteStyledSpans(buffer *canvas.Buffer, x, y int, spans []StyledSpan, maxWidth int) {
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
