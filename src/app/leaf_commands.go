package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/account"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
	"github.com/spf13/cobra"
)

func (a *Application) newNoArgsLeafCommand(use string, short string, helpText string, run func(context.Context) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  helpText,
	}

	return a.newLeafCommand(cmd, run)
}

func (a *Application) newLeafCommand(cmd *cobra.Command, run func(context.Context) error) *cobra.Command {
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if isLeafHelpArg(args) {
			return a.writeCommandHelp(cmd)
		}

		if len(args) > 0 {
			return unexpectedLeafArgsError(cmd, args)
		}

		return run(cmd.Context())
	}

	return cmd
}

func (a *Application) newAccountCreateCommand() *cobra.Command {
	var code, name, accountType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new account",
		Long:  accountCreateHelpText,
	}
	cmd.Flags().StringVar(&code, "code", "", "account code")
	cmd.Flags().StringVar(&name, "name", "", "account name")
	cmd.Flags().StringVar(&accountType, "type", "", "account type (asset, liability, equity, income, expense)")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if code == "" {
			return fmt.Errorf("--code is required")
		}
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		if accountType == "" {
			return fmt.Errorf("--type is required")
		}

		return account.RunCreate(ctx, a.handlerContext(), code, name, ledger.AccountType(accountType))
	})
}

func (a *Application) newAccountRenameCommand() *cobra.Command {
	var code, name string

	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename an existing account",
		Long:  accountRenameHelpText,
	}
	cmd.Flags().StringVar(&code, "code", "", "account code")
	cmd.Flags().StringVar(&name, "name", "", "new account name")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if code == "" {
			return fmt.Errorf("--code is required")
		}
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		return account.RunRename(ctx, a.handlerContext(), code, name)
	})
}

func (a *Application) newAccountDeactivateCommand() *cobra.Command {
	var code string

	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate an account",
		Long:  accountDeactivateHelpText,
	}
	cmd.Flags().StringVar(&code, "code", "", "account code")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if code == "" {
			return fmt.Errorf("--code is required")
		}

		return account.RunDeactivate(ctx, a.handlerContext(), code)
	})
}

func (a *Application) newAccountActivateCommand() *cobra.Command {
	var code string

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate an account",
		Long:  accountActivateHelpText,
	}
	cmd.Flags().StringVar(&code, "code", "", "account code")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if code == "" {
			return fmt.Errorf("--code is required")
		}

		return account.RunActivate(ctx, a.handlerContext(), code)
	})
}

func (a *Application) newAccountDeleteCommand() *cobra.Command {
	var code string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an unused account",
		Long:  accountDeleteHelpText,
	}
	cmd.Flags().StringVar(&code, "code", "", "account code")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if code == "" {
			return fmt.Errorf("--code is required")
		}

		return account.RunDelete(ctx, a.handlerContext(), code)
	})
}

func (a *Application) newAccountLedgerCommand() *cobra.Command {
	var code string

	cmd := &cobra.Command{
		Use:   "ledger",
		Short: "Show a single-account ledger report",
		Long:  accountLedgerHelpText,
	}
	cmd.Flags().StringVar(&code, "code", "", "account code")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if strings.TrimSpace(code) == "" {
			return fmt.Errorf("--code is required")
		}

		return journal.RunAccountLedger(ctx, a.handlerContext(), code)
	})
}

func (a *Application) newJournalPostCommand() *cobra.Command {
	var entryDate, description string
	var debitSpecs, creditSpecs []string

	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post a balanced journal entry",
		Long:  journalPostHelpText,
	}
	cmd.Flags().StringVar(&entryDate, "date", "", "entry date in YYYY-MM-DD")
	cmd.Flags().StringVar(&description, "description", "", "journal entry description")
	cmd.Flags().StringArrayVar(&debitSpecs, "debit", nil, "debit line in CODE:AMOUNT[:MEMO] format")
	cmd.Flags().StringArrayVar(&creditSpecs, "credit", nil, "credit line in CODE:AMOUNT[:MEMO] format")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		input, err := journal.BuildJournalPostInput(entryDate, description, debitSpecs, creditSpecs)
		if err != nil {
			return fmt.Errorf("%s\n\n%s", err, journalPostHelpText)
		}

		return journal.RunPost(ctx, a.handlerContext(), input)
	})
}

