package app

import (
	"context"
	"io"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/account"
	"github.com/OskarLeirvaag/Lootsheet/src/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
	"github.com/spf13/cobra"
)

func (a *Application) executeRootCommand(ctx context.Context, args []string) error {
	root := a.newRootCommand()
	root.SetArgs(args)
	root.SetOut(a.stdout)
	root.SetErr(io.Discard)

	return root.ExecuteContext(ctx)
}

func (a *Application) newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:               "lootsheet",
		Short:             "Local-first D&D 5e double-entry bookkeeping CLI/TUI",
		Long:              rootHelpText,
		SilenceErrors:     true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_ = a.writeCommandHelp(cmd)
	})

	root.AddCommand(
		a.newDatabaseCommand(),
		a.newInitCommand(),
		a.newTUICommand(),
		a.newAccountCommand(),
		a.newJournalCommand(),
		a.newQuestCommand(),
		a.newLootCommand(),
		a.newReportCommand(),
	)

	return root
}

func (a *Application) newDatabaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Inspect database state and run schema migrations",
		Long:  dbHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	cmd.AddCommand(
		a.newNoArgsLeafCommand("status", "Show configured database lifecycle state", dbStatusHelpText, func(ctx context.Context) error {
			return a.runDatabase(ctx, []string{"status"})
		}),
		a.newNoArgsLeafCommand("migrate", "Apply pending embedded schema migrations", dbMigrateHelpText, func(ctx context.Context) error {
			return a.runDatabase(ctx, []string{"migrate"})
		}),
	)

	return cmd
}

func (a *Application) newInitCommand() *cobra.Command {
	return a.newNoArgsLeafCommand("init", "Initialize a fresh LootSheet database", initHelpText, func(ctx context.Context) error {
		return a.runInit(ctx)
	})
}

func (a *Application) newAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage chart-of-accounts records",
		Long:  accountHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	cmd.AddCommand(
		a.newNoArgsLeafCommand("list", "Show the chart of accounts", accountListHelpText, func(ctx context.Context) error {
			return account.RunList(ctx, a.handlerContext())
		}),
		a.newAccountCreateCommand(),
		a.newAccountRenameCommand(),
		a.newAccountDeactivateCommand(),
		a.newAccountActivateCommand(),
		a.newAccountDeleteCommand(),
		a.newAccountLedgerCommand(),
	)

	return cmd
}

func (a *Application) newJournalCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "journal",
		Short: "Post and reverse balanced journal entries",
		Long:  journalHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	cmd.AddCommand(
		a.newJournalPostCommand(),
		a.newJournalReverseCommand(),
	)

	return cmd
}

func (a *Application) newQuestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quest",
		Short: "Track promised, earned, and collected quest rewards",
		Long:  questHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	cmd.AddCommand(
		a.newQuestCreateCommand(),
		a.newNoArgsLeafCommand("list", "List quest register entries", questListHelpText, func(ctx context.Context) error {
			return quest.RunList(ctx, a.handlerContext())
		}),
		a.newQuestAcceptCommand(),
		a.newQuestCompleteCommand(),
		a.newQuestCollectCommand(),
		a.newQuestWriteoffCommand(),
	)

	return cmd
}

func (a *Application) newLootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loot",
		Short: "Track loot appraisal, recognition, and sale workflows",
		Long:  lootHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	cmd.AddCommand(
		a.newLootCreateCommand(),
		a.newNoArgsLeafCommand("list", "List tracked loot items", lootListHelpText, func(ctx context.Context) error {
			return loot.RunList(ctx, a.handlerContext())
		}),
		a.newLootAppraiseCommand(),
		a.newLootRecognizeCommand(),
		a.newLootSellCommand(),
	)

	return cmd
}

func (a *Application) newReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Run read-only accounting and register reports",
		Long:  reportHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	cmd.AddCommand(
		a.newNoArgsLeafCommand("trial-balance", "Show the trial balance", reportTrialBalanceHelpText, func(ctx context.Context) error {
			return report.RunTrialBalance(ctx, a.handlerContext())
		}),
		a.newNoArgsLeafCommand("quest-receivables", "Show earned but unpaid quest rewards", reportQuestReceivablesHelpText, func(ctx context.Context) error {
			return report.RunQuestReceivables(ctx, a.handlerContext())
		}),
		a.newNoArgsLeafCommand("promised-quests", "Show promised but unearned quests", reportPromisedQuestsHelpText, func(ctx context.Context) error {
			return report.RunPromisedQuests(ctx, a.handlerContext())
		}),
		a.newNoArgsLeafCommand("loot-summary", "Show held and recognized loot", reportLootSummaryHelpText, func(ctx context.Context) error {
			return report.RunLootSummary(ctx, a.handlerContext())
		}),
		a.newReportWriteoffCandidatesCommand(),
	)

	return cmd
}

func (a *Application) writeCommandHelp(cmd *cobra.Command) error {
	helpText := strings.TrimSpace(cmd.Long)
	if helpText == "" {
		helpText = strings.TrimSpace(cmd.Short)
	}
	if helpText == "" {
		return nil
	}

	_, err := io.WriteString(a.stdout, helpText+"\n")
	return err
}
