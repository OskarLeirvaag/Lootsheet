package model

// Section identifies a top-level TUI screen.
type Section int

const (
	SectionDashboard Section = iota
	SectionLedger
	SectionJournal
	SectionQuests
	SectionLoot
	SectionAssets
	SectionCodex
	SectionNotes
	SectionSettings
)

// Virtual sections for settings tabs — not in OrderedSections.
const (
	SettingsTabAccounts Section = 100 + iota
	SettingsTabCodexTypes
	SettingsTabCampaigns
)

// SettingsTabs lists the virtual settings sections.
var SettingsTabs = []Section{SettingsTabAccounts, SettingsTabCodexTypes, SettingsTabCampaigns}

// SearchableSections lists sections that appear in the search modal.
var SearchableSections = []Section{
	SectionJournal, SectionQuests, SectionLoot,
	SectionAssets, SectionCodex, SectionNotes,
}

// OrderedSections lists sections in tab order.
var OrderedSections = []Section{
	SectionDashboard,
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
	case SectionLedger:
		return "Ledger"
	case SettingsTabAccounts:
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
	case SectionSettings:
		return "Settings"
	case SettingsTabCodexTypes:
		return "Codex Types"
	case SettingsTabCampaigns:
		return "Campaigns"
	default:
		return "Dashboard"
	}
}

// Next returns the following section in tab order.
func (s Section) Next() Section {
	for index, current := range OrderedSections {
		if current == s {
			return OrderedSections[(index+1)%len(OrderedSections)]
		}
	}
	return SectionDashboard
}

// Previous returns the preceding section in tab order.
func (s Section) Previous() Section {
	for index, current := range OrderedSections {
		if current == s {
			return OrderedSections[(index+len(OrderedSections)-1)%len(OrderedSections)]
		}
	}
	return SectionDashboard
}

// Scrollable reports whether this section has a scrollable list.
func (s Section) Scrollable() bool {
	return s != SectionDashboard
}

// MaxSectionTitleWidth returns the longest section title length.
func MaxSectionTitleWidth() int {
	width := 0
	for _, section := range OrderedSections {
		titleWidth := len(section.Title())
		if titleWidth > width {
			width = titleWidth
		}
	}
	return width
}
