package render

import (
	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// Type aliases re-export model section types.
type Section = model.Section

const (
	SectionDashboard = model.SectionDashboard
	SectionAccounts  = model.SectionAccounts
	SectionJournal   = model.SectionJournal
	SectionQuests    = model.SectionQuests
	SectionLoot      = model.SectionLoot
	SectionAssets    = model.SectionAssets
	SectionCodex     = model.SectionCodex
	SectionNotes     = model.SectionNotes
	SectionSettings  = model.SectionSettings
)

const (
	settingsTabAccounts   = model.SettingsTabAccounts
	settingsTabCodexTypes = model.SettingsTabCodexTypes
)

var (
	settingsTabs       = model.SettingsTabs
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
	case SectionAccounts:
		return SectionStyle{
			Accent:        theme.SectionAccounts,
			ScatterGlyphs: scatterAccounts,
			ScatterStyle:  &theme.ScatterAccounts,
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
	case SectionSettings, settingsTabAccounts, settingsTabCodexTypes:
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
