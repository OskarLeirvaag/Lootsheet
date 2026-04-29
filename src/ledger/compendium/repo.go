package compendium

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// ListMonstersForCampaign returns monsters from sources enabled for the given
// campaign, optionally filtered by a full-text query. Used by the TUI data
// loader so each campaign only sees its own source books.
func ListMonstersForCampaign(ctx context.Context, databasePath string, campaignID string, query string) ([]Monster, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Monster, error) {
		var rows *sql.Rows
		var err error
		if query != "" {
			rows, err = db.QueryContext(ctx, `
				SELECT DISTINCT m.id, m.ddb_id, m.name, m.cr, m.type, m.size, m.hp, m.ac, m.source_name, '', m.synced_at
				FROM compendium_monsters m
				JOIN compendium_monsters_fts f ON f.rowid = m.id
				JOIN compendium_sources cs ON cs.name = m.source_name
				JOIN compendium_campaign_sources ccs ON ccs.source_id = cs.id AND ccs.campaign_id = ? AND ccs.enabled = 1
				WHERE compendium_monsters_fts MATCH ?
				ORDER BY m.name`, campaignID, ftsQuery(query))
		} else {
			rows, err = db.QueryContext(ctx, `
				SELECT DISTINCT m.id, m.ddb_id, m.name, m.cr, m.type, m.size, m.hp, m.ac, m.source_name, '', m.synced_at
				FROM compendium_monsters m
				JOIN compendium_sources cs ON cs.name = m.source_name
				JOIN compendium_campaign_sources ccs ON ccs.source_id = cs.id AND ccs.campaign_id = ? AND ccs.enabled = 1
				ORDER BY m.name`, campaignID)
		}
		if err != nil {
			return nil, fmt.Errorf("list monsters for campaign: %w", err)
		}
		defer rows.Close()
		var result []Monster
		for rows.Next() {
			var m Monster
			if err := rows.Scan(&m.ID, &m.DdbID, &m.Name, &m.CR, &m.Type, &m.Size, &m.HP, &m.AC, &m.SourceName, &m.DetailJSON, &m.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan monster: %w", err)
			}
			result = append(result, m)
		}
		return result, rows.Err()
	})
}

// ListSpellsForCampaign returns spells from sources enabled for the given campaign.
func ListSpellsForCampaign(ctx context.Context, databasePath string, campaignID string, query string) ([]Spell, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Spell, error) {
		var rows *sql.Rows
		var err error
		if query != "" {
			rows, err = db.QueryContext(ctx, `
				SELECT DISTINCT s.id, s.ddb_id, s.name, s.level, s.school, s.casting_time, s.range, s.components, s.duration, s.classes, s.source_name, '', s.synced_at
				FROM compendium_spells s
				JOIN compendium_spells_fts f ON f.rowid = s.id
				JOIN compendium_sources cs ON cs.name = s.source_name
				JOIN compendium_campaign_sources ccs ON ccs.source_id = cs.id AND ccs.campaign_id = ? AND ccs.enabled = 1
				WHERE compendium_spells_fts MATCH ?
				ORDER BY s.level, s.name`, campaignID, ftsQuery(query))
		} else {
			rows, err = db.QueryContext(ctx, `
				SELECT DISTINCT s.id, s.ddb_id, s.name, s.level, s.school, s.casting_time, s.range, s.components, s.duration, s.classes, s.source_name, '', s.synced_at
				FROM compendium_spells s
				JOIN compendium_sources cs ON cs.name = s.source_name
				JOIN compendium_campaign_sources ccs ON ccs.source_id = cs.id AND ccs.campaign_id = ? AND ccs.enabled = 1
				ORDER BY s.level, s.name`, campaignID)
		}
		if err != nil {
			return nil, fmt.Errorf("list spells for campaign: %w", err)
		}
		defer rows.Close()
		var result []Spell
		for rows.Next() {
			var s Spell
			if err := rows.Scan(&s.ID, &s.DdbID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration, &s.Classes, &s.SourceName, &s.DetailJSON, &s.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan spell: %w", err)
			}
			result = append(result, s)
		}
		return result, rows.Err()
	})
}

