package ledger

import "testing"

func TestFTSQuery(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"   ", ""},
		{"dragon", `"dragon"*`},
		{"Bryn Sander", `"Bryn"* "Sander"*`},
		{`say "hi"`, `"say"* """hi"""*`},          // escape embedded quotes
		{"drag*", `"drag*"*`},                     // already-* words are quoted literally
		{"person/Bryn", `"person/Bryn"*`},         // slashes are safe inside quotes
		{"@[person", `"@[person"*`},               // brackets are safe inside quotes
		{"AND OR NOT", `"AND"* "OR"* "NOT"*`},     // FTS5 operators disarmed
	}
	for _, tt := range tests {
		got := FTSQuery(tt.input)
		if got != tt.want {
			t.Errorf("FTSQuery(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLIKEPattern(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "%%"},
		{"dragon", "%dragon%"},
		{"50%", `%50\%%`},       // percent escaped
		{"under_score", `%under\_score%`},
		{`back\slash`, `%back\\slash%`},
	}
	for _, tt := range tests {
		got := LIKEPattern(tt.input)
		if got != tt.want {
			t.Errorf("LIKEPattern(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
