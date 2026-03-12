package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (s *Shell) footerHelpText(keymap KeyMap) string {
	if s.editor != nil {
		return ":w save  :q quit  :wq save+quit  i insert  Esc normal/close  Tab title/body"
	}
	if s.input != nil {
		return "Enter submit  Backspace delete  Ctrl+U clear  Esc cancel  q cancel"
	}
	if s.glossary != nil {
		return "? close  Esc cancel  q cancel"
	}
	if s.search != nil {
		return "Enter select  \u2191\u2193 navigate  \u2190\u2192 filter  Ctrl+U clear  Esc close"
	}
	if s.compose != nil {
		return s.composeHelpText()
	}

	if s.confirm != nil {
		return "Enter confirm  Esc cancel  q cancel"
	}

	help := keymap.HelpTextFor(ActionNextSection, ActionShowDashboard)
	if s.Section.scrollable() {
		help = joinHelp(help, keymap.HelpTextFor(ActionMoveDown))
	}

	if labels := s.currentActionLabels(); labels != "" {
		help = joinHelp(help, labels)
	}

	help = joinHelp(help, s.sectionLauncherHelpText())
	help = joinHelp(help, "/ search", "? terms", "q quit", "Ctrl+L refresh")

	return help
}

const helpAddEdit = "a add  u edit"

func (s *Shell) sectionLauncherHelpText() string {
	switch s.Section {
	case SectionAccounts:
		return "a add"
	case SectionSettings:
		if s.activeSettingsSection() == settingsTabCodexTypes {
			return "a add codex type"
		}
		return "a add account"
	case SectionJournal:
		return "e/i entry"
	case SectionQuests, SectionLoot, SectionAssets, SectionCodex, SectionNotes:
		return helpAddEdit
	default:
		return "e/i/a entry"
	}
}

func (s *Shell) currentActionLabels() string {
	item := s.currentSelectedItem(s.listSection())
	if item == nil || len(item.Actions) == 0 {
		return ""
	}

	labels := make([]string, 0, len(item.Actions))
	for index := range item.Actions {
		action := item.Actions[index]
		if strings.TrimSpace(action.Label) == "" {
			continue
		}
		labels = append(labels, action.Label)
	}

	return strings.Join(labels, "  ")
}

func (s *Shell) headerLines() []string {
	source := s.currentHeaderLines()
	lines := make([]string, 0, len(source)+2)
	lines = append(lines, fmt.Sprintf("Section: %s", s.Section.Title()))
	lines = append(lines, source...)
	lines = append(lines, "Sections: "+s.tabsLine())

	return lines
}

func (s *Shell) currentHeaderLines() []string {
	switch s.Section {
	case SectionAccounts:
		return append([]string{}, s.Data.Accounts.HeaderLines...)
	case SectionJournal:
		return append([]string{}, s.Data.Journal.HeaderLines...)
	case SectionQuests:
		return append([]string{}, s.Data.Quests.HeaderLines...)
	case SectionLoot:
		return append([]string{}, s.Data.Loot.HeaderLines...)
	case SectionAssets:
		return append([]string{}, s.Data.Assets.HeaderLines...)
	case SectionCodex:
		return append([]string{}, s.Data.Codex.HeaderLines...)
	case SectionNotes:
		return append([]string{}, s.Data.Notes.HeaderLines...)
	case SectionSettings:
		data := s.listDataForSection(s.activeSettingsSection())
		if data != nil {
			return append([]string{}, data.HeaderLines...)
		}
		return nil
	default:
		return append([]string{}, resolveDashboardData(&s.Data.Dashboard).HeaderLines...)
	}
}

func (s *Shell) tabsLine() string {
	width := maxSectionTitleWidth() + 2
	var line strings.Builder
	for index, section := range orderedSections {
		if index > 0 {
			line.WriteString("  ")
		}

		label := section.Title()
		if section == s.Section {
			label = "[" + label + "]"
		} else {
			label = " " + label + " "
		}
		fmt.Fprintf(&line, "%-*s", width, label)
	}

	return line.String()
}

