package ddb

import (
	"regexp"
	"strings"
)

// HTMLToMarkdown converts simple DDB HTML to markdown for TUI display.
// Handles: p, strong, em, h1-h3, ul/ol/li, hr, br. Strips other tags.
func HTMLToMarkdown(html string) string {
	if html == "" {
		return ""
	}

	s := html

	// Normalize line breaks.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Block elements → newlines.
	s = reReplace(s, `<br\s*/?>`, "\n")
	s = reReplace(s, `<hr\s*/?>`, "\n---\n")

	// Headings.
	s = reReplace(s, `<h1[^>]*>`, "\n# ")
	s = reReplace(s, `</h1>`, "\n")
	s = reReplace(s, `<h2[^>]*>`, "\n## ")
	s = reReplace(s, `</h2>`, "\n")
	s = reReplace(s, `<h3[^>]*>`, "\n### ")
	s = reReplace(s, `</h3>`, "\n")

	// Lists.
	s = reReplace(s, `<ul[^>]*>`, "\n")
	s = reReplace(s, `</ul>`, "\n")
	s = reReplace(s, `<ol[^>]*>`, "\n")
	s = reReplace(s, `</ol>`, "\n")
	s = reReplace(s, `<li[^>]*>`, "- ")
	s = reReplace(s, `</li>`, "\n")

	// Inline formatting.
	s = reReplace(s, `<strong[^>]*>`, "**")
	s = reReplace(s, `</strong>`, "**")
	s = reReplace(s, `<em[^>]*>`, "*")
	s = reReplace(s, `</em>`, "*")
	s = reReplace(s, `<b[^>]*>`, "**")
	s = reReplace(s, `</b>`, "**")
	s = reReplace(s, `<i[^>]*>`, "*")
	s = reReplace(s, `</i>`, "*")

	// Paragraphs → double newline.
	s = reReplace(s, `<p[^>]*>`, "\n")
	s = reReplace(s, `</p>`, "\n")

	// Strip remaining tags.
	s = reReplace(s, `<[^>]+>`, "")

	// Decode common HTML entities.
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&ndash;", "–")
	s = strings.ReplaceAll(s, "&mdash;", "—")
	s = strings.ReplaceAll(s, "&nbsp;", " ")

	// Clean up excessive whitespace.
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")
	s = strings.TrimSpace(s)

	return s
}

func reReplace(s, pattern, replacement string) string {
	return regexp.MustCompile(pattern).ReplaceAllString(s, replacement)
}
