package ledger

import "strings"

// FTSQuery converts a user search string into a safe FTS5 prefix query.
// Each word is quoted (with embedded quotes escaped) and given a `*` suffix
// for prefix matching. This prevents FTS5 operator injection (AND, OR, NOT,
// NEAR, column filters, etc.) and makes partial-word queries work naturally
// — typing "drag" matches "dragon", "Moon" matches "Moonwhisper".
//
// Returns an empty string if the query has no searchable words.
func FTSQuery(q string) string {
	words := strings.Fields(q)
	if len(words) == 0 {
		return ""
	}
	parts := make([]string, 0, len(words))
	for _, w := range words {
		w = strings.ReplaceAll(w, `"`, `""`)
		parts = append(parts, `"`+w+`"*`)
	}
	return strings.Join(parts, " ")
}

// LIKEPattern returns a SQL LIKE pattern matching the query as a substring.
// Backslash, percent, and underscore are escaped. Use with ESCAPE '\' in SQL.
func LIKEPattern(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, "%", `\%`)
	q = strings.ReplaceAll(q, "_", `\_`)
	return "%" + q + "%"
}
