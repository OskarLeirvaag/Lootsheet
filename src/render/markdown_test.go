package render

import (
	"strings"
	"testing"
)

func TestParseMarkdownLinesHeading(t *testing.T) {
	theme := DefaultTheme()
	lines := parseMarkdownLines("# Session 5", 40, &theme)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if len(lines[0].Spans) == 0 {
		t.Fatal("expected spans in heading line")
	}
	if lines[0].Spans[0].Text != "Session 5" {
		t.Fatalf("heading text = %q, want 'Session 5'", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesSubheading(t *testing.T) {
	theme := DefaultTheme()
	lines := parseMarkdownLines("## Action Items", 40, &theme)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if lines[0].Spans[0].Text != "Action Items" {
		t.Fatalf("subheading text = %q, want 'Action Items'", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesListItem(t *testing.T) {
	theme := DefaultTheme()
	lines := parseMarkdownLines("- Buy arrows", 40, &theme)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	// First span is the bullet.
	if lines[0].Spans[0].Text != "- " {
		t.Fatalf("bullet = %q, want '- '", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesBlockquote(t *testing.T) {
	theme := DefaultTheme()
	lines := parseMarkdownLines("> Important note", 40, &theme)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if lines[0].Spans[0].Text != "> " {
		t.Fatalf("blockquote prefix = %q, want '> '", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesCodeFence(t *testing.T) {
	theme := DefaultTheme()
	body := "```\nfoo := bar\n```"
	lines := parseMarkdownLines(body, 40, &theme)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for code fence, got %d", len(lines))
	}
	if lines[1].Spans[0].Text != "foo := bar" {
		t.Fatalf("code line = %q", lines[1].Spans[0].Text)
	}
}

func TestParseInlineSpansBold(t *testing.T) {
	theme := DefaultTheme()
	spans := parseInlineSpans("hello **world**", theme.Text)
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
	if spans[0].Text != "hello " {
		t.Fatalf("span 0 = %q", spans[0].Text)
	}
	if spans[1].Text != "world" {
		t.Fatalf("span 1 = %q", spans[1].Text)
	}
}

func TestParseInlineSpansItalic(t *testing.T) {
	theme := DefaultTheme()
	spans := parseInlineSpans("an *italic* word", theme.Text)
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}
	if spans[1].Text != "italic" {
		t.Fatalf("italic span = %q", spans[1].Text)
	}
}

func TestParseInlineSpansCode(t *testing.T) {
	theme := DefaultTheme()
	spans := parseInlineSpans("run `go test`", theme.Text)
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
	if spans[1].Text != "go test" {
		t.Fatalf("code span = %q", spans[1].Text)
	}
}

func TestParseInlineSpansReference(t *testing.T) {
	theme := DefaultTheme()
	spans := parseInlineSpans("see @quest/Clear the Tower", theme.Text)
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d: %v", len(spans), spans)
	}
	if spans[1].Text != "@quest/Clear" {
		t.Fatalf("reference span = %q", spans[1].Text)
	}
}

func TestWrapSpansShortLine(t *testing.T) {
	spans := []styledSpan{{Text: "short"}}
	lines := wrapSpans(spans, 40, 0)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
}

func TestWrapSpansLongLine(t *testing.T) {
	text := strings.Repeat("abcde ", 10)
	spans := []styledSpan{{Text: text}}
	lines := wrapSpans(spans, 20, 0)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %d lines", len(lines))
	}
}

func TestWrapSpansWithIndent(t *testing.T) {
	text := strings.Repeat("word ", 12)
	spans := []styledSpan{{Text: text}}
	lines := wrapSpans(spans, 20, 2)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %d lines", len(lines))
	}
	// Continuation lines should start with 2-space indent.
	if len(lines) > 1 && len(lines[1].Spans) > 0 {
		first := lines[1].Spans[0].Text
		if !strings.HasPrefix(first, "  ") {
			t.Fatalf("continuation should start with indent, got %q", first)
		}
	}
}

func TestParseMarkdownLinesEmptyBody(t *testing.T) {
	theme := DefaultTheme()
	lines := parseMarkdownLines("", 40, &theme)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for empty body, got %d", len(lines))
	}
}

func TestParseMarkdownLinesNumberedList(t *testing.T) {
	theme := DefaultTheme()
	lines := parseMarkdownLines("1. First item", 40, &theme)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if lines[0].Spans[0].Text != "1. " {
		t.Fatalf("numbered list prefix = %q, want '1. '", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesMultipleBlocks(t *testing.T) {
	theme := DefaultTheme()
	body := "# Title\n\n- Item one\n- Item two\n\n> A quote"
	lines := parseMarkdownLines(body, 40, &theme)
	if len(lines) < 5 {
		t.Fatalf("expected at least 5 lines, got %d", len(lines))
	}
}

func TestDrawStyledPanelProducesOutput(t *testing.T) {
	theme := DefaultTheme()
	buffer := NewBuffer(50, 15, theme.Base)

	body := "# Test Heading\n\nSome **bold** text and @quest/Clear reference."
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
