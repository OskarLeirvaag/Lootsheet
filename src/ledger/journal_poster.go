package ledger

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// PostJournalWithinTx validates a JournalPostInput, resolves account codes,
// allocates an entry number, and inserts the journal entry and lines within
// the provided transaction. The caller is responsible for committing.
//
// Read-only queries (account resolution, next entry number) use db so they
// do not widen the write transaction's lock scope. All inserts use tx.
func PostJournalWithinTx(ctx context.Context, db *sql.DB, tx *sql.Tx, campaignID string, input JournalPostInput) (PostedJournalEntry, error) {
	validated, err := ValidateJournalPostInput(input)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	accountIDsByCode, err := ResolveActiveAccountIDsByCode(ctx, db, campaignID, validated.Lines)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	entryNumber, err := NextJournalEntryNumber(ctx, db, campaignID)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	entryID := uuid.NewString()

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries (id, campaign_id, entry_number, status, entry_date, description, posted_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		entryID, campaignID, entryNumber, "posted", validated.EntryDate, validated.Description,
	); err != nil {
		return PostedJournalEntry{}, fmt.Errorf("insert journal entry: %w", err)
	}

	for index, line := range validated.Lines {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
			uuid.NewString(), entryID, index+1, accountIDsByCode[line.AccountCode], line.Memo, line.DebitAmount, line.CreditAmount,
		); err != nil {
			return PostedJournalEntry{}, fmt.Errorf("insert journal line %d: %w", index+1, err)
		}
	}

	return PostedJournalEntry{
		ID:          entryID,
		EntryNumber: entryNumber,
		EntryDate:   validated.EntryDate,
		Description: validated.Description,
		LineCount:   len(validated.Lines),
		DebitTotal:  validated.Totals.DebitAmount,
		CreditTotal: validated.Totals.CreditAmount,
	}, nil
}
