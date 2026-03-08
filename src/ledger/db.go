package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite" // register sqlite driver for database/sql
)

// OpenDBForTest opens a raw database connection for use in tests outside
// the ledger package. It does not set any pragmas beyond the defaults.
func OpenDBForTest(databasePath string) (*sql.DB, error) {
	return sql.Open("sqlite", databasePath)
}

// OpenDB opens a database connection with standard pragmas applied.
func OpenDB(databasePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.ExecContext(context.Background(), pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("set database pragma: %w", err)
		}
	}

	return db, nil
}

type databaseState struct {
	Exists             bool
	LifecycleState     DatabaseLifecycleState
	Detail             string
	UserTableCount     int
	SchemaVersion      string
	AppliedMigrations  []AppliedMigration
	UsesLegacyMetadata bool
}

// InspectSQLiteDatabase examines the database file at the given path and returns
// its current state, including whether it exists, its user table count, schema
// version, and applied migrations. It handles both legacy (settings-table) and
// current (schema_migrations-table) metadata formats.
func InspectSQLiteDatabase(ctx context.Context, databasePath string) (databaseState, error) {
	state := databaseState{LifecycleState: DatabaseStateUninitialized}

	info, err := os.Stat(databasePath)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, fmt.Errorf("inspect database path: %w", err)
	}
	state.Exists = true
	if info.IsDir() {
		state.LifecycleState = DatabaseStateDamaged
		state.Detail = fmt.Sprintf("path %q is a directory, not a SQLite database file", databasePath)
		return state, nil
	}

	db, err := OpenDB(databasePath)
	if err != nil {
		if detail, ok := classifyDamagedDatabaseError(err); ok {
			state.LifecycleState = DatabaseStateDamaged
			state.Detail = detail
			return state, nil
		}
		return state, err
	}
	defer db.Close()

	var userTableCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'",
	).Scan(&userTableCount); err != nil {
		if detail, ok := classifyDamagedDatabaseError(err); ok {
			state.LifecycleState = DatabaseStateDamaged
			state.Detail = detail
			return state, nil
		}
		return state, fmt.Errorf("count user tables: %w", err)
	}
	state.UserTableCount = userTableCount

	quickCheck, err := quickCheckSQLiteDatabase(ctx, db)
	if err != nil {
		if detail, ok := classifyDamagedDatabaseError(err); ok {
			state.LifecycleState = DatabaseStateDamaged
			state.Detail = detail
			return state, nil
		}
		return state, fmt.Errorf("run sqlite quick_check: %w", err)
	}
	if quickCheck != "ok" {
		state.LifecycleState = DatabaseStateDamaged
		state.Detail = fmt.Sprintf("sqlite quick_check reported %q", quickCheck)
		return state, nil
	}

	appliedMigrations, err := LoadAppliedMigrations(ctx, db)
	if err == nil {
		state.AppliedMigrations = appliedMigrations
		if len(appliedMigrations) > 0 {
			state.SchemaVersion = appliedMigrations[len(appliedMigrations)-1].Version
			return state, nil
		}

		if userTableCount == 0 {
			return state, nil
		}

		state.LifecycleState = DatabaseStateForeign
		state.Detail = "database has LootSheet migration tables but no applied schema version"
		return state, nil
	}

	if !isMissingTableError(err, "schema_migrations") {
		if detail, ok := classifyDamagedDatabaseError(err); ok {
			state.LifecycleState = DatabaseStateDamaged
			state.Detail = detail
			return state, nil
		}
		return state, fmt.Errorf("load applied migrations: %w", err)
	}

	var schemaVersion string
	queryErr := db.QueryRowContext(ctx,
		"SELECT value FROM settings WHERE key = 'schema_version'",
	).Scan(&schemaVersion)
	if queryErr != nil {
		if isMissingTableError(queryErr, "settings") {
			if userTableCount == 0 {
				return state, nil
			}

			state.LifecycleState = DatabaseStateForeign
			state.Detail = "database has user tables but is missing LootSheet migration metadata"
			return state, nil
		}
		if detail, ok := classifyDamagedDatabaseError(queryErr); ok {
			state.LifecycleState = DatabaseStateDamaged
			state.Detail = detail
			return state, nil
		}
		return state, fmt.Errorf("query schema version: %w", queryErr)
	}

	state.SchemaVersion = strings.TrimSpace(schemaVersion)
	if state.SchemaVersion == "" {
		if userTableCount == 0 {
			return state, nil
		}

		state.LifecycleState = DatabaseStateForeign
		state.Detail = "database settings are present but schema_version is empty"
		return state, nil
	}

	state.UsesLegacyMetadata = true
	return state, nil
}

// EnsureInitializedDatabase checks that the database at the given path
// has been initialized with the LootSheet schema.
func EnsureInitializedDatabase(ctx context.Context, databasePath string) error {
	state, err := InspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return err
	}

	switch state.LifecycleState {
	case DatabaseStateDamaged:
		return fmt.Errorf("database %q is damaged: %s", databasePath, blankDatabaseDetail(state.Detail))
	case DatabaseStateForeign:
		return fmt.Errorf("database %q is foreign: %s", databasePath, blankDatabaseDetail(state.Detail))
	case DatabaseStateUninitialized, DatabaseStateCurrent, DatabaseStateUpgradeable:
	}

	if state.SchemaVersion == "" {
		return fmt.Errorf("database %q is not initialized; run `lootsheet init`", databasePath)
	}

	return nil
}

// LoadAppliedMigrations queries the schema_migrations table for applied migrations.
func LoadAppliedMigrations(ctx context.Context, db *sql.DB) ([]AppliedMigration, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT version, name, applied_at FROM schema_migrations ORDER BY CAST(version AS INTEGER), name",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := []AppliedMigration{}
	for rows.Next() {
		var m AppliedMigration
		if err := rows.Scan(&m.Version, &m.Name, &m.AppliedAt); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		migrations = append(migrations, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return migrations, nil
}

func quickCheckSQLiteDatabase(ctx context.Context, db *sql.DB) (string, error) {
	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&result); err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

func classifyDamagedDatabaseError(err error) (string, bool) {
	message := strings.ToLower(err.Error())

	switch {
	case strings.Contains(message, "file is not a database"):
		return "file is not a valid SQLite database", true
	case strings.Contains(message, "database disk image is malformed"):
		return "database disk image is malformed", true
	case strings.Contains(message, "database corrupt"):
		return "database is corrupt", true
	case strings.Contains(message, "malformed"):
		return "database appears malformed", true
	}

	return "", false
}

func isMissingTableError(err error, table string) bool {
	return strings.Contains(strings.ToLower(err.Error()), "no such table: "+strings.ToLower(table))
}

func blankDatabaseDetail(detail string) string {
	if strings.TrimSpace(detail) == "" {
		return "no further detail available"
	}

	return detail
}
