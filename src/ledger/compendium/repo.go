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

// ListSources returns all source books.
func ListSources(ctx context.Context, databasePath string) ([]Source, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Source, error) {
		rows, err := db.QueryContext(ctx, `SELECT id, name, enabled, owned, has_spells, has_items, is_released, category_id FROM compendium_sources ORDER BY name`)
		if err != nil {
			return nil, fmt.Errorf("list sources: %w", err)
		}
		defer rows.Close()

		var result []Source
		for rows.Next() {
			var s Source
			if err := rows.Scan(&s.ID, &s.Name, &s.Enabled, &s.Owned, &s.HasSpells, &s.HasItems, &s.IsReleased, &s.CategoryID); err != nil {
				return nil, fmt.Errorf("scan source: %w", err)
			}
			result = append(result, s)
		}
		return result, rows.Err()
	})
}

// EnabledSourceIDs returns the IDs of all enabled source books.
func EnabledSourceIDs(ctx context.Context, databasePath string) ([]int, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]int, error) {
		rows, err := db.QueryContext(ctx, `SELECT id FROM compendium_sources WHERE enabled = 1`)
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

// ToggleSource flips the enabled state of a source book.
func ToggleSource(ctx context.Context, databasePath string, sourceID int) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		_, err := db.ExecContext(ctx, `UPDATE compendium_sources SET enabled = 1 - enabled WHERE id = ?`, sourceID)
		return err
	})
}

// UpsertSources inserts or updates source books (from DDB config). Existing
// rows preserve their `enabled`, `owned`, `has_spells`, `has_items` user state;
// only catalogue fields (name, is_released, category_id) refresh.
func UpsertSources(ctx context.Context, databasePath string, sources []Source) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO compendium_sources (id, name, enabled, is_released, category_id)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
			     name = excluded.name,
			     is_released = excluded.is_released,
			     category_id = excluded.category_id`)
		if err != nil {
			return fmt.Errorf("prepare upsert sources: %w", err)
		}
		defer stmt.Close()

		for _, s := range sources {
			enabled := 0
			if s.Enabled {
				enabled = 1
			}
			released := 0
			if s.IsReleased {
				released = 1
			}
			if _, err := stmt.ExecContext(ctx, s.ID, s.Name, enabled, released, s.CategoryID); err != nil {
				return fmt.Errorf("upsert source %d: %w", s.ID, err)
			}
		}
		return tx.Commit()
	})
}

// SetSourceOwnership marks the listed source IDs as owned (1) and every other
// known source as locked (2). Sources not yet probed should be filtered before
// calling this — any row not in `ownedIDs` will be marked locked.
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

// SetSourceHasSpells stores whether a source produced any spells in the most
// recent Phase B fetch. Subsequent Phase B runs skip the spell pass entirely
// when no enabled source has has_spells=1.
func SetSourceHasSpells(ctx context.Context, databasePath string, sourceID int, hasSpells bool) error {
	return setSourceContentFlag(ctx, databasePath, sourceID, "has_spells", hasSpells)
}

// SetSourceHasItems stores whether a source produced any items in the most
// recent Phase B fetch.
func SetSourceHasItems(ctx context.Context, databasePath string, sourceID int, hasItems bool) error {
	return setSourceContentFlag(ctx, databasePath, sourceID, "has_items", hasItems)
}

func setSourceContentFlag(ctx context.Context, databasePath string, sourceID int, column string, value bool) error {
	if column != "has_spells" && column != "has_items" {
		return fmt.Errorf("invalid content column: %q", column)
	}
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		v := 0
		if value {
			v = 1
		}
		query := "UPDATE compendium_sources SET " + column + " = ? WHERE id = ?" //nolint:gosec // column validated.
		if _, err := db.ExecContext(ctx, query, v, sourceID); err != nil {
			return fmt.Errorf("update %s for source %d: %w", column, sourceID, err)
		}
		return nil
	})
}

// GetLastSyncedAt returns the timestamp of the last successful Phase B sync,
// or the zero time if no sync has completed yet.
func GetLastSyncedAt(ctx context.Context, databasePath string) (time.Time, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (time.Time, error) {
		var raw sql.NullString
		err := db.QueryRowContext(ctx, `SELECT last_synced_at FROM compendium_sync_state WHERE id = 1`).Scan(&raw)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return time.Time{}, nil
			}
			return time.Time{}, fmt.Errorf("get last_synced_at: %w", err)
		}
		if !raw.Valid || raw.String == "" {
			return time.Time{}, nil
		}
		// SQLite datetime('now') format: "2006-01-02 15:04:05".
		t, err := time.Parse("2006-01-02 15:04:05", raw.String)
		if err != nil {
			// Try RFC3339 as a fallback.
			t, err = time.Parse(time.RFC3339, raw.String)
			if err != nil {
				return time.Time{}, fmt.Errorf("parse last_synced_at %q: %w", raw.String, err)
			}
		}
		return t, nil
	})
}

// RecordSyncCompleted updates compendium_sync_state.last_synced_at to now.
func RecordSyncCompleted(ctx context.Context, databasePath string) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		_, err := db.ExecContext(ctx,
			`UPDATE compendium_sync_state SET last_synced_at = datetime('now'), last_phase = 'sync', in_progress = 0 WHERE id = 1`)
		if err != nil {
			return fmt.Errorf("record sync completed: %w", err)
		}
		return nil
	})
}

// SpellBearingEnabledSourceIDs returns enabled sources where has_spells=1.
func SpellBearingEnabledSourceIDs(ctx context.Context, databasePath string) ([]int, error) {
	return enabledSourceIDsByContent(ctx, databasePath, "has_spells")
}

// ItemBearingEnabledSourceIDs returns enabled sources where has_items=1.
func ItemBearingEnabledSourceIDs(ctx context.Context, databasePath string) ([]int, error) {
	return enabledSourceIDsByContent(ctx, databasePath, "has_items")
}

func enabledSourceIDsByContent(ctx context.Context, databasePath string, column string) ([]int, error) {
	if column != "has_spells" && column != "has_items" {
		return nil, fmt.Errorf("invalid content column: %q", column)
	}
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]int, error) {
		query := "SELECT id FROM compendium_sources WHERE enabled = 1 AND " + column + " = 1" //nolint:gosec // column is validated above.
		rows, err := db.QueryContext(ctx, query)
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