// ListItemsForCampaign returns items from sources enabled for the given campaign.
func ListItemsForCampaign(ctx context.Context, databasePath string, campaignID string, query string) ([]Item, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Item, error) {
		var rows *sql.Rows
		var err error
		if query != "" {
			rows, err = db.QueryContext(ctx, `
				SELECT DISTINCT i.id, i.ddb_id, i.name, i.type, i.rarity, i.attunement, i.source_name, '', i.synced_at
				FROM compendium_items i
				JOIN compendium_items_fts f ON f.rowid = i.id
				JOIN compendium_sources cs ON cs.name = i.source_name
				JOIN compendium_campaign_sources ccs ON ccs.source_id = cs.id AND ccs.campaign_id = ? AND ccs.enabled = 1
				WHERE compendium_items_fts MATCH ?
				ORDER BY i.name`, campaignID, ftsQuery(query))
		} else {
			rows, err = db.QueryContext(ctx, `
				SELECT DISTINCT i.id, i.ddb_id, i.name, i.type, i.rarity, i.attunement, i.source_name, '', i.synced_at
				FROM compendium_items i
				JOIN compendium_sources cs ON cs.name = i.source_name
				JOIN compendium_campaign_sources ccs ON ccs.source_id = cs.id AND ccs.campaign_id = ? AND ccs.enabled = 1
				ORDER BY i.name`, campaignID)
		}
		if err != nil {
			return nil, fmt.Errorf("list items for campaign: %w", err)
		}
		defer rows.Close()
		var result []Item
		for rows.Next() {
			var i Item
			if err := rows.Scan(&i.ID, &i.DdbID, &i.Name, &i.Type, &i.Rarity, &i.Attunement, &i.SourceName, &i.DetailJSON, &i.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan item: %w", err)
			}
			result = append(result, i)
		}
		return result, rows.Err()
	})
}

// ListMonsters returns all monsters, optionally filtered by a search query (FTS).
func ListMonsters(ctx context.Context, databasePath string, query string) ([]Monster, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Monster, error) {
		q := `SELECT id, ddb_id, name, cr, type, size, hp, ac, source_name, '', synced_at
		      FROM compendium_monsters ORDER BY name`
		args := []any{}
		if query != "" {
			q = `SELECT m.id, m.ddb_id, m.name, m.cr, m.type, m.size, m.hp, m.ac, m.source_name, '', m.synced_at
			     FROM compendium_monsters m
			     JOIN compendium_monsters_fts f ON f.rowid = m.id
			     WHERE compendium_monsters_fts MATCH ?
			     ORDER BY m.name`
			args = append(args, ftsQuery(query))
		}

		rows, err := db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, fmt.Errorf("list monsters: %w", err)
		}
		defer rows.Close()

		var result []Monster
		for rows.Next() {
			var m Monster
			if err := rows.Scan(&m.ID, &m.DdbID, &m.Name, &m.CR, &m.Type, &m.Size, &m.HP, &m.AC, &m.SourceName, &m.DetailJSON, &m.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan monster: %w", err)
			}
			result = append(result, m)
		}
		return result, rows.Err()
	})
}

// ListSpells returns all spells, optionally filtered by a search query (FTS).
func ListSpells(ctx context.Context, databasePath string, query string) ([]Spell, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Spell, error) {
		q := `SELECT id, ddb_id, name, level, school, casting_time, range, components, duration, classes, source_name, '', synced_at
		      FROM compendium_spells ORDER BY level, name`
		args := []any{}
		if query != "" {
			q = `SELECT s.id, s.ddb_id, s.name, s.level, s.school, s.casting_time, s.range, s.components, s.duration, s.classes, s.source_name, '', s.synced_at
			     FROM compendium_spells s
			     JOIN compendium_spells_fts f ON f.rowid = s.id
			     WHERE compendium_spells_fts MATCH ?
			     ORDER BY s.level, s.name`
			args = append(args, ftsQuery(query))
		}

		rows, err := db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, fmt.Errorf("list spells: %w", err)
		}
		defer rows.Close()

		var result []Spell
		for rows.Next() {
			var s Spell
			if err := rows.Scan(&s.ID, &s.DdbID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range, &s.Components, &s.Duration, &s.Classes, &s.SourceName, &s.DetailJSON, &s.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan spell: %w", err)
			}
			result = append(result, s)
		}
		return result, rows.Err()
	})
}

