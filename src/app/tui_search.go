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
		if items, handled, err := searchCompendiumIfApplicable(ctx, loader, section, query); handled {
			return items, err
		}

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

func searchCompendium(ctx context.Context, loader TUIDataLoader, section render.Section, query string) ([]render.ListItemData, error) {
	items, handled, err := searchCompendiumIfApplicable(ctx, loader, section, query)
	if !handled {
		return nil, nil
	}
	return items, err
}

func searchCompendiumIfApplicable(ctx context.Context, loader TUIDataLoader, section render.Section, query string) ([]render.ListItemData, bool, error) {
	switch section {
	case render.CompendiumTabMonsters:
		records, err := loader.ListCompendiumMonsters(ctx, query)
		if err != nil {
			return nil, true, err
		}
		return buildCompendiumMonsterItems(records), true, nil

	case render.CompendiumTabSpells:
		records, err := loader.ListCompendiumSpells(ctx, query)
		if err != nil {
			return nil, true, err
		}
		return buildCompendiumSpellItems(records), true, nil

	case render.CompendiumTabItems:
		records, err := loader.ListCompendiumItems(ctx, query)
		if err != nil {
			return nil, true, err
		}
		return buildCompendiumItemItems(records), true, nil

	case render.CompendiumTabRules:
		records, err := loader.ListCompendiumRules(ctx, query)
		if err != nil {
			return nil, true, err
		}
		return buildCompendiumRuleItems(records), true, nil

	case render.CompendiumTabConditions:
		records, err := loader.ListCompendiumConditions(ctx, query)
		if err != nil {
			return nil, true, err
		}
		return buildCompendiumConditionItems(records), true, nil

	default:
		return nil, false, nil
	}
}
