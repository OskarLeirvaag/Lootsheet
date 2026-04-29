package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ddb"
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
			Row:         fmt.Sprintf("%-5s %-12s %s", m.CR, truncateField(m.Type, colWidthMonsterType), m.Name),
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
			Row:         fmt.Sprintf("%-5s %-14s %s", lvl, truncateField(s.School, colWidthSpellSchool), s.Name),
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
			Row:         fmt.Sprintf("%-12s %-16s %s", truncateField(it.Rarity, colWidthItemRarity), truncateField(it.Type, colWidthItemType), it.Name),
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
			Row:         fmt.Sprintf("%-18s %s", truncateField(r.Category, colWidthRuleCategory), r.Name),
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
		status := "[ ]"
		if s.Enabled {
			status = "[x]"
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
	// strings.Builder.Write never errors; discard Fprintf return values.
	_, _ = fmt.Fprintf(&b, "## %s\n", m.Name)
	_, _ = fmt.Fprintf(&b, "*%s %s, CR %s*\n\n", m.Size, m.Type, m.CR)
	_, _ = fmt.Fprintf(&b, "**AC** %s  |  **HP** %s\n\n", m.AC, m.HP)
	if m.SourceName != "" {
		_, _ = fmt.Fprintf(&b, "Source: %s\n", m.SourceName)
	}
	return b.String()
}

