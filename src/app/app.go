package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/repo"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
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
		return a.runAccountList(ctx)
	case "create":
		return a.runAccountCreate(ctx, args[1:])
	case "rename":
		return a.runAccountRename(ctx, args[1:])
	case "deactivate":
		return a.runAccountDeactivate(ctx, args[1:])
	case "activate":
		return a.runAccountActivate(ctx, args[1:])
	default:
		return fmt.Errorf("unknown account subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runAccountList(ctx context.Context) error {
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
}

func (a *Application) runAccountCreate(ctx context.Context, args []string) error {
	var code, name, accountType string

	flagSet := flag.NewFlagSet("account create", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")
	flagSet.StringVar(&name, "name", "", "account name")
	flagSet.StringVar(&accountType, "type", "", "account type (asset, liability, equity, income, expense)")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, usageText)
	}

	a.log.logger.InfoContext(ctx, "creating account",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("code", code),
		slog.String("name", name),
		slog.String("type", accountType),
	)

	result, err := repo.CreateAccount(ctx, a.config.Paths.DatabasePath, code, name, service.AccountType(accountType))
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to create account", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "created account", slog.String("code", result.Code), slog.String("id", result.ID))

	if _, err := fmt.Fprintf(
		a.stdout,
		"Created account %s\nCode: %s\nName: %s\nType: %s\n",
		result.ID,
		result.Code,
		result.Name,
		string(result.Type),
	); err != nil {
		return fmt.Errorf("write account output: %w", err)
	}

	return nil
}

func (a *Application) runAccountRename(ctx context.Context, args []string) error {
	var code, name string

	flagSet := flag.NewFlagSet("account rename", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")
	flagSet.StringVar(&name, "name", "", "new account name")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, usageText)
	}

	a.log.logger.InfoContext(ctx, "renaming account",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("code", code),
		slog.String("name", name),
	)

	if err := repo.RenameAccount(ctx, a.config.Paths.DatabasePath, code, name); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to rename account", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "renamed account", slog.String("code", code), slog.String("name", name))

	if _, err := fmt.Fprintf(a.stdout, "Renamed account %s to %q\n", code, name); err != nil {
		return fmt.Errorf("write rename output: %w", err)
	}

	return nil
}

func (a *Application) runAccountDeactivate(ctx context.Context, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account deactivate", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, usageText)
	}

	a.log.logger.InfoContext(ctx, "deactivating account",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("code", code),
	)

	if err := repo.DeactivateAccount(ctx, a.config.Paths.DatabasePath, code); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to deactivate account", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "deactivated account", slog.String("code", code))

	if _, err := fmt.Fprintf(a.stdout, "Deactivated account %s\n", code); err != nil {
		return fmt.Errorf("write deactivate output: %w", err)
	}

	return nil
}

func (a *Application) runAccountActivate(ctx context.Context, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account activate", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, usageText)
	}

	a.log.logger.InfoContext(ctx, "activating account",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("code", code),
	)

	if err := repo.ActivateAccount(ctx, a.config.Paths.DatabasePath, code); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to activate account", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "activated account", slog.String("code", code))

	if _, err := fmt.Fprintf(a.stdout, "Activated account %s\n", code); err != nil {
		return fmt.Errorf("write activate output: %w", err)
	}

	return nil
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
  lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]
  lootsheet help
`
