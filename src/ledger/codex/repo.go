package codex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
)

var validFormIDs = map[string]bool{
	"player":     true,
	"npc":        true,
	"settlement": true,
}

// CreateType inserts a new codex type.
func CreateType(ctx context.Context, databasePath string, id, name, formID string) (CodexType, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	formID = strings.TrimSpace(formID)

	if id == "" {
		return CodexType{}, errors.New("codex type ID is required")
	}
	if name == "" {
		return CodexType{}, errors.New("codex type name is required")
	}
	if !validFormIDs[formID] {
		return CodexType{}, errors.New("form_id must be one of: npc, player, settlement")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (CodexType, error) {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO codex_types (id, name, form_id) VALUES (?, ?, ?)`,
			id, name, formID,
		); err != nil {
			return CodexType{}, fmt.Errorf("insert codex type: %w", err)
		}
		return CodexType{ID: id, Name: name, FormID: formID}, nil
	})
}

// RenameType updates the display name of a codex type.
func RenameType(ctx context.Context, databasePath string, typeID, newName string) error {
	typeID = strings.TrimSpace(typeID)
	newName = strings.TrimSpace(newName)

	if typeID == "" {
		return errors.New("codex type ID is required")
	}
	if newName == "" {
		return errors.New("new name is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		result, err := db.ExecContext(ctx, `UPDATE codex_types SET name = ? WHERE id = ?`, newName, typeID)
		if err != nil {
			return fmt.Errorf("rename codex type: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check rename result: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("codex type %q does not exist", typeID)
		}
		return nil
	})
}

// DeleteType removes a codex type, refusing if entries reference it.
func DeleteType(ctx context.Context, databasePath string, typeID string) error {
	typeID = strings.TrimSpace(typeID)
	if typeID == "" {
		return errors.New("codex type ID is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM codex_entries WHERE type_id = ?`, typeID).Scan(&count); err != nil {
			return fmt.Errorf("check codex entries: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("cannot delete codex type %q: %d entries still reference it", typeID, count)
		}

		result, err := db.ExecContext(ctx, `DELETE FROM codex_types WHERE id = ?`, typeID)
		if err != nil {
			return fmt.Errorf("delete codex type: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check delete result: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("codex type %q does not exist", typeID)
		}
		return nil
	})
}

// ListTypes returns all codex types ordered by name.
func ListTypes(ctx context.Context, databasePath string) ([]CodexType, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]CodexType, error) {
		rows, err := db.QueryContext(ctx, `SELECT id, name, form_id FROM codex_types ORDER BY name`)
		if err != nil {
			return nil, fmt.Errorf("query codex types: %w", err)
		}
		defer rows.Close()

		types := []CodexType{}
		for rows.Next() {
			var t CodexType
			if err := rows.Scan(&t.ID, &t.Name, &t.FormID); err != nil {
				return nil, fmt.Errorf("scan codex type row: %w", err)
			}
			types = append(types, t)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate codex type rows: %w", err)
		}

		return types, nil
	})
}

