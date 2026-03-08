package service

import (
	"testing"
)

func TestParseAmountSingleDenominations(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100CP", 100},
		{"25cp", 25},
		{"3SP", 30},
		{"5sp", 50},
		{"1EP", 50},
		{"2ep", 100},
		{"1GP", 100},
		{"5gp", 500},
		{"1PP", 1000},
		{"3pp", 3000},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAmountDecimalDenominations(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"23.432PP", 23432},
		{"5.5GP", 550},
		{"0.5SP", 5},
		{"0.1GP", 10},
		{"2.5EP", 125},
		{"1.5PP", 1500},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAmountMixedDenominations(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"2gp5sp", 250},
		{"1pp2gp3sp5cp", 1235},
		{"2GP 5SP", 250},
		{"1PP 2GP 3SP 5CP", 1235},
		{"10gp20sp5cp", 1205},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAmountBareIntegers(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100", 100},
		{"0", 0},
		{"1", 1},
		{"999999", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAmountCaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"5gp", 500},
		{"5GP", 500},
		{"5Gp", 500},
		{"5gP", 500},
		{"1Pp", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAmountRejectsInvalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"", "empty string"},
		{"abc", "non-numeric"},
		{"-5GP", "negative amount"},
		{"2.31EP", "fractional CP from EP"},
		{"0.33SP", "fractional CP from SP"},
		{"0.0015PP", "fractional CP from PP"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := ParseAmount(tt.input)
			if err == nil {
				t.Fatalf("ParseAmount(%q) should have returned an error for %s", tt.input, tt.desc)
			}
		})
	}
}

func TestParseAmountWhitespace(t *testing.T) {
	got, err := ParseAmount("  5GP  ")
	if err != nil {
		t.Fatalf("ParseAmount with whitespace error: %v", err)
	}
	if got != 500 {
		t.Fatalf("ParseAmount with whitespace = %d, want 500", got)
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		cp   int64
		want string
	}{
		{0, "0 CP"},
		{1, "1 CP"},
		{10, "1 SP"},
		{50, "5 SP"},
		{100, "1 GP"},
		{550, "5 GP 5 SP"},
		{1000, "1 PP"},
		{1235, "1 PP 2 GP 3 SP 5 CP"},
		{23432, "23 PP 4 GP 3 SP 2 CP"},
		{25, "2 SP 5 CP"},
		{500, "5 GP"},
		{200, "2 GP"},
		{1100, "1 PP 1 GP"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatAmount(tt.cp)
			if got != tt.want {
				t.Fatalf("FormatAmount(%d) = %q, want %q", tt.cp, got, tt.want)
			}
		})
	}
}

func TestFormatAmountNegative(t *testing.T) {
	got := FormatAmount(-550)
	if got != "-5 GP 5 SP" {
		t.Fatalf("FormatAmount(-550) = %q, want %q", got, "-5 GP 5 SP")
	}
}

func TestParseFormatRoundTrip(t *testing.T) {
	// Parse a string, format the result, parse again, verify same value.
	tests := []struct {
		input string
		cp    int64
	}{
		{"23.432PP", 23432},
		{"5.5GP", 550},
		{"2gp5sp", 250},
		{"1pp2gp3sp5cp", 1235},
		{"100", 100},
		{"25cp", 25},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsed, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount(%q) error: %v", tt.input, err)
			}
			if parsed != tt.cp {
				t.Fatalf("ParseAmount(%q) = %d, want %d", tt.input, parsed, tt.cp)
			}

			formatted := FormatAmount(parsed)

			// Parse the formatted output back.
			reparsed, err := ParseAmount(formatted)
			if err != nil {
				t.Fatalf("ParseAmount(%q) (round-trip) error: %v", formatted, err)
			}
			if reparsed != parsed {
				t.Fatalf("round-trip mismatch: %q -> %d -> %q -> %d", tt.input, parsed, formatted, reparsed)
			}
		})
	}
}
