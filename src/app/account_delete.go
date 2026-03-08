package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
)

func (a *Application) runAccountDelete(ctx context.Context, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account delete", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, usageText)
	}

	a.log.logger.InfoContext(ctx, "deleting account",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("code", code),
	)

	if err := repo.DeleteAccount(ctx, a.config.Paths.DatabasePath, code); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to delete account", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(ctx, "deleted account", slog.String("code", code))

	if _, err := fmt.Fprintf(a.stdout, "Deleted account %s\n", code); err != nil {
		return fmt.Errorf("write delete output: %w", err)
	}

	return nil
}
