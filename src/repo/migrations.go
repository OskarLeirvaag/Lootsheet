package repo

import (
	"context"
	"fmt"
	"slices"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
)

func GetDatabaseStatusWithAssets(ctx context.Context, databasePath string, assets config.InitAssets) (DatabaseStatus, error) {
	state, err := inspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return DatabaseStatus{}, err
	}

	status := DatabaseStatus{
		Exists:              state.Exists,
		Initialized:         state.SchemaVersion != "",
		State:               DatabaseStateUninitialized,
		UserTableCount:      state.UserTableCount,
		SchemaVersion:       state.SchemaVersion,
		TargetSchemaVersion: assets.SchemaVersion,
		AppliedMigrations:   state.AppliedMigrations,
	}

	if state.SchemaVersion == "" {
		return status, nil
	}

	_, pendingMigrations, err := splitMigrationsAtVersion(assets.Migrations, state.SchemaVersion)
	if err != nil {
		status.State = DatabaseStateUnknown
		return status, nil
	}

	status.PendingMigrations = toPendingMigrations(pendingMigrations)
	if len(status.PendingMigrations) > 0 {
		status.State = DatabaseStateUpgradeable
		return status, nil
	}

	status.State = DatabaseStateCurrent
	return status, nil
}

func MigrateSQLiteDatabase(ctx context.Context, databasePath string, assets config.InitAssets) (MigrationResult, error) {
	state, err := inspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return MigrationResult{}, err
	}

	switch {
	case state.SchemaVersion == "" && state.UserTableCount == 0:
		return MigrationResult{}, fmt.Errorf("database %q is not initialized; run `lootsheet init`", databasePath)
	case state.SchemaVersion == "":
		return MigrationResult{}, fmt.Errorf("database %q has tables but is missing LootSheet migration metadata", databasePath)
	}

	appliedMigrations, pendingMigrations, err := splitMigrationsAtVersion(assets.Migrations, state.SchemaVersion)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("database schema version %q is not recognized by this build", state.SchemaVersion)
	}

	result := MigrationResult{
		FromSchemaVersion: state.SchemaVersion,
		ToSchemaVersion:   state.SchemaVersion,
		AppliedMigrations: toPendingMigrations(pendingMigrations),
	}

	if len(pendingMigrations) > 0 {
		result.Migrated = true
		result.ToSchemaVersion = pendingMigrations[len(pendingMigrations)-1].Version
	}

	if state.UsesLegacyMetadata {
		result.MetadataRepaired = true
	}

	if !result.Migrated && !result.MetadataRepaired {
		return result, nil
	}

	db, err := openDB(databasePath)
	if err != nil {
		return MigrationResult{}, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	if state.UsesLegacyMetadata {
		if _, err := tx.ExecContext(ctx, schemaMigrationsTableSQL); err != nil {
			return MigrationResult{}, fmt.Errorf("create schema_migrations table: %w", err)
		}

		for _, migration := range appliedMigrations {
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
				migration.Version, migration.Name,
			); err != nil {
				return MigrationResult{}, fmt.Errorf("backfill migration record %s: %w", migration.Name, err)
			}
		}
	}

	for _, migration := range pendingMigrations {
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			return MigrationResult{}, fmt.Errorf("execute migration %s: %w", migration.Name, err)
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version, migration.Name,
		); err != nil {
			return MigrationResult{}, fmt.Errorf("record migration %s: %w", migration.Name, err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP",
		"schema_version", result.ToSchemaVersion,
	); err != nil {
		return MigrationResult{}, fmt.Errorf("update schema version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return MigrationResult{}, fmt.Errorf("commit migration transaction: %w", err)
	}

	return result, nil
}

func splitMigrationsAtVersion(
	migrations []config.InitMigration,
	version string,
) ([]config.InitMigration, []config.InitMigration, error) {
	for index, migration := range migrations {
		if migration.Version != version {
			continue
		}

		return slices.Clone(migrations[:index+1]), slices.Clone(migrations[index+1:]), nil
	}

	return nil, nil, fmt.Errorf("schema version %q not found", version)
}

func toPendingMigrations(migrations []config.InitMigration) []PendingMigration {
	pending := make([]PendingMigration, 0, len(migrations))
	for _, migration := range migrations {
		pending = append(pending, PendingMigration{
			Version: migration.Version,
			Name:    migration.Name,
		})
	}

	return pending
}

const schemaMigrationsTableSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