// CreateEntry inserts a new codex entry and rebuilds references.
func CreateEntry(ctx context.Context, databasePath string, campaignID string, input *CreateInput) (CodexEntry, error) {
	if input == nil {
		return CodexEntry{}, errors.New("codex entry input is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return CodexEntry{}, errors.New("codex entry name is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (CodexEntry, error) {
		id := uuid.NewString()
		typeID := strings.TrimSpace(input.TypeID)
		if typeID == "" {
			typeID = "npc"
		}

		partyMember := 0
		if typeID == "player" {
			partyMember = 1
		}

		notes := strings.TrimSpace(input.Notes)

		if _, err := db.ExecContext(ctx,
			`INSERT INTO codex_entries (id, campaign_id, type_id, name, title, location, faction, disposition, party_member,
			                            player_name, class, race, background, description, notes)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, campaignID, typeID, name,
			strings.TrimSpace(input.Title), strings.TrimSpace(input.Location),
			strings.TrimSpace(input.Faction), strings.TrimSpace(input.Disposition),
			partyMember,
			strings.TrimSpace(input.PlayerName),
			strings.TrimSpace(input.Class), strings.TrimSpace(input.Race),
			strings.TrimSpace(input.Background), strings.TrimSpace(input.Description),
			notes,
		); err != nil {
			return CodexEntry{}, fmt.Errorf("insert codex entry: %w", err)
		}

		if err := rebuildReferences(ctx, db, id, campaignID, name, notes); err != nil {
			return CodexEntry{}, err
		}

		return getEntryByID(ctx, db, id)
	})
}

// UpdateEntry edits a codex entry's fields and rebuilds references.
func UpdateEntry(ctx context.Context, databasePath string, campaignID string, entryID string, input *UpdateInput) (CodexEntry, error) {
	entryID = strings.TrimSpace(entryID)
	if entryID == "" {
		return CodexEntry{}, errors.New("codex entry ID is required")
	}
	if input == nil {
		return CodexEntry{}, errors.New("codex entry input is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return CodexEntry{}, errors.New("codex entry name is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (CodexEntry, error) {
		var exists int
		if err := db.QueryRowContext(ctx, "SELECT 1 FROM codex_entries WHERE id = ? AND campaign_id = ?", entryID, campaignID).Scan(&exists); err != nil {
			if err == sql.ErrNoRows {
				return CodexEntry{}, fmt.Errorf("codex entry %q does not exist", entryID)
			}
			return CodexEntry{}, fmt.Errorf("query codex entry: %w", err)
		}

		typeID := strings.TrimSpace(input.TypeID)
		if typeID == "" {
			typeID = "npc"
		}

		partyMember := 0
		if typeID == "player" {
			partyMember = 1
		}

		notes := strings.TrimSpace(input.Notes)

		if _, err := db.ExecContext(ctx,
			`UPDATE codex_entries
			 SET type_id = ?, name = ?, title = ?, location = ?, faction = ?, disposition = ?,
			     party_member = ?, player_name = ?, class = ?, race = ?, background = ?, description = ?,
			     notes = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			typeID, name, strings.TrimSpace(input.Title), strings.TrimSpace(input.Location),
			strings.TrimSpace(input.Faction), strings.TrimSpace(input.Disposition),
			partyMember, strings.TrimSpace(input.PlayerName),
			strings.TrimSpace(input.Class), strings.TrimSpace(input.Race),
			strings.TrimSpace(input.Background), strings.TrimSpace(input.Description),
			notes, entryID,
		); err != nil {
			return CodexEntry{}, fmt.Errorf("update codex entry: %w", err)
		}

		if err := rebuildReferences(ctx, db, entryID, campaignID, name, notes); err != nil {
			return CodexEntry{}, err
		}

		return getEntryByID(ctx, db, entryID)
	})
}

// DeleteEntry removes a codex entry and its outbound entity_references rows.
func DeleteEntry(ctx context.Context, databasePath string, campaignID string, entryID string) error {
	entryID = strings.TrimSpace(entryID)
	if entryID == "" {
		return errors.New("codex entry ID is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		result, err := db.ExecContext(ctx, "DELETE FROM codex_entries WHERE id = ?", entryID)
		if err != nil {
			return fmt.Errorf("delete codex entry: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check delete result: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("codex entry %q does not exist", entryID)
		}
		return refs.DeleteBySource(ctx, db, "codex", entryID)
	})
}

// ListEntries returns all codex entries joined with their type, ordered by type name then entry name.
func ListEntries(ctx context.Context, databasePath string, campaignID string) ([]CodexEntry, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]CodexEntry, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT e.id, e.type_id, t.name, e.name, e.title, e.location, e.faction, e.disposition,
			       e.party_member, e.player_name, e.class, e.race, e.background, e.description, e.notes,
			       e.created_at, e.updated_at
			FROM codex_entries e
			JOIN codex_types t ON t.id = e.type_id
			WHERE e.campaign_id = ?
			ORDER BY e.party_member DESC, t.name ASC, e.name ASC
		`, campaignID)
		if err != nil {
			return nil, fmt.Errorf("query codex entries: %w", err)
		}
		defer rows.Close()

		entries := []CodexEntry{}
		for rows.Next() {
			var e CodexEntry
			var partyMember int

			if err := rows.Scan(
				&e.ID, &e.TypeID, &e.TypeName, &e.Name, &e.Title, &e.Location, &e.Faction,
				&e.Disposition, &partyMember, &e.PlayerName, &e.Class, &e.Race, &e.Background,
				&e.Description, &e.Notes, &e.CreatedAt, &e.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("scan codex entry row: %w", err)
			}

			e.PartyMember = partyMember != 0
			entries = append(entries, e)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate codex entry rows: %w", err)
		}

		return entries, nil
	})
}

// SearchEntries returns codex entries matching the query.
// First tries an FTS5 prefix search; if that yields no results, falls back to
// a case-insensitive LIKE substring match across all searchable fields. The
// LIKE fallback catches cases where the FTS tokenizer misses the term.
func SearchEntries(ctx context.Context, databasePath string, campaignID string, query string) ([]CodexEntry, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return ListEntries(ctx, databasePath, campaignID)
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]CodexEntry, error) {
		direct, err := searchEntriesFTS(ctx, db, campaignID, query)
		if err != nil || len(direct) == 0 {
			// Fall back to LIKE so the user still gets results on FTS tokenizer
			// quirks or FTS index inconsistencies.
			direct, err = searchEntriesLIKE(ctx, db, campaignID, query)
			if err != nil {
				return nil, err
			}
		}

		// Also include codex entries whose refs have a target_name matching
		// the query — so "Bryn" returns codex entries that reference Bryn.
		indirect, err := searchEntriesByReference(ctx, db, campaignID, query)
		if err != nil {
			return direct, nil //nolint:nilerr // reference expansion is best-effort
		}

		return mergeCodexEntries(direct, indirect), nil
	})
}

// searchEntriesByReference returns codex entries that have a ref whose
// target_name matches the query.
func searchEntriesByReference(ctx context.Context, db *sql.DB, campaignID, query string) ([]CodexEntry, error) {
	ids, err := refs.FindSourcesReferencingTarget(ctx, db, campaignID, "codex", ledger.LIKEPattern(query))
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, campaignID)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	// Only '?' placeholders are concatenated; actual IDs are bound via args.
	//nolint:gosec // G202: only '?' placeholders are concatenated, never user input
	sqlStr := `
		SELECT e.id, e.type_id, t.name, e.name, e.title, e.location, e.faction, e.disposition,
		       e.party_member, e.player_name, e.class, e.race, e.background, e.description, e.notes,
		       e.created_at, e.updated_at
		FROM codex_entries e
		JOIN codex_types t ON t.id = e.type_id
		WHERE e.campaign_id = ? AND e.id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY e.party_member DESC, t.name ASC, e.name ASC`

	rows, err := db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("search codex entries (refs): %w", err)
	}
	defer rows.Close()
	return scanCodexEntries(rows)
}

// mergeCodexEntries combines two entry slices, deduplicating by ID.
func mergeCodexEntries(a, b []CodexEntry) []CodexEntry {
	if len(b) == 0 {
		return a
	}
	seen := make(map[string]bool, len(a))
	for i := range a {
		seen[a[i].ID] = true
	}
	result := slices.Clone(a)
	for i := range b {
		if !seen[b[i].ID] {
			seen[b[i].ID] = true
			result = append(result, b[i])
		}
	}
	return result
}

// searchEntriesFTS runs an FTS5 prefix query across all indexed codex columns.
func searchEntriesFTS(ctx context.Context, db *sql.DB, campaignID, query string) ([]CodexEntry, error) {
	ftsQuery := ledger.FTSQuery(query)
	if ftsQuery == "" {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT e.id, e.type_id, t.name, e.name, e.title, e.location, e.faction, e.disposition,
		       e.party_member, e.player_name, e.class, e.race, e.background, e.description, e.notes,
		       e.created_at, e.updated_at
		FROM codex_entries e
		JOIN codex_types t ON t.id = e.type_id
		JOIN codex_fts f ON f.rowid = e.rowid
		WHERE codex_fts MATCH ?
		   AND e.campaign_id = ?
		ORDER BY e.party_member DESC, t.name ASC, e.name ASC
	`, ftsQuery, campaignID)
	if err != nil {
		return nil, fmt.Errorf("search codex entries (fts): %w", err)
	}
	defer rows.Close()

	return scanCodexEntries(rows)
}