func (a *Application) newJournalReverseCommand() *cobra.Command {
	var entryID, date, description string

	cmd := &cobra.Command{
		Use:   "reverse",
		Short: "Reverse a posted journal entry",
		Long:  journalReverseHelpText,
	}
	cmd.Flags().StringVar(&entryID, "entry-id", "", "UUID of the journal entry to reverse")
	cmd.Flags().StringVar(&date, "date", "", "reversal date in YYYY-MM-DD")
	cmd.Flags().StringVar(&description, "description", "", "optional reversal description")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if entryID == "" {
			return fmt.Errorf("--entry-id is required\n\n%s", journalReverseHelpText)
		}
		if date == "" {
			return fmt.Errorf("--date is required\n\n%s", journalReverseHelpText)
		}

		return journal.RunReverse(ctx, a.handlerContext(), entryID, date, description)
	})
}

func (a *Application) newEntryExpenseCommand() *cobra.Command {
	var accountCode, paidFromCode, amountStr, date, description, memo string

	cmd := &cobra.Command{
		Use:   "expense",
		Short: "Record a guided expense entry",
		Long:  entryExpenseHelpText,
	}
	cmd.Flags().StringVar(&accountCode, "account", "", "expense account code (required)")
	cmd.Flags().StringVar(&paidFromCode, "paid-from", "1000", "funding account code")
	cmd.Flags().StringVar(&amountStr, "amount", "", "expense amount (required)")
	cmd.Flags().StringVar(&date, "date", "", "entry date in YYYY-MM-DD (defaults to today)")
	cmd.Flags().StringVar(&description, "description", "", "journal entry description (required)")
	cmd.Flags().StringVar(&memo, "memo", "", "optional line memo")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if strings.TrimSpace(accountCode) == "" {
			return fmt.Errorf("--account is required")
		}
		if strings.TrimSpace(amountStr) == "" {
			return fmt.Errorf("--amount is required")
		}
		if strings.TrimSpace(description) == "" {
			return fmt.Errorf("--description is required")
		}
		amount, err := tools.ParseAmount(amountStr)
		if err != nil {
			return fmt.Errorf("invalid amount %q: %w", amountStr, err)
		}
		if strings.TrimSpace(date) == "" {
			date = tuiToday()
		}

		return journal.RunExpense(ctx, a.handlerContext(), &journal.ExpenseEntryInput{
			Date:               date,
			Description:        description,
			ExpenseAccountCode: accountCode,
			FundingAccountCode: paidFromCode,
			Amount:             amount,
			Memo:               memo,
		})
	})
}

func (a *Application) newEntryIncomeCommand() *cobra.Command {
	var accountCode, depositToCode, amountStr, date, description, memo string

	cmd := &cobra.Command{
		Use:   "income",
		Short: "Record a guided income entry",
		Long:  entryIncomeHelpText,
	}
	cmd.Flags().StringVar(&accountCode, "account", "", "income account code (required)")
	cmd.Flags().StringVar(&depositToCode, "deposit-to", "1000", "deposit account code")
	cmd.Flags().StringVar(&amountStr, "amount", "", "income amount (required)")
	cmd.Flags().StringVar(&date, "date", "", "entry date in YYYY-MM-DD (defaults to today)")
	cmd.Flags().StringVar(&description, "description", "", "journal entry description (required)")
	cmd.Flags().StringVar(&memo, "memo", "", "optional line memo")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if strings.TrimSpace(accountCode) == "" {
			return fmt.Errorf("--account is required")
		}
		if strings.TrimSpace(amountStr) == "" {
			return fmt.Errorf("--amount is required")
		}
		if strings.TrimSpace(description) == "" {
			return fmt.Errorf("--description is required")
		}
		amount, err := tools.ParseAmount(amountStr)
		if err != nil {
			return fmt.Errorf("invalid amount %q: %w", amountStr, err)
		}
		if strings.TrimSpace(date) == "" {
			date = tuiToday()
		}

		return journal.RunIncome(ctx, a.handlerContext(), &journal.IncomeEntryInput{
			Date:               date,
			Description:        description,
			IncomeAccountCode:  accountCode,
			DepositAccountCode: depositToCode,
			Amount:             amount,
			Memo:               memo,
		})
	})
}

