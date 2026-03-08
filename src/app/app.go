// Package app wires together configuration, logging, and CLI command routing
// for the LootSheet application.
package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/OskarLeirvaag/Lootsheet/src/account"
	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
)

// Application holds the runtime state for a single CLI invocation, including
// resolved configuration, output destination, and structured logger.
type Application struct {
	config config.Config
	stdout io.Writer
	log    *appLogger
}

// Run is the top-level entry point that loads configuration, creates an
// Application, and dispatches to the appropriate subcommand.
func Run(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	application, err := New(cfg, stdout, os.Stderr)
	if err != nil {
		return err
	}

	return application.Run(ctx, args)
}

// New creates a new Application with the given configuration, stdout destination,
// and log output writer. It validates the configuration and initializes the
// OTel-backed structured logger.
func New(cfg config.Config, stdout io.Writer, logOutput io.Writer) (*Application, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if stdout == nil {
		stdout = os.Stdout
	}

	logger, err := newAppLogger(logOutput)
	if err != nil {
		return nil, err
	}

	return &Application{
		config: cfg,
		stdout: stdout,
		log:    logger,
	}, nil
}

func (a *Application) handlerContext() ledger.HandlerContext {
	return ledger.HandlerContext{
		DatabasePath: a.config.Paths.DatabasePath,
		Stdout:       a.stdout,
	}
}

// Run parses the top-level command from args and dispatches to the appropriate
// subcommand handler. It shuts down the logger on return.
func (a *Application) Run(ctx context.Context, args []string) error {
	defer func() {
		if a.log != nil && a.log.shutdown != nil {
			_ = a.log.shutdown(ctx)
		}
	}()

	a.log.logger.DebugContext(ctx, "command start", slog.String("args", fmt.Sprint(args)))

	if len(args) == 0 {
		return a.printUsage()
	}

	switch args[0] {
	case "help", "-h", "--help":
		return a.printUsage()
	case "db":
		return a.runDatabase(ctx, args[1:])
	case "init":
		return a.runInit(ctx)
	case "account":
		return a.runAccount(ctx, args[1:])
	case "journal":
		return a.runJournal(ctx, args[1:])
	case "quest":
		return a.runQuest(ctx, args[1:])
	case "loot":
		return a.runLoot(ctx, args[1:])
	case "report":
		return a.runReport(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runInit(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "initializing database", slog.String("database_path", a.config.Paths.DatabasePath))

	if err := a.config.EnsureDirectories(); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to prepare directories", slog.String("error", err.Error()))
		return err
	}

	initAssets, err := config.LoadInitAssets()
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to load init assets", slog.String("error", err.Error()))
		return err
	}

	result, err := ledger.EnsureSQLiteInitialized(ctx, a.config.Paths.DatabasePath, initAssets)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to initialize sqlite database", slog.String("error", err.Error()))
		return err
	}

	status := "already initialized"
	seededAccounts := 0
	if result.Initialized {
		status = "initialized"
		seededAccounts = result.SeededCounts.Accounts
	}

	a.log.logger.InfoContext(
		ctx,
		"database initialization finished",
		slog.String("status", status),
		slog.Int("seeded_accounts", seededAccounts),
	)

	if _, err := fmt.Fprintf(
		a.stdout,
		"LootSheet %s\nConfig: %s\nDatabase: %s\nSeeded accounts: %d\n",
		status,
		a.config.Paths.ConfigFile,
		a.config.Paths.DatabasePath,
		seededAccounts,
	); err != nil {
		return fmt.Errorf("write init output: %w", err)
	}

	return nil
}

