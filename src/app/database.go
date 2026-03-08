package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

func (a *Application) runDatabase(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing db subcommand\n\n%s", dbHelpText)
	}

	switch args[0] {
	case "status":
		return a.runDatabaseStatus(ctx)
	case "migrate":
		return a.runDatabaseMigrate(ctx)
	default:
		return fmt.Errorf("unknown db subcommand %q\n\n%s", args[0], dbHelpText)
	}
}

func (a *Application) runDatabaseStatus(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "reading database status", slog.String("database_path", a.config.Paths.DatabasePath))

	initAssets, err := config.LoadInitAssets()
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to load init assets", slog.String("error", err.Error()))
		return err
	}

	status, err := ledger.GetDatabaseStatusWithAssets(ctx, a.config.Paths.DatabasePath, initAssets)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to read database status", slog.String("error", err.Error()))
		return err
	}

	existsLabel := "no"
	if status.Exists {
		existsLabel = "yes"
	}

	if _, err := fmt.Fprintf(
		a.stdout,
		"Database: %s\nExists: %s\nState: %s\nDetail: %s\nSchema version: %s\nTarget schema version: %s\nApplied migrations: %d\nPending migrations: %d\n",
		a.config.Paths.DatabasePath,
		existsLabel,
		status.State,
		blankIfEmpty(status.Detail),
		blankIfEmpty(status.SchemaVersion),
		blankIfEmpty(status.TargetSchemaVersion),
		len(status.AppliedMigrations),
		len(status.PendingMigrations),
	); err != nil {
		return fmt.Errorf("write database status output: %w", err)
	}

	if len(status.AppliedMigrations) > 0 {
		if _, err := fmt.Fprintln(a.stdout, "Applied:"); err != nil {
			return fmt.Errorf("write applied migrations header: %w", err)
		}
	}

	for _, migration := range status.AppliedMigrations {
		if _, err := fmt.Fprintf(a.stdout, "%s  %s  %s\n", migration.Version, migration.Name, migration.AppliedAt); err != nil {
			return fmt.Errorf("write migration status row: %w", err)
		}
	}

	if len(status.PendingMigrations) > 0 {
		if _, err := fmt.Fprintln(a.stdout, "Pending:"); err != nil {
			return fmt.Errorf("write pending migrations header: %w", err)
		}
	}

	for _, migration := range status.PendingMigrations {
		if _, err := fmt.Fprintf(a.stdout, "%s  %s\n", migration.Version, migration.Name); err != nil {
			return fmt.Errorf("write pending migration row: %w", err)
		}
	}

	a.log.logger.InfoContext(
		ctx,
		"read database status",
		slog.Bool("exists", status.Exists),
		slog.Bool("initialized", status.Initialized),
		slog.String("state", string(status.State)),
		slog.String("detail", status.Detail),
		slog.String("schema_version", status.SchemaVersion),
		slog.String("target_schema_version", status.TargetSchemaVersion),
		slog.Int("applied_migrations", len(status.AppliedMigrations)),
		slog.Int("pending_migrations", len(status.PendingMigrations)),
	)

	return nil
}

func (a *Application) runDatabaseMigrate(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "migrating database", slog.String("database_path", a.config.Paths.DatabasePath))

	if err := a.config.EnsureDirectories(); err != nil {
		a.log.logger.ErrorContext(ctx, "failed to prepare directories", slog.String("error", err.Error()))
		return err
	}

	initAssets, err := config.LoadInitAssets()
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to load init assets", slog.String("error", err.Error()))
		return err
	}

	result, err := ledger.MigrateSQLiteDatabase(ctx, a.config.Paths.DatabasePath, a.config.Paths.BackupDir, initAssets)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to migrate database", slog.String("error", err.Error()))
		return err
	}

	stateLabel := "current"
	switch {
	case result.Migrated:
		stateLabel = "migrated"
	case result.MetadataRepaired:
		stateLabel = "metadata repaired"
	}

	if _, err := fmt.Fprintf(
		a.stdout,
		"Database: %s\nState: %s\nBackup: %s\nFrom schema version: %s\nTo schema version: %s\nApplied migrations: %d\n",
		a.config.Paths.DatabasePath,
		stateLabel,
		blankIfEmpty(result.BackupPath),
		blankIfEmpty(result.FromSchemaVersion),
		blankIfEmpty(result.ToSchemaVersion),
		len(result.AppliedMigrations),
	); err != nil {
		return fmt.Errorf("write database migrate output: %w", err)
	}

	for _, migration := range result.AppliedMigrations {
		if _, err := fmt.Fprintf(a.stdout, "%s  %s\n", migration.Version, migration.Name); err != nil {
			return fmt.Errorf("write applied migration row: %w", err)
		}
	}

	a.log.logger.InfoContext(
		ctx,
		"database migration finished",
		slog.Bool("migrated", result.Migrated),
		slog.Bool("metadata_repaired", result.MetadataRepaired),
		slog.String("backup_path", result.BackupPath),
		slog.String("from_schema_version", result.FromSchemaVersion),
		slog.String("to_schema_version", result.ToSchemaVersion),
		slog.Int("applied_migrations", len(result.AppliedMigrations)),
	)

	return nil
}

func blankIfEmpty(value string) string {
	if value == "" {
		return "-"
	}

	return value
}
