package render

import "github.com/gdamore/tcell/v2"

// Section identifies a top-level TUI screen.
type Section int

const (
	SectionDashboard Section = iota
	SectionAccounts
	SectionJournal
	SectionQuests
	SectionLoot
	SectionAssets
	SectionCodex
	SectionNotes
)

var orderedSections = []Section{
	SectionDashboard,
	SectionAccounts,
	SectionJournal,
	SectionQuests,
	SectionLoot,
	SectionAssets,
	SectionCodex,
	SectionNotes,
}

// Title returns the user-facing section name.
func (s Section) Title() string {
	switch s {
	case SectionAccounts:
		return "Accounts"
	case SectionJournal:
		return "Journal"
	case SectionQuests:
		return "Quests"
	case SectionLoot:
		return "Loot"
	case SectionAssets:
		return "Assets"
	case SectionCodex:
		return "Codex"
	case SectionNotes:
		return "Notes"
	default:
		return "Dashboard"
	}
}

func (s Section) next() Section {
	for index, current := range orderedSections {
		if current == s {
			return orderedSections[(index+1)%len(orderedSections)]
		}
	}
	return SectionDashboard
}

func (s Section) previous() Section {
	for index, current := range orderedSections {
		if current == s {
			return orderedSections[(index+len(orderedSections)-1)%len(orderedSections)]
		}
	}
	return SectionDashboard
}

func (s Section) scrollable() bool {
	return s != SectionDashboard
}

func maxSectionTitleWidth() int {
	width := 0
	for _, section := range orderedSections {
		titleWidth := len(section.Title())
		if titleWidth > width {
			width = titleWidth
		}
	}
	return width
}

// SectionStyle bundles visual properties derived from a Section.
type SectionStyle struct {
	Accent        tcell.Style
	Texture       PanelTexture
	Borders       *BorderSet
	ScatterGlyphs []rune
	ScatterStyle  *tcell.Style
}

// Style returns the visual properties for this section under the given theme.
func (s Section) Style(theme *Theme) SectionStyle {
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
