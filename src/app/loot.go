package app

import (
	"context"
	"flag"
	"fmt"
	"github.com/OskarLeirvaag/Lootsheet/src/repo"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
	"io"
	"log/slog"
)

func (a *Application) runLoot(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing loot subcommand\n\n%s", usageText)
	}

	switch args[0] {
	case "create":
		return a.runLootCreate(ctx, args[1:])
	case "list":
		return a.runLootList(ctx)
	case "appraise":
		return a.runLootAppraise(ctx, args[1:])
	case "recognize":
		return a.runLootRecognize(ctx, args[1:])
	case "sell":
		return a.runLootSell(ctx, args[1:])
	default:
		return fmt.Errorf("unknown loot subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runLootCreate(ctx context.Context, args []string) error {
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
		return fmt.Errorf("%s\n\n%s", err, lootCreateUsageText)
	}

	if name == "" {
		return fmt.Errorf("--name is required\n\n%s", lootCreateUsageText)
	}

	a.log.logger.InfoContext(ctx, "creating loot item",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("name", name),
	)

	result, err := repo.CreateLootItem(ctx, a.config.Paths.DatabasePath, name, source, quantity, holder, notes)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to create loot item", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "created loot item", slog.String("id", result.ID), slog.String("name", result.Name))

	if _, err := fmt.Fprintf(
		a.stdout,
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

func (a *Application) runLootList(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "listing loot items", slog.String("database_path", a.config.Paths.DatabasePath))

	items, err := repo.ListLootItems(ctx, a.config.Paths.DatabasePath)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to list loot items", slog.String("error", err.Error()))
		return err
	}

	if _, err := fmt.Fprintln(a.stdout, "STATUS       QTY  NAME"); err != nil {
		return fmt.Errorf("write loot header: %w", err)
	}

	for _, item := range items {
		if _, err := fmt.Fprintf(
			a.stdout,
			"%-12s %3d  %s\n",
			string(item.Status),
			item.Quantity,
			item.Name,
		); err != nil {
			return fmt.Errorf("write loot row: %w", err)
		}
	}

	a.log.logger.InfoContext(ctx, "listed loot items", slog.Int("count", len(items)))
	return nil
}

func (a *Application) runLootAppraise(ctx context.Context, args []string) error {
	var id, appraiser, date, notes, valueStr string

	flagSet := flag.NewFlagSet("loot appraise", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "loot item ID (required)")
	flagSet.StringVar(&valueStr, "value", "", "appraised value (required)")
	flagSet.StringVar(&appraiser, "appraiser", "", "who appraised the item")
	flagSet.StringVar(&date, "date", "", "appraisal date in YYYY-MM-DD (required)")
	flagSet.StringVar(&notes, "notes", "", "appraisal notes")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, lootAppraiseUsageText)
	}

	if id == "" {
		return fmt.Errorf("--id is required\n\n%s", lootAppraiseUsageText)
	}

	if valueStr == "" {
		return fmt.Errorf("--value is required\n\n%s", lootAppraiseUsageText)
	}

	value, err := service.ParseAmount(valueStr)
	if err != nil {
		return fmt.Errorf("invalid value %q: %w\n\n%s", valueStr, err, lootAppraiseUsageText)
	}

	if date == "" {
		return fmt.Errorf("--date is required\n\n%s", lootAppraiseUsageText)
	}

	a.log.logger.InfoContext(ctx, "appraising loot item",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("loot_item_id", id),
		slog.Int64("value", value),
	)

	result, err := repo.AppraiseLootItem(ctx, a.config.Paths.DatabasePath, id, value, appraiser, date, notes)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to appraise loot item", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "appraised loot item",
		slog.String("appraisal_id", result.ID),
		slog.Int64("value", result.AppraisedValue),
	)

	if _, err := fmt.Fprintf(
		a.stdout,
		"Appraised loot item %s\nAppraisal ID: %s\nValue: %s\nDate: %s\n",
		id,
		result.ID,
		service.FormatAmount(result.AppraisedValue),
		result.AppraisedAt,
	); err != nil {
		return fmt.Errorf("write loot appraise output: %w", err)
	}

	return nil
}

