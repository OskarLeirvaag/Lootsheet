package render

import (
	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// Type aliases re-export model section types.
type Section = model.Section

const (
	SectionDashboard = model.SectionDashboard
	SectionLedger    = model.SectionLedger
	SectionJournal   = model.SectionJournal
	SectionQuests    = model.SectionQuests
	SectionLoot      = model.SectionLoot
	SectionAssets    = model.SectionAssets
	SectionCodex     = model.SectionCodex
	SectionNotes      = model.SectionNotes
	SectionCompendium = model.SectionCompendium
	SectionSettings   = model.SectionSettings
)

const (
	settingsTabAccounts   = model.SettingsTabAccounts
	settingsTabCodexTypes = model.SettingsTabCodexTypes
	settingsTabCampaigns   = model.SettingsTabCampaigns
	settingsTabCompendium  = model.SettingsTabCompendium

	compendiumTabMonsters   = model.CompendiumTabMonsters
	compendiumTabSpells     = model.CompendiumTabSpells
	compendiumTabItems      = model.CompendiumTabItems
	compendiumTabRules      = model.CompendiumTabRules
	compendiumTabConditions = model.CompendiumTabConditions

	// Exported for use by the app package (search handler).
	CompendiumTabMonsters   = model.CompendiumTabMonsters
	CompendiumTabSpells     = model.CompendiumTabSpells
	CompendiumTabItems      = model.CompendiumTabItems
	CompendiumTabRules      = model.CompendiumTabRules
	CompendiumTabConditions = model.CompendiumTabConditions
)

var (
	settingsTabs       = model.SettingsTabs
	compendiumTabs     = model.CompendiumTabs
	searchableSections = model.SearchableSections
	orderedSections    = model.OrderedSections
)

// SectionStyle bundles visual properties derived from a Section.
type SectionStyle struct {
	Accent        tcell.Style
	Texture       PanelTexture
	Borders       *BorderSet
	ScatterGlyphs []rune
	ScatterStyle  *tcell.Style
}

// sectionStyleFor returns the visual properties for a section under the given theme.
func sectionStyleFor(s Section, theme *Theme) SectionStyle {
	switch s {
	case SectionLedger:
		return SectionStyle{
			Accent:        theme.SectionLedger,
			ScatterGlyphs: scatterLedger,
			ScatterStyle:  &theme.ScatterLedger,
		}
	case SectionJournal:
		return SectionStyle{
			Accent:        theme.SectionJournal,
			ScatterGlyphs: scatterJournal,
			ScatterStyle:  &theme.ScatterJournal,
		}
	case SectionQuests:
		return SectionStyle{
			Accent:        theme.SectionQuests,
			ScatterGlyphs: scatterQuests,
			ScatterStyle:  &theme.ScatterQuests,
		}
	case SectionLoot:
		return SectionStyle{
			Accent:        theme.SectionLoot,
			ScatterGlyphs: scatterLoot,
			ScatterStyle:  &theme.ScatterLoot,
		}
	case SectionAssets:
		return SectionStyle{
			Accent:        theme.SectionAssets,
			Texture:       PanelTextureLeaf,
			Borders:       &runicBorders,
			ScatterGlyphs: scatterGlyphs,
			ScatterStyle:  &theme.ScatterAssets,
		}
	case SectionCodex:
		return SectionStyle{
			Accent:        theme.SectionCodex,
			ScatterGlyphs: scatterPeople,
			ScatterStyle:  &theme.ScatterCodex,
		}
	case SectionNotes:
		return SectionStyle{
			Accent:        theme.SectionNotes,
			ScatterGlyphs: scatterNotes,
			ScatterStyle:  &theme.ScatterNotes,
		}
	case SectionCompendium, compendiumTabMonsters, compendiumTabSpells, compendiumTabItems, compendiumTabRules, compendiumTabConditions:
		return SectionStyle{
			Accent:        theme.SectionCompendium,
			ScatterGlyphs: scatterCompendium,
			ScatterStyle:  &theme.ScatterCompendium,
		}
	case SectionSettings, settingsTabAccounts, settingsTabCodexTypes, settingsTabCampaigns, settingsTabCompendium:
		return SectionStyle{
			Accent:        theme.SectionSettings,
			ScatterGlyphs: scatterSettings,
			ScatterStyle:  &theme.ScatterSettings,
		}
	default:
		return SectionStyle{
			Accent: theme.SectionDashboard,
		}
	}
}

// Panel returns a Panel pre-filled with section visual properties.
func (ss *SectionStyle) Panel(title string, lines []string) Panel {
	return Panel{
		Title:         title,
		Lines:         lines,
		BorderStyle:   &ss.Accent,
		TitleStyle:    &ss.Accent,
		Texture:       ss.Texture,
		Borders:       ss.Borders,
		ScatterGlyphs: ss.ScatterGlyphs,
		ScatterStyle:  ss.ScatterStyle,
	}
}