// ListItems returns all items, optionally filtered by a search query (FTS).
func ListItems(ctx context.Context, databasePath string, query string) ([]Item, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Item, error) {
		q := `SELECT id, ddb_id, name, type, rarity, attunement, source_name, '', synced_at
		      FROM compendium_items ORDER BY name`
		args := []any{}
		if query != "" {
			q = `SELECT i.id, i.ddb_id, i.name, i.type, i.rarity, i.attunement, i.source_name, '', i.synced_at
			     FROM compendium_items i
			     JOIN compendium_items_fts f ON f.rowid = i.id
			     WHERE compendium_items_fts MATCH ?
			     ORDER BY i.name`
			args = append(args, ftsQuery(query))
		}

		rows, err := db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, fmt.Errorf("list items: %w", err)
		}
		defer rows.Close()

		var result []Item
		for rows.Next() {
			var i Item
			if err := rows.Scan(&i.ID, &i.DdbID, &i.Name, &i.Type, &i.Rarity, &i.Attunement, &i.SourceName, &i.DetailJSON, &i.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan item: %w", err)
			}
			result = append(result, i)
		}
		return result, rows.Err()
	})
}

// ListRules returns all rules, optionally filtered by a search query (FTS).
func ListRules(ctx context.Context, databasePath string, query string) ([]Rule, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Rule, error) {
		q := `SELECT id, ddb_id, name, category, description, synced_at
		      FROM compendium_rules ORDER BY category, name`
		args := []any{}
		if query != "" {
			q = `SELECT r.id, r.ddb_id, r.name, r.category, r.description, r.synced_at
			     FROM compendium_rules r
			     JOIN compendium_rules_fts f ON f.rowid = r.id
			     WHERE compendium_rules_fts MATCH ?
			     ORDER BY r.category, r.name`
			args = append(args, ftsQuery(query))
		}

		rows, err := db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, fmt.Errorf("list rules: %w", err)
		}
		defer rows.Close()

		var result []Rule
		for rows.Next() {
			var r Rule
			if err := rows.Scan(&r.ID, &r.DdbID, &r.Name, &r.Category, &r.Description, &r.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan rule: %w", err)
			}
			result = append(result, r)
		}
		return result, rows.Err()
	})
}

// ListConditions returns all conditions, optionally filtered by a search query (FTS).
func ListConditions(ctx context.Context, databasePath string, query string) ([]Condition, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Condition, error) {
		q := `SELECT id, ddb_id, name, description, synced_at
		      FROM compendium_conditions ORDER BY name`
		args := []any{}
		if query != "" {
			q = `SELECT c.id, c.ddb_id, c.name, c.description, c.synced_at
			     FROM compendium_conditions c
			     JOIN compendium_conditions_fts f ON f.rowid = c.id
			     WHERE compendium_conditions_fts MATCH ?
			     ORDER BY c.name`
			args = append(args, ftsQuery(query))
		}

		rows, err := db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, fmt.Errorf("list conditions: %w", err)
		}
		defer rows.Close()

		var result []Condition
		for rows.Next() {
			var c Condition
			if err := rows.Scan(&c.ID, &c.DdbID, &c.Name, &c.Description, &c.SyncedAt); err != nil {
				return nil, fmt.Errorf("scan condition: %w", err)
			}
			result = append(result, c)
		}
		return result, rows.Err()
	})
}

