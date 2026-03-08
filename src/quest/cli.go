package quest

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

// HandleCreate parses flags and creates a new quest.
func HandleCreate(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var title, patron, description, bonus, status, acceptedOn string
	var rewardStr, advanceStr string

	flagSet := flag.NewFlagSet("quest create", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&title, "title", "", "quest title (required)")
	flagSet.StringVar(&patron, "patron", "", "quest patron")
	flagSet.StringVar(&description, "description", "", "quest description")
	flagSet.StringVar(&rewardStr, "reward", "0", "promised base reward")
	flagSet.StringVar(&advanceStr, "advance", "0", "partial advance received")
	flagSet.StringVar(&bonus, "bonus", "", "bonus conditions")
	flagSet.StringVar(&status, "status", "offered", "initial status (offered or accepted)")
	flagSet.StringVar(&acceptedOn, "accepted-on", "", "accepted date (required if status=accepted)")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	reward, err := tools.ParseAmount(rewardStr)
	if err != nil {
		return fmt.Errorf("invalid reward %q: %w", rewardStr, err)
	}

	advance, err := tools.ParseAmount(advanceStr)
	if err != nil {
		return fmt.Errorf("invalid advance %q: %w", advanceStr, err)
	}

	return RunCreate(ctx, hctx, &CreateQuestInput{
		Title:              title,
		Patron:             patron,
		Description:        description,
		PromisedBaseReward: reward,
		PartialAdvance:     advance,
		BonusConditions:    bonus,
		Status:             status,
		AcceptedOn:         acceptedOn,
	})
}

// RunCreate creates a quest and writes the CLI output.
func RunCreate(ctx context.Context, hctx ledger.HandlerContext, input *CreateQuestInput) error {
	if input == nil {
		return fmt.Errorf("quest input is required")
	}

	result, err := CreateQuest(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Created quest %s\nTitle: %s\nStatus: %s\nReward: %s\n",
		result.ID,
		result.Title,
		string(result.Status),
		tools.FormatAmount(result.PromisedBaseReward),
	); err != nil {
		return fmt.Errorf("write quest output: %w", err)
	}

	return nil
}

// HandleList writes the quest listing.
func HandleList(ctx context.Context, hctx ledger.HandlerContext) error {
	quests, err := ListQuests(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "STATUS           REWARD                  TITLE"); err != nil {
		return fmt.Errorf("write quests header: %w", err)
	}

	for i := range quests {
		if _, err := fmt.Fprintf(
			hctx.Stdout,
			"%-16s %-22s  %s\n",
			string(quests[i].Status),
			tools.FormatAmount(quests[i].PromisedBaseReward),
			quests[i].Title,
		); err != nil {
			return fmt.Errorf("write quest row: %w", err)
		}
	}

	return nil
}

// HandleAccept parses flags and accepts a quest.
func HandleAccept(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var id, date string

	flagSet := flag.NewFlagSet("quest accept", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "quest ID (required)")
	flagSet.StringVar(&date, "date", "", "accepted date in YYYY-MM-DD (required)")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("--id is required")
	}

	if date == "" {
		return fmt.Errorf("--date is required")
	}

	return RunAccept(ctx, hctx, id, date)
}

// RunAccept accepts a quest and writes the CLI output.
func RunAccept(ctx context.Context, hctx ledger.HandlerContext, id string, date string) error {
	if err := AcceptQuest(ctx, hctx.DatabasePath, id, date); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Accepted quest %s\nDate: %s\n", id, date); err != nil {
		return fmt.Errorf("write accept output: %w", err)
	}

	return nil
}

// HandleComplete parses flags and completes a quest.
func HandleComplete(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var id, date string

	flagSet := flag.NewFlagSet("quest complete", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "quest ID (required)")
	flagSet.StringVar(&date, "date", "", "completed date in YYYY-MM-DD (required)")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("--id is required")
	}

	if date == "" {
		return fmt.Errorf("--date is required")
	}

	return RunComplete(ctx, hctx, id, date)
}

// RunComplete completes a quest and writes the CLI output.
func RunComplete(ctx context.Context, hctx ledger.HandlerContext, id string, date string) error {
	if err := CompleteQuest(ctx, hctx.DatabasePath, id, date); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Completed quest %s\nDate: %s\n", id, date); err != nil {
		return fmt.Errorf("write complete output: %w", err)
	}

	return nil
}

// HandleCollect parses flags and collects a quest payment.
func HandleCollect(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var id, date, description, amountStr string

	flagSet := flag.NewFlagSet("quest collect", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "quest ID (required)")
	flagSet.StringVar(&amountStr, "amount", "", "payment amount (required)")
	flagSet.StringVar(&date, "date", "", "payment date in YYYY-MM-DD (required)")
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

	return RunCollect(ctx, hctx, CollectQuestPaymentInput{
		QuestID:     id,
		Amount:      amount,
		Date:        date,
		Description: description,
	})
}

// RunCollect collects quest payment and writes the CLI output.
func RunCollect(ctx context.Context, hctx ledger.HandlerContext, input CollectQuestPaymentInput) error {
	result, err := CollectQuestPayment(ctx, hctx.DatabasePath, CollectQuestPaymentInput{
		QuestID:     input.QuestID,
		Amount:      input.Amount,
		Date:        input.Date,
		Description: input.Description,
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Collected quest payment as journal entry #%d\nDate: %s\nDescription: %s\nAmount: %s\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		tools.FormatAmount(input.Amount),
		tools.FormatAmount(result.DebitTotal),
		tools.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write collect output: %w", err)
	}

	return nil
}

// HandleWriteoff parses flags and writes off a quest.
func HandleWriteoff(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var id, date, description string

	flagSet := flag.NewFlagSet("quest writeoff", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&id, "id", "", "quest ID (required)")
	flagSet.StringVar(&date, "date", "", "write-off date in YYYY-MM-DD (required)")
	flagSet.StringVar(&description, "description", "", "optional description")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("--id is required")
	}

	if date == "" {
		return fmt.Errorf("--date is required")
	}

	return RunWriteoff(ctx, hctx, WriteOffQuestInput{
		QuestID:     id,
		Date:        date,
		Description: description,
	})
}

// RunWriteoff writes off a quest balance and writes the CLI output.
func RunWriteoff(ctx context.Context, hctx ledger.HandlerContext, input WriteOffQuestInput) error {
	result, err := WriteOffQuest(ctx, hctx.DatabasePath, WriteOffQuestInput{
		QuestID:     input.QuestID,
		Date:        input.Date,
		Description: input.Description,
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Wrote off quest as journal entry #%d\nDate: %s\nDescription: %s\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		tools.FormatAmount(result.DebitTotal),
		tools.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write writeoff output: %w", err)
	}

	return nil
}
