package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/account"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/notes"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
	"github.com/spf13/cobra"
)

// defaultMinAgeDays is the default minimum completed age in days for write-off candidates.
const defaultMinAgeDays = 30

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
			return fmt.Errorf("%w\n\n%s", err, journalPostHelpText)
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
	return a.newGuidedEntryCommand(&entryCommandParams{
		use:            "expense",
		short:          "Record a guided expense entry",
		helpText:       entryExpenseHelpText,
		offsetFlagName: "paid-from",
		offsetFlagHelp: "funding account code",
		post: func(ctx context.Context, hctx ledger.HandlerContext, date, description, accountCode, offsetCode string, amount int64, memo string) error {
			return journal.RunExpense(ctx, hctx, &journal.ExpenseEntryInput{
				Date:               date,
				Description:        description,
				ExpenseAccountCode: accountCode,
				FundingAccountCode: offsetCode,
				Amount:             amount,
				Memo:               memo,
			})
		},
	})
}

func (a *Application) newEntryIncomeCommand() *cobra.Command {
	return a.newGuidedEntryCommand(&entryCommandParams{
		use:            "income",
		short:          "Record a guided income entry",
		helpText:       entryIncomeHelpText,
		offsetFlagName: "deposit-to",
		offsetFlagHelp: "deposit account code",
		post: func(ctx context.Context, hctx ledger.HandlerContext, date, description, accountCode, offsetCode string, amount int64, memo string) error {
			return journal.RunIncome(ctx, hctx, &journal.IncomeEntryInput{
				Date:               date,
				Description:        description,
				IncomeAccountCode:  accountCode,
				DepositAccountCode: offsetCode,
				Amount:             amount,
				Memo:               memo,
			})
		},
	})
}

type entryCommandParams struct {
	use            string
	short          string
	helpText       string
	offsetFlagName string
	offsetFlagHelp string
	post           func(ctx context.Context, hctx ledger.HandlerContext, date, description, accountCode, offsetCode string, amount int64, memo string) error
}

func (a *Application) newGuidedEntryCommand(p *entryCommandParams) *cobra.Command {
	var accountCode, offsetCode, amountStr, date, description, memo string

	cmd := &cobra.Command{
		Use:   p.use,
		Short: p.short,
		Long:  p.helpText,
	}
	cmd.Flags().StringVar(&accountCode, "account", "", p.use+" account code (required)")
	cmd.Flags().StringVar(&offsetCode, p.offsetFlagName, "1000", p.offsetFlagHelp)
	cmd.Flags().StringVar(&amountStr, "amount", "", p.use+" amount (required)")
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
		amount, err := currency.ParseAmount(amountStr)
		if err != nil {
			return fmt.Errorf("invalid amount %q: %w", amountStr, err)
		}
		if strings.TrimSpace(date) == "" {
			date = tuiToday()
		}

		return p.post(ctx, a.handlerContext(), date, description, accountCode, offsetCode, amount, memo)
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
			return fmt.Errorf("%w\n\n%s", err, entryCustomHelpText)
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
		reward, err := currency.ParseAmount(rewardStr)
		if err != nil {
			return fmt.Errorf("invalid reward %q: %w", rewardStr, err)
		}
		advance, err := currency.ParseAmount(advanceStr)
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

		amount, err := currency.ParseAmount(amountStr)
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

		value, err := currency.ParseAmount(valueStr)
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

		amount, err := currency.ParseAmount(amountStr)
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
	cmd.Flags().IntVar(&minAgeDays, "min-age-days", defaultMinAgeDays, "minimum completed age in days")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		return report.RunWriteOffCandidates(ctx, a.handlerContext(), report.WriteOffCandidateFilter{
			AsOfDate:   asOfDate,
			MinAgeDays: minAgeDays,
		})
	})
}

func (a *Application) newCodexCreateCommand() *cobra.Command {
	var name, typeID, title, location, faction, disposition, class, race, background, description, notes string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new codex entry",
		Long:  codexCreateHelpText,
	}
	cmd.Flags().StringVar(&name, "name", "", "entry name (required)")
	cmd.Flags().StringVar(&typeID, "type", "npc", "entry type ID (e.g. player, npc)")
	cmd.Flags().StringVar(&title, "title", "", "title or role")
	cmd.Flags().StringVar(&location, "location", "", "where the entry can be found")
	cmd.Flags().StringVar(&faction, "faction", "", "faction or group affiliation")
	cmd.Flags().StringVar(&disposition, "disposition", "", "friendly, neutral, hostile, etc.")
	cmd.Flags().StringVar(&class, "class", "", "character class (player type)")
	cmd.Flags().StringVar(&race, "race", "", "character race (player type)")
	cmd.Flags().StringVar(&background, "background", "", "character background (player type)")
	cmd.Flags().StringVar(&description, "description", "", "physical or personality description")
	cmd.Flags().StringVar(&notes, "notes", "", "notes text; use @type/name for cross-references")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		return codex.RunCreate(ctx, a.handlerContext(), &codex.CreateInput{
			TypeID:      typeID,
			Name:        name,
			Title:       title,
			Location:    location,
			Faction:     faction,
			Disposition: disposition,
			Class:       class,
			Race:        race,
			Background:  background,
			Description: description,
			Notes:       notes,
		})
	})
}

func (a *Application) newNotesCreateCommand() *cobra.Command {
	var title, body string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new note",
		Long:  notesCreateHelpText,
	}
	cmd.Flags().StringVar(&title, "title", "", "note title (required)")
	cmd.Flags().StringVar(&body, "body", "", "note body; use @type/name for cross-references")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if title == "" {
			return fmt.Errorf("--title is required")
		}

		return notes.RunCreate(ctx, a.handlerContext(), &notes.CreateNoteInput{
			Title: title,
			Body:  body,
		})
	})
}

func (a *Application) newNotesSearchCommand() *cobra.Command {
	var query string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search notes by title or body content",
		Long:  notesSearchHelpText,
	}
	cmd.Flags().StringVar(&query, "query", "", "search query (required)")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if query == "" {
			return fmt.Errorf("--query is required")
		}

		return notes.RunSearch(ctx, a.handlerContext(), query)
	})
}

func (a *Application) newCodexSearchCommand() *cobra.Command {
	var query string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search codex entries by name, title, location, faction, class, race, or notes",
		Long:  codexSearchHelpText,
	}
	cmd.Flags().StringVar(&query, "query", "", "search query (required)")

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		if query == "" {
			return fmt.Errorf("--query is required")
		}

		return codex.RunSearch(ctx, a.handlerContext(), query)
	})
}

func isLeafHelpArg(args []string) bool {
	return len(args) == 1 && args[0] == "help"
}

func unexpectedLeafArgsError(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("unexpected arguments for %s: %s\n\n%s", cmd.CommandPath(), strings.Join(args, " "), cmd.Long)
}
