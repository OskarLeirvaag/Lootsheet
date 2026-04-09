package model

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
	InputRequired string
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

// CampaignOption is a campaign choice for the campaign picker.
type CampaignOption struct {
	ID   string
	Name string
}

// CodexTypeOption is a type choice for the codex picker.
type CodexTypeOption struct {
	ID     string
	Name   string
	FormID string
}

// DashboardData is the read-only view model rendered by the dashboard shell.
type DashboardData struct {
	HeaderLines     []string
	AccountsLines   []string
	JournalLines    []string
	HoardLines      []string
	QuickEntryLines []string
	LedgerLines     []string
	QuestLines      []string
	LootLines       []string
	AssetLines      []string
}

// LedgerDetailEntry is a single posting in an account's ledger.
type LedgerDetailEntry struct {
	EntryNumber    int
	Date           string
	Description    string
	Debit          string
	Credit         string
	RunningBalance string
}

// LedgerAccountDetail is the drill-down data for a single account.
type LedgerAccountDetail struct {
	AccountCode  string
	AccountName  string
	AccountType  string
	Entries      []LedgerDetailEntry
	TotalDebits  string
	TotalCredits string
	Balance      string
}

// LedgerViewRow is a single account row in the trial balance.
type LedgerViewRow struct {
	AccountCode  string
	AccountName  string
	AccountType  string
	TotalDebits  string
	TotalCredits string
	Balance      string
}

// LedgerViewData is the structured data for the full-screen ledger overlay.
type LedgerViewData struct {
	Rows          []LedgerViewRow
	AccountDetail map[string]LedgerAccountDetail
	TotalDebits   string
	TotalCredits  string
	Balanced      bool
	Available     bool
}

// ShellData contains the full TUI snapshot.
type ShellData struct {
	Dashboard          DashboardData
	Ledger             ListScreenData
	LedgerReport       LedgerViewData
	Journal            ListScreenData
	Quests             ListScreenData
	Loot               ListScreenData
	Assets             ListScreenData
	Codex              ListScreenData
	Notes              ListScreenData
	SettingsAccounts      ListScreenData
	SettingsCodexTypes    ListScreenData
	SettingsCampaigns     ListScreenData
	SettingsCompendium    ListScreenData
	CompendiumMonsters    ListScreenData
	CompendiumSpells      ListScreenData
	CompendiumItems       ListScreenData
	CompendiumRules       ListScreenData
	CompendiumConditions  ListScreenData
	EntryCatalog       EntryCatalog
	CodexTypes         []CodexTypeOption
	CampaignName       string           // active campaign name for header
	Campaigns          []CampaignOption // for campaign picker modal
}
