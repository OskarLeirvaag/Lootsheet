package ddb

import (
	"html"
	"regexp"
	"strings"
)

// Pre-compiled regex patterns for HTML→markdown conversion.
var (
	reBR         = regexp.MustCompile(`<br\s*/?>`)
	reHR         = regexp.MustCompile(`<hr\s*/?>`)
	reH1Open     = regexp.MustCompile(`<h1[^>]*>`)
	reH1Close    = regexp.MustCompile(`</h1>`)
	reH2Open     = regexp.MustCompile(`<h2[^>]*>`)
	reH2Close    = regexp.MustCompile(`</h2>`)
	reH3Open     = regexp.MustCompile(`<h3[^>]*>`)
	reH3Close    = regexp.MustCompile(`</h3>`)
	reULOpen     = regexp.MustCompile(`<ul[^>]*>`)
	reULClose    = regexp.MustCompile(`</ul>`)
	reOLOpen     = regexp.MustCompile(`<ol[^>]*>`)
	reOLClose    = regexp.MustCompile(`</ol>`)
	reLIOpen     = regexp.MustCompile(`<li[^>]*>`)
	reLIClose    = regexp.MustCompile(`</li>`)
	reStrong     = regexp.MustCompile(`</?strong[^>]*>`)
	reEm         = regexp.MustCompile(`</?em[^>]*>`)
	reBold       = regexp.MustCompile(`</?b[^>]*>`)
	reItalic     = regexp.MustCompile(`</?i[^>]*>`)
	rePOpen      = regexp.MustCompile(`<p[^>]*>`)
	rePClose     = regexp.MustCompile(`</p>`)
	reStripTags  = regexp.MustCompile(`<[^>]+>`)
	reExcessNL   = regexp.MustCompile(`\n{3,}`)
)

// HTMLToMarkdown converts simple DDB HTML to markdown for TUI display.
// Handles: p, strong, em, h1-h3, ul/ol/li, hr, br. Strips other tags.
func HTMLToMarkdown(s string) string {
	if s == "" {
		return ""
	}

	// Normalize line breaks.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Block elements → newlines.
	s = reBR.ReplaceAllString(s, "\n")
	s = reHR.ReplaceAllString(s, "\n---\n")

	// Headings.
	s = reH1Open.ReplaceAllString(s, "\n# ")
	s = reH1Close.ReplaceAllString(s, "\n")
	s = reH2Open.ReplaceAllString(s, "\n## ")
	s = reH2Close.ReplaceAllString(s, "\n")
	s = reH3Open.ReplaceAllString(s, "\n### ")
	s = reH3Close.ReplaceAllString(s, "\n")

	// Lists.
	s = reULOpen.ReplaceAllString(s, "\n")
	s = reULClose.ReplaceAllString(s, "\n")
	s = reOLOpen.ReplaceAllString(s, "\n")
	s = reOLClose.ReplaceAllString(s, "\n")
	s = reLIOpen.ReplaceAllString(s, "- ")
	s = reLIClose.ReplaceAllString(s, "\n")

	// Inline formatting.
	s = reStrong.ReplaceAllString(s, "**")
	s = reEm.ReplaceAllString(s, "*")
	s = reBold.ReplaceAllString(s, "**")
	s = reItalic.ReplaceAllString(s, "*")

	// Paragraphs → double newline.
	s = rePOpen.ReplaceAllString(s, "\n")
	s = rePClose.ReplaceAllString(s, "\n")

	// Strip remaining tags.
	s = reStripTags.ReplaceAllString(s, "")

	// Decode all HTML entities (named, numeric, hex).
	s = html.UnescapeString(s)

	// Clean up excessive whitespace.
	s = reExcessNL.ReplaceAllString(s, "\n\n")
	s = strings.TrimSpace(s)

	return s
}
