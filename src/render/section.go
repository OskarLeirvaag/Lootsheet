package render

// Section identifies a top-level TUI screen.
type Section int

const (
	SectionDashboard Section = iota
	SectionAccounts
	SectionJournal
	SectionQuests
	SectionLoot
	SectionAssets
)

var orderedSections = []Section{
	SectionDashboard,
	SectionAccounts,
	SectionJournal,
	SectionQuests,
	SectionLoot,
	SectionAssets,
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
