package report

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

const reportDateLayout = "2006-01-02"

// WriteOffCandidateFilter controls which stale receivables are returned.
type WriteOffCandidateFilter struct {
	AsOfDate   string
	MinAgeDays int
}

// WriteOffCandidateRow represents a completed quest with an old, uncollected
// balance that may be a write-off candidate.
type WriteOffCandidateRow struct {
	QuestID        string
	Title          string
	Patron         string
	Status         ledger.QuestStatus
	CompletedOn    string
	PromisedReward int64
	TotalPaid      int64
	Outstanding    int64
	AgeDays        int
}

// GetWriteOffCandidates returns completed, collectible, or partially paid
// quests with outstanding balances whose completed date is at least the
// requested age threshold.
func GetWriteOffCandidates(ctx context.Context, databasePath string, filter WriteOffCandidateFilter) ([]WriteOffCandidateRow, error) {
	asOfDate := strings.TrimSpace(filter.AsOfDate)
	if asOfDate == "" {
		return nil, fmt.Errorf("as_of date is required")
	}
	if filter.MinAgeDays < 0 {
		return nil, fmt.Errorf("min_age_days must be non-negative")
	}

	asOf, err := time.Parse(reportDateLayout, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("parse as_of date %q: %w", asOfDate, err)
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]WriteOffCandidateRow, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT
				q.id,
				q.title,
				q.patron,
				q.status,
				COALESCE(q.completed_on, ''),
				q.promised_base_reward,
				COALESCE((
					SELECT SUM(jl.debit_amount)
					FROM journal_lines jl
					JOIN journal_entries je ON je.id = jl.journal_entry_id
					JOIN accounts a ON a.id = jl.account_id
					WHERE je.status = 'posted'
					  AND a.code = '1000'
					  AND jl.memo = 'Quest payment: ' || q.title
				), 0) AS total_paid
			FROM quests q
			WHERE q.status IN ('completed', 'collectible', 'partially_paid')
			  AND q.completed_on IS NOT NULL
			  AND q.promised_base_reward > 0
			ORDER BY q.completed_on, q.title
		`)
		if err != nil {
			return nil, fmt.Errorf("query write-off candidates: %w", err)
		}
		defer rows.Close()

		var result []WriteOffCandidateRow
		for rows.Next() {
			var row WriteOffCandidateRow
			var status string

			if err := rows.Scan(
				&row.QuestID,
				&row.Title,
				&row.Patron,
				&status,
				&row.CompletedOn,
				&row.PromisedReward,
				&row.TotalPaid,
			); err != nil {
				return nil, fmt.Errorf("scan write-off candidate row: %w", err)
			}

			row.Status = ledger.QuestStatus(status)
			if !row.Status.Valid() {
				return nil, fmt.Errorf("scan write-off candidate row: invalid quest status %q", status)
			}

			completedOn, err := time.Parse(reportDateLayout, row.CompletedOn)
			if err != nil {
				return nil, fmt.Errorf("parse completed_on date %q for quest %q: %w", row.CompletedOn, row.Title, err)
			}
			if completedOn.After(asOf) {
				continue
			}

			row.AgeDays = int(asOf.Sub(completedOn).Hours() / 24)
			row.Outstanding = row.PromisedReward - row.TotalPaid
			if row.Outstanding <= 0 || row.AgeDays < filter.MinAgeDays {
				continue
			}

			result = append(result, row)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate write-off candidate rows: %w", err)
		}

		return result, nil
	})
}
