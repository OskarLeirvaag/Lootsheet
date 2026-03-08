// Package tools provides shared utility functions for the LootSheet application,
// including D&D 5e currency parsing and formatting.
package tools

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Currency denominations and their copper piece (CP) equivalents.
const (
	// CPPerCP is the number of copper pieces in one copper piece.
	CPPerCP = 1
	// CPPerSP is the number of copper pieces in one silver piece.
	CPPerSP = 10
	// CPPerEP is the number of copper pieces in one electrum piece.
	CPPerEP = 50
	// CPPerGP is the number of copper pieces in one gold piece.
	CPPerGP = 100
	// CPPerPP is the number of copper pieces in one platinum piece.
	CPPerPP = 1000
)

// denominationPattern matches a single denomination token like "23.432PP" or "5SP".
var denominationPattern = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(pp|gp|ep|sp|cp)`)

// bareIntegerPattern matches a string that is only digits (bare integer = CP).
var bareIntegerPattern = regexp.MustCompile(`^\d+$`)

// ParseAmount parses a D&D 5e currency string into copper pieces (int64).
//
// Supported formats:
//   - Single denomination: "23.432PP", "5.5GP", "100CP", "3SP"
//   - Mixed denominations: "2gp5sp", "1pp2gp3sp5cp", "2GP 5SP"
//   - Bare integer (backwards compatible): "100" treated as CP
//   - Case insensitive
//
// Returns an error if the input is empty, negative, or results in fractional CP.
func ParseAmount(input string) (int64, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, fmt.Errorf("amount is required")
	}

	// Reject negative values.
	if strings.HasPrefix(trimmed, "-") {
		return 0, fmt.Errorf("amount must not be negative: %q", input)
	}

	// Bare integer: treat as CP.
	if bareIntegerPattern.MatchString(trimmed) {
		cp, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount %q: %w", input, err)
		}
		if cp < 0 {
			return 0, fmt.Errorf("amount must not be negative: %q", input)
		}
		return cp, nil
	}

	// Try matching denomination tokens.
	matches := denominationPattern.FindAllStringSubmatchIndex(trimmed, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid amount %q: expected a number with denomination (PP, GP, EP, SP, CP) or a bare integer", input)
	}

	// Verify that the matches cover the entire string (ignoring spaces).
	covered := make([]bool, len(trimmed))
	for _, match := range matches {
		for i := match[0]; i < match[1]; i++ {
			covered[i] = true
		}
	}
	for i, c := range trimmed {
		if !covered[i] && c != ' ' {
			return 0, fmt.Errorf("invalid amount %q: unexpected character at position %d", input, i)
		}
	}

	var totalCP float64
	for _, match := range matches {
		numStr := trimmed[match[2]:match[3]]
		denomStr := strings.ToUpper(trimmed[match[4]:match[5]])

		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount %q: bad number %q", input, numStr)
		}

		if num < 0 {
			return 0, fmt.Errorf("amount must not be negative: %q", input)
		}

		var cpMultiplier int
		switch denomStr {
		case "PP":
			cpMultiplier = CPPerPP
		case "GP":
			cpMultiplier = CPPerGP
		case "EP":
			cpMultiplier = CPPerEP
		case "SP":
			cpMultiplier = CPPerSP
		case "CP":
			cpMultiplier = CPPerCP
		default:
			return 0, fmt.Errorf("invalid denomination %q in amount %q", denomStr, input)
		}

		totalCP += num * float64(cpMultiplier)
	}

	// Check that the result is a whole number of CP.
	rounded := math.Round(totalCP)
	if math.Abs(totalCP-rounded) > 0.0001 {
		return 0, fmt.Errorf("amount %q does not resolve to a whole number of copper pieces (got %.2f CP)", input, totalCP)
	}

	result := int64(rounded)
	if result < 0 {
		return 0, fmt.Errorf("amount must not be negative: %q", input)
	}

	return result, nil
}

// FormatAmount converts copper pieces to a readable mixed denomination string.
// Uses PP, GP, SP, CP for output (EP is accepted on input but not used for display).
// Greedy breakdown: PP first, then GP, SP, CP. Zero denominations are skipped.
func FormatAmount(cp int64) string {
	if cp == 0 {
		return "0 CP"
	}

	negative := false
	if cp < 0 {
		negative = true
		cp = -cp
	}

	var parts []string

	if pp := cp / CPPerPP; pp > 0 {
		parts = append(parts, fmt.Sprintf("%d PP", pp))
		cp %= CPPerPP
	}

	if gp := cp / CPPerGP; gp > 0 {
		parts = append(parts, fmt.Sprintf("%d GP", gp))
		cp %= CPPerGP
	}

	if sp := cp / CPPerSP; sp > 0 {
		parts = append(parts, fmt.Sprintf("%d SP", sp))
		cp %= CPPerSP
	}

	if cp > 0 {
		parts = append(parts, fmt.Sprintf("%d CP", cp))
	}

	result := strings.Join(parts, " ")
	if negative {
		result = "-" + result
	}

	return result
}
