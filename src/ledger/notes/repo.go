package notes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// CreateNote inserts a new note into the database and rebuilds references.
func CreateNote(ctx context.Context, databasePath string, input *CreateNoteInput) (NoteRecord, error) {
	if input == nil {
		return NoteRecord{}, fmt.Errorf("note input is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return NoteRecord{}, fmt.Errorf("note title is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (NoteRecord, error) {
		id := uuid.NewString()
		body := strings.TrimSpace(input.Body)

		if _, err := db.ExecContext(ctx,
			`INSERT INTO notes (id, title, body)
			 VALUES (?, ?, ?)`,
			id, title, body,
		); err != nil {
			return NoteRecord{}, fmt.Errorf("insert note: %w", err)
		}

		if err := rebuildReferences(ctx, db, id, body); err != nil {
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
func UpdateNote(ctx context.Context, databasePath string, noteID string, input *UpdateNoteInput) (NoteRecord, error) {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return NoteRecord{}, fmt.Errorf("note ID is required")
	}
	if input == nil {
		return NoteRecord{}, fmt.Errorf("note input is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return NoteRecord{}, fmt.Errorf("note title is required")
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

		if err := rebuildReferences(ctx, db, noteID, body); err != nil {
			return NoteRecord{}, err
		}

		return getNoteByID(ctx, db, noteID)
	})
}

// DeleteNote removes a note from the database. References cascade.
func DeleteNote(ctx context.Context, databasePath string, noteID string) error {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return fmt.Errorf("note ID is required")
	}

	return ledger.WithDB(ctx, databasePath, func(db *sql.DB) error {
		result, err := db.ExecContext(ctx, "DELETE FROM notes WHERE id = ?", noteID)
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
		return nil
	})
}

// ListNotes returns all notes ordered by updated_at DESC.
func ListNotes(ctx context.Context, databasePath string) ([]NoteRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]NoteRecord, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT id, title, body, created_at, updated_at
			FROM notes
			ORDER BY updated_at DESC
		`)
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

// SearchNotes returns notes matching a LIKE query across title and body.
func SearchNotes(ctx context.Context, databasePath string, query string) ([]NoteRecord, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return ListNotes(ctx, databasePath)
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]NoteRecord, error) {
		pattern := "%" + query + "%"
		rows, err := db.QueryContext(ctx, `
			SELECT id, title, body, created_at, updated_at
			FROM notes
			WHERE title LIKE ? OR body LIKE ?
			ORDER BY updated_at DESC
		`, pattern, pattern)
		if err != nil {
			return nil, fmt.Errorf("search notes: %w", err)
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

// ListAllReferences returns all notes_references rows grouped by note_id.
func ListAllReferences(ctx context.Context, databasePath string) (map[string][]ReferenceRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (map[string][]ReferenceRecord, error) {
		refs, err := queryReferences(ctx, db, "SELECT id, note_id, target_type, target_name, created_at FROM notes_references ORDER BY created_at")
		if err != nil {
			return nil, err
		}

		result := make(map[string][]ReferenceRecord)
		for _, ref := range refs {
			result[ref.NoteID] = append(result[ref.NoteID], ref)
		}
		return result, nil
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

func queryReferences(ctx context.Context, db *sql.DB, query string, args ...any) ([]ReferenceRecord, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query references: %w", err)
	}
	defer rows.Close()

	refs := []ReferenceRecord{}
	for rows.Next() {
		var r ReferenceRecord
		if err := rows.Scan(&r.ID, &r.NoteID, &r.TargetType, &r.TargetName, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan reference row: %w", err)
		}
		refs = append(refs, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reference rows: %w", err)
	}

	return refs, nil
}
