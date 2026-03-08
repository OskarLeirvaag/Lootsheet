package app

import (
	"context"
	"fmt"
	"io"
	"strings"

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
		Short:             "Local-first D&D 5e double-entry bookkeeping CLI",
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
			return a.runAccount(ctx, []string{"list"})
		}),
		a.newPassthroughLeafCommand("create", "Create a new account", accountCreateHelpText, func(ctx context.Context, args []string) error {
			return a.runAccount(ctx, append([]string{"create"}, args...))
		}),
		a.newPassthroughLeafCommand("rename", "Rename an existing account", accountRenameHelpText, func(ctx context.Context, args []string) error {
			return a.runAccount(ctx, append([]string{"rename"}, args...))
		}),
		a.newPassthroughLeafCommand("deactivate", "Deactivate an account", accountDeactivateHelpText, func(ctx context.Context, args []string) error {
			return a.runAccount(ctx, append([]string{"deactivate"}, args...))
		}),
		a.newPassthroughLeafCommand("activate", "Activate an account", accountActivateHelpText, func(ctx context.Context, args []string) error {
			return a.runAccount(ctx, append([]string{"activate"}, args...))
		}),
		a.newPassthroughLeafCommand("delete", "Delete an unused account", accountDeleteHelpText, func(ctx context.Context, args []string) error {
			return a.runAccount(ctx, append([]string{"delete"}, args...))
		}),
		a.newPassthroughLeafCommand("ledger", "Show a single-account ledger report", accountLedgerHelpText, func(ctx context.Context, args []string) error {
			return a.runAccount(ctx, append([]string{"ledger"}, args...))
		}),
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
		a.newPassthroughLeafCommand("post", "Post a balanced journal entry", journalPostHelpText, func(ctx context.Context, args []string) error {
			return a.runJournal(ctx, append([]string{"post"}, args...))
		}),
		a.newPassthroughLeafCommand("reverse", "Reverse a posted journal entry", journalReverseHelpText, func(ctx context.Context, args []string) error {
			return a.runJournal(ctx, append([]string{"reverse"}, args...))
		}),
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
		a.newPassthroughLeafCommand("create", "Create a quest register entry", questCreateHelpText, func(ctx context.Context, args []string) error {
			return a.runQuest(ctx, append([]string{"create"}, args...))
		}),
		a.newNoArgsLeafCommand("list", "List quest register entries", questListHelpText, func(ctx context.Context) error {
			return a.runQuest(ctx, []string{"list"})
		}),
		a.newPassthroughLeafCommand("accept", "Mark an offered quest as accepted", questAcceptHelpText, func(ctx context.Context, args []string) error {
			return a.runQuest(ctx, append([]string{"accept"}, args...))
		}),
		a.newPassthroughLeafCommand("complete", "Recognize a quest as earned", questCompleteHelpText, func(ctx context.Context, args []string) error {
			return a.runQuest(ctx, append([]string{"complete"}, args...))
		}),
		a.newPassthroughLeafCommand("collect", "Collect cash against an earned quest reward", questCollectHelpText, func(ctx context.Context, args []string) error {
			return a.runQuest(ctx, append([]string{"collect"}, args...))
		}),
		a.newPassthroughLeafCommand("writeoff", "Write off an uncollectible quest reward", questWriteoffHelpText, func(ctx context.Context, args []string) error {
			return a.runQuest(ctx, append([]string{"writeoff"}, args...))
		}),
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
		a.newPassthroughLeafCommand("create", "Create a loot register entry", lootCreateHelpText, func(ctx context.Context, args []string) error {
			return a.runLoot(ctx, append([]string{"create"}, args...))
		}),
		a.newNoArgsLeafCommand("list", "List tracked loot items", lootListHelpText, func(ctx context.Context) error {
			return a.runLoot(ctx, []string{"list"})
		}),
		a.newPassthroughLeafCommand("appraise", "Record an off-ledger appraisal", lootAppraiseHelpText, func(ctx context.Context, args []string) error {
			return a.runLoot(ctx, append([]string{"appraise"}, args...))
		}),
		a.newPassthroughLeafCommand("recognize", "Recognize an appraisal on-ledger", lootRecognizeHelpText, func(ctx context.Context, args []string) error {
			return a.runLoot(ctx, append([]string{"recognize"}, args...))
		}),
		a.newPassthroughLeafCommand("sell", "Record a loot sale", lootSellHelpText, func(ctx context.Context, args []string) error {
			return a.runLoot(ctx, append([]string{"sell"}, args...))
		}),
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
			return a.runReport(ctx, []string{"trial-balance"})
		}),
		a.newNoArgsLeafCommand("quest-receivables", "Show earned but unpaid quest rewards", reportQuestReceivablesHelpText, func(ctx context.Context) error {
			return a.runReport(ctx, []string{"quest-receivables"})
		}),
		a.newNoArgsLeafCommand("promised-quests", "Show promised but unearned quests", reportPromisedQuestsHelpText, func(ctx context.Context) error {
			return a.runReport(ctx, []string{"promised-quests"})
		}),
		a.newNoArgsLeafCommand("loot-summary", "Show held and recognized loot", reportLootSummaryHelpText, func(ctx context.Context) error {
			return a.runReport(ctx, []string{"loot-summary"})
		}),
		a.newPassthroughLeafCommand("writeoff-candidates", "Show stale outstanding quest balances", reportWriteoffCandidatesHelpText, func(ctx context.Context, args []string) error {
			return a.runReport(ctx, append([]string{"writeoff-candidates"}, args...))
		}),
	)

	return cmd
}

func (a *Application) newNoArgsLeafCommand(use string, short string, helpText string, run func(context.Context) error) *cobra.Command {
	return &cobra.Command{
		Use:                use,
		Short:              short,
		Long:               helpText,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if hasHelpToken(args) {
				return a.writeCommandHelp(cmd)
			}

			if len(args) > 0 {
				return fmt.Errorf("unexpected arguments for %s: %s\n\n%s", cmd.CommandPath(), strings.Join(args, " "), helpText)
			}

			return run(cmd.Context())
		},
	}
}

func (a *Application) newPassthroughLeafCommand(use string, short string, helpText string, run func(context.Context, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:                use,
		Short:              short,
		Long:               helpText,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if hasHelpToken(args) {
				return a.writeCommandHelp(cmd)
			}

			return run(cmd.Context(), args)
		},
	}
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

func hasHelpToken(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "help", "-h", "--help":
			return true
		}
	}

	return false
}