// searchEntriesLIKE runs a substring LIKE match across searchable fields.
func searchEntriesLIKE(ctx context.Context, db *sql.DB, campaignID, query string) ([]CodexEntry, error) {
	pattern := ledger.LIKEPattern(query)
	rows, err := db.QueryContext(ctx, `
		SELECT e.id, e.type_id, t.name, e.name, e.title, e.location, e.faction, e.disposition,
		       e.party_member, e.player_name, e.class, e.race, e.background, e.description, e.notes,
		       e.created_at, e.updated_at
		FROM codex_entries e
		JOIN codex_types t ON t.id = e.type_id
		WHERE e.campaign_id = ?
		  AND (e.name LIKE ? ESCAPE '\'
		    OR e.title LIKE ? ESCAPE '\'
		    OR e.location LIKE ? ESCAPE '\'
		    OR e.faction LIKE ? ESCAPE '\'
		    OR e.player_name LIKE ? ESCAPE '\'
		    OR e.class LIKE ? ESCAPE '\'
		    OR e.race LIKE ? ESCAPE '\'
		    OR e.background LIKE ? ESCAPE '\'
		    OR e.description LIKE ? ESCAPE '\'
		    OR e.notes LIKE ? ESCAPE '\')
		ORDER BY e.party_member DESC, t.name ASC, e.name ASC
	`, campaignID, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("search codex entries (like): %w", err)
	}
	defer rows.Close()

	return scanCodexEntries(rows)
}

