package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
)

func (a *Application) runDatabase(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing db subcommand\n\n%s", usageText)
	}

	switch args[0] {
	case "status":
		return a.runDatabaseStatus(ctx)
	default:
		return fmt.Errorf("unknown db subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runDatabaseStatus(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "reading database status", slog.String("database_path", a.config.Paths.DatabasePath))

	status, err := repo.GetDatabaseStatus(ctx, a.config.Paths.DatabasePath)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to read database status", slog.String("error", err.Error()))
		return err
	}

	stateLabel := "uninitialized"
	if status.Initialized {
		stateLabel = "initialized"
	}

	existsLabel := "no"
	if status.Exists {
		existsLabel = "yes"
	}

	if _, err := fmt.Fprintf(
		a.stdout,
		"Database: %s\nExists: %s\nState: %s\nSchema version: %s\nApplied migrations: %d\n",
		a.config.Paths.DatabasePath,
		existsLabel,
		stateLabel,
		blankIfEmpty(status.SchemaVersion),
		len(status.AppliedMigrations),
	); err != nil {
		return fmt.Errorf("write database status output: %w", err)
	}

	for _, migration := range status.AppliedMigrations {
		if _, err := fmt.Fprintf(a.stdout, "%s  %s  %s\n", migration.Version, migration.Name, migration.AppliedAt); err != nil {
			return fmt.Errorf("write migration status row: %w", err)
		}
	}

	a.log.logger.InfoContext(
		ctx,
		"read database status",
		slog.Bool("exists", status.Exists),
		slog.Bool("initialized", status.Initialized),
		slog.String("schema_version", status.SchemaVersion),
		slog.Int("applied_migrations", len(status.AppliedMigrations)),
	)

	return nil
}

func blankIfEmpty(value string) string {
	if value == "" {
		return "-"
	}

	return value
}
