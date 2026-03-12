package app

import (
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func summarizeCodex(entries []codex.CodexEntry) []string {
	typeCounts := make(map[string]int)
	for i := range entries {
		typeCounts[entries[i].TypeName]++
	}

	lines := make([]string, 0, 1+len(typeCounts))
	lines = append(lines, fmt.Sprintf("Codex entries: %d", len(entries)))
	for typeName, count := range typeCounts {
		lines = append(lines, fmt.Sprintf("  %s: %d", typeName, count))
	}

	return lines
}

func buildCodexItems(entries []codex.CodexEntry, refs map[string][]codex.Reference) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(entries))
	for index := range entries {
		e := &entries[index]

		secondary := codexSecondary(e)

		detailLines := []string{
			"Name: " + e.Name,
			"Type: " + e.TypeName,
		}
		if strings.TrimSpace(e.Title) != "" {
			detailLines = append(detailLines, "Title: "+e.Title)
		}
		if strings.TrimSpace(e.Class) != "" {
			detailLines = append(detailLines, "Class: "+e.Class)
		}
		if strings.TrimSpace(e.Race) != "" {
			detailLines = append(detailLines, "Race: "+e.Race)
		}
		if strings.TrimSpace(e.Background) != "" {
			detailLines = append(detailLines, "Background: "+e.Background)
		}
		if strings.TrimSpace(e.Location) != "" {
			detailLines = append(detailLines, "Location: "+e.Location)
		}
		if strings.TrimSpace(e.Faction) != "" {
			detailLines = append(detailLines, "Faction: "+e.Faction)
		}
		if strings.TrimSpace(e.Disposition) != "" {
			detailLines = append(detailLines, "Disposition: "+e.Disposition)
		}
		if strings.TrimSpace(e.Description) != "" {
			detailLines = append(detailLines, "Description: "+e.Description)
		}
		if strings.TrimSpace(e.Notes) != "" {
			detailLines = append(detailLines, "Notes: "+e.Notes)
		}

		// Show parsed references.
		if entryRefs, ok := refs[e.ID]; ok && len(entryRefs) > 0 {
			detailLines = append(detailLines, "", "References:")
			for _, ref := range entryRefs {
				detailLines = append(detailLines, fmt.Sprintf("  @%s/%s", ref.TargetType, ref.TargetName))
			}
		}

		// Build compose fields for editing — include _form_id and _type_id
		// so the compose system can pick the correct form.
		composeFields := map[string]string{
			"_form_id":    codexFormIDForType(e.TypeID),
			"_type_id":    e.TypeID,
			"name":        e.Name,
			"title":       e.Title,
			"location":    e.Location,
			"faction":     e.Faction,
			"disposition": e.Disposition,
			"class":       e.Class,
			"race":        e.Race,
			"background":  e.Background,
			"description": e.Description,
			"notes":       e.Notes,
		}

		actions := []render.ItemActionData{
			{
				Trigger:       render.ActionEdit,
				ID:            tuiCommandCodexUpdate,
				Label:         "u edit",
				Mode:          render.ItemActionModeCompose,
				ComposeMode:   "codex",
				ComposeTitle:  "Edit " + e.TypeName,
				ComposeFields: composeFields,
			},
			{
				Trigger:      render.ActionDelete,
				ID:           tuiCommandCodexDelete,
				Label:        "d delete",
				Mode:         render.ItemActionModeConfirm,
				ConfirmTitle: fmt.Sprintf("Delete %q?", e.Name),
				ConfirmLines: []string{
					"This will permanently remove this codex entry and its references.",
				},
			},
		}

		items = append(items, render.ListItemData{
			Key:         e.ID,
			Row:         fmt.Sprintf("%-8s %-14s %s", e.TypeName, secondary, e.Name),
			DetailTitle: e.Name,
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

// codexSecondary returns the type-specific secondary field for list display.
func codexSecondary(e *codex.CodexEntry) string {
	var s string
	switch e.TypeID {
	case "player":
		s = e.Class
	case "settlement":
		s = e.Location
	default:
		s = e.Disposition
	}
	if strings.TrimSpace(s) == "" {
		s = "-"
	}
	return s
}

// codexFormIDForType maps a type ID to its form ID.
func codexFormIDForType(typeID string) string {
	switch typeID {
	case "player":
		return "player"
	case "settlement":
		return "settlement"
	default:
		return "npc"
	}
}

// buildMentionedByLines returns "Mentioned by:" detail lines for an entity.
func buildMentionedByLines(allRefs map[string][]codex.Reference, entries []codex.CodexEntry, targetType, targetName string) []string {
	entriesByID := make(map[string]*codex.CodexEntry, len(entries))
	for i := range entries {
		entriesByID[entries[i].ID] = &entries[i]
	}

	var mentioners []string
	for entryID, refs := range allRefs {
		for _, ref := range refs {
			if ref.TargetType == targetType && strings.EqualFold(ref.TargetName, targetName) {
				if e, ok := entriesByID[entryID]; ok {
					mentioners = append(mentioners, e.Name)
				}
				break
			}
		}
	}

	if len(mentioners) == 0 {
		return nil
	}

	return []string{"", "Mentioned by: " + strings.Join(mentioners, ", ")}
}
