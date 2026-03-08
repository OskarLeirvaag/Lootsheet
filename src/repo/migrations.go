package repo

import (
	"context"
	"fmt"
	"slices"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
)

func GetDatabaseStatusWithAssets(ctx context.Context, databasePath string, assets config.InitAssets) (DatabaseStatus, error) {
	if err := ensureSQLiteAvailable(); err != nil {
		return DatabaseStatus{}, err
	}

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
	if err := ensureSQLiteAvailable(); err != nil {
		return MigrationResult{}, err
	}

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

	if err := runSQLiteScript(
		ctx,
		databasePath,
		buildMigrationScript(state, appliedMigrations, pendingMigrations, result.ToSchemaVersion),
	); err != nil {
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

func buildMigrationScript(
	state databaseState,
	appliedMigrations []config.InitMigration,
	pendingMigrations []config.InitMigration,
	targetSchemaVersion string,
) string {
	statements := make([]string, 0, len(appliedMigrations)+len(pendingMigrations)*2+1)

	if state.UsesLegacyMetadata {
		statements = append(statements, buildSchemaMigrationsTableStatement())

		for _, migration := range appliedMigrations {
			statements = append(statements, buildInsertStatement(
				"schema_migrations",
				[]string{"version", "name"},
				[]string{sqlString(migration.Version), sqlString(migration.Name)},
			))
		}
	}

	for _, migration := range pendingMigrations {
		statements = append(statements, migration.SQL)
		statements = append(statements, buildInsertStatement(
			"schema_migrations",
			[]string{"version", "name"},
			[]string{sqlString(migration.Version), sqlString(migration.Name)},
		))
	}

	statements = append(statements, buildUpsertStatement(
		"settings",
		[]string{"key", "value"},
		[]string{sqlString("schema_version"), sqlString(targetSchemaVersion)},
		[]string{"key"},
		[]string{"value = excluded.value", "updated_at = CURRENT_TIMESTAMP"},
	))

	return buildTransactionScript(statements...)
}

func buildSchemaMigrationsTableStatement() string {
	return `CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
}
