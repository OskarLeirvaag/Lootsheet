package render

// ledgerFilter selects which account types to show in the ledger overlay.
type ledgerFilter int

const (
	ledgerFilterAll ledgerFilter = iota
	ledgerFilterAssets
	ledgerFilterLiabilities
	ledgerFilterEquity
	ledgerFilterIncome
	ledgerFilterExpenses
)

// ledgerViewMode distinguishes between list and detail drill-down.
type ledgerViewMode int

const (
	ledgerModeList ledgerViewMode = iota
	ledgerModeDetail
)

// ledgerFilterLabels maps each filter to a human-readable display label.
var ledgerFilterLabels = []string{
	"All",
	"Assets",
	"Liabilities",
	"Equity",
	"Income",
	"Expenses",
}

// ledgerFilterAccountTypes maps a non-All filter to the corresponding AccountType string.
var ledgerFilterAccountTypes = map[ledgerFilter]string{
	ledgerFilterAssets:      "asset",
	ledgerFilterLiabilities: "liability",
	ledgerFilterEquity:      "equity",
	ledgerFilterIncome:      "income",
	ledgerFilterExpenses:    "expense",
}

// ledgerViewState holds the transient UI state for the ledger overlay.
type ledgerViewState struct {
	filter          ledgerFilter
	mode            ledgerViewMode
	selectedIndex   int
	scroll          int
	rows            []LedgerViewRow
	detail          *LedgerAccountDetail
	detailScroll    int
	previousSection Section
}

// filterLabel returns the display label for the given filter.
func filterLabel(f ledgerFilter) string {
	if int(f) >= 0 && int(f) < len(ledgerFilterLabels) {
		return ledgerFilterLabels[f]
	}
	return "All"
}

// openLedgerView initialises the ledger overlay and stores the current section
// so it can be restored on close. Returns false if ledger data is unavailable.
func (s *Shell) openLedgerView() bool {
	s.ledgerView = &ledgerViewState{
		filter:          ledgerFilterAll,
		mode:            ledgerModeList,
		previousSection: s.Section,
	}
	s.computeLedgerRows()
	return true
}

// computeLedgerRows rebuilds the filtered row slice from the full ledger report.
func (s *Shell) computeLedgerRows() {
	lv := s.ledgerView
	if lv == nil {
		return
	}

	all := s.Data.LedgerReport.Rows

	if lv.filter == ledgerFilterAll {
		lv.rows = make([]LedgerViewRow, len(all))
		copy(lv.rows, all)
	} else {
		want, ok := ledgerFilterAccountTypes[lv.filter]
		if !ok {
			lv.rows = nil
		} else {
			filtered := make([]LedgerViewRow, 0, len(all))
			for _, r := range all {
				if r.AccountType == want {
					filtered = append(filtered, r)
				}
			}
			lv.rows = filtered
		}
	}

	// Clamp selection and scroll to the new row count.
	if len(lv.rows) == 0 {
		lv.selectedIndex = 0
		lv.scroll = 0
	} else {
		if lv.selectedIndex >= len(lv.rows) {
			lv.selectedIndex = len(lv.rows) - 1
		}
		if lv.scroll > lv.selectedIndex {
			lv.scroll = lv.selectedIndex
		}
	}
}
