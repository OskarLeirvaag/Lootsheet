package codex

import (
	"context"
	"database/sql"
	"fmt"
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
		return CodexType{}, fmt.Errorf("codex type ID is required")
	}
	if name == "" {
		return CodexType{}, fmt.Errorf("codex type name is required")
	}
	if !validFormIDs[formID] {
		return CodexType{}, fmt.Errorf("form_id must be one of: npc, player, settlement")
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
		return fmt.Errorf("codex type ID is required")
	}
	if newName == "" {
		return fmt.Errorf("new name is required")
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
		return fmt.Errorf("codex type ID is required")
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
		return CodexEntry{}, fmt.Errorf("codex entry input is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return CodexEntry{}, fmt.Errorf("codex entry name is required")
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
			                            class, race, background, description, notes)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, campaignID, typeID, name,
			strings.TrimSpace(input.Title), strings.TrimSpace(input.Location),
			strings.TrimSpace(input.Faction), strings.TrimSpace(input.Disposition),
			partyMember,
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
		return CodexEntry{}, fmt.Errorf("codex entry ID is required")
	}
	if input == nil {
		return CodexEntry{}, fmt.Errorf("codex entry input is required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return CodexEntry{}, fmt.Errorf("codex entry name is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (CodexEntry, error) {
		var exists int
		if err := db.QueryRowContext(ctx, "SELECT 1 FROM codex_entries WHERE id = ?", entryID).Scan(&exists); err != nil {
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
			     party_member = ?, class = ?, race = ?, background = ?, description = ?,
			     notes = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			typeID, name, strings.TrimSpace(input.Title), strings.TrimSpace(input.Location),
			strings.TrimSpace(input.Faction), strings.TrimSpace(input.Disposition),
			partyMember,
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
		return fmt.Errorf("codex entry ID is required")
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
			       e.party_member, e.class, e.race, e.background, e.description, e.notes,
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
				&e.Disposition, &partyMember, &e.Class, &e.Race, &e.Background,
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

// SearchEntries returns codex entries matching a LIKE query.
func SearchEntries(ctx context.Context, databasePath string, campaignID string, query string) ([]CodexEntry, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return ListEntries(ctx, databasePath, campaignID)
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]CodexEntry, error) {
		pattern := "%" + query + "%"
		rows, err := db.QueryContext(ctx, `
			SELECT e.id, e.type_id, t.name, e.name, e.title, e.location, e.faction, e.disposition,
			       e.party_member, e.class, e.race, e.background, e.description, e.notes,
			       e.created_at, e.updated_at
			FROM codex_entries e
			JOIN codex_types t ON t.id = e.type_id
			WHERE (e.name LIKE ? OR e.title LIKE ? OR e.location LIKE ? OR e.faction LIKE ?
			   OR e.notes LIKE ? OR e.class LIKE ? OR e.race LIKE ? OR e.description LIKE ?)
			   AND e.campaign_id = ?
			ORDER BY e.party_member DESC, t.name ASC, e.name ASC
		`, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, campaignID)
		if err != nil {
			return nil, fmt.Errorf("search codex entries: %w", err)
		}
		defer rows.Close()

		entries := []CodexEntry{}
		for rows.Next() {
			var e CodexEntry
			var partyMember int

			if err := rows.Scan(
				&e.ID, &e.TypeID, &e.TypeName, &e.Name, &e.Title, &e.Location, &e.Faction,
				&e.Disposition, &partyMember, &e.Class, &e.Race, &e.Background,
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
		       e.party_member, e.class, e.race, e.background, e.description, e.notes,
		       e.created_at, e.updated_at
		FROM codex_entries e
		JOIN codex_types t ON t.id = e.type_id
		WHERE e.id = ?
	`, entryID).Scan(
		&record.ID, &record.TypeID, &record.TypeName, &record.Name, &record.Title, &record.Location,
		&record.Faction, &record.Disposition, &partyMember, &record.Class, &record.Race,
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
