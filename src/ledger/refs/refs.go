// Package refs provides shared cross-reference parsing and storage for all
// entity types (codex, notes, quests, loot).
package refs

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// ParsedRef represents a single @type/name mention extracted from text.
type ParsedRef struct {
	TargetType string
	TargetName string
}

// EntityReference represents a row in the entity_references table.
type EntityReference struct {
	ID         string
	SourceType string
	SourceID   string
	SourceName string
	TargetType string
	TargetName string
	CreatedAt  string
}

var refPattern = regexp.MustCompile(`@\[(quest|loot|asset|person|note)/([^\]]+)\]`)

// ParseReferences extracts @type/name pairs from text.
func ParseReferences(text string) []ParsedRef {
	matches := refPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	refs := make([]ParsedRef, 0, len(matches))
	for _, match := range matches {
		targetType := strings.TrimSpace(match[1])
		targetName := strings.TrimSpace(match[2])
		// Trim trailing punctuation that was captured as part of the name.
		targetName = strings.TrimRight(targetName, ".,;:!?")
		targetName = strings.TrimSpace(targetName)
		if targetType == "" || targetName == "" {
			continue
		}
		refs = append(refs, ParsedRef{
			TargetType: targetType,
			TargetName: targetName,
		})
	}

	return refs
}

// RebuildReferences deletes old references for a source entity and inserts new
// ones parsed from the text. Must be called within a transaction or single
// connection context.
func RebuildReferences(ctx context.Context, db *sql.DB, sourceType, sourceID, campaignID, sourceName, text string) error {
	if _, err := db.ExecContext(ctx,
		"DELETE FROM entity_references WHERE source_type = ? AND source_id = ?", sourceType, sourceID,
	); err != nil {
		return fmt.Errorf("delete old references: %w", err)
	}

	refs := ParseReferences(text)
	for _, ref := range refs {
		id := uuid.NewString()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO entity_references (id, campaign_id, source_type, source_id, source_name, target_type, target_name)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			id, campaignID, sourceType, sourceID, sourceName, ref.TargetType, ref.TargetName,
		); err != nil {
			return fmt.Errorf("insert reference: %w", err)
		}
	}

	return nil
}

// ListAllByTarget returns all entity references indexed by "target_type:lower(target_name)".
func ListAllByTarget(ctx context.Context, db *sql.DB, campaignID string) (map[string][]EntityReference, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, source_type, source_id, source_name, target_type, target_name, created_at
		 FROM entity_references WHERE campaign_id = ? ORDER BY created_at`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("query entity references: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]EntityReference)
	for rows.Next() {
		var r EntityReference
		if err := rows.Scan(&r.ID, &r.SourceType, &r.SourceID, &r.SourceName, &r.TargetType, &r.TargetName, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan entity reference row: %w", err)
		}
		key := r.TargetType + ":" + strings.ToLower(r.TargetName)
		result[key] = append(result[key], r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity reference rows: %w", err)
	}

	return result, nil
}

// ListBySource returns all entity references for a given source type, indexed by source_id.
func ListBySource(ctx context.Context, db *sql.DB, sourceType string, campaignID string) (map[string][]EntityReference, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, source_type, source_id, source_name, target_type, target_name, created_at
		 FROM entity_references WHERE source_type = ? AND campaign_id = ? ORDER BY created_at`, sourceType, campaignID)
	if err != nil {
		return nil, fmt.Errorf("query entity references: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]EntityReference)
	for rows.Next() {
		var r EntityReference
		if err := rows.Scan(&r.ID, &r.SourceType, &r.SourceID, &r.SourceName, &r.TargetType, &r.TargetName, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan entity reference row: %w", err)
		}
		result[r.SourceID] = append(result[r.SourceID], r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity reference rows: %w", err)
	}

	return result, nil
}

// DeleteBySource removes all references for a given source entity.
func DeleteBySource(ctx context.Context, db *sql.DB, sourceType, sourceID string) error {
	if _, err := db.ExecContext(ctx,
		"DELETE FROM entity_references WHERE source_type = ? AND source_id = ?", sourceType, sourceID,
	); err != nil {
		return fmt.Errorf("delete references for source: %w", err)
	}
	return nil
}