func truncateField(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// Command constants for compendium operations.
const (
	tuiCommandCompendiumToggleSource = "compendium.toggle_source"
	tuiCommandCompendiumSync         = "compendium.sync"
)

// Column widths for list row formatting.
const (
	colWidthMonsterType  = 12
	colWidthSpellSchool  = 14
	colWidthItemRarity   = 12
	colWidthItemType     = 16
	colWidthRuleCategory = 18
)

// DDB ID offsets to avoid collisions when merging rules, actions, and weapon properties.
const (
	ddbIDOffsetBasicActions     = 10000
	ddbIDOffsetWeaponProperties = 20000
)

// DDB activation types for spell casting time.
const (
	ddbActivationAction      = 1
	ddbActivationBonusAction = 3
	ddbActivationReaction    = 4
	ddbActivationOneMinute   = 6
	ddbActivationTenMinutes  = 7
	ddbActivationOneHour     = 8
)

// --- DDB → domain converters ---

func convertDDBSources(sources []ddb.ConfigSource) []compendium.Source {
	result := make([]compendium.Source, len(sources))
	for i, s := range sources {
		result[i] = compendium.Source{ID: s.ID, Name: s.Description}
	}
	return result
}

func convertDDBConditions(entries []ddb.ConfigConditionEntry) []compendium.Condition {
	result := make([]compendium.Condition, len(entries))
	for i, e := range entries {
		result[i] = compendium.Condition{
			DdbID:       e.Definition.ID,
			Name:        e.Definition.Name,
			Description: ddb.HTMLToMarkdown(e.Definition.Description),
		}
	}
	return result
}

func convertDDBRules(cfg *ddb.Config) []compendium.Rule {
	result := make([]compendium.Rule, 0, len(cfg.Rules)+len(cfg.BasicActions)+len(cfg.WeaponProperties))

	for _, r := range cfg.Rules {
		result = append(result, compendium.Rule{
			DdbID:       r.ID,
			Name:        r.Name,
			Category:    "Rule",
			Description: ddb.HTMLToMarkdown(r.Description),
		})
	}
	for _, a := range cfg.BasicActions {
		result = append(result, compendium.Rule{
			DdbID:       ddbIDOffsetBasicActions + a.ID,
			Name:        a.Name,
			Category:    "Action",
			Description: ddb.HTMLToMarkdown(a.Description),
		})
	}
	for _, wp := range cfg.WeaponProperties {
		result = append(result, compendium.Rule{
			DdbID:       ddbIDOffsetWeaponProperties + wp.ID,
			Name:        wp.Name,
			Category:    "Weapon Property",
			Description: ddb.HTMLToMarkdown(wp.Description),
		})
	}

	return result
}

func convertDDBMonsters(monsters []ddb.RawMonster, cfg *ddb.Config) []compendium.Monster {
	result := make([]compendium.Monster, len(monsters))
	for i := range monsters {
		m := &monsters[i]
		sourceName := ""
		if len(m.Sources) > 0 {
			sourceName = cfg.SourceName(m.Sources[0].SourceID)
		}
		rawJSON := "{}"
		if m.RawJSON != nil {
			rawJSON = string(m.RawJSON)
		}
		result[i] = compendium.Monster{
			DdbID:      m.ID,
			Name:       m.Name,
			CR:         cfg.ChallengeRatingLabel(m.ChallengeRatingID),
			Type:       cfg.MonsterTypeName(m.TypeID),
			Size:       cfg.CreatureSizeName(m.SizeID),
			HP:         ddb.FormatMonsterHP(m),
			AC:         ddb.FormatMonsterAC(m),
			SourceName: sourceName,
			DetailJSON: rawJSON,
		}
	}
	return result
}

func convertDDBSpells(spells []ddb.RawSpellEntry, spellClasses map[int][]string, cfg *ddb.Config) []compendium.Spell {
	result := make([]compendium.Spell, len(spells))
	for i := range spells {
		def := &spells[i].Definition
		sourceName := ""
		if len(def.Sources) > 0 {
			sourceName = cfg.SourceName(def.Sources[0].SourceID)
		}
		rawJSON := "{}"
		if def.RawJSON != nil {
			rawJSON = string(def.RawJSON)
		}
		classes := ""
		if names, ok := spellClasses[def.ID]; ok {
			classes = strings.Join(names, ", ")
		}
		result[i] = compendium.Spell{
			DdbID:       def.ID,
			Name:        def.Name,
			Level:       def.Level,
			School:      def.School,
			CastingTime: formatActivation(def),
			Range:       ddb.FormatSpellRange(def),
			Components:  ddb.FormatSpellComponents(def),
			Duration:    ddb.FormatSpellDuration(def),
			Classes:     classes,
			SourceName:  sourceName,
			DetailJSON:  rawJSON,
		}
	}
	return result
}

func convertDDBItems(items []ddb.RawItem, cfg *ddb.Config) []compendium.Item {
	result := make([]compendium.Item, len(items))
	for i := range items {
		item := &items[i]
		sourceName := ""
		if len(item.Sources) > 0 {
			sourceName = cfg.SourceName(item.Sources[0].SourceID)
		}
		rawJSON := "{}"
		if item.RawJSON != nil {
			rawJSON = string(item.RawJSON)
		} else {
			// Fallback: marshal the parsed struct
			if b, err := json.Marshal(item); err == nil {
				rawJSON = string(b)
			}
		}
		result[i] = compendium.Item{
			DdbID:      item.ID,
			Name:       item.Name,
			Type:       ddb.ItemTypeName(item),
			Rarity:     item.Rarity,
			Attunement: item.CanAttune,
			SourceName: sourceName,
			DetailJSON: rawJSON,
		}
	}
	return result
}

func filterDDBSpellsBySource(spells []ddb.RawSpellEntry, sourceIDs []int) []ddb.RawSpellEntry {
	if len(sourceIDs) == 0 {
		return nil
	}
	enabled := sourceIDSet(sourceIDs)
	filtered := make([]ddb.RawSpellEntry, 0, len(spells))
	for i := range spells {
		if hasEnabledSource(spells[i].Definition.Sources, enabled) {
			filtered = append(filtered, spells[i])
		}
	}
	return filtered
}

func filterDDBItemsBySource(items []ddb.RawItem, sourceIDs []int) []ddb.RawItem {
	if len(sourceIDs) == 0 {
		return nil
	}
	enabled := sourceIDSet(sourceIDs)
	filtered := make([]ddb.RawItem, 0, len(items))
	for i := range items {
		if hasEnabledSource(items[i].Sources, enabled) {
			filtered = append(filtered, items[i])
		}
	}
	return filtered
}

func monsterDDBIDs(monsters []compendium.Monster) []int {
	ids := make([]int, len(monsters))
	for i := range monsters {
		ids[i] = monsters[i].DdbID
	}
	return ids
}

func spellDDBIDs(spells []compendium.Spell) []int {
	ids := make([]int, len(spells))
	for i := range spells {
		ids[i] = spells[i].DdbID
	}
	return ids
}

func itemDDBIDs(items []compendium.Item) []int {
	ids := make([]int, len(items))
	for i := range items {
		ids[i] = items[i].DdbID
	}
	return ids
}

func sourceIDSet(sourceIDs []int) map[int]struct{} {
	set := make(map[int]struct{}, len(sourceIDs))
	for _, id := range sourceIDs {
		set[id] = struct{}{}
	}
	return set
}

func hasEnabledSource(sources []ddb.SourceRef, enabled map[int]struct{}) bool {
	for _, source := range sources {
		if _, ok := enabled[source.SourceID]; ok {
			return true
		}
	}
	return false
}

func formatActivation(def *ddb.RawSpellDef) string {
	base := "1 action"
	switch def.Activation.ActivationType {
	case ddbActivationAction:
		// default
	case ddbActivationBonusAction:
		base = "1 bonus action"
	case ddbActivationReaction:
		base = "1 reaction"
	case ddbActivationOneMinute:
		base = "1 minute"
	case ddbActivationTenMinutes:
		base = "10 minutes"
	case ddbActivationOneHour:
		base = "1 hour"
	default:
		// unknown activation type, keep default
	}
	if def.Ritual {
		base += " (ritual)"
	}
	return base
}