func (a *Application) newEntryCustomCommand() *cobra.Command {
	var entryDate, description string
	var debitSpecs, creditSpecs []string

	cmd := &cobra.Command{
		Use:   "custom",
		Short: "Record a guided custom journal entry",
		Long:  entryCustomHelpText,
	}
	cmd.Flags().StringVar(&entryDate, "date", "", "entry date in YYYY-MM-DD (defaults to today)")
	cmd.Flags().StringVar(&description, "description", "", "journal entry description (required)")
	cmd.Flags().StringArrayVar(&debitSpecs, "debit", nil, "debit line in CODE:AMOUNT[:MEMO] format")
	cmd.Flags().StringArrayVar(&creditSpecs, "credit", nil, "credit line in CODE:AMOUNT[:MEMO] format")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if strings.TrimSpace(description) == "" {
			return fmt.Errorf("--description is required")
		}
		if strings.TrimSpace(entryDate) == "" {
			entryDate = tuiToday()
		}

		input, err := journal.BuildJournalPostInput(entryDate, description, debitSpecs, creditSpecs)
		if err != nil {
			return fmt.Errorf("%s\n\n%s", err, entryCustomHelpText)
		}

		return journal.RunCustom(ctx, a.handlerContext(), input)
	})
}

func (a *Application) newQuestCreateCommand() *cobra.Command {
	var title, patron, description, rewardStr, advanceStr, bonus, status, acceptedOn string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a quest register entry",
		Long:  questCreateHelpText,
	}
	cmd.Flags().StringVar(&title, "title", "", "quest title (required)")
	cmd.Flags().StringVar(&patron, "patron", "", "quest patron")
	cmd.Flags().StringVar(&description, "description", "", "quest description")
	cmd.Flags().StringVar(&rewardStr, "reward", "0", "promised base reward")
	cmd.Flags().StringVar(&advanceStr, "advance", "0", "partial advance received")
	cmd.Flags().StringVar(&bonus, "bonus", "", "bonus conditions")
	cmd.Flags().StringVar(&status, "status", "offered", "initial status (offered or accepted)")
	cmd.Flags().StringVar(&acceptedOn, "accepted-on", "", "accepted date (required if status=accepted)")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		reward, err := tools.ParseAmount(rewardStr)
		if err != nil {
			return fmt.Errorf("invalid reward %q: %w", rewardStr, err)
		}
		advance, err := tools.ParseAmount(advanceStr)
		if err != nil {
			return fmt.Errorf("invalid advance %q: %w", advanceStr, err)
		}

		return quest.RunCreate(ctx, a.handlerContext(), &quest.CreateQuestInput{
			Title:              title,
			Patron:             patron,
			Description:        description,
			PromisedBaseReward: reward,
			PartialAdvance:     advance,
			BonusConditions:    bonus,
			Status:             status,
			AcceptedOn:         acceptedOn,
		})
	})
}

func (a *Application) newQuestAcceptCommand() *cobra.Command {
	var id, date string

	cmd := &cobra.Command{
		Use:   "accept",
		Short: "Mark an offered quest as accepted",
		Long:  questAcceptHelpText,
	}
	cmd.Flags().StringVar(&id, "id", "", "quest ID (required)")
	cmd.Flags().StringVar(&date, "date", "", "accepted date in YYYY-MM-DD (required)")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if id == "" {
			return fmt.Errorf("--id is required")
		}
		if date == "" {
			return fmt.Errorf("--date is required")
		}

		return quest.RunAccept(ctx, a.handlerContext(), id, date)
	})
}

func (a *Application) newQuestCompleteCommand() *cobra.Command {
	var id, date string

	cmd := &cobra.Command{
		Use:   "complete",
		Short: "Recognize a quest as earned",
		Long:  questCompleteHelpText,
	}
	cmd.Flags().StringVar(&id, "id", "", "quest ID (required)")
	cmd.Flags().StringVar(&date, "date", "", "completed date in YYYY-MM-DD (required)")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if id == "" {
			return fmt.Errorf("--id is required")
		}
		if date == "" {
			return fmt.Errorf("--date is required")
		}

		return quest.RunComplete(ctx, a.handlerContext(), id, date)
	})
}

func (a *Application) newQuestCollectCommand() *cobra.Command {
	var id, amountStr, date, description string

	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect cash against an earned quest reward",
		Long:  questCollectHelpText,
	}
	cmd.Flags().StringVar(&id, "id", "", "quest ID (required)")
	cmd.Flags().StringVar(&amountStr, "amount", "", "payment amount (required)")
	cmd.Flags().StringVar(&date, "date", "", "payment date in YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&description, "description", "", "optional description")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
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

		return quest.RunCollect(ctx, a.handlerContext(), quest.CollectQuestPaymentInput{
			QuestID:     id,
			Amount:      amount,
			Date:        date,
			Description: description,
		})
	})
}