// scanCodexEntries decodes rows shared by FTS and LIKE search paths.
func scanCodexEntries(rows *sql.Rows) ([]CodexEntry, error) {
	entries := []CodexEntry{}
	for rows.Next() {
		var e CodexEntry
		var partyMember int

		if err := rows.Scan(
			&e.ID, &e.TypeID, &e.TypeName, &e.Name, &e.Title, &e.Location, &e.Faction,
			&e.Disposition, &partyMember, &e.PlayerName, &e.Class, &e.Race, &e.Background,
			&e.Description, &e.Notes, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan codex entry row: %w", err)
		}

		e.PartyMember = partyMember != 0
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate codex entry rows: %w", err)
	}
	return entries, nil
}

// ListAllReferences returns all entity_references rows for codex source type, grouped by source_id.
func ListAllReferences(ctx context.Context, databasePath string, campaignID string) (map[string][]refs.EntityReference, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (map[string][]refs.EntityReference, error) {
		return refs.ListBySource(ctx, db, "codex", campaignID)
	})
}

func getEntryByID(ctx context.Context, db *sql.DB, entryID string) (CodexEntry, error) {
	var record CodexEntry
	var partyMember int

	if err := db.QueryRowContext(ctx, `
		SELECT e.id, e.type_id, t.name, e.name, e.title, e.location, e.faction, e.disposition,
		       e.party_member, e.player_name, e.class, e.race, e.background, e.description, e.notes,
		       e.created_at, e.updated_at
		FROM codex_entries e
		JOIN codex_types t ON t.id = e.type_id
		WHERE e.id = ?
	`, entryID).Scan(
		&record.ID, &record.TypeID, &record.TypeName, &record.Name, &record.Title, &record.Location,
		&record.Faction, &record.Disposition, &partyMember, &record.PlayerName, &record.Class, &record.Race,
		&record.Background, &record.Description, &record.Notes, &record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return CodexEntry{}, fmt.Errorf("codex entry %q does not exist", entryID)
		}
		return CodexEntry{}, fmt.Errorf("query codex entry: %w", err)
	}

	record.PartyMember = partyMember != 0
	return record, nil
}
