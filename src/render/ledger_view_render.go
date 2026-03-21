package render

import (
	"fmt"
	"strings"
)

// renderLedgerView draws the full-screen ledger overlay inside the body rect,
// replacing the normal section content. It supports list mode (trial balance)
// and detail mode (account drill-down).
func (s *Shell) renderLedgerView(buffer *Buffer, rect Rect, theme *Theme) { //nolint:revive // TUI renderer; cyclomatic complexity is inherent
	if s.ledgerView == nil || rect.Empty() {
		return
	}

	switch s.ledgerView.mode {
	case ledgerModeDetail:
		s.renderLedgerDetail(buffer, rect, theme)
	default:
		s.renderLedgerList(buffer, rect, theme)
	}
}

// renderLedgerList draws the list-mode trial balance view.
func (s *Shell) renderLedgerList(buffer *Buffer, rect Rect, theme *Theme) { //nolint:revive // TUI renderer
	lv := s.ledgerView
	style := sectionStyleFor(SectionLedger, theme)
	panel := style.Panel("General Ledger", nil)
	DrawPanel(buffer, rect, theme, panel)
	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	// --- Filter tab bar (row 0) ---
	x := content.X
	for i, label := range ledgerFilterLabels {
		if i > 0 {
			x += buffer.WriteString(x, content.Y, theme.Muted, "  ")
		}
		if ledgerFilter(i) == lv.filter {
			x += buffer.WriteString(x, content.Y, theme.SectionLedger, "["+label+"]")
		} else {
			x += buffer.WriteString(x, content.Y, theme.Muted, " "+label+" ")
		}
	}

	// --- Column header (row 2, after a blank row) ---
	headerY := content.Y + 2
	header := fmt.Sprintf("%-4s  %-30s  %10s  %10s  %10s", "CODE", "NAME", "DEBITS", "CREDITS", "BALANCE")
	buffer.WriteString(content.X, headerY, theme.Muted, clipText(header, content.W))

	// --- Account rows ---
	rowStartY := headerY + 1
	footerH := 2
	availH := max(1, content.Y+content.H-rowStartY-footerH)

	if lv.filter == ledgerFilterAll {
		s.renderLedgerGroupedRows(buffer, content, theme, rowStartY, availH)
	} else {
		s.renderLedgerFlatRows(buffer, content, theme, rowStartY, availH)
	}

	// --- Footer summary (last 2 rows) ---
	footerY := content.Y + content.H - footerH
	report := &s.Data.LedgerReport
	statusLabel := "Balanced"
	statusStyle := theme.StatusOK
	if !report.Balanced {
		statusLabel = "UNBALANCED"
		statusStyle = theme.StatusError
	}

	summaryLine := fmt.Sprintf("Total debits: %s    Total credits: %s    Status: ", report.TotalDebits, report.TotalCredits)
	w := buffer.WriteString(content.X, footerY, theme.Text, clipText(summaryLine, content.W))
	remaining := content.W - w
	if remaining > 0 {
		buffer.WriteString(content.X+w, footerY, statusStyle, clipText(statusLabel, remaining))
	}
}

// accountTypeGroup maps a type string to a group label and code range.
var accountTypeGroup = map[string]struct {
	label    string
	codeHint string
}{
	"asset":     {label: "Assets", codeHint: "1XXX"},
	"liability": {label: "Liabilities", codeHint: "2XXX"},
	"equity":    {label: "Equity", codeHint: "3XXX"},
	"income":    {label: "Income", codeHint: "4XXX"},
	"expense":   {label: "Expenses", codeHint: "5XXX"},
}

// displayLine is a virtual line in the ledger list that can be either a data
// row, a group header, or a subtotal.
type displayLine struct {
	kind     int // 0 = data row, 1 = group header, 2 = subtotal
	text     string
	dataIdx  int // index into lv.rows (only valid for kind == 0)
	typeKey  string
	dbTotal  string
	crTotal  string
	balTotal string
}

const (
	dlData     = 0
	dlHeader   = 1
	dlSubtotal = 2
)

