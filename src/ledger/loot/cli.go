package loot

import (
	"context"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// RunCreate creates a loot item and writes the CLI output.
func RunCreate(ctx context.Context, hctx ledger.HandlerContext, name string, source string, quantity int, holder string, notes string, itemType string) error {
	result, err := CreateLootItem(ctx, hctx.DatabasePath, name, source, quantity, holder, notes, itemType)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Created loot item %s\nName: %s\nStatus: %s\nQuantity: %d\n",
		result.ID,
		result.Name,
		string(result.Status),
		result.Quantity,
	); err != nil {
		return fmt.Errorf("write loot create output: %w", err)
	}

	return nil
}

// RunList writes the loot item listing.
func RunList(ctx context.Context, hctx ledger.HandlerContext, itemType string) error {
	items, err := ListLootItems(ctx, hctx.DatabasePath, itemType)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "STATUS       QTY  NAME"); err != nil {
		return fmt.Errorf("write loot header: %w", err)
	}

	for i := range items {
		if _, err := fmt.Fprintf(
			hctx.Stdout,
			"%-12s %3d  %s\n",
			string(items[i].Status),
			items[i].Quantity,
			items[i].Name,
		); err != nil {
			return fmt.Errorf("write loot row: %w", err)
		}
	}

	return nil
}

// RunAppraise records a loot appraisal and writes the CLI output.
func RunAppraise(ctx context.Context, hctx ledger.HandlerContext, id string, value int64, appraiser string, date string, notes string) error {
	result, err := AppraiseLootItem(ctx, hctx.DatabasePath, id, value, appraiser, date, notes)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Appraised loot item %s\nAppraisal ID: %s\nValue: %s\nDate: %s\n",
		id,
		result.ID,
		currency.FormatAmount(result.AppraisedValue),
		result.AppraisedAt,
	); err != nil {
		return fmt.Errorf("write loot appraise output: %w", err)
	}

	return nil
}

// RunRecognize recognizes a loot appraisal and writes the CLI output.
func RunRecognize(ctx context.Context, hctx ledger.HandlerContext, appraisalID string, date string, description string) error {
	result, err := RecognizeLootAppraisal(ctx, hctx.DatabasePath, appraisalID, date, description)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Recognized loot appraisal as journal entry #%d\nDate: %s\nDescription: %s\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		currency.FormatAmount(result.DebitTotal),
		currency.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write loot recognize output: %w", err)
	}

	return nil
}

// RunSell records a loot sale and writes the CLI output.
func RunSell(ctx context.Context, hctx ledger.HandlerContext, id string, amount int64, date string, description string) error {
	result, err := SellLootItem(ctx, hctx.DatabasePath, id, amount, date, description)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Sold loot item as journal entry #%d\nDate: %s\nDescription: %s\nAmount: %s\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		currency.FormatAmount(amount),
		currency.FormatAmount(result.DebitTotal),
		currency.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write loot sell output: %w", err)
	}

	return nil
}
