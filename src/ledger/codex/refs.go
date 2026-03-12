package codex

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

var refPattern = regexp.MustCompile(`@(quest|loot|asset|person)/([^\s@]+(?:\s+[^\s@]+)*)`)

// ParseReferences extracts @type/name pairs from notes text.
func ParseReferences(notes string) []parsedRef {
	matches := refPattern.FindAllStringSubmatch(notes, -1)
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

// rebuildReferences deletes old references for an entry and inserts new ones
// parsed from the notes text. Must be called within a transaction or single
// connection context.
func rebuildReferences(ctx context.Context, db *sql.DB, entryID string, notes string) error {
	if _, err := db.ExecContext(ctx,
		"DELETE FROM codex_references WHERE entry_id = ?", entryID,
	); err != nil {
		return fmt.Errorf("delete old references: %w", err)
	}

	refs := ParseReferences(notes)
	for _, ref := range refs {
		id := uuid.NewString()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO codex_references (id, entry_id, target_type, target_name)
			 VALUES (?, ?, ?, ?)`,
			id, entryID, ref.TargetType, ref.TargetName,
		); err != nil {
			return fmt.Errorf("insert reference: %w", err)
		}
	}

	return nil
}
