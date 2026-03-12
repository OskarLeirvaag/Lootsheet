package app

import (
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/notes"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func summarizeNotes(records []notes.NoteRecord) []string {
	return []string{
		fmt.Sprintf("Total notes: %d", len(records)),
	}
}

func buildNotesItems(records []notes.NoteRecord, refs map[string][]notes.ReferenceRecord) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(records))
	for index := range records {
		n := &records[index]
		updated := n.UpdatedAt
		if len(updated) > 10 {
			updated = updated[:10]
		}

		detailLines := []string{
			"Updated: " + n.UpdatedAt,
		}

		// Show parsed references.
		if noteRefs, ok := refs[n.ID]; ok && len(noteRefs) > 0 {
			detailLines = append(detailLines, "", "References:")
			for _, ref := range noteRefs {
				detailLines = append(detailLines, fmt.Sprintf("  @%s/%s", ref.TargetType, ref.TargetName))
			}
		}

		actions := []render.ItemActionData{
			{
				Trigger:      render.ActionEdit,
				ID:           tuiCommandNotesUpdate,
				Label:        "u edit",
				Mode:         render.ItemActionModeCompose,
				ComposeMode:  "notes",
				ComposeTitle: "Edit Note",
				ComposeFields: map[string]string{
					"title": n.Title,
					"body":  n.Body,
				},
			},
			{
				Trigger:      render.ActionDelete,
				ID:           tuiCommandNotesDelete,
				Label:        "d delete",
				Mode:         render.ItemActionModeConfirm,
				ConfirmTitle: fmt.Sprintf("Delete %q?", n.Title),
				ConfirmLines: []string{
					"This will permanently remove this note and its references.",
				},
			},
		}

		items = append(items, render.ListItemData{
			Key:         n.ID,
			Row:         fmt.Sprintf("%-11s %s", updated, n.Title),
			DetailTitle: n.Title,
			DetailLines: detailLines,
			DetailBody:  n.Body,
			Actions:     actions,
		})
	}

	return items
}

// buildNoteReferencedInLines returns "Referenced in:" detail lines for an entity
// that is mentioned in notes.
func buildNoteReferencedInLines(allRefs map[string][]notes.ReferenceRecord, allNotes []notes.NoteRecord, targetType, targetName string) []string {
	notesByID := make(map[string]*notes.NoteRecord, len(allNotes))
	for i := range allNotes {
		notesByID[allNotes[i].ID] = &allNotes[i]
	}

	var mentioners []string
	for noteID, refs := range allRefs {
		for _, ref := range refs {
			if ref.TargetType == targetType && strings.EqualFold(ref.TargetName, targetName) {
				if n, ok := notesByID[noteID]; ok {
					mentioners = append(mentioners, n.Title)
				}
				break
			}
		}
	}

	if len(mentioners) == 0 {
		return nil
	}

	return []string{"Referenced in: " + strings.Join(mentioners, ", ")}
}
