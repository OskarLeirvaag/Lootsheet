package report

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// QuestReceivableRow represents a quest with an outstanding receivable balance.
type QuestReceivableRow struct {
	QuestID        string
	Title          string
	Patron         string
	Status         ledger.QuestStatus
	PromisedReward int64
	TotalPaid      int64
	Outstanding    int64
}

// GetQuestReceivables returns quests in completed/collectible/partially_paid status
// that have outstanding receivable balances (promised reward minus total collected > 0).
// Total collected is computed by summing debit amounts against account 1000 (Party Cash)
// from posted journal entries whose description matches "Quest payment: <title>%".
func GetQuestReceivables(ctx context.Context, databasePath string) ([]QuestReceivableRow, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]QuestReceivableRow, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT
				q.id,
				q.title,
				q.patron,
				q.status,
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
			  AND q.promised_base_reward > 0
			ORDER BY q.status, q.title
		`)
		if err != nil {
			return nil, fmt.Errorf("query quest receivables: %w", err)
		}
		defer rows.Close()

		var result []QuestReceivableRow
		for rows.Next() {
			var r QuestReceivableRow
			var status string

			if err := rows.Scan(
				&r.QuestID, &r.Title, &r.Patron, &status,
				&r.PromisedReward, &r.TotalPaid,
			); err != nil {
				return nil, fmt.Errorf("scan quest receivable row: %w", err)
			}

			r.Status = ledger.QuestStatus(status)
			r.Outstanding = r.PromisedReward - r.TotalPaid
			if r.Outstanding <= 0 {
				continue
			}

			result = append(result, r)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate quest receivable rows: %w", err)
		}

		return result, nil
	})
}
