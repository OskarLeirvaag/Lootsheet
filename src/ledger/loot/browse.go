package loot

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	ID                       string
	Name                     string
	Source                   string
	Status                   ledger.LootStatus
	ItemType                 string
	Quantity                 int
	Holder                   string
	Notes                    string
	AppraisalCount           int
	HasRecognizedAppraisal   bool
	RecognizedAppraisalValue int64
	LatestAppraisal          *BrowseAppraisalRecord
	TemplateLines            []AssetTemplateLineRecord
}

// ListBrowseItems returns held and recognized items of the given type with latest-appraisal detail.
func ListBrowseItems(ctx context.Context, databasePath string, itemType string) ([]BrowseItemRecord, error) {
	itemType = strings.TrimSpace(itemType)
	if itemType == "" {
		itemType = "loot"
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]BrowseItemRecord, error) {
		itemRows, err := db.QueryContext(ctx, `
			SELECT
				id,
				name,
				source,
				status,
				item_type,
				quantity,
				holder,
				notes
			FROM loot_items
			WHERE item_type = ? AND status IN ('held', 'recognized')
			ORDER BY status, name, created_at, id
		`, itemType)
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
				&item.ItemType,
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

		recognizedRows, err := db.QueryContext(ctx, `
			SELECT
				la.loot_item_id,
				la.appraised_value
			FROM loot_appraisals la
			JOIN loot_items li ON li.id = la.loot_item_id
			WHERE li.status IN ('held', 'recognized')
			  AND la.recognized_entry_id IS NOT NULL
			ORDER BY la.loot_item_id, la.appraised_at DESC, la.created_at DESC, la.id DESC
		`)
		if err != nil {
			return nil, fmt.Errorf("query loot browse recognized appraisals: %w", err)
		}
		defer recognizedRows.Close()

		for recognizedRows.Next() {
			var itemID string
			var appraisedValue int64

			if err := recognizedRows.Scan(&itemID, &appraisedValue); err != nil {
				return nil, fmt.Errorf("scan loot browse recognized appraisal: %w", err)
			}

			index, ok := indexesByID[itemID]
			if !ok || items[index].HasRecognizedAppraisal {
				continue
			}

			items[index].HasRecognizedAppraisal = true
			items[index].RecognizedAppraisalValue = appraisedValue
		}

		if err := recognizedRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate loot browse recognized appraisals: %w", err)
		}

		// Load asset template lines when browsing assets.
		if itemType == "asset" && len(items) > 0 {
			templateQuery := `SELECT id, loot_item_id, side, account_code, sort_order FROM asset_template_lines WHERE loot_item_id IN (` + placeholders(len(items)) + `) ORDER BY loot_item_id, sort_order` //nolint:gosec // placeholders generates safe ?,? strings
			templateRows, templateErr := db.QueryContext(ctx, templateQuery, itemIDs(items)...)
			if templateErr != nil {
				return nil, fmt.Errorf("query asset template lines: %w", templateErr)
			}
			defer templateRows.Close()

			for templateRows.Next() {
				var line AssetTemplateLineRecord
				var itemID string
				if err := templateRows.Scan(&line.ID, &itemID, &line.Side, &line.AccountCode, &line.SortOrder); err != nil {
					return nil, fmt.Errorf("scan asset template line: %w", err)
				}
				if index, ok := indexesByID[itemID]; ok {
					items[index].TemplateLines = append(items[index].TemplateLines, line)
				}
			}
			if err := templateRows.Err(); err != nil {
				return nil, fmt.Errorf("iterate asset template lines: %w", err)
			}
		}

		return items, nil
	})
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n-1) + "?"
}

func itemIDs(items []BrowseItemRecord) []any {
	ids := make([]any, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	return ids
}
