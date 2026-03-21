package app

import (
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
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

func buildCodexItems(entries []codex.CodexEntry, entryRefs map[string][]refs.EntityReference) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(entries))
	for index := range entries {
		e := &entries[index]

		secondary := codexSecondary(e)
		detailLines := buildCodexDetailLines(e, entryRefs)

		// Build compose fields for editing — include _form_id and _type_id
		// so the compose system can pick the correct form.
		composeFields := map[string]string{
			"_form_id":    codexFormIDForType(e.TypeID),
			"_type_id":    e.TypeID,
			"name":        e.Name,
			"player_name": e.PlayerName,
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

func buildCodexDetailLines(e *codex.CodexEntry, entryRefs map[string][]refs.EntityReference) []string {
	detailLines := []string{
		"Name: " + e.Name,
		"Type: " + e.TypeName,
	}

	for _, pair := range []struct{ label, value string }{
		{"Title", e.Title},
		{"Player", e.PlayerName},
		{"Class", e.Class},
		{"Race", e.Race},
		{"Background", e.Background},
		{"Location", e.Location},
		{"Faction", e.Faction},
		{"Disposition", e.Disposition},
		{"Description", e.Description},
		{"Notes", e.Notes},
	} {
		if strings.TrimSpace(pair.value) != "" {
			detailLines = append(detailLines, pair.label+": "+pair.value)
		}
	}

	// Show parsed references.
	if eRefs, ok := entryRefs[e.ID]; ok && len(eRefs) > 0 {
		detailLines = append(detailLines, "", "References:")
		for _, ref := range eRefs {
			detailLines = append(detailLines, fmt.Sprintf("  @%s/%s", ref.TargetType, ref.TargetName))
		}
	}

	return detailLines
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

// buildLinkedFromLines returns "Linked from:" detail lines for an entity
// using the unified entity_references indexed by target key.
func buildLinkedFromLines(allRefsByTarget map[string][]refs.EntityReference, targetType, targetName string) []string {
	key := targetType + ":" + strings.ToLower(targetName)
	targetRefs, ok := allRefsByTarget[key]
	if !ok || len(targetRefs) == 0 {
		return nil
	}

	var names []string
	for _, ref := range targetRefs {
		if ref.SourceName != "" {
			names = append(names, ref.SourceName)
		}
	}

	if len(names) == 0 {
		return nil
	}

	return []string{"", "Linked from: " + strings.Join(names, ", ")}
}
