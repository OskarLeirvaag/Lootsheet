package compendium

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// ListMonsters returns all monsters, optionally filtered by a search query (FTS).
func ListMonsters(ctx context.Context, databasePath string, query string) ([]Monster, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]Monster, error) {
		q := `SELECT id, ddb_id, name, cr, type, size, hp, ac, source_name, detail_json, synced_at
		      FROM compendium_monsters ORDER BY name`
		args := []any{}
		if query != "" {
			q = `SELECT m.id, m.ddb_id, m.name, m.cr, m.type, m.size, m.hp, m.ac, m.source_name, m.detail_json, m.synced_at
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
		q := `SELECT id, ddb_id, name, level, school, casting_time, range, components, duration, classes, source_name, detail_json, synced_at
		      FROM compendium_spells ORDER BY level, name`
		args := []any{}
		if query != "" {
			q = `SELECT s.id, s.ddb_id, s.name, s.level, s.school, s.casting_time, s.range, s.components, s.duration, s.classes, s.source_name, s.detail_json, s.synced_at
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
		q := `SELECT id, ddb_id, name, type, rarity, attunement, source_name, detail_json, synced_at
		      FROM compendium_items ORDER BY name`
		args := []any{}
		if query != "" {
			q = `SELECT i.id, i.ddb_id, i.name, i.type, i.rarity, i.attunement, i.source_name, i.detail_json, i.synced_at
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
		rows, err := db.QueryContext(ctx, `SELECT id, name, enabled FROM compendium_sources ORDER BY name`)
		if err != nil {
			return nil, fmt.Errorf("list sources: %w", err)
		}
		defer rows.Close()

		var result []Source
		for rows.Next() {
			var s Source
			if err := rows.Scan(&s.ID, &s.Name, &s.Enabled); err != nil {
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

// UpsertSources inserts or updates source books (from DDB config).
func UpsertSources(ctx context.Context, databasePath string, sources []Source) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		stmt, err := db.PrepareContext(ctx,
			`INSERT INTO compendium_sources (id, name, enabled) VALUES (?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET name = excluded.name`)
		if err != nil {
			return fmt.Errorf("prepare upsert sources: %w", err)
		}
		defer stmt.Close()

		for _, s := range sources {
			enabled := 0
			if s.Enabled {
				enabled = 1
			}
			if _, err := stmt.ExecContext(ctx, s.ID, s.Name, enabled); err != nil {
				return fmt.Errorf("upsert source %d: %w", s.ID, err)
			}
		}
		return nil
	})
}

// UpsertMonsters inserts or replaces monsters (from DDB sync).
func UpsertMonsters(ctx context.Context, databasePath string, monsters []Monster) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		stmt, err := db.PrepareContext(ctx,
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

		for _, m := range monsters {
			if _, err := stmt.ExecContext(ctx, m.DdbID, m.Name, m.CR, m.Type, m.Size, m.HP, m.AC, m.SourceName, m.DetailJSON); err != nil {
				return fmt.Errorf("upsert monster %d: %w", m.DdbID, err)
			}
		}
		return nil
	})
}

// UpsertSpells inserts or replaces spells (from DDB sync).
func UpsertSpells(ctx context.Context, databasePath string, spells []Spell) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		stmt, err := db.PrepareContext(ctx,
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

		for _, s := range spells {
			if _, err := stmt.ExecContext(ctx, s.DdbID, s.Name, s.Level, s.School, s.CastingTime, s.Range, s.Components, s.Duration, s.Classes, s.SourceName, s.DetailJSON); err != nil {
				return fmt.Errorf("upsert spell %d: %w", s.DdbID, err)
			}
		}
		return nil
	})
}

// UpsertItems inserts or replaces items (from DDB sync).
func UpsertItems(ctx context.Context, databasePath string, items []Item) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		stmt, err := db.PrepareContext(ctx,
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
		return nil
	})
}

// UpsertRules inserts or replaces rules (from DDB config sync).
func UpsertRules(ctx context.Context, databasePath string, rules []Rule) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		stmt, err := db.PrepareContext(ctx,
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
		return nil
	})
}

// UpsertConditions inserts or replaces conditions (from DDB config sync).
func UpsertConditions(ctx context.Context, databasePath string, conditions []Condition) error {
	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		stmt, err := db.PrepareContext(ctx,
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
		return nil
	})
}

// ftsQuery converts a user search string into an FTS5 prefix query.
func ftsQuery(q string) string {
	words := strings.Fields(q)
	for i, w := range words {
		words[i] = w + "*"
	}
	return strings.Join(words, " ")
}
