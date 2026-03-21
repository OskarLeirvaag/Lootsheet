package ledger

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// JournalPostInput holds the raw, unvalidated fields for posting a journal entry.
type JournalPostInput struct {
	EntryDate     string
	Description   string
	SourceQuestID string
	Lines         []JournalLineInput
}

// JournalLineInput represents a single line in a journal entry before validation.
// Exactly one of DebitAmount or CreditAmount must be positive.
type JournalLineInput struct {
	AccountCode  string
	Memo         string
	DebitAmount  int64
	CreditAmount int64
}

// JournalPostTotals holds the aggregate debit and credit amounts for a journal entry.
type JournalPostTotals struct {
	DebitAmount  int64
	CreditAmount int64
}

// ValidatedJournalPost is the result of successfully validating a JournalPostInput.
// It guarantees the entry is balanced and all fields are sanitized.
type ValidatedJournalPost struct {
	EntryDate   string
	Description string
	Lines       []JournalLineInput
	Totals      JournalPostTotals
}

// ValidateJournalPostInput checks that the journal post input is well-formed:
// the date is valid YYYY-MM-DD, a description is present, there are at least
// two lines with both debit and credit sides, and the entry balances.
func ValidateJournalPostInput(input JournalPostInput) (ValidatedJournalPost, error) {
	entryDate := strings.TrimSpace(input.EntryDate)
	if entryDate == "" {
		return ValidatedJournalPost{}, errors.New("journal entry date is required")
	}

	parsedDate, err := time.Parse("2006-01-02", entryDate)
	if err != nil {
		return ValidatedJournalPost{}, errors.New("journal entry date must use YYYY-MM-DD")
	}

	description := strings.TrimSpace(input.Description)
	if description == "" {
		return ValidatedJournalPost{}, errors.New("journal entry description is required")
	}

	if len(input.Lines) < 2 {
		return ValidatedJournalPost{}, errors.New("journal entry must have at least 2 lines")
	}

	lines := make([]JournalLineInput, 0, len(input.Lines))
	totals := JournalPostTotals{}
	debitLines := 0
	creditLines := 0

	for index, line := range input.Lines {
		accountCode := strings.TrimSpace(line.AccountCode)
		if accountCode == "" {
			return ValidatedJournalPost{}, fmt.Errorf("journal line %d account code is required", index+1)
		}

		if line.DebitAmount < 0 {
			return ValidatedJournalPost{}, fmt.Errorf("journal line %d debit amount must be positive", index+1)
		}

		if line.CreditAmount < 0 {
			return ValidatedJournalPost{}, fmt.Errorf("journal line %d credit amount must be positive", index+1)
		}

		switch {
		case line.DebitAmount > 0 && line.CreditAmount > 0:
			return ValidatedJournalPost{}, fmt.Errorf("journal line %d cannot contain both debit and credit amounts", index+1)
		case line.DebitAmount == 0 && line.CreditAmount == 0:
			return ValidatedJournalPost{}, fmt.Errorf("journal line %d must contain either a debit or credit amount", index+1)
		case line.DebitAmount > 0:
			debitLines++
			totals.DebitAmount += line.DebitAmount
		case line.CreditAmount > 0:
			creditLines++
			totals.CreditAmount += line.CreditAmount
		default:
		}

		lines = append(lines, JournalLineInput{
			AccountCode:  accountCode,
			Memo:         strings.TrimSpace(line.Memo),
			DebitAmount:  line.DebitAmount,
			CreditAmount: line.CreditAmount,
		})
	}

	if debitLines == 0 {
		return ValidatedJournalPost{}, errors.New("journal entry must contain at least one debit line")
	}

	if creditLines == 0 {
		return ValidatedJournalPost{}, errors.New("journal entry must contain at least one credit line")
	}

	if totals.DebitAmount != totals.CreditAmount {
		return ValidatedJournalPost{}, fmt.Errorf(
			"journal entry is not balanced: debit total %d != credit total %d",
			totals.DebitAmount,
			totals.CreditAmount,
		)
	}

	return ValidatedJournalPost{
		EntryDate:   parsedDate.Format("2006-01-02"),
		Description: description,
		Lines:       lines,
		Totals:      totals,
	}, nil
}
