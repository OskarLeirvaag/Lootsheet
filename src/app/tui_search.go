package app

import (
	"context"

	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

// buildSearchHandler returns a SearchHandler that delegates to SQL-based search
// for sections that support it (Codex, Notes) and returns nil for all others,
// causing the render layer to fall back to client-side filtering.
//
// The captured ctx lives for the entire TUI session. Per-query cancellation or
// timeouts should be added here when the data source moves to a remote server.
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

		case render.CompendiumTabMonsters:
			records, err := loader.ListCompendiumMonsters(ctx, query)
			if err != nil {
				return nil, err
			}
			return buildCompendiumMonsterItems(records), nil

		case render.CompendiumTabSpells:
			records, err := loader.ListCompendiumSpells(ctx, query)
			if err != nil {
				return nil, err
			}
			return buildCompendiumSpellItems(records), nil

		case render.CompendiumTabItems:
			records, err := loader.ListCompendiumItems(ctx, query)
			if err != nil {
				return nil, err
			}
			return buildCompendiumItemItems(records), nil

		case render.CompendiumTabRules:
			records, err := loader.ListCompendiumRules(ctx, query)
			if err != nil {
				return nil, err
			}
			return buildCompendiumRuleItems(records), nil

		case render.CompendiumTabConditions:
			records, err := loader.ListCompendiumConditions(ctx, query)
			if err != nil {
				return nil, err
			}
			return buildCompendiumConditionItems(records), nil

		default:
			return nil, nil
		}
	}
}
