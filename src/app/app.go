package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/repo"
)

type Application struct {
	config config.Config
	stdout io.Writer
	log    *appLogger
}

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

	result, err := repo.EnsureSQLiteInitialized(ctx, a.config.Paths.DatabasePath, initAssets)
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

	switch args[0] {
	case "list":
		a.log.logger.InfoContext(ctx, "listing accounts", slog.String("database_path", a.config.Paths.DatabasePath))

		accounts, err := repo.ListAccounts(ctx, a.config.Paths.DatabasePath)
		if err != nil {
			a.log.logger.ErrorContext(ctx, "failed to list accounts", slog.String("error", err.Error()))
			return err
		}

		if _, err := fmt.Fprintln(a.stdout, "CODE  TYPE       ACTIVE  NAME"); err != nil {
			return fmt.Errorf("write accounts header: %w", err)
		}

		for _, account := range accounts {
			activeLabel := "no"
			if account.Active {
				activeLabel = "yes"
			}

			if _, err := fmt.Fprintf(
				a.stdout,
				"%-4s  %-10s %-6s  %s\n",
				account.Code,
				string(account.Type),
				activeLabel,
				account.Name,
			); err != nil {
				return fmt.Errorf("write account row: %w", err)
			}
		}

		a.log.logger.InfoContext(ctx, "listed accounts", slog.Int("count", len(accounts)))
		return nil
	default:
		return fmt.Errorf("unknown account subcommand %q\n\n%s", args[0], usageText)
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
  lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]
  lootsheet help
`