// ListSources returns all source books for a campaign, joining the global
// catalogue with the campaign's per-source selection state.
func ListSources(ctx context.Context, databasePath string, campaignID string) ([]Source, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Source, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT cs.id, cs.name, cs.owned, cs.shared, cs.is_released, cs.category_id,
			       COALESCE(ccs.enabled, 0),
			       COALESCE(ccs.has_spells, 1),
			       COALESCE(ccs.has_items, 1)
			FROM compendium_sources cs
			LEFT JOIN compendium_campaign_sources ccs ON ccs.source_id = cs.id AND ccs.campaign_id = ?
			ORDER BY cs.name`, campaignID)
		if err != nil {
			return nil, fmt.Errorf("list sources: %w", err)
		}
		defer rows.Close()

		var result []Source
		for rows.Next() {
			var s Source
			if err := rows.Scan(&s.ID, &s.Name, &s.Owned, &s.Shared, &s.IsReleased, &s.CategoryID,
				&s.Enabled, &s.HasSpells, &s.HasItems); err != nil {
				return nil, fmt.Errorf("scan source: %w", err)
			}
			result = append(result, s)
		}
		return result, rows.Err()
	})
}

// EnabledSourceIDs returns the IDs of enabled source books for the given campaign.
func EnabledSourceIDs(ctx context.Context, databasePath string, campaignID string) ([]int, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]int, error) {
		rows, err := db.QueryContext(ctx,
			`SELECT source_id FROM compendium_campaign_sources WHERE campaign_id = ? AND enabled = 1`,
			campaignID)
		if err != nil {
			return nil, fmt.Errorf("enabled sources: %w", err)
		}
		defer rows.Close()

		var ids []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return nil, fmt.Errorf("scan source id: %w", err)
			}
			ids = append(ids, id)
		}
		return ids, rows.Err()
	})
}

// ToggleSource flips the enabled state of a source book for the given campaign.
// Upserts the row so campaigns that never had an explicit selection work correctly.
func ToggleSource(ctx context.Context, databasePath string, campaignID string, sourceID int) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO compendium_campaign_sources (campaign_id, source_id, enabled)
			VALUES (?, ?, 1)
			ON CONFLICT(campaign_id, source_id) DO UPDATE SET enabled = 1 - enabled`,
			campaignID, sourceID)
		return err
	})
}

