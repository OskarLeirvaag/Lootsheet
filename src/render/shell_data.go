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
	ComposeLines  []CommandLine
}

// ListItemData is a structured row plus detail content for list-style screens.
type ListItemData struct {
	Key         string
	Row         string
	DetailTitle string
	DetailLines []string
	DetailBody  string // if set, rendered as styled markdown in the detail pane
	Actions     []ItemActionData
}

// ListScreenData is the neutral view model for list-style TUI sections.
type ListScreenData struct {
	HeaderLines   []string
	SummaryLines  []string
	ListHeaderRow string
	Items         []ListItemData
	EmptyLines    []string
}

// ShellData contains the full TUI snapshot.
type ShellData struct {
	Dashboard          DashboardData
	Accounts           ListScreenData
	Journal            ListScreenData
	Quests             ListScreenData
	Loot               ListScreenData
	Assets             ListScreenData
	Codex              ListScreenData
	Notes              ListScreenData
	SettingsAccounts   ListScreenData
	SettingsCodexTypes ListScreenData
	EntryCatalog       EntryCatalog
	CodexTypes         []CodexTypeOption
}

// DefaultShellData returns placeholder content for the full shell.
func DefaultShellData() ShellData {
	return ShellData{
		Dashboard:          DefaultDashboardData(),
		Accounts:           defaultListScreenData(SectionAccounts),
		Journal:            defaultListScreenData(SectionJournal),
		Quests:             defaultListScreenData(SectionQuests),
		Loot:               defaultListScreenData(SectionLoot),
		Assets:             defaultListScreenData(SectionAssets),
		Codex:              defaultListScreenData(SectionCodex),
		Notes:              defaultListScreenData(SectionNotes),
		SettingsAccounts:   defaultListScreenData(settingsTabAccounts),
		SettingsCodexTypes: defaultListScreenData(settingsTabCodexTypes),
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
		Codex: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Codex data unavailable.", detail},
			EmptyLines:   []string{"No codex rows loaded.", detail},
		},
		Notes: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Notes data unavailable.", detail},
			EmptyLines:   []string{"No notes rows loaded.", detail},
		},
		SettingsAccounts: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Account settings unavailable.", detail},
			EmptyLines:   []string{"No account rows loaded.", detail},
		},
		SettingsCodexTypes: ListScreenData{
			HeaderLines:  []string{summary, detail},
			SummaryLines: []string{"Codex type settings unavailable.", detail},
			EmptyLines:   []string{"No codex type rows loaded.", detail},
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
	if listScreenDataEmpty(&resolved.Codex) {
		resolved.Codex = defaultListScreenData(SectionCodex)
	}
	if listScreenDataEmpty(&resolved.Notes) {
		resolved.Notes = defaultListScreenData(SectionNotes)
	}
	if listScreenDataEmpty(&resolved.SettingsAccounts) {
		resolved.SettingsAccounts = defaultListScreenData(settingsTabAccounts)
	}
	if listScreenDataEmpty(&resolved.SettingsCodexTypes) {
		resolved.SettingsCodexTypes = defaultListScreenData(settingsTabCodexTypes)
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
		listScreenDataEmpty(&data.Assets) &&
		listScreenDataEmpty(&data.Codex) &&
		listScreenDataEmpty(&data.Notes) &&
		listScreenDataEmpty(&data.SettingsAccounts) &&
		listScreenDataEmpty(&data.SettingsCodexTypes)
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
	case SectionCodex:
		return ListScreenData{
			HeaderLines: []string{
				"Codex shell.",
				"Players, NPCs, and contacts with type-specific forms and cross-references.",
			},
			SummaryLines: []string{
				"Codex entries can reference quests, loot, assets, and other people.",
				"Use @type/name syntax in notes for cross-references.",
			},
			EmptyLines: []string{
				"No codex entries loaded yet.",
				"App-side adapters fill this screen with live codex data and actions.",
			},
		}
	case SectionNotes:
		return ListScreenData{
			HeaderLines: []string{
				"Campaign notes shell.",
				"General-purpose session and campaign notes with cross-references.",
			},
			SummaryLines: []string{
				"Notes can reference quests, loot, assets, people, and other notes.",
				"Use @type/name syntax in body text for cross-references.",
			},
			EmptyLines: []string{
				"No notes loaded yet.",
				"App-side adapters fill this screen with live note data and actions.",
			},
		}
	case settingsTabAccounts:
		return ListScreenData{
			HeaderLines: []string{
				"Account settings.",
				"Chart of accounts used by the ledger. `a` adds, `u` renames, `d` deletes, `t` toggles.",
			},
			SummaryLines: []string{
				"Account codes remain immutable.",
				"Used accounts may be marked inactive.",
				"Accounts with postings cannot be deleted.",
			},
			EmptyLines: []string{
				"No account rows loaded yet.",
				"App-side adapters fill this tab with live account data.",
			},
		}
	case settingsTabCodexTypes:
		return ListScreenData{
			HeaderLines: []string{
				"Codex type settings.",
				"Entry categories for the codex. `a` adds, `u` renames, `d` deletes.",
			},
			SummaryLines: []string{
				"Each codex type uses a form template (player, npc, settlement).",
				"Types with existing entries cannot be deleted.",
			},
			EmptyLines: []string{
				"No codex type rows loaded yet.",
				"App-side adapters fill this tab with live codex type data.",
			},
		}
	default:
		return ListScreenData{}
	}
}
