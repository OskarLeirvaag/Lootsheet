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
	info, err := os.Stat(databasePath)
	if errors.Is(err, os.ErrNotExist) {
		return databaseState{}, nil
	}
	if err != nil {
		return databaseState{}, fmt.Errorf("inspect database path: %w", err)
	}
	if info.IsDir() {
		return databaseState{}, fmt.Errorf("database path %q is a directory", databasePath)
	}

	db, err := OpenDB(databasePath)
	if err != nil {
		return databaseState{}, err
	}
	defer db.Close()

	var userTableCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'",
	).Scan(&userTableCount); err != nil {
		return databaseState{}, fmt.Errorf("count user tables: %w", err)
	}

	appliedMigrations, err := LoadAppliedMigrations(ctx, db)
	if err == nil {
		schemaVersion := ""
		if len(appliedMigrations) > 0 {
			schemaVersion = appliedMigrations[len(appliedMigrations)-1].Version
		}

		return databaseState{
			Exists:            true,
			UserTableCount:    userTableCount,
			SchemaVersion:     schemaVersion,
			AppliedMigrations: appliedMigrations,
		}, nil
	}

	if !strings.Contains(err.Error(), "no such table: schema_migrations") {
		return databaseState{}, err
	}

	var schemaVersion string
	queryErr := db.QueryRowContext(ctx,
		"SELECT value FROM settings WHERE key = 'schema_version'",
	).Scan(&schemaVersion)
	if queryErr != nil {
		if strings.Contains(queryErr.Error(), "no such table: settings") {
			return databaseState{Exists: true, UserTableCount: userTableCount}, nil
		}
		return databaseState{}, queryErr
	}

	return databaseState{
		Exists:             true,
		UserTableCount:     userTableCount,
		SchemaVersion:      schemaVersion,
		UsesLegacyMetadata: true,
	}, nil
}

// EnsureInitializedDatabase checks that the database at the given path
// has been initialized with the LootSheet schema.
func EnsureInitializedDatabase(ctx context.Context, databasePath string) error {
	state, err := InspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return err
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