// UpsertSources inserts or updates source books (from DDB config). Only global
// catalogue fields (name, is_released, category_id) are written; per-campaign
// selection lives in compendium_campaign_sources.
func UpsertSources(ctx context.Context, databasePath string, sources []Source) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO compendium_sources (id, name, is_released, category_id)
			 VALUES (?, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
			     name = excluded.name,
			     is_released = excluded.is_released,
			     category_id = excluded.category_id`)
		if err != nil {
			return fmt.Errorf("prepare upsert sources: %w", err)
		}
		defer stmt.Close()

		for _, s := range sources {
			released := 0
			if s.IsReleased {
				released = 1
			}
			if _, err := stmt.ExecContext(ctx, s.ID, s.Name, released, s.CategoryID); err != nil {
				return fmt.Errorf("upsert source %d: %w", s.ID, err)
			}
		}
		return tx.Commit()
	})
}

// SetSourceOwnership marks the listed source IDs as owned (1) and every other
// known source as locked (2). Sources not yet probed should be filtered before
// calling this — any row not in `ownedIDs` will be marked locked.
//
// `shared` is left untouched here; it's populated lazily during Phase B sync
// when content actually comes back for a non-owned source.
func SetSourceOwnership(ctx context.Context, databasePath string, ownedIDs []int) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck

		// First mark everything as locked, then re-mark owned IDs as owned.
		if _, err := tx.ExecContext(ctx, `UPDATE compendium_sources SET owned = ?`, OwnershipLocked); err != nil {
			return fmt.Errorf("set sources locked: %w", err)
		}
		for _, id := range ownedIDs {
			if _, err := tx.ExecContext(ctx, `UPDATE compendium_sources SET owned = ? WHERE id = ?`, OwnershipOwned, id); err != nil {
				return fmt.Errorf("set source %d owned: %w", id, err)
			}
		}
		return tx.Commit()
	})
}

// MarkSourceShared flips `shared = 1` for the given IDs that aren't already
// marked owned. Used in Phase B when content for a non-owned source actually
// returns from DDB — implying the active DDB campaign grants access.
func MarkSourceShared(ctx context.Context, databasePath string, sourceIDs []int) error {
	if len(sourceIDs) == 0 {
		return nil
	}
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck

		for _, id := range sourceIDs {
			if _, err := tx.ExecContext(ctx,
				`UPDATE compendium_sources SET shared = 1 WHERE id = ? AND owned != ?`,
				id, OwnershipOwned); err != nil {
				return fmt.Errorf("mark source %d shared: %w", id, err)
			}
		}
		return tx.Commit()
	})
}

// SetSourceHasSpells records observed spell presence for a source in the given
// campaign. Used after Phase B to update the skip-guard for subsequent syncs.
func SetSourceHasSpells(ctx context.Context, databasePath string, campaignID string, sourceID int, hasSpells bool) error {
	return setCampaignSourceFlag(ctx, databasePath, campaignID, sourceID, "has_spells", hasSpells)
}

// SetSourceHasItems records observed item presence for a source in the given campaign.
func SetSourceHasItems(ctx context.Context, databasePath string, campaignID string, sourceID int, hasItems bool) error {
	return setCampaignSourceFlag(ctx, databasePath, campaignID, sourceID, "has_items", hasItems)
}

func setCampaignSourceFlag(ctx context.Context, databasePath string, campaignID string, sourceID int, column string, value bool) error {
	if column != "has_spells" && column != "has_items" {
		return fmt.Errorf("invalid content column: %q", column)
	}
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		v := 0
		if value {
			v = 1
		}
		// Upsert: ensure the row exists, then set the flag.
		if column == "has_spells" {
			_, err := db.ExecContext(ctx, `
				INSERT INTO compendium_campaign_sources (campaign_id, source_id, has_spells)
				VALUES (?, ?, ?)
				ON CONFLICT(campaign_id, source_id) DO UPDATE SET has_spells = excluded.has_spells`,
				campaignID, sourceID, v)
			return err
		}
		_, err := db.ExecContext(ctx, `
			INSERT INTO compendium_campaign_sources (campaign_id, source_id, has_items)
			VALUES (?, ?, ?)
			ON CONFLICT(campaign_id, source_id) DO UPDATE SET has_items = excluded.has_items`,
			campaignID, sourceID, v)
		return err
	})
}

// GetLastSyncedAt returns the last successful Phase B sync timestamp for a
// campaign, or the zero time if it has never synced.
func GetLastSyncedAt(ctx context.Context, databasePath string, campaignID string) (time.Time, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (time.Time, error) {
		var raw sql.NullString
		err := db.QueryRowContext(ctx,
			`SELECT last_synced_at FROM compendium_campaign_sync WHERE campaign_id = ?`, campaignID).Scan(&raw)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return time.Time{}, nil
			}
			return time.Time{}, fmt.Errorf("get last_synced_at: %w", err)
		}
		if !raw.Valid || raw.String == "" {
			return time.Time{}, nil
		}
		t, err := time.Parse("2006-01-02 15:04:05", raw.String)
		if err != nil {
			t, err = time.Parse(time.RFC3339, raw.String)
			if err != nil {
				return time.Time{}, fmt.Errorf("parse last_synced_at %q: %w", raw.String, err)
			}
		}
		return t, nil
	})
}

// RecordSyncCompleted updates last_synced_at to now for the given campaign.
func RecordSyncCompleted(ctx context.Context, databasePath string, campaignID string) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO compendium_campaign_sync (campaign_id, last_synced_at, last_phase, in_progress)
			VALUES (?, datetime('now'), 'sync', 0)
			ON CONFLICT(campaign_id) DO UPDATE SET
			    last_synced_at = datetime('now'), last_phase = 'sync', in_progress = 0`,
			campaignID)
		if err != nil {
			return fmt.Errorf("record sync completed: %w", err)
		}
		return nil
	})
}

// GetDDBCampaignID returns the user's selected D&D Beyond campaign ID for a
// Lootsheet campaign, or 0 if not yet set.
func GetDDBCampaignID(ctx context.Context, databasePath string, campaignID string) (int, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (int, error) {
		var id int
		err := db.QueryRowContext(ctx,
			`SELECT ddb_campaign_id FROM compendium_campaign_sync WHERE campaign_id = ?`, campaignID).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, nil
			}
			return 0, fmt.Errorf("get ddb_campaign_id: %w", err)
		}
		return id, nil
	})
}

