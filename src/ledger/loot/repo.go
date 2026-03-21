package loot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// CreateLootItem inserts a new loot item with status='held'.
func CreateLootItem(ctx context.Context, databasePath string, campaignID string, name string, source string, quantity int, holder string, notes string, itemType string) (LootItemRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return LootItemRecord{}, errors.New("loot item name is required")
	}

	if quantity <= 0 {
		quantity = 1
	}

	itemType = strings.TrimSpace(itemType)
	if itemType == "" {
		itemType = "loot"
	}
	if itemType != "loot" && itemType != "asset" {
		return LootItemRecord{}, errors.New("item type must be loot or asset")
	}

	notes = strings.TrimSpace(notes)

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (LootItemRecord, error) {
		id := uuid.NewString()

		if _, err := db.ExecContext(ctx,
			`INSERT INTO loot_items (id, campaign_id, name, source, status, quantity, holder, notes, item_type)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, campaignID, name, strings.TrimSpace(source), string(ledger.LootStatusHeld),
			quantity, strings.TrimSpace(holder), notes, itemType,
		); err != nil {
			return LootItemRecord{}, fmt.Errorf("insert loot item: %w", err)
		}

		if err := rebuildReferences(ctx, db, id, campaignID, name, notes); err != nil {
			return LootItemRecord{}, err
		}

		return LootItemRecord{
			ID:       id,
			Name:     name,
			Source:   strings.TrimSpace(source),
			Status:   ledger.LootStatusHeld,
			ItemType: itemType,
			Quantity: quantity,
			Holder:   strings.TrimSpace(holder),
			Notes:    notes,
		}, nil
	})
}

// UpdateLootItem edits descriptive loot fields. Quantity may only change while held.
func UpdateLootItem(ctx context.Context, databasePath string, campaignID string, lootItemID string, input *UpdateLootItemInput) (LootItemRecord, error) {
	lootItemID = strings.TrimSpace(lootItemID)
	if lootItemID == "" {
		return LootItemRecord{}, errors.New("loot item ID is required")
	}
	if input == nil {
		return LootItemRecord{}, errors.New("loot item input is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return LootItemRecord{}, errors.New("loot item name is required")
	}
	if input.Quantity <= 0 {
		return LootItemRecord{}, errors.New("quantity must be positive")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (LootItemRecord, error) {
		current, err := getLootItemByID(ctx, db, lootItemID)
		if err != nil {
			return LootItemRecord{}, err
		}

		if current.Status != ledger.LootStatusHeld && input.Quantity != current.Quantity {
			return LootItemRecord{}, errors.New("quantity can only be edited while loot is held")
		}

		notes := strings.TrimSpace(input.Notes)

		if _, err := db.ExecContext(ctx,
			`UPDATE loot_items
			 SET name = ?, source = ?, quantity = ?, holder = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			name,
			strings.TrimSpace(input.Source),
			input.Quantity,
			strings.TrimSpace(input.Holder),
			notes,
			lootItemID,
		); err != nil {
			return LootItemRecord{}, fmt.Errorf("update loot item: %w", err)
		}

		if err := rebuildReferences(ctx, db, lootItemID, campaignID, name, notes); err != nil {
			return LootItemRecord{}, err
		}

		return getLootItemByID(ctx, db, lootItemID)
	})
}

