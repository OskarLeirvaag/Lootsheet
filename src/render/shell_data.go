package render

// ListScreenData is the neutral read-only model for list-style TUI sections.
type ListScreenData struct {
	HeaderLines  []string
	SummaryLines []string
	RowLines     []string
	EmptyLines   []string
}

// ShellData contains the full read-only TUI snapshot.
type ShellData struct {
	Dashboard DashboardData
	Accounts  ListScreenData
	Journal   ListScreenData
	Quests    ListScreenData
	Loot      ListScreenData
}

// DefaultShellData returns placeholder content for the full read-only shell.
func DefaultShellData() ShellData {
	return ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts:  defaultListScreenData(SectionAccounts),
		Journal:   defaultListScreenData(SectionJournal),
		Quests:    defaultListScreenData(SectionQuests),
		Loot:      defaultListScreenData(SectionLoot),
	}
}

// ErrorShellData keeps the TUI open while surfacing a loader failure.
func ErrorShellData(summary string, detail string) ShellData {
	if summary == "" {
		summary = "TUI data unavailable."
	}
	if detail == "" {
		detail = "No additional detail."
	}

	return ShellData{
		Dashboard: ErrorDashboardData(summary, detail),
		Accounts: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Account data unavailable.", detail},
			EmptyLines:   []string{"No account rows loaded.", detail},
		},
		Journal: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Journal data unavailable.", detail},
			EmptyLines:   []string{"No journal rows loaded.", detail},
		},
		Quests: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Quest data unavailable.", detail},
			EmptyLines:   []string{"No quest rows loaded.", detail},
		},
		Loot: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Loot data unavailable.", detail},
			EmptyLines:   []string{"No loot rows loaded.", detail},
		},
	}
}

func resolveShellData(data *ShellData) ShellData {
	if data == nil || shellDataEmpty(data) {
		return DefaultShellData()
	}

	resolved := *data
	if dashboardDataEmpty(&resolved.Dashboard) {
		resolved.Dashboard = DefaultDashboardData()
	}

	if listScreenDataEmpty(&resolved.Accounts) {
		resolved.Accounts = defaultListScreenData(SectionAccounts)
	}
	if listScreenDataEmpty(&resolved.Journal) {
		resolved.Journal = defaultListScreenData(SectionJournal)
	}
	if listScreenDataEmpty(&resolved.Quests) {
		resolved.Quests = defaultListScreenData(SectionQuests)
	}
	if listScreenDataEmpty(&resolved.Loot) {
		resolved.Loot = defaultListScreenData(SectionLoot)
	}

	return resolved
}

func shellDataEmpty(data *ShellData) bool {
	if data == nil {
		return true
	}

	return dashboardDataEmpty(&data.Dashboard) &&
		listScreenDataEmpty(&data.Accounts) &&
		listScreenDataEmpty(&data.Journal) &&
		listScreenDataEmpty(&data.Quests) &&
		listScreenDataEmpty(&data.Loot)
}

func listScreenDataEmpty(data *ListScreenData) bool {
	if data == nil {
		return true
	}

	return len(data.HeaderLines) == 0 &&
		len(data.SummaryLines) == 0 &&
		len(data.RowLines) == 0 &&
		len(data.EmptyLines) == 0
}

func defaultListScreenData(section Section) ListScreenData {
	switch section {
	case SectionAccounts:
		return ListScreenData{
			HeaderLines: []string{
				"Chart of accounts shell.",
				"Read-only list now works; editing still stays in the CLI.",
			},
			SummaryLines: []string{
				"Account codes remain immutable.",
				"Used accounts may be marked inactive.",
				"Accounts with postings cannot be deleted.",
			},
			EmptyLines: []string{
				"No account rows loaded yet.",
				"App-side adapters fill this screen with live account data.",
			},
		}
	case SectionJournal:
		return ListScreenData{
			HeaderLines: []string{
				"Journal browser shell.",
				"Read-only browsing works; corrections still happen by reversal.",
			},
			SummaryLines: []string{
				"Posted journal entries remain immutable.",
				"Reversed entries stay visible in the audit trail.",
				"Interactive editing flows are intentionally deferred.",
			},
			EmptyLines: []string{
				"No journal rows loaded yet.",
				"App-side adapters fill this screen with live journal data.",
			},
		}
	case SectionQuests:
		return ListScreenData{
			HeaderLines: []string{
				"Quest register shell.",
				"Promised rewards stay off-ledger until earned.",
			},
			SummaryLines: []string{
				"Accepted, completed, and collectible quests stay visible.",
				"Receivables still belong to the formal ledger reports.",
			},
			EmptyLines: []string{
				"No quest rows loaded yet.",
				"App-side adapters fill this screen with live quest data.",
			},
		}
	case SectionLoot:
		return ListScreenData{
			HeaderLines: []string{
				"Unrealized loot register shell.",
				"Appraisals stay off-ledger until explicitly recognized.",
			},
			SummaryLines: []string{
				"Recognition and sale flows remain in the domain layer.",
				"This screen is read-only in the first interactive slice.",
			},
			EmptyLines: []string{
				"No loot rows loaded yet.",
				"App-side adapters fill this screen with live loot data.",
			},
		}
	default:
		return ListScreenData{}
	}
}