// SetDDBCampaignID stores the active D&D Beyond campaign ID for a Lootsheet campaign.
func SetDDBCampaignID(ctx context.Context, databasePath string, campaignID string, ddbCampaignID int) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO compendium_campaign_sync (campaign_id, ddb_campaign_id)
			VALUES (?, ?)
			ON CONFLICT(campaign_id) DO UPDATE SET ddb_campaign_id = excluded.ddb_campaign_id`,
			campaignID, ddbCampaignID)
		if err != nil {
			return fmt.Errorf("set ddb_campaign_id: %w", err)
		}
		return nil
	})
}

// SpellBearingEnabledSourceIDs returns IDs of enabled sources where has_spells=1.
func SpellBearingEnabledSourceIDs(ctx context.Context, databasePath string, campaignID string) ([]int, error) {
	return enabledSourceIDsByContent(ctx, databasePath, campaignID, "has_spells")
}

// ItemBearingEnabledSourceIDs returns IDs of enabled sources where has_items=1.
func ItemBearingEnabledSourceIDs(ctx context.Context, databasePath string, campaignID string) ([]int, error) {
	return enabledSourceIDsByContent(ctx, databasePath, campaignID, "has_items")
}

func enabledSourceIDsByContent(ctx context.Context, databasePath string, campaignID string, column string) ([]int, error) {
	if column != "has_spells" && column != "has_items" {
		return nil, fmt.Errorf("invalid content column: %q", column)
	}
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]int, error) {
		query := "SELECT source_id FROM compendium_campaign_sources WHERE campaign_id = ? AND enabled = 1 AND " + column + " = 1" //nolint:gosec // column validated.
		rows, err := db.QueryContext(ctx, query, campaignID)
		if err != nil {
			return nil, fmt.Errorf("enabled %s: %w", column, err)
		}
		defer rows.Close()

		var ids []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return nil, fmt.Errorf("scan %s id: %w", column, err)
			}
			ids = append(ids, id)
		}
		return ids, rows.Err()
	})
}

// UpsertMonsters inserts or replaces monsters (from DDB sync).
func UpsertMonsters(ctx context.Context, databasePath string, monsters []Monster) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO compendium_monsters (ddb_id, name, cr, type, size, hp, ac, source_name, detail_json)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(ddb_id) DO UPDATE SET
			     name = excluded.name, cr = excluded.cr, type = excluded.type, size = excluded.size,
			     hp = excluded.hp, ac = excluded.ac, source_name = excluded.source_name,
			     detail_json = excluded.detail_json, synced_at = datetime('now')`)
		if err != nil {
			return fmt.Errorf("prepare upsert monsters: %w", err)
		}
		defer stmt.Close()

		for i := range monsters {
			m := &monsters[i]
			if _, err := stmt.ExecContext(ctx, m.DdbID, m.Name, m.CR, m.Type, m.Size, m.HP, m.AC, m.SourceName, m.DetailJSON); err != nil {
				return fmt.Errorf("upsert monster %d: %w", m.DdbID, err)
			}
		}
		return tx.Commit()
	})
}

func PruneMonsters(ctx context.Context, databasePath string, keepDDBIDs []int) error {
	return pruneByDDBID(ctx, databasePath, "compendium_monsters", keepDDBIDs)
}

// UpsertSpells inserts or replaces spells (from DDB sync).
func UpsertSpells(ctx context.Context, databasePath string, spells []Spell) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO compendium_spells (ddb_id, name, level, school, casting_time, range, components, duration, classes, source_name, detail_json)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(ddb_id) DO UPDATE SET
			     name = excluded.name, level = excluded.level, school = excluded.school,
			     casting_time = excluded.casting_time, range = excluded.range, components = excluded.components,
			     duration = excluded.duration, classes = excluded.classes, source_name = excluded.source_name,
			     detail_json = excluded.detail_json, synced_at = datetime('now')`)
		if err != nil {
			return fmt.Errorf("prepare upsert spells: %w", err)
		}
		defer stmt.Close()

		for i := range spells {
			s := &spells[i]
			if _, err := stmt.ExecContext(ctx, s.DdbID, s.Name, s.Level, s.School, s.CastingTime, s.Range, s.Components, s.Duration, s.Classes, s.SourceName, s.DetailJSON); err != nil {
				return fmt.Errorf("upsert spell %d: %w", s.DdbID, err)
			}
		}
		return tx.Commit()
	})
}

func PruneSpells(ctx context.Context, databasePath string, keepDDBIDs []int) error {
	return pruneByDDBID(ctx, databasePath, "compendium_spells", keepDDBIDs)
}

// UpsertItems inserts or replaces items (from DDB sync).
func UpsertItems(ctx context.Context, databasePath string, items []Item) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO compendium_items (ddb_id, name, type, rarity, attunement, source_name, detail_json)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(ddb_id) DO UPDATE SET
			     name = excluded.name, type = excluded.type, rarity = excluded.rarity,
			     attunement = excluded.attunement, source_name = excluded.source_name,
			     detail_json = excluded.detail_json, synced_at = datetime('now')`)
		if err != nil {
			return fmt.Errorf("prepare upsert items: %w", err)
		}
		defer stmt.Close()

		for _, i := range items {
			attunement := 0
			if i.Attunement {
				attunement = 1
			}
			if _, err := stmt.ExecContext(ctx, i.DdbID, i.Name, i.Type, i.Rarity, attunement, i.SourceName, i.DetailJSON); err != nil {
				return fmt.Errorf("upsert item %d: %w", i.DdbID, err)
			}
		}
		return tx.Commit()
	})
}