func (a *Application) newQuestWriteoffCommand() *cobra.Command {
	var id, date, description string

	cmd := &cobra.Command{
		Use:   "writeoff",
		Short: "Write off an uncollectible quest reward",
		Long:  questWriteoffHelpText,
	}
	cmd.Flags().StringVar(&id, "id", "", "quest ID (required)")
	cmd.Flags().StringVar(&date, "date", "", "write-off date in YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&description, "description", "", "optional description")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if id == "" {
			return fmt.Errorf("--id is required")
		}
		if date == "" {
			return fmt.Errorf("--date is required")
		}

		return quest.RunWriteoff(ctx, a.handlerContext(), quest.WriteOffQuestInput{
			QuestID:     id,
			Date:        date,
			Description: description,
		})
	})
}

func (a *Application) newLootCreateCommand() *cobra.Command {
	var name, source, holder, notes string
	var quantity int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a loot register entry",
		Long:  lootCreateHelpText,
	}
	cmd.Flags().StringVar(&name, "name", "", "item name (required)")
	cmd.Flags().StringVar(&source, "source", "", "where the item was found")
	cmd.Flags().IntVar(&quantity, "quantity", 1, "item quantity")
	cmd.Flags().StringVar(&holder, "holder", "", "who is carrying the item")
	cmd.Flags().StringVar(&notes, "notes", "", "additional notes")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		return loot.RunCreate(ctx, a.handlerContext(), name, source, quantity, holder, notes, "loot")
	})
}

func (a *Application) newLootAppraiseCommand() *cobra.Command {
	var id, valueStr, appraiser, date, notes string

	cmd := &cobra.Command{
		Use:   "appraise",
		Short: "Record an off-ledger appraisal",
		Long:  lootAppraiseHelpText,
	}
	cmd.Flags().StringVar(&id, "id", "", "loot item ID (required)")
	cmd.Flags().StringVar(&valueStr, "value", "", "appraised value (required)")
	cmd.Flags().StringVar(&appraiser, "appraiser", "", "who appraised the item")
	cmd.Flags().StringVar(&date, "date", "", "appraisal date in YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&notes, "notes", "", "appraisal notes")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
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

		return loot.RunAppraise(ctx, a.handlerContext(), id, value, appraiser, date, notes)
	})
}

func (a *Application) newLootRecognizeCommand() *cobra.Command {
	var appraisalID, date, description string

	cmd := &cobra.Command{
		Use:   "recognize",
		Short: "Recognize an appraisal on-ledger",
		Long:  lootRecognizeHelpText,
	}
	cmd.Flags().StringVar(&appraisalID, "appraisal-id", "", "appraisal ID (required)")
	cmd.Flags().StringVar(&date, "date", "", "recognition date in YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&description, "description", "", "optional description")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if appraisalID == "" {
			return fmt.Errorf("--appraisal-id is required")
		}
		if date == "" {
			return fmt.Errorf("--date is required")
		}

		return loot.RunRecognize(ctx, a.handlerContext(), appraisalID, date, description)
	})
}

func (a *Application) newLootSellCommand() *cobra.Command {
	var id, amountStr, date, description string

	cmd := &cobra.Command{
		Use:   "sell",
		Short: "Record a loot sale",
		Long:  lootSellHelpText,
	}
	cmd.Flags().StringVar(&id, "id", "", "loot item ID (required)")
	cmd.Flags().StringVar(&amountStr, "amount", "", "sale amount (required)")
	cmd.Flags().StringVar(&date, "date", "", "sale date in YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&description, "description", "", "optional description")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
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

		return loot.RunSell(ctx, a.handlerContext(), id, amount, date, description)
	})
}

func (a *Application) newReportWriteoffCandidatesCommand() *cobra.Command {
	var asOfDate string
	var minAgeDays int

	cmd := &cobra.Command{
		Use:   "writeoff-candidates",
		Short: "Show stale outstanding quest balances",
		Long:  reportWriteoffCandidatesHelpText,
	}
	cmd.Flags().StringVar(&asOfDate, "as-of", time.Now().Format("2006-01-02"), "report date in YYYY-MM-DD")
	cmd.Flags().IntVar(&minAgeDays, "min-age-days", 30, "minimum completed age in days")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		return report.RunWriteOffCandidates(ctx, a.handlerContext(), report.WriteOffCandidateFilter{
			AsOfDate:   asOfDate,
			MinAgeDays: minAgeDays,
		})
	})
}

func isLeafHelpArg(args []string) bool {
	return len(args) == 1 && args[0] == "help"
}

func unexpectedLeafArgsError(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("unexpected arguments for %s: %s\n\n%s", cmd.CommandPath(), strings.Join(args, " "), cmd.Long)
}
