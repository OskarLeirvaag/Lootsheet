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
	Amount      string
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
				`INSERT INTO asset_template_lines (id, loot_item_id, side, account_code, amount, sort_order)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				id, itemID, strings.TrimSpace(line.Side), strings.TrimSpace(line.AccountCode), strings.TrimSpace(line.Amount), i,
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

