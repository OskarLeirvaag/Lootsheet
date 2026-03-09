package loot

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// AssetTemplateLineRecord represents a single template line stored against an asset.
type AssetTemplateLineRecord struct {
	ID          string
	LootItemID  string
	Side        string // "debit" or "credit"
	AccountCode string
	SortOrder   int
}

// SaveAssetTemplate replaces all template lines for an asset in a single transaction.
func SaveAssetTemplate(ctx context.Context, databasePath string, itemID string, lines []AssetTemplateLineRecord) error {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return fmt.Errorf("item ID is required")
	}

	for i, line := range lines {
		side := strings.TrimSpace(line.Side)
		if side != "debit" && side != "credit" {
			return fmt.Errorf("line %d side must be debit or credit", i+1)
		}
		if strings.TrimSpace(line.AccountCode) == "" {
			return fmt.Errorf("line %d account code is required", i+1)
		}
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		// Verify item exists and is an asset.
		var itemType string
		if err := db.QueryRowContext(ctx,
			"SELECT item_type FROM loot_items WHERE id = ?", itemID,
		).Scan(&itemType); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("item %q does not exist", itemID)
			}
			return fmt.Errorf("query item type: %w", err)
		}
		if itemType != "asset" {
			return fmt.Errorf("item %q is not an asset (type=%q)", itemID, itemType)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin template save transaction: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		if _, err := tx.ExecContext(ctx,
			"DELETE FROM asset_template_lines WHERE loot_item_id = ?", itemID,
		); err != nil {
			return fmt.Errorf("delete existing template lines: %w", err)
		}

		for i, line := range lines {
			id := uuid.NewString()
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO asset_template_lines (id, loot_item_id, side, account_code, sort_order)
				 VALUES (?, ?, ?, ?, ?)`,
				id, itemID, strings.TrimSpace(line.Side), strings.TrimSpace(line.AccountCode), i,
			); err != nil {
				return fmt.Errorf("insert template line %d: %w", i+1, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit template save transaction: %w", err)
		}

		return nil
	})
}

// ListAssetTemplateLines returns all template lines for an item ordered by sort_order.
func ListAssetTemplateLines(ctx context.Context, databasePath string, itemID string) ([]AssetTemplateLineRecord, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return nil, fmt.Errorf("item ID is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]AssetTemplateLineRecord, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT id, loot_item_id, side, account_code, sort_order
			FROM asset_template_lines
			WHERE loot_item_id = ?
			ORDER BY sort_order
		`, itemID)
		if err != nil {
			return nil, fmt.Errorf("query template lines: %w", err)
		}
		defer rows.Close()

		var lines []AssetTemplateLineRecord
		for rows.Next() {
			var line AssetTemplateLineRecord
			if err := rows.Scan(&line.ID, &line.LootItemID, &line.Side, &line.AccountCode, &line.SortOrder); err != nil {
				return nil, fmt.Errorf("scan template line: %w", err)
			}
			lines = append(lines, line)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate template lines: %w", err)
		}

		return lines, nil
	})
}

// DeleteAssetTemplate removes all template lines for an item.
func DeleteAssetTemplate(ctx context.Context, databasePath string, itemID string) error {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return fmt.Errorf("item ID is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		if _, err := db.ExecContext(ctx,
			"DELETE FROM asset_template_lines WHERE loot_item_id = ?", itemID,
		); err != nil {
			return fmt.Errorf("delete template lines: %w", err)
		}
		return nil
	})
}
