package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
)

// GetDatabaseStatusWithAssets returns the full database status including lifecycle
// state and pending migrations computed by comparing the current schema version
// against the provided init assets.
func GetDatabaseStatusWithAssets(ctx context.Context, databasePath string, assets config.InitAssets) (DatabaseStatus, error) {
	state, err := InspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return DatabaseStatus{}, err
	}

	status := DatabaseStatus{
		Exists:              state.Exists,
		Initialized:         state.SchemaVersion != "",
		State:               state.LifecycleState,
		Detail:              state.Detail,
		UserTableCount:      state.UserTableCount,
		SchemaVersion:       state.SchemaVersion,
		TargetSchemaVersion: assets.SchemaVersion,
		AppliedMigrations:   state.AppliedMigrations,
	}

	switch state.LifecycleState {
	case DatabaseStateDamaged, DatabaseStateForeign:
		return status, nil
	case DatabaseStateUninitialized, DatabaseStateCurrent, DatabaseStateUpgradeable:
	}

	if state.SchemaVersion == "" {
		return status, nil
	}

	_, pendingMigrations, err := splitMigrationsAtVersion(assets.Migrations, state.SchemaVersion)
	if err != nil {
		status.State = DatabaseStateForeign
		status.Detail = fmt.Sprintf("schema version %q is not recognized by this build", state.SchemaVersion)
		return status, nil //nolint:nilerr // unknown schema version is a valid state, not a caller error
	}

	status.PendingMigrations = toPendingMigrations(pendingMigrations)
	if len(status.PendingMigrations) > 0 {
		status.State = DatabaseStateUpgradeable
		return status, nil
	}

	status.State = DatabaseStateCurrent
	return status, nil
}

// MigrateSQLiteDatabase applies any pending schema migrations to an existing
// LootSheet database. It also repairs legacy metadata if the database uses
// the old settings-only format. All changes are applied in a single transaction.
func MigrateSQLiteDatabase(ctx context.Context, databasePath string, backupDir string, assets config.InitAssets) (MigrationResult, error) {
	state, err := InspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return MigrationResult{}, err
	}

	switch {
	case state.LifecycleState == DatabaseStateDamaged:
		return MigrationResult{}, fmt.Errorf("database %q is damaged: %s", databasePath, blankDatabaseDetail(state.Detail))
	case state.LifecycleState == DatabaseStateForeign:
		return MigrationResult{}, fmt.Errorf("database %q is foreign: %s", databasePath, blankDatabaseDetail(state.Detail))
	case state.SchemaVersion == "" && state.UserTableCount == 0:
		return MigrationResult{}, fmt.Errorf("database %q is not initialized; run `lootsheet init`", databasePath)
	case state.SchemaVersion == "":
		return MigrationResult{}, fmt.Errorf("database %q is foreign: database has tables but is missing LootSheet migration metadata", databasePath)
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

	backupPath, err := CreateDatabaseBackup(databasePath, backupDir)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("create database backup: %w", err)
	}
	result.BackupPath = backupPath

	db, err := OpenDB(databasePath)
	if err != nil {
		return MigrationResult{}, err
	}
	defer db.Close()

	// Disable FK checks for table-rebuild migrations (DROP+RENAME requires
	// foreign_keys=OFF which cannot run inside a transaction). The migration
	// SQL itself is still executed inside a transaction for atomicity. After
	// commit we run foreign_key_check and re-enable foreign_keys.
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return MigrationResult{}, fmt.Errorf("disable foreign keys for migration: %w", err)
	}

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

	// Verify FK integrity and re-enable foreign keys after table rebuilds.
	if err := verifyForeignKeys(ctx, db); err != nil {
		return MigrationResult{}, err
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

func verifyForeignKeys(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("foreign key check: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var table, rowid, parent, fkid string
		if scanErr := rows.Scan(&table, &rowid, &parent, &fkid); scanErr == nil {
			return fmt.Errorf("foreign key violation after migration: table=%s rowid=%s parent=%s", table, rowid, parent)
		}
		return fmt.Errorf("foreign key violation detected after migration")
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("foreign key check iteration: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("re-enable foreign keys: %w", err)
	}

	return nil
}

const schemaMigrationsTableSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
