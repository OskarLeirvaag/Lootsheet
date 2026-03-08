package report

import (
	"context"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// LootSummaryRow represents a loot item with its latest appraisal.
type LootSummaryRow struct {
	ItemID               string
	Name                 string
	Source               string
	Status               ledger.LootStatus
	Quantity             int
	LatestAppraisalValue int64
	AppraisedAt          string
}

// GetLootSummary returns loot items with status 'held' or 'recognized',
// along with their latest appraisal value (if any).
func GetLootSummary(ctx context.Context, databasePath string) ([]LootSummaryRow, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return nil, err
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT
			li.id,
			li.name,
			li.source,
			li.status,
			li.quantity,
			COALESCE(la.appraised_value, 0) AS latest_appraisal_value,
			COALESCE(la.appraised_at, '') AS appraised_at
		FROM loot_items li
		LEFT JOIN loot_appraisals la ON la.loot_item_id = li.id
			AND la.appraised_at = (
				SELECT MAX(la2.appraised_at)
				FROM loot_appraisals la2
				WHERE la2.loot_item_id = li.id
			)
		WHERE li.status IN ('held', 'recognized')
		ORDER BY li.status, li.name
	`)
	if err != nil {
		return nil, fmt.Errorf("query loot summary: %w", err)
	}
	defer rows.Close()

	var result []LootSummaryRow
	for rows.Next() {
		var r LootSummaryRow
		var status string

		if err := rows.Scan(
			&r.ItemID, &r.Name, &r.Source, &status,
			&r.Quantity, &r.LatestAppraisalValue, &r.AppraisedAt,
		); err != nil {
			return nil, fmt.Errorf("scan loot summary row: %w", err)
		}

		r.Status = ledger.LootStatus(status)
		result = append(result, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate loot summary rows: %w", err)
	}

	return result, nil
}
