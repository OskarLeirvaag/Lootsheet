package render

// DashboardData is the read-only view model rendered by the dashboard shell.
type DashboardData struct {
	HeaderLines   []string
	AccountsLines []string
	JournalLines  []string
	LedgerLines   []string
	QuestLines    []string
	LootLines     []string
}

// DefaultDashboardData returns the placeholder content used when no adapter is wired yet.
func DefaultDashboardData() DashboardData {
	return DashboardData{
		HeaderLines: []string{
			"Accounting shell slice 1: alternate screen, resize-safe layout, boxed panels, and footer help.",
			"Read-only placeholder view. Domain adapters and navigation land in the next slices.",
		},
		AccountsLines: []string{
			"Chart of accounts screen comes next.",
			"Codes stay immutable; names remain editable.",
			"Deletion protection stays in the domain layer.",
		},
		JournalLines: []string{
			"Posted entries remain immutable.",
			"Corrections continue to happen by reversal or adjustment.",
			"Interactive browsing lands after the dashboard shell.",
		},
		LedgerLines: []string{
			"Read-only data adapters are intentionally deferred.",
			"This slice proves the screen lifecycle before wiring reports.",
			"No raw SQL belongs in src/render.",
		},
		QuestLines: []string{
			"Promised rewards stay off-ledger until earned.",
			"Dashboard drill-down is planned after report adapters.",
		},
		LootLines: []string{
			"Unrealized appraisals stay off-ledger until recognized.",
			"Sales and losses will remain visible once wired into views.",
		},
	}
}

// ErrorDashboardData returns a dashboard model that keeps the TUI open while surfacing an error.
func ErrorDashboardData(summary string, detail string) DashboardData {
	if summary == "" {
		summary = "Dashboard data unavailable."
	}
	if detail == "" {
		detail = "No additional detail."
	}

	return DashboardData{
		HeaderLines:   []string{summary, detail},
		AccountsLines: []string{"Accounts unavailable.", detail},
		JournalLines:  []string{"Journal unavailable.", detail},
		LedgerLines:   []string{"Ledger snapshot unavailable.", detail},
		QuestLines:    []string{"Quest register unavailable.", detail},
		LootLines:     []string{"Loot register unavailable.", detail},
	}
}

func resolveDashboardData(data *DashboardData) DashboardData {
	if data == nil {
		return DefaultDashboardData()
	}

	if len(data.HeaderLines) == 0 &&
		len(data.AccountsLines) == 0 &&
		len(data.JournalLines) == 0 &&
		len(data.LedgerLines) == 0 &&
		len(data.QuestLines) == 0 &&
		len(data.LootLines) == 0 {
		return DefaultDashboardData()
	}

	return *data
}
