package journal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// BrowseEntryLine is a journal line enriched for TUI detail rendering.
type BrowseEntryLine struct {
	LineNumber   int
	AccountCode  string
	AccountName  string
	Memo         string
	DebitAmount  int64
	CreditAmount int64
}

// BrowseEntryRecord is a journal row enriched with line detail and reversal linkage.
type BrowseEntryRecord struct {
	ID                    string
	EntryNumber           int
	Status                ledger.JournalEntryStatus
	EntryDate             string
	Description           string
	ReversesEntryID       string
	ReversesEntryNumber   int
	ReversedByEntryID     string
	ReversedByEntryNumber int
	Lines                 []BrowseEntryLine
}

// ListBrowseEntries returns journal entries ordered newest first with line detail
// and reversal linkage for TUI browsing.
func ListBrowseEntries(ctx context.Context, databasePath string, campaignID string) ([]BrowseEntryRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]BrowseEntryRecord, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT
				je.id,
				je.entry_number,
				je.status,
				je.entry_date,
				je.description,
				COALESCE(je.reverses_entry_id, ''),
				COALESCE(reversed_entry.entry_number, 0),
				COALESCE(reversal_entry.id, ''),
				COALESCE(reversal_entry.entry_number, 0)
			FROM journal_entries je
			LEFT JOIN journal_entries reversed_entry ON reversed_entry.id = je.reverses_entry_id
			LEFT JOIN journal_entries reversal_entry ON reversal_entry.reverses_entry_id = je.id
			WHERE je.campaign_id = ?
			ORDER BY je.entry_number DESC, je.id DESC
		`, campaignID)
		if err != nil {
			return nil, fmt.Errorf("query journal browse entries: %w", err)
		}
		defer rows.Close()

		entries := []BrowseEntryRecord{}
		indexByID := map[string]int{}
		for rows.Next() {
			var entry BrowseEntryRecord
			var status string

			if err := rows.Scan(
				&entry.ID,
				&entry.EntryNumber,
				&status,
				&entry.EntryDate,
				&entry.Description,
				&entry.ReversesEntryID,
				&entry.ReversesEntryNumber,
				&entry.ReversedByEntryID,
				&entry.ReversedByEntryNumber,
			); err != nil {
				return nil, fmt.Errorf("scan journal browse entry: %w", err)
			}

			entry.Status = ledger.JournalEntryStatus(status)
			if !entry.Status.Valid() {
				return nil, fmt.Errorf("scan journal browse entry: invalid journal status %q", status)
			}

			indexByID[entry.ID] = len(entries)
			entries = append(entries, entry)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate journal browse entries: %w", err)
		}
		if len(entries) == 0 {
			return entries, nil
		}

		lineRows, err := db.QueryContext(ctx, `
			SELECT
				jl.journal_entry_id,
				jl.line_number,
				a.code,
				a.name,
				jl.memo,
				jl.debit_amount,
				jl.credit_amount
			FROM journal_lines jl
			JOIN accounts a ON a.id = jl.account_id
			JOIN journal_entries je ON je.id = jl.journal_entry_id
			WHERE je.campaign_id = ?
			ORDER BY je.entry_number DESC, je.id DESC, jl.line_number ASC
		`, campaignID)
		if err != nil {
			return nil, fmt.Errorf("query journal browse lines: %w", err)
		}
		defer lineRows.Close()

		for lineRows.Next() {
			var entryID string
			var line BrowseEntryLine

			if err := lineRows.Scan(
				&entryID,
				&line.LineNumber,
				&line.AccountCode,
				&line.AccountName,
				&line.Memo,
				&line.DebitAmount,
				&line.CreditAmount,
			); err != nil {
				return nil, fmt.Errorf("scan journal browse line: %w", err)
			}

			index, ok := indexByID[entryID]
			if !ok {
				return nil, fmt.Errorf("scan journal browse line: entry %q missing from browse set", entryID)
			}

			entries[index].Lines = append(entries[index].Lines, line)
		}
		if err := lineRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate journal browse lines: %w", err)
		}

		return entries, nil
	})
}
