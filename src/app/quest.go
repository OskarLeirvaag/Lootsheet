package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
)

func (a *Application) runQuest(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing quest subcommand\n\n%s", usageText)
	}

	switch args[0] {
	case "create":
		return a.runQuestCreate(ctx, args[1:])
	case "list":
		return a.runQuestList(ctx)
	case "accept":
		return a.runQuestAccept(ctx, args[1:])
	case "complete":
		return a.runQuestComplete(ctx, args[1:])
	case "collect":
		return a.runQuestCollect(ctx, args[1:])
	default:
		return fmt.Errorf("unknown quest subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runQuestCreate(ctx context.Context, args []string) error {
	var title, patron, description, bonus, status, acceptedOn string
	var reward, advance int64

	flagSet := flag.NewFlagSet("quest create", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&title, "title", "", "quest title (required)")
	flagSet.StringVar(&patron, "patron", "", "quest patron")
	flagSet.StringVar(&description, "description", "", "quest description")
	flagSet.Int64Var(&reward, "reward", 0, "promised base reward")
	flagSet.Int64Var(&advance, "advance", 0, "partial advance received")
	flagSet.StringVar(&bonus, "bonus", "", "bonus conditions")
	flagSet.StringVar(&status, "status", "offered", "initial status (offered or accepted)")
	flagSet.StringVar(&acceptedOn, "accepted-on", "", "accepted date (required if status=accepted)")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, questCreateUsageText)
	}

	a.log.logger.InfoContext(ctx, "creating quest",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("title", title),
		slog.String("status", status),
	)

	result, err := repo.CreateQuest(ctx, a.config.Paths.DatabasePath, repo.CreateQuestInput{
		Title:              title,
		Patron:             patron,
		Description:        description,
		PromisedBaseReward: reward,
		PartialAdvance:     advance,
		BonusConditions:    bonus,
		Status:             status,
		AcceptedOn:         acceptedOn,
	})
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to create quest", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "created quest", slog.String("id", result.ID), slog.String("title", result.Title))

	if _, err := fmt.Fprintf(
		a.stdout,
		"Created quest %s\nTitle: %s\nStatus: %s\nReward: %d\n",
		result.ID,
		result.Title,
		string(result.Status),
		result.PromisedBaseReward,
	); err != nil {
		return fmt.Errorf("write quest output: %w", err)
	}

	return nil
}

func (a *Application) runQuestList(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "listing quests", slog.String("database_path", a.config.Paths.DatabasePath))

	quests, err := repo.ListQuests(ctx, a.config.Paths.DatabasePath)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to list quests", slog.String("error", err.Error()))
		return err
	}

	if _, err := fmt.Fprintln(a.stdout, "STATUS           REWARD  TITLE"); err != nil {
		return fmt.Errorf("write quests header: %w", err)
	}

	for _, quest := range quests {
		if _, err := fmt.Fprintf(
			a.stdout,
			"%-16s %6d  %s\n",
			string(quest.Status),
			quest.PromisedBaseReward,
			quest.Title,
		); err != nil {
			return fmt.Errorf("write quest row: %w", err)
		}
	}

	a.log.logger.InfoContext(ctx, "listed quests", slog.Int("count", len(quests)))
	return nil
}

func (a *Application) runQuestAccept(ctx context.Context, args []string) error {
	var id, date string

	flagSet := flag.NewFlagSet("quest accept", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "quest ID (required)")
	flagSet.StringVar(&date, "date", "", "accepted date in YYYY-MM-DD (required)")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, questAcceptUsageText)
	}

	if id == "" {
		return fmt.Errorf("--id is required\n\n%s", questAcceptUsageText)
	}

	if date == "" {
		return fmt.Errorf("--date is required\n\n%s", questAcceptUsageText)
	}

	a.log.logger.InfoContext(ctx, "accepting quest",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("quest_id", id),
		slog.String("date", date),
	)

	if err := repo.AcceptQuest(ctx, a.config.Paths.DatabasePath, id, date); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to accept quest", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "accepted quest", slog.String("quest_id", id))

	if _, err := fmt.Fprintf(a.stdout, "Accepted quest %s\nDate: %s\n", id, date); err != nil {
		return fmt.Errorf("write accept output: %w", err)
	}

	return nil
}

