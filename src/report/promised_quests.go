package report

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// PromisedQuestRow represents a quest promise that remains off-ledger because
// the quest has not yet been earned.
type PromisedQuestRow struct {
	QuestID         string
	Title           string
	Patron          string
	Status          ledger.QuestStatus
	PromisedReward  int64
	PartialAdvance  int64
	BonusConditions string
}

// GetPromisedQuests returns quests that are still in offered or accepted
// status, showing the promise details still tracked in the quest register.
func GetPromisedQuests(ctx context.Context, databasePath string) ([]PromisedQuestRow, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]PromisedQuestRow, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT
				q.id,
				q.title,
				q.patron,
				q.status,
				q.promised_base_reward,
				q.partial_advance,
				q.bonus_conditions
			FROM quests q
			WHERE q.status IN ('offered', 'accepted')
			ORDER BY
			  CASE q.status
			    WHEN 'accepted' THEN 1
			    WHEN 'offered' THEN 2
			  END,
			  COALESCE(q.accepted_on, q.created_at),
			  q.title
		`)
		if err != nil {
			return nil, fmt.Errorf("query promised quests: %w", err)
		}
		defer rows.Close()

		var result []PromisedQuestRow
		for rows.Next() {
			var row PromisedQuestRow
			var status string

			if err := rows.Scan(
				&row.QuestID,
				&row.Title,
				&row.Patron,
				&status,
				&row.PromisedReward,
				&row.PartialAdvance,
				&row.BonusConditions,
			); err != nil {
				return nil, fmt.Errorf("scan promised quest row: %w", err)
			}

			row.Status = ledger.QuestStatus(status)
			if !row.Status.Valid() {
				return nil, fmt.Errorf("scan promised quest row: invalid quest status %q", status)
			}

			result = append(result, row)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate promised quest rows: %w", err)
		}

		return result, nil
	})
}
