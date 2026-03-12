package dashboard

import "github.com/OskarLeirvaag/Lootsheet/src/render/model"

// DefaultData returns placeholder content used when no adapter is wired yet.
func DefaultData() model.DashboardData {
	return model.DashboardData{
		HeaderLines: []string{
			"Read-only TUI shell: boxed dashboard, live navigation, and app-facing adapters.",
			"Use section navigation to move between dashboard, accounts, journal, quests, and loot.",
		},
		AccountsLines: []string{
			"Chart of accounts screen is live in read-only mode.",
			"Codes stay immutable; names remain editable.",
			"Deletion protection stays in the domain layer.",
		},
		JournalLines: []string{
			"Posted entries remain immutable.",
			"Corrections continue to happen by reversal or adjustment.",
			"Read-only browsing is available before editing flows.",
		},
		HoardLines: []string{
			"To share now: awaiting ledger data.",
			"Unsold loot stays out of the split until sold.",
		},
		QuickEntryLines: []string{
			"e  I have an expense",
			"i  I have income",
			"a  Add custom entry",
		},
		LedgerLines: []string{
			"Dashboard summaries stay read-only in this slice.",
			"Drill-down screens use app-side adapters instead of raw SQL.",
			"No raw SQL belongs in src/render.",
		},
		QuestLines: []string{
			"Promised rewards stay off-ledger until earned.",
			"Quest register now supports collect and write-off actions.",
		},
		LootLines: []string{
			"Unrealized appraisals stay off-ledger until recognized.",
			"Loot recognition is the next interactive TUI workflow.",
		},
		AssetLines: []string{
			"High-value items the party keeps.",
			"Transfer to loot when ready to sell.",
		},
	}
}

// ErrorData returns a dashboard model that keeps the TUI open while surfacing an error.
func ErrorData(summary string, detail string) model.DashboardData {
	if summary == "" {
		summary = "Dashboard data unavailable."
	}
	if detail == "" {
		detail = "No additional detail."
	}

	return model.DashboardData{
		HeaderLines:     []string{summary, detail},
		AccountsLines:   []string{"Accounts unavailable.", detail},
		JournalLines:    []string{"Journal unavailable.", detail},
		HoardLines:      []string{"Dragon hoard unavailable.", detail},
		QuickEntryLines: []string{"Quick entry unavailable.", detail},
		LedgerLines:     []string{"Ledger snapshot unavailable.", detail},
		QuestLines:      []string{"Quest register unavailable.", detail},
		LootLines:       []string{"Loot register unavailable.", detail},
		AssetLines:      []string{"Asset register unavailable.", detail},
	}
}

// ResolveData returns the given data if populated, or DefaultData otherwise.
func ResolveData(data *model.DashboardData) model.DashboardData {
	if data == nil || DataEmpty(data) {
		return DefaultData()
	}

	return *data
}

// DataEmpty reports whether all dashboard fields are empty.
func DataEmpty(data *model.DashboardData) bool {
	if data == nil {
		return true
	}

	return len(data.HeaderLines) == 0 &&
		len(data.AccountsLines) == 0 &&
		len(data.JournalLines) == 0 &&
		len(data.HoardLines) == 0 &&
		len(data.QuickEntryLines) == 0 &&
		len(data.LedgerLines) == 0 &&
		len(data.QuestLines) == 0 &&
		len(data.LootLines) == 0 &&
		len(data.AssetLines) == 0
}
