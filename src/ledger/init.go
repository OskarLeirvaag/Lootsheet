package ledger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
)

// EnsureSQLiteInitialized creates and initializes a new LootSheet database if one
// does not already exist at the given path. It applies all schema migrations and
// seeds the default accounts within a single transaction. If the database is already
// initialized, it returns a zero-value InitResult with Initialized=false.
func EnsureSQLiteInitialized(ctx context.Context, databasePath string, assets config.InitAssets) (InitResult, error) {
	state, err := InspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return InitResult{}, err
	}

	switch {
	case state.LifecycleState == DatabaseStateDamaged:
		return InitResult{}, fmt.Errorf("database %q is damaged: %s", databasePath, blankDatabaseDetail(state.Detail))
	case state.LifecycleState == DatabaseStateForeign:
		return InitResult{}, fmt.Errorf("database %q is foreign: %s", databasePath, blankDatabaseDetail(state.Detail))
	case state.SchemaVersion != "":
		return InitResult{}, nil
	case state.UserTableCount > 0:
		return InitResult{}, fmt.Errorf("database %q is foreign: database already has tables but is missing LootSheet init metadata", databasePath)
	}

	if err := os.MkdirAll(filepath.Dir(databasePath), dirPerm); err != nil {
		return InitResult{}, fmt.Errorf("create database directory: %w", err)
	}

	db, err := OpenDB(databasePath)
	if err != nil {
		return InitResult{}, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return InitResult{}, fmt.Errorf("begin init transaction: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range assets.Migrations {
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			return InitResult{}, fmt.Errorf("execute init migration %s: %w", migration.Name, err)
		}
	}

	for _, migration := range assets.Migrations {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version, migration.Name,
		); err != nil {
			return InitResult{}, fmt.Errorf("record init migration %s: %w", migration.Name, err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO settings (key, value) VALUES (?, ?)",
		"schema_version", assets.SchemaVersion,
	); err != nil {
		return InitResult{}, fmt.Errorf("record schema version: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO settings (key, value) VALUES (?, CURRENT_TIMESTAMP)",
		"initialized_at",
	); err != nil {
		return InitResult{}, fmt.Errorf("record initialization timestamp: %w", err)
	}

	for _, account := range assets.Accounts {
		active := 0
		if account.Active {
			active = 1
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO accounts (id, code, name, type, active) VALUES (?, ?, ?, ?, ?)",
			account.ID, account.Code, account.Name, account.Type, active,
		); err != nil {
			return InitResult{}, fmt.Errorf("seed account %s: %w", account.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return InitResult{}, fmt.Errorf("commit init transaction: %w", err)
	}

	return InitResult{
		Initialized: true,
		SeededCounts: SeededCounts{
			Accounts: len(assets.Accounts),
		},
	}, nil
}

// GetDatabaseStatus returns basic database status without comparing against
// available migrations. For a status that includes pending migration info,
// use GetDatabaseStatusWithAssets instead.
func GetDatabaseStatus(ctx context.Context, databasePath string) (DatabaseStatus, error) {
	state, err := InspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return DatabaseStatus{}, err
	}

	stateLabel := state.LifecycleState
	if state.SchemaVersion != "" && stateLabel == DatabaseStateUninitialized {
		stateLabel = DatabaseStateCurrent
	}

	return DatabaseStatus{
		Exists:            state.Exists,
		Initialized:       state.SchemaVersion != "",
		State:             stateLabel,
		Detail:            state.Detail,
		UserTableCount:    state.UserTableCount,
		SchemaVersion:     state.SchemaVersion,
		AppliedMigrations: state.AppliedMigrations,
	}, nil
}
