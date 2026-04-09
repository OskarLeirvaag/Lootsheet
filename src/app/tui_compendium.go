package app

import (
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/compendium"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func summarizeCompendiumMonsters(records []compendium.Monster) []string {
	return []string{fmt.Sprintf("Total monsters: %d", len(records))}
}

func summarizeCompendiumSpells(records []compendium.Spell) []string {
	return []string{fmt.Sprintf("Total spells: %d", len(records))}
}

func summarizeCompendiumItems(records []compendium.Item) []string {
	return []string{fmt.Sprintf("Total items: %d", len(records))}
}

func summarizeCompendiumRules(records []compendium.Rule) []string {
	return []string{fmt.Sprintf("Total rules: %d", len(records))}
}

func summarizeCompendiumConditions(records []compendium.Condition) []string {
	return []string{fmt.Sprintf("Total conditions: %d", len(records))}
}

func summarizeCompendiumSources(records []compendium.Source) []string {
	enabled := 0
	for _, s := range records {
		if s.Enabled {
			enabled++
		}
	}
	return []string{fmt.Sprintf("Sources: %d enabled / %d total", enabled, len(records))}
}

func buildCompendiumMonsterItems(records []compendium.Monster) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(records))
	for i := range records {
		m := &records[i]
		items = append(items, render.ListItemData{
			Key:         fmt.Sprintf("monster-%d", m.DdbID),
			Row:         fmt.Sprintf("%-5s %-12s %s", m.CR, truncateField(m.Type, 12), m.Name),
			DetailTitle: m.Name,
			DetailLines: []string{
				fmt.Sprintf("CR: %s  |  Type: %s  |  Size: %s", m.CR, m.Type, m.Size),
				fmt.Sprintf("HP: %s  |  AC: %s", m.HP, m.AC),
				fmt.Sprintf("Source: %s", m.SourceName),
			},
			DetailBody: buildMonsterDetailBody(m),
		})
	}
	return items
}

func buildCompendiumSpellItems(records []compendium.Spell) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(records))
	for i := range records {
		s := &records[i]
		lvl := "Cntrp"
		if s.Level > 0 {
			lvl = fmt.Sprintf("Lv %d", s.Level)
		}
		items = append(items, render.ListItemData{
			Key:         fmt.Sprintf("spell-%d", s.DdbID),
			Row:         fmt.Sprintf("%-5s %-14s %s", lvl, truncateField(s.School, 14), s.Name),
			DetailTitle: s.Name,
			DetailLines: []string{
				fmt.Sprintf("Level: %d  |  School: %s", s.Level, s.School),
				fmt.Sprintf("Casting Time: %s  |  Range: %s", s.CastingTime, s.Range),
				fmt.Sprintf("Components: %s  |  Duration: %s", s.Components, s.Duration),
				fmt.Sprintf("Classes: %s  |  Source: %s", s.Classes, s.SourceName),
			},
		})
	}
	return items
}

func buildCompendiumItemItems(records []compendium.Item) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(records))
	for i := range records {
		it := &records[i]
		attune := ""
		if it.Attunement {
			attune = " (attunement)"
		}
		items = append(items, render.ListItemData{
			Key:         fmt.Sprintf("item-%d", it.DdbID),
			Row:         fmt.Sprintf("%-12s %-16s %s", truncateField(it.Rarity, 12), truncateField(it.Type, 16), it.Name),
			DetailTitle: it.Name,
			DetailLines: []string{
				fmt.Sprintf("Type: %s  |  Rarity: %s%s", it.Type, it.Rarity, attune),
				fmt.Sprintf("Source: %s", it.SourceName),
			},
		})
	}
	return items
}

func buildCompendiumRuleItems(records []compendium.Rule) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(records))
	for i := range records {
		r := &records[i]
		items = append(items, render.ListItemData{
			Key:         fmt.Sprintf("rule-%d", r.DdbID),
			Row:         fmt.Sprintf("%-18s %s", truncateField(r.Category, 18), r.Name),
			DetailTitle: r.Name,
			DetailLines: []string{
				fmt.Sprintf("Category: %s", r.Category),
			},
			DetailBody: r.Description,
		})
	}
	return items
}

func buildCompendiumConditionItems(records []compendium.Condition) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(records))
	for i := range records {
		c := &records[i]
		items = append(items, render.ListItemData{
			Key:         fmt.Sprintf("condition-%d", c.DdbID),
			Row:         c.Name,
			DetailTitle: c.Name,
			DetailBody:  c.Description,
		})
	}
	return items
}

func buildCompendiumSourceItems(records []compendium.Source) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(records))
	for i := range records {
		s := &records[i]
		status := "  "
		if s.Enabled {
			status = "[x]"
		} else {
			status = "[ ]"
		}
		items = append(items, render.ListItemData{
			Key:         fmt.Sprintf("source-%d", s.ID),
			Row:         fmt.Sprintf("%s %s", status, s.Name),
			DetailTitle: s.Name,
			DetailLines: []string{
				fmt.Sprintf("ID: %d", s.ID),
				fmt.Sprintf("Enabled: %v", s.Enabled),
			},
			Actions: []render.ItemActionData{
				{
					Trigger:      render.ActionToggle,
					ID:           tuiCommandCompendiumToggleSource,
					Label:        "t toggle",
					Mode:         render.ItemActionModeConfirm,
					ConfirmTitle: fmt.Sprintf("Toggle %q?", s.Name),
					ConfirmLines: []string{"Enable or disable this source for compendium sync."},
				},
			},
		})
	}
	return items
}

func buildMonsterDetailBody(m *compendium.Monster) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s\n", m.Name))
	b.WriteString(fmt.Sprintf("*%s %s, CR %s*\n\n", m.Size, m.Type, m.CR))
	b.WriteString(fmt.Sprintf("**AC** %s  |  **HP** %s\n\n", m.AC, m.HP))
	if m.SourceName != "" {
		b.WriteString(fmt.Sprintf("Source: %s\n", m.SourceName))
	}
	return b.String()
}

func truncateField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Command constants for compendium operations.
const (
	tuiCommandCompendiumToggleSource = "compendium.toggle_source"
	tuiCommandCompendiumSync         = "compendium.sync"
)
