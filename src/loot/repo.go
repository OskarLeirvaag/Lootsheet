package loot

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// CreateLootItem inserts a new loot item with status='held'.
func CreateLootItem(ctx context.Context, databasePath string, name string, source string, quantity int, holder string, notes string) (LootItemRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return LootItemRecord{}, fmt.Errorf("loot item name is required")
	}

	if quantity <= 0 {
		quantity = 1
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (LootItemRecord, error) {
		id := uuid.NewString()

		if _, err := db.ExecContext(ctx,
			`INSERT INTO loot_items (id, name, source, status, quantity, holder, notes)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			id, name, strings.TrimSpace(source), string(ledger.LootStatusHeld),
			quantity, strings.TrimSpace(holder), strings.TrimSpace(notes),
		); err != nil {
			return LootItemRecord{}, fmt.Errorf("insert loot item: %w", err)
		}

		return LootItemRecord{
			ID:       id,
			Name:     name,
			Source:   strings.TrimSpace(source),
			Status:   ledger.LootStatusHeld,
			Quantity: quantity,
			Holder:   strings.TrimSpace(holder),
			Notes:    strings.TrimSpace(notes),
		}, nil
	})
}

// UpdateLootItem edits descriptive loot fields. Quantity may only change while held.
func UpdateLootItem(ctx context.Context, databasePath string, lootItemID string, input *UpdateLootItemInput) (LootItemRecord, error) {
	lootItemID = strings.TrimSpace(lootItemID)
	if lootItemID == "" {
		return LootItemRecord{}, fmt.Errorf("loot item ID is required")
	}
	if input == nil {
		return LootItemRecord{}, fmt.Errorf("loot item input is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return LootItemRecord{}, fmt.Errorf("loot item name is required")
	}
	if input.Quantity <= 0 {
		return LootItemRecord{}, fmt.Errorf("quantity must be positive")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (LootItemRecord, error) {
		current, err := getLootItemByID(ctx, db, lootItemID)
		if err != nil {
			return LootItemRecord{}, err
		}

		if current.Status != ledger.LootStatusHeld && input.Quantity != current.Quantity {
			return LootItemRecord{}, fmt.Errorf("quantity can only be edited while loot is held")
		}

		if _, err := db.ExecContext(ctx,
			`UPDATE loot_items
			 SET name = ?, source = ?, quantity = ?, holder = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			name,
			strings.TrimSpace(input.Source),
			input.Quantity,
			strings.TrimSpace(input.Holder),
			strings.TrimSpace(input.Notes),
			lootItemID,
		); err != nil {
			return LootItemRecord{}, fmt.Errorf("update loot item: %w", err)
		}

		return getLootItemByID(ctx, db, lootItemID)
	})
}

