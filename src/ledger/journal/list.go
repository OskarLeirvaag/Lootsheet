package journal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// EntryRecord is a compact read-only journal row for list browsing.
type EntryRecord struct {
	ID              string
	EntryNumber     int
	Status          ledger.JournalEntryStatus
	EntryDate       string
	Description     string
	ReversesEntryID string
}

// ListEntries returns journal entries ordered newest first for read-only browsing.
func ListEntries(ctx context.Context, databasePath string) ([]EntryRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]EntryRecord, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT id, entry_number, status, entry_date, description, COALESCE(reverses_entry_id, '')
			FROM journal_entries
			ORDER BY entry_number DESC, id DESC
		`)
		if err != nil {
			return nil, fmt.Errorf("query journal entries: %w", err)
		}
		defer rows.Close()

		entries := []EntryRecord{}
		for rows.Next() {
			var entry EntryRecord
			var status string

			if err := rows.Scan(
				&entry.ID,
				&entry.EntryNumber,
				&status,
				&entry.EntryDate,
				&entry.Description,
				&entry.ReversesEntryID,
			); err != nil {
				return nil, fmt.Errorf("scan journal entry row: %w", err)
			}

			entry.Status = ledger.JournalEntryStatus(status)
			if !entry.Status.Valid() {
				return nil, fmt.Errorf("scan journal entry row: invalid journal status %q", status)
			}

			entries = append(entries, entry)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate journal entry rows: %w", err)
		}

		return entries, nil
	})
}