func (s *Shell) renderLedgerGroupedRows(buffer *Buffer, content Rect, theme *Theme, startY, availH int) { //nolint:revive // TUI renderer
	lv := s.ledgerView
	nameW := max(10, content.W-60)

	// Build display lines with group headers and subtotals.
	var lines []displayLine
	var currentType string
	var groupDebits, groupCredits, groupBalance string

	flushSubtotal := func() {
		if currentType != "" {
			lines = append(lines, displayLine{
				kind:     dlSubtotal,
				typeKey:  currentType,
				dbTotal:  groupDebits,
				crTotal:  groupCredits,
				balTotal: groupBalance,
			})
		}
	}

	for i, row := range lv.rows {
		if row.AccountType != currentType {
			flushSubtotal()
			currentType = row.AccountType
			groupDebits = ""
			groupCredits = ""
			groupBalance = ""

			grp, ok := accountTypeGroup[currentType]
			if !ok {
				grp = struct {
					label    string
					codeHint string
				}{label: currentType, codeHint: "????"}
			}
			headerText := fmt.Sprintf("── %s (%s) ", grp.label, grp.codeHint)
			headerText += strings.Repeat("─", max(0, content.W-len([]rune(headerText))))
			lines = append(lines, displayLine{kind: dlHeader, text: headerText, typeKey: currentType})
		}

		name := clipText(row.AccountName, nameW)
		text := fmt.Sprintf("%-4s  %-*s  %10s  %10s  %10s",
			row.AccountCode, nameW, name, row.TotalDebits, row.TotalCredits, row.Balance)
		lines = append(lines, displayLine{kind: dlData, text: text, dataIdx: i})

		// Track last row values as group subtotal (server sends cumulative or we just show last).
		groupDebits = row.TotalDebits
		groupCredits = row.TotalCredits
		groupBalance = row.Balance
	}
	flushSubtotal()

	// Map selectedIndex to the display-line index.
	selectedDisplayIdx := 0
	for i, dl := range lines {
		if dl.kind == dlData && dl.dataIdx == lv.selectedIndex {
			selectedDisplayIdx = i
			break
		}
	}

	// Ensure scroll keeps selected item visible.
	if lv.scroll > selectedDisplayIdx {
		lv.scroll = selectedDisplayIdx
	}
	if selectedDisplayIdx >= lv.scroll+availH {
		lv.scroll = selectedDisplayIdx - availH + 1
	}
	maxScroll := max(0, len(lines)-availH)
	lv.scroll = clampInt(lv.scroll, 0, maxScroll)

	// Render visible lines.
	for row := range availH {
		lineIdx := lv.scroll + row
		if lineIdx >= len(lines) {
			break
		}
		dl := lines[lineIdx]
		y := startY + row

		switch dl.kind {
		case dlHeader:
			buffer.WriteString(content.X, y, theme.Muted, clipText(dl.text, content.W))
		case dlSubtotal:
			grp := accountTypeGroup[dl.typeKey]
			sub := fmt.Sprintf("%*s  %10s  %10s  %10s", nameW+6, grp.label+" total:", dl.dbTotal, dl.crTotal, dl.balTotal)
			buffer.WriteString(content.X, y, theme.Muted, clipText(sub, content.W))
		default:
			lineRect := Rect{X: content.X, Y: y, W: content.W, H: 1}
			st := theme.Text
			if dl.dataIdx == lv.selectedIndex {
				buffer.FillRect(lineRect, ' ', theme.SelectedRow)
				st = theme.SelectedRow
			}
			buffer.WriteString(content.X, y, st, clipText(dl.text, content.W))
		}
	}
}

