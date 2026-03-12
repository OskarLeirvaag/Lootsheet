// Package notes provides repository and CLI handler functions for managing
// campaign/session notes with cross-reference support.
package notes

// NoteRecord represents a note row from the database.
type NoteRecord struct {
	ID        string
	Title     string
	Body      string
	CreatedAt string
	UpdatedAt string
}

// CreateNoteInput holds the parameters for creating a new note.
type CreateNoteInput struct {
	Title string
	Body  string
}

// UpdateNoteInput holds the parameters for editing a note.
type UpdateNoteInput struct {
	Title string
	Body  string
}

// ReferenceRecord represents a parsed @type/name mention from a note's body.
type ReferenceRecord struct {
	ID         string
	NoteID     string
	TargetType string
	TargetName string
	CreatedAt  string
}
