package journal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// Summary describes the current journal activity at a glance.
type Summary struct {
	TotalEntries      int
	PostedEntries     int
	ReversedEntries   int
	ReversalEntries   int
	LatestEntryNumber int
	LatestEntryDate   string
	LatestDescription string
}

// GetSummary returns a compact journal summary for read-only dashboards.
func GetSummary(ctx context.Context, databasePath string) (Summary, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (Summary, error) {
		var summary Summary

		if err := db.QueryRowContext(ctx, `
			SELECT
				COUNT(*),
				COALESCE(SUM(CASE WHEN status = 'posted' THEN 1 ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN status = 'reversed' THEN 1 ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN reverses_entry_id IS NOT NULL THEN 1 ELSE 0 END), 0)
			FROM journal_entries
		`).Scan(
			&summary.TotalEntries,
			&summary.PostedEntries,
			&summary.ReversedEntries,
			&summary.ReversalEntries,
		); err != nil {
			return Summary{}, fmt.Errorf("query journal summary: %w", err)
		}

		if summary.TotalEntries == 0 {
			return summary, nil
		}

		if err := db.QueryRowContext(ctx, `
			SELECT entry_number, entry_date, description
			FROM journal_entries
			ORDER BY entry_number DESC
			LIMIT 1
		`).Scan(
			&summary.LatestEntryNumber,
			&summary.LatestEntryDate,
			&summary.LatestDescription,
		); err != nil {
			return Summary{}, fmt.Errorf("query latest journal entry: %w", err)
		}

		return summary, nil
	})
}
