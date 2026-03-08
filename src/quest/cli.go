package quest

import (
	"context"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

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

// RunList writes the quest listing.
func RunList(ctx context.Context, hctx ledger.HandlerContext) error {
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
