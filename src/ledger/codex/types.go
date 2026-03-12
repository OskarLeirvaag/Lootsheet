// Package codex provides repository and CLI handler functions for managing
// codex entries (NPCs, players, contacts) and their cross-reference notes.
package codex

// CodexType represents a row from the codex_types table.
type CodexType struct {
	ID     string
	Name   string
	FormID string
}

// CodexEntry represents a codex_entries row joined with codex_types.
type CodexEntry struct {
	ID          string
	TypeID      string
	TypeName    string
	Name        string
	Title       string
	Location    string
	Faction     string
	Disposition string
	PartyMember bool
	Class       string
	Race        string
	Background  string
	Description string
	Notes       string
	CreatedAt   string
	UpdatedAt   string
}

// CreateInput holds the parameters for creating a new codex entry.
type CreateInput struct {
	TypeID      string
	Name        string
	Title       string
	Location    string
	Faction     string
	Disposition string
	Class       string
	Race        string
	Background  string
	Description string
	Notes       string
}

// UpdateInput holds the parameters for editing a codex entry.
type UpdateInput struct {
	TypeID      string
	Name        string
	Title       string
	Location    string
	Faction     string
	Disposition string
	Class       string
	Race        string
	Background  string
	Description string
	Notes       string
}

// Reference represents a parsed @type/name mention from a codex entry's notes.
type Reference struct {
	ID         string
	EntryID    string
	TargetType string
	TargetName string
	CreatedAt  string
}
