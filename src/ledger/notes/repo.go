package notes

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

// CreateNote inserts a new note into the database and rebuilds references.
func CreateNote(ctx context.Context, databasePath string, campaignID string, input *CreateNoteInput) (NoteRecord, error) {
	if input == nil {
		return NoteRecord{}, errors.New("note input is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return NoteRecord{}, errors.New("note title is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (NoteRecord, error) {
		id := uuid.NewString()
		body := strings.TrimSpace(input.Body)

		if _, err := db.ExecContext(ctx,
			`INSERT INTO notes (id, campaign_id, title, body)
			 VALUES (?, ?, ?, ?)`,
			id, campaignID, title, body,
		); err != nil {
			return NoteRecord{}, fmt.Errorf("insert note: %w", err)
		}

		if err := rebuildReferences(ctx, db, id, campaignID, title, body); err != nil {
			return NoteRecord{}, err
		}

		return NoteRecord{
			ID:    id,
			Title: title,
			Body:  body,
		}, nil
	})
}

// UpdateNote edits a note's fields and rebuilds references.
func UpdateNote(ctx context.Context, databasePath string, campaignID string, noteID string, input *UpdateNoteInput) (NoteRecord, error) {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return NoteRecord{}, errors.New("note ID is required")
	}
	if input == nil {
		return NoteRecord{}, errors.New("note input is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return NoteRecord{}, errors.New("note title is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (NoteRecord, error) {
		// Verify note exists.
		var exists int
		if err := db.QueryRowContext(ctx, "SELECT 1 FROM notes WHERE id = ?", noteID).Scan(&exists); err != nil {
			if err == sql.ErrNoRows {
				return NoteRecord{}, fmt.Errorf("note %q does not exist", noteID)
			}
			return NoteRecord{}, fmt.Errorf("query note: %w", err)
		}

		body := strings.TrimSpace(input.Body)

		if _, err := db.ExecContext(ctx,
			`UPDATE notes
			 SET title = ?, body = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			title, body, noteID,
		); err != nil {
			return NoteRecord{}, fmt.Errorf("update note: %w", err)
		}

		if err := rebuildReferences(ctx, db, noteID, campaignID, title, body); err != nil {
			return NoteRecord{}, err
		}

		return getNoteByID(ctx, db, noteID)
	})
}

// DeleteNote removes a note and its outbound entity_references rows.
func DeleteNote(ctx context.Context, databasePath string, campaignID string, noteID string) error {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return errors.New("note ID is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		result, err := db.ExecContext(ctx, "DELETE FROM notes WHERE id = ? AND campaign_id = ?", noteID, campaignID)
		if err != nil {
			return fmt.Errorf("delete note: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("check delete result: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("note %q does not exist", noteID)
		}
		return refs.DeleteBySource(ctx, db, "note", noteID)
	})
}

// ListNotes returns all notes ordered by updated_at DESC.
func ListNotes(ctx context.Context, databasePath string, campaignID string) ([]NoteRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]NoteRecord, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT id, title, body, created_at, updated_at
			FROM notes
			WHERE campaign_id = ?
			ORDER BY updated_at DESC
		`, campaignID)
		if err != nil {
			return nil, fmt.Errorf("query notes: %w", err)
		}
		defer rows.Close()

		notes := []NoteRecord{}
		for rows.Next() {
			var n NoteRecord
			if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt); err != nil {
				return nil, fmt.Errorf("scan note row: %w", err)
			}
			notes = append(notes, n)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate note rows: %w", err)
		}

		return notes, nil
	})
}

// SearchNotes returns notes matching the query.
// First tries an FTS5 prefix search over title and body; if that yields no
// results, falls back to a case-insensitive LIKE substring match. The LIKE
// fallback catches cases where the FTS tokenizer misses the term.
func SearchNotes(ctx context.Context, databasePath string, campaignID string, query string) ([]NoteRecord, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return ListNotes(ctx, databasePath, campaignID)
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]NoteRecord, error) {
		direct, err := searchNotesFTS(ctx, db, campaignID, query)
		if err != nil || len(direct) == 0 {
			direct, err = searchNotesLIKE(ctx, db, campaignID, query)
			if err != nil {
				return nil, err
			}
		}

		// Also include notes whose refs have a target_name matching the query.
		indirect, err := searchNotesByReference(ctx, db, campaignID, query)
		if err != nil {
			return direct, nil //nolint:nilerr // reference expansion is best-effort
		}

		return mergeNotes(direct, indirect), nil
	})
}

func searchNotesByReference(ctx context.Context, db *sql.DB, campaignID, query string) ([]NoteRecord, error) {
	ids, err := refs.FindSourcesReferencingTarget(ctx, db, campaignID, "note", ledger.LIKEPattern(query))
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
		SELECT id, title, body, created_at, updated_at
		FROM notes
		WHERE campaign_id = ? AND id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY updated_at DESC`

	rows, err := db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("search notes (refs): %w", err)
	}
	defer rows.Close()
	return scanNotes(rows)
}

func mergeNotes(a, b []NoteRecord) []NoteRecord {
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

func searchNotesFTS(ctx context.Context, db *sql.DB, campaignID, query string) ([]NoteRecord, error) {
	ftsQuery := ledger.FTSQuery(query)
	if ftsQuery == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT n.id, n.title, n.body, n.created_at, n.updated_at
		FROM notes n
		JOIN notes_fts f ON f.rowid = n.rowid
		WHERE notes_fts MATCH ?
		  AND n.campaign_id = ?
		ORDER BY n.updated_at DESC
	`, ftsQuery, campaignID)
	if err != nil {
		return nil, fmt.Errorf("search notes (fts): %w", err)
	}
	defer rows.Close()
	return scanNotes(rows)
}

func searchNotesLIKE(ctx context.Context, db *sql.DB, campaignID, query string) ([]NoteRecord, error) {
	pattern := ledger.LIKEPattern(query)
	rows, err := db.QueryContext(ctx, `
		SELECT id, title, body, created_at, updated_at
		FROM notes
		WHERE campaign_id = ?
		  AND (title LIKE ? ESCAPE '\' OR body LIKE ? ESCAPE '\')
		ORDER BY updated_at DESC
	`, campaignID, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("search notes (like): %w", err)
	}
	defer rows.Close()
	return scanNotes(rows)
}

func scanNotes(rows *sql.Rows) ([]NoteRecord, error) {
	notes := []NoteRecord{}
	for rows.Next() {
		var n NoteRecord
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note row: %w", err)
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate note rows: %w", err)
	}
	return notes, nil
}

// ListAllReferences returns all entity_references rows for note source type, grouped by source_id.
func ListAllReferences(ctx context.Context, databasePath string, campaignID string) (map[string][]refs.EntityReference, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (map[string][]refs.EntityReference, error) {
		return refs.ListBySource(ctx, db, "note", campaignID)
	})
}

func getNoteByID(ctx context.Context, db *sql.DB, noteID string) (NoteRecord, error) {
	var record NoteRecord

	if err := db.QueryRowContext(ctx, `
		SELECT id, title, body, created_at, updated_at
		FROM notes
		WHERE id = ?
	`, noteID).Scan(
		&record.ID, &record.Title, &record.Body, &record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return NoteRecord{}, fmt.Errorf("note %q does not exist", noteID)
		}
		return NoteRecord{}, fmt.Errorf("query note: %w", err)
	}

	return record, nil
}
