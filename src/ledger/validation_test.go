package ledger

import (
	"strings"
	"testing"
)

func TestValidateJournalPostInputNormalizesBalancedEntry(t *testing.T) {
	validated, err := ValidateJournalPostInput(JournalPostInput{
		EntryDate:   " 2026-03-08 ",
		Description: " Restock arrows ",
		Lines: []JournalLineInput{
			{AccountCode: " 5100 ", DebitAmount: 25, Memo: " Quiver refill "},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("validate journal post input: %v", err)
	}

	if validated.EntryDate != "2026-03-08" {
		t.Fatalf("entry date = %q, want 2026-03-08", validated.EntryDate)
	}

	if validated.Description != "Restock arrows" {
		t.Fatalf("description = %q, want Restock arrows", validated.Description)
	}

	if validated.Lines[0].AccountCode != "5100" {
		t.Fatalf("first account code = %q, want 5100", validated.Lines[0].AccountCode)
	}

	if validated.Lines[0].Memo != "Quiver refill" {
		t.Fatalf("first memo = %q, want Quiver refill", validated.Lines[0].Memo)
	}

	if validated.Totals.DebitAmount != 25 || validated.Totals.CreditAmount != 25 {
		t.Fatalf("totals = %+v, want 25/25", validated.Totals)
	}
}

func TestValidateJournalPostInputRejectsUnbalancedEntry(t *testing.T) {
	_, err := ValidateJournalPostInput(JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Unbalanced test",
		Lines: []JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 20},
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "journal entry is not balanced") {
		t.Fatalf("error = %q, want balance error", err)
	}
}

func TestValidateJournalPostInputRejectsLineWithoutAmounts(t *testing.T) {
	_, err := ValidateJournalPostInput(JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Missing amount",
		Lines: []JournalLineInput{
			{AccountCode: "5100"},
			{AccountCode: "1000", CreditAmount: 10},
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "must contain either a debit or credit amount") {
		t.Fatalf("error = %q, want missing amount error", err)
	}
}