// ListLootItems returns all loot items ordered by status priority then name.
func ListLootItems(ctx context.Context, databasePath string) ([]LootItemRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]LootItemRecord, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT id, name, source, status, quantity, holder, notes, created_at, updated_at
			FROM loot_items
			ORDER BY
			  CASE status
			    WHEN 'held' THEN 1
			    WHEN 'recognized' THEN 2
			    WHEN 'assigned' THEN 3
			    WHEN 'sold' THEN 4
			    WHEN 'consumed' THEN 5
			    WHEN 'discarded' THEN 6
			  END,
			  name
		`)
		if err != nil {
			return nil, fmt.Errorf("query loot items: %w", err)
		}
		defer rows.Close()

		items := []LootItemRecord{}
		for rows.Next() {
			var item LootItemRecord
			var status string

			if err := rows.Scan(
				&item.ID, &item.Name, &item.Source, &status,
				&item.Quantity, &item.Holder, &item.Notes,
				&item.CreatedAt, &item.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("scan loot item row: %w", err)
			}

			item.Status = ledger.LootStatus(status)
			if !item.Status.Valid() {
				return nil, fmt.Errorf("scan loot item row: invalid loot status %q", status)
			}

			items = append(items, item)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate loot item rows: %w", err)
		}

		return items, nil
	})
}

// AppraiseLootItem adds an appraisal to a held loot item.
// The appraisal stays off-ledger (recognized_entry_id is NULL).
func AppraiseLootItem(ctx context.Context, databasePath string, lootItemID string, appraisedValue int64, appraiser string, appraisedDate string, notes string) (LootAppraisalRecord, error) {
	lootItemID = strings.TrimSpace(lootItemID)
	if lootItemID == "" {
		return LootAppraisalRecord{}, fmt.Errorf("loot item ID is required")
	}

	if appraisedValue < 0 {
		return LootAppraisalRecord{}, fmt.Errorf("appraised value must be non-negative")
	}

	appraisedDate = strings.TrimSpace(appraisedDate)
	if appraisedDate == "" {
		return LootAppraisalRecord{}, fmt.Errorf("appraisal date is required")
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
func RecognizeLootAppraisal(ctx context.Context, databasePath string, appraisalID string, date string, description string) (ledger.PostedJournalEntry, error) {
	appraisalID = strings.TrimSpace(appraisalID)
	if appraisalID == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("appraisal ID is required")
	}

	date = strings.TrimSpace(date)
	if date == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("recognition date is required")
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

		validated, err := ledger.ValidateJournalPostInput(journalInput)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		accountIDsByCode, err := ledger.ResolveActiveAccountIDsByCode(ctx, db, validated.Lines)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		entryNumber, err := ledger.NextJournalEntryNumber(ctx, db)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		entryID := uuid.NewString()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin recognition transaction: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO journal_entries (id, entry_number, status, entry_date, description, posted_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
			entryID, entryNumber, "posted", validated.EntryDate, validated.Description,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("insert recognition journal entry: %w", err)
		}

		for index, line := range validated.Lines {
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
				uuid.NewString(), entryID, index+1, accountIDsByCode[line.AccountCode], line.Memo, line.DebitAmount, line.CreditAmount,
			); err != nil {
				return ledger.PostedJournalEntry{}, fmt.Errorf("insert recognition journal line %d: %w", index+1, err)
			}
		}

		// Set recognized_entry_id on the appraisal.
		if _, err := tx.ExecContext(ctx,
			"UPDATE loot_appraisals SET recognized_entry_id = ? WHERE id = ?",
			entryID, appraisalID,
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

		return ledger.PostedJournalEntry{
			ID:          entryID,
			EntryNumber: entryNumber,
			EntryDate:   validated.EntryDate,
			Description: validated.Description,
			LineCount:   len(validated.Lines),
			DebitTotal:  validated.Totals.DebitAmount,
			CreditTotal: validated.Totals.CreditAmount,
		}, nil
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
func SellLootItem(ctx context.Context, databasePath string, lootItemID string, saleAmount int64, date string, description string) (ledger.PostedJournalEntry, error) {
	lootItemID = strings.TrimSpace(lootItemID)
	if lootItemID == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("loot item ID is required")
	}

	if saleAmount <= 0 {
		return ledger.PostedJournalEntry{}, fmt.Errorf("sale amount must be positive")
	}

	date = strings.TrimSpace(date)
	if date == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("sale date is required")
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

		validated, err := ledger.ValidateJournalPostInput(journalInput)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		accountIDsByCode, err := ledger.ResolveActiveAccountIDsByCode(ctx, db, validated.Lines)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		entryNumber, err := ledger.NextJournalEntryNumber(ctx, db)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		entryID := uuid.NewString()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin sale transaction: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO journal_entries (id, entry_number, status, entry_date, description, posted_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
			entryID, entryNumber, "posted", validated.EntryDate, validated.Description,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("insert sale journal entry: %w", err)
		}

		for index, line := range validated.Lines {
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
				uuid.NewString(), entryID, index+1, accountIDsByCode[line.AccountCode], line.Memo, line.DebitAmount, line.CreditAmount,
			); err != nil {
				return ledger.PostedJournalEntry{}, fmt.Errorf("insert sale journal line %d: %w", index+1, err)
			}
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

		return ledger.PostedJournalEntry{
			ID:          entryID,
			EntryNumber: entryNumber,
			EntryDate:   validated.EntryDate,
			Description: validated.Description,
			LineCount:   len(validated.Lines),
			DebitTotal:  validated.Totals.DebitAmount,
			CreditTotal: validated.Totals.CreditAmount,
		}, nil
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

func getLootItemByID(ctx context.Context, db *sql.DB, lootItemID string) (LootItemRecord, error) {
	var item LootItemRecord
	var status string

	if err := db.QueryRowContext(ctx,
		`SELECT id, name, source, status, quantity, holder, notes, created_at, updated_at
		 FROM loot_items
		 WHERE id = ?`,
		lootItemID,
	).Scan(
		&item.ID, &item.Name, &item.Source, &status,
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
