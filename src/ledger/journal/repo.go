// Package journal provides repository and CLI handler functions for journal
// entries: posting balanced entries, reversing posted entries, mutability
// checks, and per-account ledger reporting.
package journal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// PostJournalEntry validates, resolves accounts, and posts a balanced journal entry.
func PostJournalEntry(ctx context.Context, databasePath string, input ledger.JournalPostInput) (ledger.PostedJournalEntry, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin journal post transaction: %w", err)
		}
		defer tx.Rollback()

		posted, err := ledger.PostJournalWithinTx(ctx, db, tx, input)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		if err := tx.Commit(); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("commit journal post transaction: %w", err)
		}

		return posted, nil
	})
}

// ReverseJournalEntry creates a new posted journal entry that zeroes out the
// original entry by swapping debits and credits. The original entry's status
// is set to 'reversed' and its reversed_at timestamp is recorded.
// The original entry must exist and have status='posted'.
func ReverseJournalEntry(ctx context.Context, databasePath string, originalEntryID string, reversalDate string, description string) (ledger.PostedJournalEntry, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
		// Verify the original entry exists and is posted.
		var originalStatus string
		var originalEntryNumber int
		if err := db.QueryRowContext(ctx,
			"SELECT status, entry_number FROM journal_entries WHERE id = ?", originalEntryID,
		).Scan(&originalStatus, &originalEntryNumber); err != nil {
			if err == sql.ErrNoRows {
				return ledger.PostedJournalEntry{}, fmt.Errorf("journal entry %q does not exist", originalEntryID)
			}
			return ledger.PostedJournalEntry{}, fmt.Errorf("query original journal entry: %w", err)
		}

		if originalStatus != string(ledger.JournalEntryStatusPosted) {
			return ledger.PostedJournalEntry{}, ledger.ErrEntryNotReversible
		}

		// Load the original entry's lines.
		type originalLine struct {
			AccountID    string
			Memo         string
			DebitAmount  int64
			CreditAmount int64
		}

		rows, err := db.QueryContext(ctx,
			"SELECT account_id, memo, debit_amount, credit_amount FROM journal_lines WHERE journal_entry_id = ? ORDER BY line_number",
			originalEntryID,
		)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("query original journal lines: %w", err)
		}
		defer rows.Close()

		var lines []originalLine
		for rows.Next() {
			var l originalLine
			if err := rows.Scan(&l.AccountID, &l.Memo, &l.DebitAmount, &l.CreditAmount); err != nil {
				return ledger.PostedJournalEntry{}, fmt.Errorf("scan original journal line: %w", err)
			}
			lines = append(lines, l)
		}
		if err := rows.Err(); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("iterate original journal lines: %w", err)
		}

		if len(lines) == 0 {
			return ledger.PostedJournalEntry{}, fmt.Errorf("original journal entry %q has no lines", originalEntryID)
		}

		// Default description if not provided.
		if description == "" {
			description = fmt.Sprintf("Reversal of entry #%d", originalEntryNumber)
		}

		entryNumber, err := ledger.NextJournalEntryNumber(ctx, db)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		reversalEntryID := uuid.NewString()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin reversal transaction: %w", err)
		}
		defer tx.Rollback()

		// Create the reversal entry.
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO journal_entries (id, entry_number, status, entry_date, description, reverses_entry_id, posted_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
			reversalEntryID, entryNumber, "posted", reversalDate, description, originalEntryID,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("insert reversal journal entry: %w", err)
		}

		// Create reversed lines (debits become credits and vice versa).
		var debitTotal, creditTotal int64
		for index, line := range lines {
			swappedDebit := line.CreditAmount
			swappedCredit := line.DebitAmount
			debitTotal += swappedDebit
			creditTotal += swappedCredit

			if _, err := tx.ExecContext(ctx,
				"INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
				uuid.NewString(), reversalEntryID, index+1, line.AccountID, line.Memo, swappedDebit, swappedCredit,
			); err != nil {
				return ledger.PostedJournalEntry{}, fmt.Errorf("insert reversal journal line %d: %w", index+1, err)
			}
		}

		// Mark the original entry as reversed.
		if _, err := tx.ExecContext(ctx,
			"UPDATE journal_entries SET status = ?, reversed_at = CURRENT_TIMESTAMP WHERE id = ?",
			"reversed", originalEntryID,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("update original entry status: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("commit reversal transaction: %w", err)
		}

		return ledger.PostedJournalEntry{
			ID:          reversalEntryID,
			EntryNumber: entryNumber,
			EntryDate:   reversalDate,
			Description: description,
			LineCount:   len(lines),
			DebitTotal:  debitTotal,
			CreditTotal: creditTotal,
		}, nil
	})
}
