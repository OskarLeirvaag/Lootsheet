package app

import (
	"context"

	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

// buildSearchHandler returns a SearchHandler that delegates to SQL-based search
// for sections that support it (Codex, Notes) and returns nil for all others,
// causing the render layer to fall back to client-side filtering.
func buildSearchHandler(ctx context.Context, loader TUIDataLoader) render.SearchHandler {
	return func(section render.Section, query string) ([]render.ListItemData, error) {
		switch section {
		case render.SectionCodex:
			entries, err := loader.SearchCodexEntries(ctx, query)
			if err != nil {
				return nil, err
			}
			return buildCodexItems(entries, nil), nil

		case render.SectionNotes:
			records, err := loader.SearchNotes(ctx, query)
			if err != nil {
				return nil, err
			}
			return buildNotesItems(records, nil), nil

		default:
			return nil, nil
		}
	}
}