func (a *Application) runQuestComplete(ctx context.Context, args []string) error {
	var id, date string

	flagSet := flag.NewFlagSet("quest complete", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "quest ID (required)")
	flagSet.StringVar(&date, "date", "", "completed date in YYYY-MM-DD (required)")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, questCompleteUsageText)
	}

	if id == "" {
		return fmt.Errorf("--id is required\n\n%s", questCompleteUsageText)
	}

	if date == "" {
		return fmt.Errorf("--date is required\n\n%s", questCompleteUsageText)
	}

	a.log.logger.InfoContext(ctx, "completing quest",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("quest_id", id),
		slog.String("date", date),
	)

	if err := repo.CompleteQuest(ctx, a.config.Paths.DatabasePath, id, date); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to complete quest", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "completed quest", slog.String("quest_id", id))

	if _, err := fmt.Fprintf(a.stdout, "Completed quest %s\nDate: %s\n", id, date); err != nil {
		return fmt.Errorf("write complete output: %w", err)
	}

	return nil
}

func (a *Application) runQuestCollect(ctx context.Context, args []string) error {
	var id, date, description, amountStr string

	flagSet := flag.NewFlagSet("quest collect", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "quest ID (required)")
	flagSet.StringVar(&amountStr, "amount", "", "payment amount (required)")
	flagSet.StringVar(&date, "date", "", "payment date in YYYY-MM-DD (required)")
	flagSet.StringVar(&description, "description", "", "optional description")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, questCollectUsageText)
	}

	if id == "" {
		return fmt.Errorf("--id is required\n\n%s", questCollectUsageText)
	}

	if amountStr == "" {
		return fmt.Errorf("--amount is required\n\n%s", questCollectUsageText)
	}

	amount, err := strconv.ParseInt(strings.TrimSpace(amountStr), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w\n\n%s", amountStr, err, questCollectUsageText)
	}

	if date == "" {
		return fmt.Errorf("--date is required\n\n%s", questCollectUsageText)
	}

	a.log.logger.InfoContext(ctx, "collecting quest payment",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("quest_id", id),
		slog.Int64("amount", amount),
		slog.String("date", date),
	)

	result, err := repo.CollectQuestPayment(ctx, a.config.Paths.DatabasePath, repo.CollectQuestPaymentInput{
		QuestID:     id,
		Amount:      amount,
		Date:        date,
		Description: description,
	})
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to collect quest payment", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "collected quest payment",
		slog.String("quest_id", id),
		slog.Int("entry_number", result.EntryNumber),
	)

	if _, err := fmt.Fprintf(
		a.stdout,
		"Collected quest payment as journal entry #%d\nDate: %s\nDescription: %s\nAmount: %d\nDebits: %d\nCredits: %d\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		amount,
		result.DebitTotal,
		result.CreditTotal,
	); err != nil {
		return fmt.Errorf("write collect output: %w", err)
	}

	return nil
}

const questCreateUsageText = `LootSheet CLI

Usage:
  lootsheet quest create --title TEXT [--patron TEXT] [--description TEXT] [--reward AMOUNT] [--advance AMOUNT] [--bonus TEXT] [--status offered|accepted] [--accepted-on DATE]

Examples:
  lootsheet quest create --title "Clear the Goblin Cave" --patron "Mayor Thornton" --reward 500
  lootsheet quest create --title "Escort the Merchant" --status accepted --accepted-on 2026-03-01 --reward 200
`

const questAcceptUsageText = `LootSheet CLI

Usage:
  lootsheet quest accept --id ID --date YYYY-MM-DD
`

const questCompleteUsageText = `LootSheet CLI

Usage:
  lootsheet quest complete --id ID --date YYYY-MM-DD
`

const questCollectUsageText = `LootSheet CLI

Usage:
  lootsheet quest collect --id ID --amount AMOUNT --date YYYY-MM-DD [--description TEXT]
`
