package campaign

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// Create inserts a new campaign and seeds the default chart of accounts.
func Create(ctx context.Context, databasePath string, name string, accounts []config.SeedAccount) (Record, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Record{}, errors.New("campaign name is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (Record, error) {
		id := uuid.NewString()

		if _, err := db.ExecContext(ctx,
			"INSERT INTO campaigns (id, name) VALUES (?, ?)",
			id, name,
		); err != nil {
			return Record{}, fmt.Errorf("insert campaign: %w", err)
		}

		for _, acct := range accounts {
			active := 0
			if acct.Active {
				active = 1
			}

			if _, err := db.ExecContext(ctx,
				"INSERT INTO accounts (id, campaign_id, code, name, type, active) VALUES (?, ?, ?, ?, ?, ?)",
				uuid.NewString(), id, acct.Code, acct.Name, acct.Type, active,
			); err != nil {
				return Record{}, fmt.Errorf("seed account %s for campaign: %w", acct.Code, err)
			}
		}

		var record Record
		if err := db.QueryRowContext(ctx,
			"SELECT id, name, created_at, updated_at FROM campaigns WHERE id = ?", id,
		).Scan(&record.ID, &record.Name, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return Record{}, fmt.Errorf("query created campaign: %w", err)
		}

		return record, nil
	})
}

// List returns all campaigns ordered by name.
func List(ctx context.Context, databasePath string) ([]Record, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Record, error) {
		rows, err := db.QueryContext(ctx,
			"SELECT id, name, created_at, updated_at FROM campaigns ORDER BY name, id")
		if err != nil {
			return nil, fmt.Errorf("query campaigns: %w", err)
		}
		defer rows.Close()

		var campaigns []Record
		for rows.Next() {
			var r Record
			if err := rows.Scan(&r.ID, &r.Name, &r.CreatedAt, &r.UpdatedAt); err != nil {
				return nil, fmt.Errorf("scan campaign row: %w", err)
			}
			campaigns = append(campaigns, r)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate campaign rows: %w", err)
		}

		return campaigns, nil
	})
}

// Rename updates the name of an existing campaign.
func Rename(ctx context.Context, databasePath string, id string, name string) (Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Record{}, errors.New("campaign ID is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Record{}, errors.New("campaign name is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (Record, error) {
		result, err := db.ExecContext(ctx,
			"UPDATE campaigns SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			name, id,
		)
		if err != nil {
			return Record{}, fmt.Errorf("rename campaign: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return Record{}, fmt.Errorf("check rename result: %w", err)
		}
		if affected == 0 {
			return Record{}, fmt.Errorf("campaign %q does not exist", id)
		}

		var record Record
		if err := db.QueryRowContext(ctx,
			"SELECT id, name, created_at, updated_at FROM campaigns WHERE id = ?", id,
		).Scan(&record.ID, &record.Name, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return Record{}, fmt.Errorf("query renamed campaign: %w", err)
		}

		return record, nil
	})
}

// Delete removes a campaign, but only if it has no data beyond the seeded accounts.
//
//nolint:revive // cognitive-complexity: sequential safety checks before cascading delete reads best inline
func Delete(ctx context.Context, databasePath string, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("campaign ID is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		// Check for journal entries.
		var count int
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM journal_entries WHERE campaign_id = ?", id,
		).Scan(&count); err != nil {
			return fmt.Errorf("check campaign journal entries: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("cannot delete campaign: %d journal entries exist", count)
		}

		// Check for quests.
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM quests WHERE campaign_id = ?", id,
		).Scan(&count); err != nil {
			return fmt.Errorf("check campaign quests: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("cannot delete campaign: %d quests exist", count)
		}

		// Check for loot items.
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM loot_items WHERE campaign_id = ?", id,
		).Scan(&count); err != nil {
			return fmt.Errorf("check campaign loot items: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("cannot delete campaign: %d loot items exist", count)
		}

		// Delete campaign-scoped data.
		if _, err := db.ExecContext(ctx, "DELETE FROM entity_references WHERE campaign_id = ?", id); err != nil {
			return fmt.Errorf("delete campaign entity references: %w", err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM notes WHERE campaign_id = ?", id); err != nil {
			return fmt.Errorf("delete campaign notes: %w", err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM codex_entries WHERE campaign_id = ?", id); err != nil {
			return fmt.Errorf("delete campaign codex entries: %w", err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM accounts WHERE campaign_id = ?", id); err != nil {
			return fmt.Errorf("delete campaign accounts: %w", err)
		}

		result, err := db.ExecContext(ctx, "DELETE FROM campaigns WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("delete campaign: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check delete result: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("campaign %q does not exist", id)
		}

		return nil
	})
}

// GetActive reads the active_campaign_id from settings and returns the campaign record.
func GetActive(ctx context.Context, databasePath string) (Record, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (Record, error) {
		var campaignID string
		if err := db.QueryRowContext(ctx,
			"SELECT value FROM settings WHERE key = 'active_campaign_id'",
		).Scan(&campaignID); err != nil {
			return Record{}, fmt.Errorf("query active campaign setting: %w", err)
		}

		var record Record
		if err := db.QueryRowContext(ctx,
			"SELECT id, name, created_at, updated_at FROM campaigns WHERE id = ?", campaignID,
		).Scan(&record.ID, &record.Name, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return Record{}, fmt.Errorf("query active campaign: %w", err)
		}

		return record, nil
	})
}

// SetActive updates the active_campaign_id in settings.
func SetActive(ctx context.Context, databasePath string, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("campaign ID is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		// Verify campaign exists.
		var exists int
		if err := db.QueryRowContext(ctx,
			"SELECT 1 FROM campaigns WHERE id = ?", id,
		).Scan(&exists); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("campaign %q does not exist", id)
			}
			return fmt.Errorf("check campaign exists: %w", err)
		}

		if _, err := db.ExecContext(ctx,
			"INSERT INTO settings (key, value) VALUES ('active_campaign_id', ?) ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP",
			id,
		); err != nil {
			return fmt.Errorf("set active campaign: %w", err)
		}

		return nil
	})
}
