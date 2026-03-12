package notes

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type parsedRef struct {
	TargetType string
	TargetName string
}

var refPattern = regexp.MustCompile(`@(quest|loot|asset|person|note)/([^\s@]+(?:\s+[^\s@]+)*)`)

// ParseReferences extracts @type/name pairs from body text.
func ParseReferences(body string) []parsedRef {
	matches := refPattern.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}

	refs := make([]parsedRef, 0, len(matches))
	for _, match := range matches {
		targetType := strings.TrimSpace(match[1])
		targetName := strings.TrimSpace(match[2])
		// Trim trailing punctuation that was captured as part of the name.
		targetName = strings.TrimRight(targetName, ".,;:!?")
		targetName = strings.TrimSpace(targetName)
		if targetType == "" || targetName == "" {
			continue
		}
		refs = append(refs, parsedRef{
			TargetType: targetType,
			TargetName: targetName,
		})
	}

	return refs
}

// rebuildReferences deletes old references for a note and inserts new ones
// parsed from the body text. Must be called within a transaction or single
// connection context.
func rebuildReferences(ctx context.Context, db *sql.DB, noteID string, body string) error {
	if _, err := db.ExecContext(ctx,
		"DELETE FROM notes_references WHERE note_id = ?", noteID,
	); err != nil {
		return fmt.Errorf("delete old references: %w", err)
	}

	refs := ParseReferences(body)
	for _, ref := range refs {
		id := uuid.NewString()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO notes_references (id, note_id, target_type, target_name)
			 VALUES (?, ?, ?, ?)`,
			id, noteID, ref.TargetType, ref.TargetName,
		); err != nil {
			return fmt.Errorf("insert reference: %w", err)
		}
	}

	return nil
}
