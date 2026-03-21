package ledger

import (
	"context"
	"database/sql"
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

	if err := seedInitAccounts(ctx, tx, assets.Accounts); err != nil {
		return InitResult{}, err
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

// seedInitAccounts inserts seeded accounts during database initialization.
// Campaign-aware if the campaigns table exists (migration 009+), otherwise
// plain INSERT for legacy schema testing.
func seedInitAccounts(ctx context.Context, tx *sql.Tx, accounts []config.SeedAccount) error {
	var hasCampaigns int
	_ = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='campaigns'").Scan(&hasCampaigns)

	if hasCampaigns > 0 {
		const campaignID = "default"
		for _, account := range accounts {
			active := 0
			if account.Active {
				active = 1
			}
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO accounts (id, campaign_id, code, name, type, active) VALUES (?, ?, ?, ?, ?, ?)",
				account.ID, campaignID, account.Code, account.Name, account.Type, active,
			); err != nil {
				return fmt.Errorf("seed account %s: %w", account.Code, err)
			}
		}
	} else {
		for _, account := range accounts {
			active := 0
			if account.Active {
				active = 1
			}
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO accounts (id, code, name, type, active) VALUES (?, ?, ?, ?, ?)",
				account.ID, account.Code, account.Name, account.Type, active,
			); err != nil {
				return fmt.Errorf("seed account %s: %w", account.Code, err)
			}
		}
	}

	return nil
}