func (a *Application) runLootRecognize(ctx context.Context, args []string) error {
	var appraisalID, date, description string

	flagSet := flag.NewFlagSet("loot recognize", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&appraisalID, "appraisal-id", "", "appraisal ID (required)")
	flagSet.StringVar(&date, "date", "", "recognition date in YYYY-MM-DD (required)")
	flagSet.StringVar(&description, "description", "", "optional description")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, lootRecognizeUsageText)
	}

	if appraisalID == "" {
		return fmt.Errorf("--appraisal-id is required\n\n%s", lootRecognizeUsageText)
	}

	if date == "" {
		return fmt.Errorf("--date is required\n\n%s", lootRecognizeUsageText)
	}

	a.log.logger.InfoContext(ctx, "recognizing loot appraisal",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("appraisal_id", appraisalID),
		slog.String("date", date),
	)

	result, err := repo.RecognizeLootAppraisal(ctx, a.config.Paths.DatabasePath, appraisalID, date, description)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to recognize loot appraisal", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "recognized loot appraisal",
		slog.String("appraisal_id", appraisalID),
		slog.Int("entry_number", result.EntryNumber),
	)

	if _, err := fmt.Fprintf(
		a.stdout,
		"Recognized loot appraisal as journal entry #%d\nDate: %s\nDescription: %s\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		service.FormatAmount(result.DebitTotal),
		service.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write loot recognize output: %w", err)
	}

	return nil
}

func (a *Application) runLootSell(ctx context.Context, args []string) error {
	var id, date, description, amountStr string

	flagSet := flag.NewFlagSet("loot sell", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "loot item ID (required)")
	flagSet.StringVar(&amountStr, "amount", "", "sale amount (required)")
	flagSet.StringVar(&date, "date", "", "sale date in YYYY-MM-DD (required)")
	flagSet.StringVar(&description, "description", "", "optional description")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, lootSellUsageText)
	}

	if id == "" {
		return fmt.Errorf("--id is required\n\n%s", lootSellUsageText)
	}

	if amountStr == "" {
		return fmt.Errorf("--amount is required\n\n%s", lootSellUsageText)
	}

	amount, err := service.ParseAmount(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w\n\n%s", amountStr, err, lootSellUsageText)
	}

	if date == "" {
		return fmt.Errorf("--date is required\n\n%s", lootSellUsageText)
	}

	a.log.logger.InfoContext(ctx, "selling loot item",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("loot_item_id", id),
		slog.Int64("amount", amount),
		slog.String("date", date),
	)

	result, err := repo.SellLootItem(ctx, a.config.Paths.DatabasePath, id, amount, date, description)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to sell loot item", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "sold loot item",
		slog.String("loot_item_id", id),
		slog.Int("entry_number", result.EntryNumber),
	)

	if _, err := fmt.Fprintf(
		a.stdout,
		"Sold loot item as journal entry #%d\nDate: %s\nDescription: %s\nAmount: %s\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		service.FormatAmount(amount),
		service.FormatAmount(result.DebitTotal),
		service.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write loot sell output: %w", err)
	}

	return nil
}

const lootCreateUsageText = `LootSheet CLI

Usage:
  lootsheet loot create --name TEXT [--source TEXT] [--quantity N] [--holder TEXT] [--notes TEXT]

Examples:
  lootsheet loot create --name "Ruby Gemstone" --source "Dragon Hoard" --quantity 1
  lootsheet loot create --name "Silver Goblets" --quantity 3 --holder "Bard"
`

const lootAppraiseUsageText = `LootSheet CLI

Usage:
  lootsheet loot appraise --id ID --value AMOUNT --date YYYY-MM-DD [--appraiser TEXT] [--notes TEXT]
`

const lootRecognizeUsageText = `LootSheet CLI

Usage:
  lootsheet loot recognize --appraisal-id ID --date YYYY-MM-DD [--description TEXT]
`

const lootSellUsageText = `LootSheet CLI

Usage:
  lootsheet loot sell --id ID --amount AMOUNT --date YYYY-MM-DD [--description TEXT]
`