func (a *Application) runAccount(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing account subcommand\n\n%s", usageText)
	}

	hctx := a.handlerContext()

	switch args[0] {
	case "list":
		return account.HandleList(ctx, hctx)
	case "create":
		return account.HandleCreate(ctx, hctx, args[1:])
	case "rename":
		return account.HandleRename(ctx, hctx, args[1:])
	case "deactivate":
		return account.HandleDeactivate(ctx, hctx, args[1:])
	case "activate":
		return account.HandleActivate(ctx, hctx, args[1:])
	case "ledger":
		return journal.HandleAccountLedger(ctx, hctx, args[1:])
	case "delete":
		return account.HandleDelete(ctx, hctx, args[1:])
	default:
		return fmt.Errorf("unknown account subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runJournal(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing journal subcommand\n\n%s", usageText)
	}

	hctx := a.handlerContext()

	switch args[0] {
	case "post":
		return journal.HandlePost(ctx, hctx, args[1:])
	case "reverse":
		return journal.HandleReverse(ctx, hctx, args[1:])
	default:
		return fmt.Errorf("unknown journal subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runQuest(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing quest subcommand\n\n%s", usageText)
	}

	hctx := a.handlerContext()

	switch args[0] {
	case "create":
		return quest.HandleCreate(ctx, hctx, args[1:])
	case "list":
		return quest.HandleList(ctx, hctx)
	case "accept":
		return quest.HandleAccept(ctx, hctx, args[1:])
	case "complete":
		return quest.HandleComplete(ctx, hctx, args[1:])
	case "collect":
		return quest.HandleCollect(ctx, hctx, args[1:])
	case "writeoff":
		return quest.HandleWriteoff(ctx, hctx, args[1:])
	default:
		return fmt.Errorf("unknown quest subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runLoot(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing loot subcommand\n\n%s", usageText)
	}

	hctx := a.handlerContext()

	switch args[0] {
	case "create":
		return loot.HandleCreate(ctx, hctx, args[1:])
	case "list":
		return loot.HandleList(ctx, hctx)
	case "appraise":
		return loot.HandleAppraise(ctx, hctx, args[1:])
	case "recognize":
		return loot.HandleRecognize(ctx, hctx, args[1:])
	case "sell":
		return loot.HandleSell(ctx, hctx, args[1:])
	default:
		return fmt.Errorf("unknown loot subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runReport(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing report subcommand\n\n%s", usageText)
	}

	hctx := a.handlerContext()

	switch args[0] {
	case "trial-balance":
		return report.HandleTrialBalance(ctx, hctx)
	case "quest-receivables":
		return report.HandleQuestReceivables(ctx, hctx)
	case "promised-quests":
		return report.HandlePromisedQuests(ctx, hctx)
	case "loot-summary":
		return report.HandleLootSummary(ctx, hctx)
	case "writeoff-candidates":
		return report.HandleWriteOffCandidates(ctx, hctx, args[1:])
	default:
		return fmt.Errorf("unknown report subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) printUsage() error {
	_, err := io.WriteString(a.stdout, usageText)
	return err
}

const usageText = `LootSheet CLI

Usage:
  lootsheet db status
  lootsheet db migrate
  lootsheet init
  lootsheet account list
  lootsheet account create --code CODE --name NAME --type TYPE
  lootsheet account rename --code CODE --name NAME
  lootsheet account deactivate --code CODE
  lootsheet account activate --code CODE
  lootsheet account delete --code CODE
  lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]
  lootsheet journal reverse --entry-id UUID --date YYYY-MM-DD [--description TEXT]
  lootsheet quest create --title TEXT [--patron TEXT] [--description TEXT] [--reward AMOUNT] [--advance AMOUNT] [--bonus TEXT] [--status offered|accepted] [--accepted-on DATE]
  lootsheet quest list
  lootsheet quest accept --id ID --date YYYY-MM-DD
  lootsheet quest complete --id ID --date YYYY-MM-DD
  lootsheet quest collect --id ID --amount AMOUNT --date YYYY-MM-DD [--description TEXT]
  lootsheet quest writeoff --id ID --date YYYY-MM-DD [--description TEXT]
  lootsheet loot create --name TEXT [--source TEXT] [--quantity N] [--holder TEXT] [--notes TEXT]
  lootsheet loot list
  lootsheet loot appraise --id ID --value AMOUNT --date DATE [--appraiser TEXT] [--notes TEXT]
  lootsheet loot recognize --appraisal-id ID --date DATE [--description TEXT]
  lootsheet loot sell --id ID --amount AMOUNT --date DATE [--description TEXT]
  lootsheet account ledger --code CODE
  lootsheet report trial-balance
  lootsheet report quest-receivables
  lootsheet report promised-quests
  lootsheet report loot-summary
  lootsheet report writeoff-candidates [--as-of YYYY-MM-DD] [--min-age-days N]
  lootsheet help
`
