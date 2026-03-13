// Package app wires together configuration, logging, and CLI command routing
// for the LootSheet application.
package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
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

	application, err := New(&cfg, stdout, os.Stderr)
	if err != nil {
		return err
	}

	return application.Run(ctx, args)
}

// New creates a new Application with the given configuration, stdout destination,
// and log output writer. It validates the configuration and initializes the
// structured logger.
func New(cfg *config.Config, stdout io.Writer, logOutput io.Writer) (*Application, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

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
		config: *cfg,
		stdout: stdout,
		log:    logger,
	}, nil
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

	return a.executeRootCommand(ctx, args)
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