func PruneItems(ctx context.Context, databasePath string, keepDDBIDs []int) error {
	return pruneByDDBID(ctx, databasePath, "compendium_items", keepDDBIDs)
}

// UpsertRules inserts or replaces rules (from DDB config sync).
func UpsertRules(ctx context.Context, databasePath string, rules []Rule) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO compendium_rules (ddb_id, name, category, description)
			 VALUES (?, ?, ?, ?)
			 ON CONFLICT(ddb_id) DO UPDATE SET
			     name = excluded.name, category = excluded.category,
			     description = excluded.description, synced_at = datetime('now')`)
		if err != nil {
			return fmt.Errorf("prepare upsert rules: %w", err)
		}
		defer stmt.Close()

		for _, r := range rules {
			if _, err := stmt.ExecContext(ctx, r.DdbID, r.Name, r.Category, r.Description); err != nil {
				return fmt.Errorf("upsert rule %d: %w", r.DdbID, err)
			}
		}
		return tx.Commit()
	})
}

func pruneByDDBID(ctx context.Context, databasePath string, table string, keepDDBIDs []int) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		if len(keepDDBIDs) == 0 {
			if _, err := db.ExecContext(ctx, "DELETE FROM "+table); err != nil { //nolint:gosec // table is internal constant from typed wrappers.
				return fmt.Errorf("prune %s: %w", table, err)
			}
			return nil
		}

		placeholders := make([]string, len(keepDDBIDs))
		args := make([]any, len(keepDDBIDs))
		for i, id := range keepDDBIDs {
			placeholders[i] = "?"
			args[i] = id
		}
		query := "DELETE FROM " + table + " WHERE ddb_id NOT IN (" + strings.Join(placeholders, ", ") + ")" //nolint:gosec // table is internal constant from typed wrappers; placeholders are ?.
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("prune %s: %w", table, err)
		}
		return nil
	})
}

// UpsertConditions inserts or replaces conditions (from DDB config sync).
func UpsertConditions(ctx context.Context, databasePath string, conditions []Condition) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO compendium_conditions (ddb_id, name, description)
			 VALUES (?, ?, ?)
			 ON CONFLICT(ddb_id) DO UPDATE SET
			     name = excluded.name, description = excluded.description, synced_at = datetime('now')`)
		if err != nil {
			return fmt.Errorf("prepare upsert conditions: %w", err)
		}
		defer stmt.Close()

		for _, c := range conditions {
			if _, err := stmt.ExecContext(ctx, c.DdbID, c.Name, c.Description); err != nil {
				return fmt.Errorf("upsert condition %d: %w", c.DdbID, err)
			}
		}
		return tx.Commit()
	})
}

// ftsQuery converts a user search string into a safe FTS5 prefix query.
// Each word is quoted to prevent FTS5 operator injection (AND, OR, NOT, NEAR, etc.).
func ftsQuery(q string) string {
	words := strings.Fields(q)
	for i, w := range words {
		w = strings.ReplaceAll(w, `"`, `""`)
		words[i] = `"` + w + `"` + `*`
	}
	return strings.Join(words, " ")
}
