package render

// AccountOption is a type-filtered account choice for guided entry creation.
type AccountOption struct {
	Code string
	Name string
	Type string
}

// EntryCatalog contains the active account choices for guided entry creation.
type EntryCatalog struct {
	DefaultDate     string
	ExpenseAccounts []AccountOption
	IncomeAccounts  []AccountOption
	FundingAccounts []AccountOption
	DepositAccounts []AccountOption
	AllAccounts     []AccountOption
}

// ItemActionMode determines how the shell should open an item action.
type ItemActionMode string

const (
	ItemActionModeConfirm ItemActionMode = "confirm"
	ItemActionModeInput   ItemActionMode = "input"
	ItemActionModeCompose ItemActionMode = "compose"
)

// ItemActionData describes the primary action available for a list item.
type ItemActionData struct {
	Trigger       Action
	ID            string
	Label         string
	Mode          ItemActionMode
	ConfirmTitle  string
	ConfirmLines  []string
	InputTitle    string
	InputPrompt   string
	InputHelp     []string
	Placeholder   string
	ComposeMode   string
	ComposeTitle  string
	ComposeFields map[string]string
}

// ListItemData is a structured row plus detail content for list-style screens.
type ListItemData struct {
	Key         string
	Row         string
	DetailTitle string
	DetailLines []string
	Actions     []ItemActionData
}

// ListScreenData is the neutral view model for list-style TUI sections.
type ListScreenData struct {
	HeaderLines  []string
	SummaryLines []string
	Items        []ListItemData
	EmptyLines   []string
}

// ShellData contains the full TUI snapshot.
type ShellData struct {
	Dashboard    DashboardData
	Accounts     ListScreenData
	Journal      ListScreenData
	Quests       ListScreenData
	Loot         ListScreenData
	Assets       ListScreenData
	EntryCatalog EntryCatalog
}

// DefaultShellData returns placeholder content for the full shell.
func DefaultShellData() ShellData {
	return ShellData{
		Dashboard: DefaultDashboardData(),
		Accounts:  defaultListScreenData(SectionAccounts),
		Journal:   defaultListScreenData(SectionJournal),
		Quests:    defaultListScreenData(SectionQuests),
		Loot:      defaultListScreenData(SectionLoot),
		Assets:    defaultListScreenData(SectionAssets),
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
		Assets: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Asset data unavailable.", detail},
			EmptyLines:   []string{"No asset rows loaded.", detail},
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
	if listScreenDataEmpty(&resolved.Assets) {
		resolved.Assets = defaultListScreenData(SectionAssets)
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
		listScreenDataEmpty(&data.Loot) &&
		listScreenDataEmpty(&data.Assets)
}

func listScreenDataEmpty(data *ListScreenData) bool {
	if data == nil {
		return true
	}

	return len(data.HeaderLines) == 0 &&
		len(data.SummaryLines) == 0 &&
		len(data.Items) == 0 &&
		len(data.EmptyLines) == 0
}

func defaultListScreenData(section Section) ListScreenData {
	switch section {
	case SectionAccounts:
		return ListScreenData{
			HeaderLines: []string{
				"Chart of accounts shell.",
				"Selection and detail panes are ready for the first interactive slice.",
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
				"Selection and detail panes work before edit flows land.",
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
				"App-side adapters fill this screen with live quest data and actions.",
			},
		}
	case SectionLoot:
		return ListScreenData{
			HeaderLines: []string{
				"Unrealized loot register shell.",
				"Appraisals stay off-ledger until explicitly recognized.",
			},
			SummaryLines: []string{
				"Recognition and sale workflows are available from the Loot screen.",
				"Held items can be recognized; recognized items can be sold.",
			},
			EmptyLines: []string{
				"No loot rows loaded yet.",
				"App-side adapters fill this screen with live loot data and actions.",
			},
		}
	case SectionAssets:
		return ListScreenData{
			HeaderLines: []string{
				"Party asset register shell.",
				"High-value items the party intends to keep.",
			},
			SummaryLines: []string{
				"Assets share the loot appraisal system.",
				"Transfer items to the loot register when ready to sell.",
			},
			EmptyLines: []string{
				"No asset rows loaded yet.",
				"App-side adapters fill this screen with live asset data and actions.",
			},
		}
	default:
		return ListScreenData{}
	}
}
