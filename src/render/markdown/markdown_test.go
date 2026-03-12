package markdown

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func testStyles() *MarkdownStyles {
	base := tcell.StyleDefault
	return &MarkdownStyles{
		Text:       base,
		Muted:      base,
		Heading:    base.Bold(true),
		Bold:       base.Bold(true),
		Code:       base,
		Blockquote: base.Italic(true),
		Reference:  base.Foreground(tcell.NewRGBColor(140, 190, 160)),
	}
}

func TestParseMarkdownLinesHeading(t *testing.T) {
	styles := testStyles()
	lines := ParseMarkdownLines("# Session 5", 40, styles)
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
	styles := testStyles()
	lines := ParseMarkdownLines("## Action Items", 40, styles)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if lines[0].Spans[0].Text != "Action Items" {
		t.Fatalf("subheading text = %q, want 'Action Items'", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesListItem(t *testing.T) {
	styles := testStyles()
	lines := ParseMarkdownLines("- Buy arrows", 40, styles)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	// First span is the bullet.
	if lines[0].Spans[0].Text != "- " {
		t.Fatalf("bullet = %q, want '- '", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesBlockquote(t *testing.T) {
	styles := testStyles()
	lines := ParseMarkdownLines("> Important note", 40, styles)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if lines[0].Spans[0].Text != "> " {
		t.Fatalf("blockquote prefix = %q, want '> '", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesCodeFence(t *testing.T) {
	styles := testStyles()
	body := "```\nfoo := bar\n```"
	lines := ParseMarkdownLines(body, 40, styles)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for code fence, got %d", len(lines))
	}
	if lines[1].Spans[0].Text != "foo := bar" {
		t.Fatalf("code line = %q", lines[1].Spans[0].Text)
	}
}

func TestParseInlineSpansBold(t *testing.T) {
	styles := testStyles()
	spans := parseInlineSpans("hello **world**", styles.Text)
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
	styles := testStyles()
	spans := parseInlineSpans("an *italic* word", styles.Text)
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}
	if spans[1].Text != "italic" {
		t.Fatalf("italic span = %q", spans[1].Text)
	}
}

func TestParseInlineSpansCode(t *testing.T) {
	styles := testStyles()
	spans := parseInlineSpans("run `go test`", styles.Text)
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
	if spans[1].Text != "go test" {
		t.Fatalf("code span = %q", spans[1].Text)
	}
}

func TestParseInlineSpansReference(t *testing.T) {
	styles := testStyles()
	spans := parseInlineSpans("see @quest/Clear the Tower", styles.Text)
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d: %v", len(spans), spans)
	}
	if spans[1].Text != "@quest/Clear" {
		t.Fatalf("reference span = %q", spans[1].Text)
	}
}

func TestWrapSpansShortLine(t *testing.T) {
	spans := []StyledSpan{{Text: "short"}}
	lines := wrapSpans(spans, 40, 0)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
}

func TestWrapSpansLongLine(t *testing.T) {
	text := strings.Repeat("abcde ", 10)
	spans := []StyledSpan{{Text: text}}
	lines := wrapSpans(spans, 20, 0)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %d lines", len(lines))
	}
}

func TestWrapSpansWithIndent(t *testing.T) {
	text := strings.Repeat("word ", 12)
	spans := []StyledSpan{{Text: text}}
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
	styles := testStyles()
	lines := ParseMarkdownLines("", 40, styles)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for empty body, got %d", len(lines))
	}
}

func TestParseMarkdownLinesNumberedList(t *testing.T) {
	styles := testStyles()
	lines := ParseMarkdownLines("1. First item", 40, styles)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if lines[0].Spans[0].Text != "1. " {
		t.Fatalf("numbered list prefix = %q, want '1. '", lines[0].Spans[0].Text)
	}
}

func TestParseMarkdownLinesMultipleBlocks(t *testing.T) {
	styles := testStyles()
	body := "# Title\n\n- Item one\n- Item two\n\n> A quote"
	lines := ParseMarkdownLines(body, 40, styles)
	if len(lines) < 5 {
		t.Fatalf("expected at least 5 lines, got %d", len(lines))
	}
}