// AppraiseLootItem adds an appraisal to a held loot item.
// The appraisal stays off-ledger (recognized_entry_id is NULL).
func AppraiseLootItem(ctx context.Context, databasePath string, campaignID string, lootItemID string, appraisedValue int64, appraiser string, appraisedDate string, notes string) (LootAppraisalRecord, error) {
	lootItemID = strings.TrimSpace(lootItemID)
	if lootItemID == "" {
		return LootAppraisalRecord{}, errors.New("loot item ID is required")
	}

	if appraisedValue < 0 {
		return LootAppraisalRecord{}, errors.New("appraised value must be non-negative")
	}

	appraisedDate = strings.TrimSpace(appraisedDate)
	if appraisedDate == "" {
		return LootAppraisalRecord{}, errors.New("appraisal date is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (LootAppraisalRecord, error) {
		// Verify item exists and is held.
		status, err := getLootItemStatus(ctx, db, lootItemID)
		if err != nil {
			return LootAppraisalRecord{}, err
		}

		if status != ledger.LootStatusHeld {
			return LootAppraisalRecord{}, fmt.Errorf("loot item %q cannot be appraised: current status is %q, expected %q", lootItemID, status, ledger.LootStatusHeld)
		}

		id := uuid.NewString()

		if _, err := db.ExecContext(ctx,
			`INSERT INTO loot_appraisals (id, loot_item_id, appraised_value, appraiser, notes, appraised_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			id, lootItemID, appraisedValue, strings.TrimSpace(appraiser),
			strings.TrimSpace(notes), appraisedDate,
		); err != nil {
			return LootAppraisalRecord{}, fmt.Errorf("insert loot appraisal: %w", err)
		}

		return LootAppraisalRecord{
			ID:             id,
			LootItemID:     lootItemID,
			AppraisedValue: appraisedValue,
			Appraiser:      strings.TrimSpace(appraiser),
			Notes:          strings.TrimSpace(notes),
			AppraisedAt:    appraisedDate,
		}, nil
	})
}

// RecognizeLootAppraisal moves an appraisal on-ledger by creating a journal entry:
//
//	Dr Loot Inventory (1200)
//	Cr Unrealized Loot Gain (4200)
//
// Sets recognized_entry_id on the appraisal and updates the loot item status to 'recognized'.
func RecognizeLootAppraisal(ctx context.Context, databasePath string, campaignID string, appraisalID string, date string, description string) (ledger.PostedJournalEntry, error) {
	appraisalID = strings.TrimSpace(appraisalID)
	if appraisalID == "" {
		return ledger.PostedJournalEntry{}, errors.New("appraisal ID is required")
	}

	date = strings.TrimSpace(date)
	if date == "" {
		return ledger.PostedJournalEntry{}, errors.New("recognition date is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
		// Load the appraisal.
		var appraisal LootAppraisalRecord
		var recognizedEntryID sql.NullString

		if err := db.QueryRowContext(ctx,
			"SELECT id, loot_item_id, appraised_value, appraiser, recognized_entry_id FROM loot_appraisals WHERE id = ?",
			appraisalID,
		).Scan(&appraisal.ID, &appraisal.LootItemID, &appraisal.AppraisedValue, &appraisal.Appraiser, &recognizedEntryID); err != nil {
			if err == sql.ErrNoRows {
				return ledger.PostedJournalEntry{}, fmt.Errorf("appraisal %q does not exist", appraisalID)
			}
			return ledger.PostedJournalEntry{}, fmt.Errorf("query appraisal: %w", err)
		}

		if recognizedEntryID.Valid {
			return ledger.PostedJournalEntry{}, fmt.Errorf("appraisal %q is already recognized", appraisalID)
		}

		// Build the journal entry.
		if description == "" {
			description = fmt.Sprintf("Recognize loot appraisal: %s", appraisalID)
		}

		journalInput := ledger.JournalPostInput{
			EntryDate:   date,
			Description: description,
			Lines: []ledger.JournalLineInput{
				{AccountCode: "1200", DebitAmount: appraisal.AppraisedValue, Memo: "Loot inventory recognition"},
				{AccountCode: "4200", CreditAmount: appraisal.AppraisedValue, Memo: "Unrealized loot gain"},
			},
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin recognition transaction: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		posted, err := ledger.PostJournalWithinTx(ctx, db, tx, campaignID, journalInput)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		// Set recognized_entry_id on the appraisal.
		if _, err := tx.ExecContext(ctx,
			"UPDATE loot_appraisals SET recognized_entry_id = ? WHERE id = ?",
			posted.ID, appraisalID,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("update appraisal recognized_entry_id: %w", err)
		}

		// Update loot item status to 'recognized'.
		if _, err := tx.ExecContext(ctx,
			"UPDATE loot_items SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			string(ledger.LootStatusRecognized), appraisal.LootItemID,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("update loot item status to recognized: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("commit recognition transaction: %w", err)
		}

		return posted, nil
	})
}

// SellLootItem sells a recognized loot item, creating a journal entry:
//
//	Dr Party Cash (1000) for sale amount
//	Cr Loot Inventory (1200) for the recognized appraisal value
//	If sale < appraisal: Dr Loss on Sale of Loot (5400) for the difference
//	If sale > appraisal: Cr Gain on Sale of Loot (4300) for the difference
//
// Updates item status to 'sold'.
func SellLootItem(ctx context.Context, databasePath string, campaignID string, lootItemID string, saleAmount int64, date string, description string) (ledger.PostedJournalEntry, error) {
	lootItemID = strings.TrimSpace(lootItemID)
	if lootItemID == "" {
		return ledger.PostedJournalEntry{}, errors.New("loot item ID is required")
	}

	if saleAmount <= 0 {
		return ledger.PostedJournalEntry{}, errors.New("sale amount must be positive")
	}

	date = strings.TrimSpace(date)
	if date == "" {
		return ledger.PostedJournalEntry{}, errors.New("sale date is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
		// Verify item exists and is recognized.
		status, err := getLootItemStatus(ctx, db, lootItemID)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		if status != ledger.LootStatusRecognized {
			return ledger.PostedJournalEntry{}, fmt.Errorf("loot item %q cannot be sold: current status is %q, expected %q", lootItemID, status, ledger.LootStatusRecognized)
		}

		// Get the recognized appraisal value.
		var appraisedValue int64
		if err := db.QueryRowContext(ctx,
			"SELECT appraised_value FROM loot_appraisals WHERE loot_item_id = ? AND recognized_entry_id IS NOT NULL ORDER BY appraised_at DESC LIMIT 1",
			lootItemID,
		).Scan(&appraisedValue); err != nil {
			if err == sql.ErrNoRows {
				return ledger.PostedJournalEntry{}, fmt.Errorf("loot item %q has no recognized appraisal", lootItemID)
			}
			return ledger.PostedJournalEntry{}, fmt.Errorf("query recognized appraisal: %w", err)
		}

		if description == "" {
			description = fmt.Sprintf("Sale of loot item: %s", lootItemID)
		}

		// Build journal lines.
		lines := []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: saleAmount, Memo: "Loot sale proceeds"},
			{AccountCode: "1200", CreditAmount: appraisedValue, Memo: "Loot inventory disposal"},
		}

		if saleAmount < appraisedValue {
			// Loss on sale.
			loss := appraisedValue - saleAmount
			lines = append(lines, ledger.JournalLineInput{
				AccountCode: "5400", DebitAmount: loss, Memo: "Loss on loot sale",
			})
		} else if saleAmount > appraisedValue {
			// Gain on sale.
			gain := saleAmount - appraisedValue
			lines = append(lines, ledger.JournalLineInput{
				AccountCode: "4300", CreditAmount: gain, Memo: "Gain on loot sale",
			})
		}

		journalInput := ledger.JournalPostInput{
			EntryDate:   date,
			Description: description,
			Lines:       lines,
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin sale transaction: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		posted, err := ledger.PostJournalWithinTx(ctx, db, tx, campaignID, journalInput)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		// Update loot item status to 'sold'.
		if _, err := tx.ExecContext(ctx,
			"UPDATE loot_items SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			string(ledger.LootStatusSold), lootItemID,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("update loot item status to sold: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("commit sale transaction: %w", err)
		}

		return posted, nil
	})
}

// getLootItemStatus returns the current status of a loot item.
func getLootItemStatus(ctx context.Context, db *sql.DB, lootItemID string) (ledger.LootStatus, error) {
	var status string
	if err := db.QueryRowContext(ctx,
		"SELECT status FROM loot_items WHERE id = ?", lootItemID,
	).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("loot item %q does not exist", lootItemID)
		}
		return "", fmt.Errorf("query loot item status: %w", err)
	}

	s := ledger.LootStatus(status)
	if !s.Valid() {
		return "", fmt.Errorf("loot item %s has invalid status %q", lootItemID, status)
	}

	return s, nil
}

// TransferItemType changes a loot item's item_type between 'loot' and 'asset'.
func TransferItemType(ctx context.Context, databasePath string, campaignID string, itemID string, newType string) error {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return errors.New("item ID is required")
	}
	newType = strings.TrimSpace(newType)
	if newType != "loot" && newType != "asset" {
		return errors.New("new type must be loot or asset")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		var currentType string
		if err := db.QueryRowContext(ctx,
			"SELECT item_type FROM loot_items WHERE id = ?", itemID,
		).Scan(&currentType); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("item %q does not exist", itemID)
			}
			return fmt.Errorf("query item type: %w", err)
		}
		if currentType == newType {
			return fmt.Errorf("item %q is already of type %q", itemID, newType)
		}
		if _, err := db.ExecContext(ctx,
			"UPDATE loot_items SET item_type = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			newType, itemID,
		); err != nil {
			return fmt.Errorf("update item type: %w", err)
		}
		return nil
	})
}

func getLootItemByID(ctx context.Context, db *sql.DB, lootItemID string) (LootItemRecord, error) {
	var item LootItemRecord
	var status string

	if err := db.QueryRowContext(ctx,
		`SELECT id, name, source, status, item_type, quantity, holder, notes, created_at, updated_at
		 FROM loot_items
		 WHERE id = ?`,
		lootItemID,
	).Scan(
		&item.ID, &item.Name, &item.Source, &status, &item.ItemType,
		&item.Quantity, &item.Holder, &item.Notes,
		&item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return LootItemRecord{}, fmt.Errorf("loot item %q does not exist", lootItemID)
		}
		return LootItemRecord{}, fmt.Errorf("query loot item: %w", err)
	}

	item.Status = ledger.LootStatus(status)
	if !item.Status.Valid() {
		return LootItemRecord{}, fmt.Errorf("loot item %s has invalid status %q", lootItemID, status)
	}

	return item, nil
}