func (s *Shell) renderLedgerFlatRows(buffer *Buffer, content Rect, theme *Theme, startY, availH int) {
	lv := s.ledgerView
	nameW := max(10, content.W-60)

	// Ensure scroll keeps selected item visible.
	totalLines := len(lv.rows) + 1 // +1 for subtotal
	if lv.scroll > lv.selectedIndex {
		lv.scroll = lv.selectedIndex
	}
	if lv.selectedIndex >= lv.scroll+availH {
		lv.scroll = lv.selectedIndex - availH + 1
	}
	maxScroll := max(0, totalLines-availH)
	lv.scroll = clampInt(lv.scroll, 0, maxScroll)

	for row := range availH {
		lineIdx := lv.scroll + row
		y := startY + row

		if lineIdx < len(lv.rows) {
			r := lv.rows[lineIdx]
			name := clipText(r.AccountName, nameW)
			text := fmt.Sprintf("%-4s  %-*s  %10s  %10s  %10s",
				r.AccountCode, nameW, name, r.TotalDebits, r.TotalCredits, r.Balance)

			lineRect := Rect{X: content.X, Y: y, W: content.W, H: 1}
			st := theme.Text
			if lineIdx == lv.selectedIndex {
				buffer.FillRect(lineRect, ' ', theme.SelectedRow)
				st = theme.SelectedRow
			}
			buffer.WriteString(content.X, y, st, clipText(text, content.W))
		} else if lineIdx == len(lv.rows) {
			// Subtotal line.
			label := filterLabel(lv.filter) + " total:"
			sub := fmt.Sprintf("%*s  %10s  %10s  %10s", nameW+6, label,
				s.Data.LedgerReport.TotalDebits, s.Data.LedgerReport.TotalCredits, "")
			buffer.WriteString(content.X, y, theme.Muted, clipText(sub, content.W))
		}
	}
}

// renderLedgerDetail draws the detail mode for a single account.
func (s *Shell) renderLedgerDetail(buffer *Buffer, rect Rect, theme *Theme) { //nolint:revive // TUI renderer
	lv := s.ledgerView
	detail := lv.detail
	if detail == nil {
		return
	}

	style := sectionStyleFor(SectionLedger, theme)
	title := "Account " + detail.AccountCode + " \u2014 " + detail.AccountName
	panel := style.Panel(title, nil)
	DrawPanel(buffer, rect, theme, panel)
	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	// --- Account summary (rows 0-1) ---
	summaryLine := fmt.Sprintf("Type: %-12s Debits: %-12s Credits: %-12s Balance: %s",
		detail.AccountType, detail.TotalDebits, detail.TotalCredits, detail.Balance)
	buffer.WriteString(content.X, content.Y, theme.Text, clipText(summaryLine, content.W))

	// --- Column header (row 3, after a blank row) ---
	headerY := content.Y + 2
	header := fmt.Sprintf("#%-4s  %-10s  %-20s  %10s  %10s  %10s", "NUM", "DATE", "DESCRIPTION", "DEBIT", "CREDIT", "BALANCE")
	buffer.WriteString(content.X, headerY, theme.Muted, clipText(header, content.W))

	// --- Entry rows (scrollable) ---
	rowStartY := headerY + 1
	footerH := 2
	availH := max(1, content.Y+content.H-rowStartY-footerH)

	entries := detail.Entries
	descW := max(8, content.W-62)

	// Clamp detail scroll.
	if len(entries) > 0 {
		if lv.detailScroll >= len(entries) {
			lv.detailScroll = len(entries) - 1
		}
	} else {
		lv.detailScroll = 0
	}

	scroll := min(lv.detailScroll, max(0, len(entries)-availH))

	for row := range availH {
		entryIdx := scroll + row
		if entryIdx >= len(entries) {
			break
		}
		e := entries[entryIdx]
		desc := clipText(e.Description, descW)
		text := fmt.Sprintf("#%-4d  %-10s  %-*s  %10s  %10s  %10s",
			e.EntryNumber, e.Date, descW, desc, e.Debit, e.Credit, e.RunningBalance)
		buffer.WriteString(content.X, rowStartY+row, theme.Text, clipText(text, content.W))
	}

	// --- Footer: account balance ---
	footerY := content.Y + content.H - footerH
	balanceLine := fmt.Sprintf("Account balance: %s", detail.Balance)
	buffer.WriteString(content.X, footerY, theme.Text, clipText(balanceLine, content.W))
}
