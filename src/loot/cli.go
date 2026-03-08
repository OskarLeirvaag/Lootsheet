package loot

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

// HandleCreate parses flags and creates a new loot item.
func HandleCreate(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var name, source, holder, notes string
	var quantity int

	flagSet := flag.NewFlagSet("loot create", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&name, "name", "", "item name (required)")
	flagSet.StringVar(&source, "source", "", "where the item was found")
	flagSet.IntVar(&quantity, "quantity", 1, "item quantity")
	flagSet.StringVar(&holder, "holder", "", "who is carrying the item")
	flagSet.StringVar(&notes, "notes", "", "additional notes")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("--name is required")
	}

	return RunCreate(ctx, hctx, name, source, quantity, holder, notes)
}

// RunCreate creates a loot item and writes the CLI output.
func RunCreate(ctx context.Context, hctx ledger.HandlerContext, name string, source string, quantity int, holder string, notes string) error {
	result, err := CreateLootItem(ctx, hctx.DatabasePath, name, source, quantity, holder, notes)
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

// HandleList writes the loot item listing.
func HandleList(ctx context.Context, hctx ledger.HandlerContext) error {
	items, err := ListLootItems(ctx, hctx.DatabasePath)
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

// HandleAppraise parses flags and appraises a loot item.
func HandleAppraise(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var id, appraiser, date, notes, valueStr string

	flagSet := flag.NewFlagSet("loot appraise", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "loot item ID (required)")
	flagSet.StringVar(&valueStr, "value", "", "appraised value (required)")
	flagSet.StringVar(&appraiser, "appraiser", "", "who appraised the item")
	flagSet.StringVar(&date, "date", "", "appraisal date in YYYY-MM-DD (required)")
	flagSet.StringVar(&notes, "notes", "", "appraisal notes")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("--id is required")
	}

	if valueStr == "" {
		return fmt.Errorf("--value is required")
	}

	value, err := tools.ParseAmount(valueStr)
	if err != nil {
		return fmt.Errorf("invalid value %q: %w", valueStr, err)
	}

	if date == "" {
		return fmt.Errorf("--date is required")
	}

	return RunAppraise(ctx, hctx, id, value, appraiser, date, notes)
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
		tools.FormatAmount(result.AppraisedValue),
		result.AppraisedAt,
	); err != nil {
		return fmt.Errorf("write loot appraise output: %w", err)
	}

	return nil
}

// HandleRecognize parses flags and recognizes a loot appraisal on-ledger.
func HandleRecognize(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var appraisalID, date, description string

	flagSet := flag.NewFlagSet("loot recognize", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&appraisalID, "appraisal-id", "", "appraisal ID (required)")
	flagSet.StringVar(&date, "date", "", "recognition date in YYYY-MM-DD (required)")
	flagSet.StringVar(&description, "description", "", "optional description")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if appraisalID == "" {
		return fmt.Errorf("--appraisal-id is required")
	}

	if date == "" {
		return fmt.Errorf("--date is required")
	}

	return RunRecognize(ctx, hctx, appraisalID, date, description)
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
		tools.FormatAmount(result.DebitTotal),
		tools.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write loot recognize output: %w", err)
	}

	return nil
}

// HandleSell parses flags and sells a loot item.
func HandleSell(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var id, date, description, amountStr string

	flagSet := flag.NewFlagSet("loot sell", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "loot item ID (required)")
	flagSet.StringVar(&amountStr, "amount", "", "sale amount (required)")
	flagSet.StringVar(&date, "date", "", "sale date in YYYY-MM-DD (required)")
	flagSet.StringVar(&description, "description", "", "optional description")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("--id is required")
	}

	if amountStr == "" {
		return fmt.Errorf("--amount is required")
	}

	amount, err := tools.ParseAmount(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", amountStr, err)
	}

	if date == "" {
		return fmt.Errorf("--date is required")
	}

	return RunSell(ctx, hctx, id, amount, date, description)
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
		tools.FormatAmount(amount),
		tools.FormatAmount(result.DebitTotal),
		tools.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write loot sell output: %w", err)
	}

	return nil
}
