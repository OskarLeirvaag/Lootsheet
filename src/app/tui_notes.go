package app

import (
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/notes"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func summarizeNotes(records []notes.NoteRecord) []string {
	return []string{
		fmt.Sprintf("Total notes: %d", len(records)),
	}
}

func buildNotesItems(records []notes.NoteRecord, noteRefs map[string][]refs.EntityReference) []render.ListItemData {
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
		if nRefs, ok := noteRefs[n.ID]; ok && len(nRefs) > 0 {
			detailLines = append(detailLines, "", "References:")
			for _, ref := range nRefs {
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
