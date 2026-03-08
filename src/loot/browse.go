package loot

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// BrowseAppraisalRecord is the latest-appraisal view used by the TUI.
type BrowseAppraisalRecord struct {
	ID                string
	AppraisedValue    int64
	Appraiser         string
	Notes             string
	AppraisedAt       string
	RecognizedEntryID string
}

// BrowseItemRecord is the loot browse row used by the TUI.
type BrowseItemRecord struct {
	ID              string
	Name            string
	Source          string
	Status          ledger.LootStatus
	Quantity        int
	Holder          string
	Notes           string
	AppraisalCount  int
	LatestAppraisal *BrowseAppraisalRecord
}

// ListBrowseItems returns held and recognized loot items with latest-appraisal detail.
func ListBrowseItems(ctx context.Context, databasePath string) ([]BrowseItemRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]BrowseItemRecord, error) {
		itemRows, err := db.QueryContext(ctx, `
			SELECT
				id,
				name,
				source,
				status,
				quantity,
				holder,
				notes
			FROM loot_items
			WHERE status IN ('held', 'recognized')
			ORDER BY status, name, created_at, id
		`)
		if err != nil {
			return nil, fmt.Errorf("query loot browse items: %w", err)
		}
		defer itemRows.Close()

		items := make([]BrowseItemRecord, 0)
		indexesByID := make(map[string]int)
		for itemRows.Next() {
			var item BrowseItemRecord
			var status string

			if err := itemRows.Scan(
				&item.ID,
				&item.Name,
				&item.Source,
				&status,
				&item.Quantity,
				&item.Holder,
				&item.Notes,
			); err != nil {
				return nil, fmt.Errorf("scan loot browse item: %w", err)
			}

			item.Status = ledger.LootStatus(status)
			if !item.Status.Valid() {
				return nil, fmt.Errorf("scan loot browse item: invalid loot status %q", status)
			}

			indexesByID[item.ID] = len(items)
			items = append(items, item)
		}

		if err := itemRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate loot browse items: %w", err)
		}

		if len(items) == 0 {
			return items, nil
		}

		appraisalRows, err := db.QueryContext(ctx, `
			SELECT
				la.id,
				la.loot_item_id,
				la.appraised_value,
				la.appraiser,
				la.notes,
				la.appraised_at,
				COALESCE(la.recognized_entry_id, '') AS recognized_entry_id
			FROM loot_appraisals la
			JOIN loot_items li ON li.id = la.loot_item_id
			WHERE li.status IN ('held', 'recognized')
			ORDER BY la.loot_item_id, la.appraised_at DESC, la.created_at DESC, la.id DESC
		`)
		if err != nil {
			return nil, fmt.Errorf("query loot browse appraisals: %w", err)
		}
		defer appraisalRows.Close()

		for appraisalRows.Next() {
			var appraisal BrowseAppraisalRecord
			var itemID string

			if err := appraisalRows.Scan(
				&appraisal.ID,
				&itemID,
				&appraisal.AppraisedValue,
				&appraisal.Appraiser,
				&appraisal.Notes,
				&appraisal.AppraisedAt,
				&appraisal.RecognizedEntryID,
			); err != nil {
				return nil, fmt.Errorf("scan loot browse appraisal: %w", err)
			}

			index, ok := indexesByID[itemID]
			if !ok {
				continue
			}

			items[index].AppraisalCount++
			if items[index].LatestAppraisal == nil {
				appraisalCopy := appraisal
				items[index].LatestAppraisal = &appraisalCopy
			}
		}

		if err := appraisalRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate loot browse appraisals: %w", err)
		}

		return items, nil
	})
}