func (s *Shell) drawHeaderHighlights(buffer *Buffer, rect Rect, theme *Theme) {
	if buffer == nil || theme == nil {
		return
	}

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	sectionLabel := "Section: "
	buffer.WriteString(content.X, content.Y, theme.HeaderLabel, sectionLabel)
	buffer.WriteString(content.X+len([]rune(sectionLabel)), content.Y, s.sectionStyle(theme), s.Section.Title())

	if content.H < 2 {
		return
	}

	prefix := "Sections: "
	tabY := content.Y + content.H - 1
	buffer.WriteString(content.X, tabY, theme.HeaderLabel, prefix)
	x := content.X + len([]rune(prefix))
	for index, section := range orderedSections {
		if index > 0 {
			x += buffer.WriteString(x, tabY, theme.Muted, "  ")
		}

		label := section.Title()
		style := theme.TabInactive
		if section == s.Section {
			label = "[" + label + "]"
			style = section.Style(theme).Accent
		} else {
			label = " " + label + " "
		}
		width := maxSectionTitleWidth() + 2
		x += buffer.WriteString(x, tabY, style, fmt.Sprintf("%-*s", width, label))
	}
}

func (s *Shell) sectionStyle(theme *Theme) tcell.Style {
	return s.Section.Style(theme).Accent
}

func (s *Shell) glossaryTitle() string {
	return s.Section.Title() + " Terms"
}

func (s *Shell) glossaryLines() []string {
	switch s.Section {
	case SectionAccounts:
		return []string{
			"Assets: what the party owns or is owed.",
			"Liabilities: what the party owes to others.",
			"Equity: the party's accumulated net worth.",
			"Income: rewards, fees, and gains earned by the party.",
			"Expenses: costs like supplies, rations, inns, and repairs.",
			"Active: usable in new entries.",
			"Inactive: kept for history, blocked for new entries.",
			"Postings: journal lines already recorded against an account.",
		}
	case SectionJournal:
		return []string{
			"Debit / Credit: the two sides of every balanced entry.",
			"Posted: final and immutable.",
			"Reversed: corrected by a later entry instead of editing history.",
			"Reversal: a new entry that cancels an earlier posted entry.",
			"Description: the plain-language reason for the entry.",
		}
	case SectionQuests:
		return []string{
			"Off-ledger promise: discussed reward, not earned yet.",
			"Collectible: earned and ready to be paid.",
			"Receivable: reward still owed to the party.",
			"Collected so far: cash already received against the reward.",
			"Write off: accept that the remaining reward will not be collected.",
		}
	case SectionLoot:
		return []string{
			"Held: physically owned, but not yet on the books.",
			"Appraisal: estimated value of a loot item.",
			"Recognize: move an appraisal onto the ledger as inventory and gain.",
			"Recognized value: the inventory basis currently on the books.",
			"Sell: turn recognized loot into cash and record any gain or loss.",
		}
	case SectionAssets:
		return []string{
			"Asset: high-value item the party intends to keep.",
			"Transfer: move an asset to the loot register for sale, or vice versa.",
			"Appraisal: estimated value, shared with the loot system.",
			"Recognize: move an appraisal onto the ledger as inventory and gain.",
		}
	case SectionCodex:
		return []string{
			"Codex: a D&D-flavored encyclopedia of people and entities.",
			"Player: a party member with class, race, and background.",
			"NPC: a non-player character with title, location, and disposition.",
			"@type/name: inline cross-reference in notes.",
			"References: parsed @mentions linking to quests, loot, assets, or other people.",
		}
	case SectionNotes:
		return []string{
			"Note: a general-purpose campaign or session note.",
			"Title: the note's heading, shown in the list.",
			"Body: free-form text content of the note.",
			"@type/name: inline cross-reference in body text.",
			"References: parsed @mentions linking to quests, loot, assets, people, or other notes.",
		}
	case SectionSettings:
		if s.activeSettingsSection() == settingsTabCodexTypes {
			return []string{
				"Codex type: a category for codex entries (e.g. NPC, Player, Settlement).",
				"Form template: the set of fields shown when creating a codex entry of this type.",
				"Types with existing entries cannot be deleted.",
			}
		}
		return []string{
			"Account: a ledger account (asset, liability, equity, income, expense).",
			"Active: usable in new entries.",
			"Inactive: kept for history, blocked for new entries.",
			"Accounts with postings cannot be deleted.",
		}
	default:
		return []string{
			"To share now: current Party Cash balance available to split.",
			"Unsold loot: recognized inventory not yet sold for cash.",
			"Off-ledger: tracked in a register, but not yet in the ledger.",
			"Register: operational tracking that may later create entries.",
			"Ledger: the formal bookkeeping record.",
		}
	}
}

func joinHelp(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		filtered = append(filtered, part)
	}

	return strings.Join(filtered, "  ")
}
