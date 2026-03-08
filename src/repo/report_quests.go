package repo

import (
	"context"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

// QuestReceivableRow represents a quest with an outstanding receivable balance.
type QuestReceivableRow struct {
	QuestID        string
	Title          string
	Patron         string
	Status         service.QuestStatus
	PromisedReward int64
	TotalPaid      int64
	Outstanding    int64
}

// GetQuestReceivables returns quests in completed/collectible/partially_paid status
// that have outstanding receivable balances (promised reward minus total collected > 0).
// Total collected is computed by summing debit amounts against account 1000 (Party Cash)
// from posted journal entries whose description matches "Quest payment: <title>%".
func GetQuestReceivables(ctx context.Context, databasePath string) ([]QuestReceivableRow, error) {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return nil, err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

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
				WHERE je.description LIKE 'Quest payment: ' || q.title || '%'
				  AND je.status = 'posted'
				  AND jl.account_id = (SELECT id FROM accounts WHERE code = '1000')
			), 0) AS total_paid
		FROM quests q
		WHERE q.status IN ('completed', 'collectible', 'partially_paid')
		  AND q.promised_base_reward > 0
		HAVING (q.promised_base_reward - total_paid) > 0
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

		r.Status = service.QuestStatus(status)
		r.Outstanding = r.PromisedReward - r.TotalPaid
		if r.Outstanding < 0 {
			r.Outstanding = 0
		}

		result = append(result, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quest receivable rows: %w", err)
	}

	return result, nil
}
