package render

import "github.com/gdamore/tcell/v2"

const ledgerFilterCount = 6

func (s *Shell) handleLedgerViewKeyEvent(event *tcell.EventKey, action Action) (HandleResult, bool) {
	if s.ledgerView == nil || event == nil {
		return HandleResult{}, false
	}

	lv := s.ledgerView

	switch lv.mode {
	case ledgerModeList:
		return s.handleLedgerListKey(event, action)
	case ledgerModeDetail:
		return s.handleLedgerDetailKey(event, action)
	default:
		return HandleResult{}, true
	}
}

func (s *Shell) handleLedgerListKey(event *tcell.EventKey, action Action) (HandleResult, bool) { //nolint:revive // key event dispatcher; cyclomatic complexity inherent
	lv := s.ledgerView

	// Export shortcuts (raw rune keys, not semantic actions).
	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case 'x':
			return HandleResult{Command: &Command{ID: "ledger.export.excel"}}, true
		case 'p':
			return HandleResult{Command: &Command{ID: "ledger.export.pdf"}}, true
		case 'c':
			return HandleResult{Command: &Command{ID: "ledger.export.csv"}}, true
		}
	}

	switch action {
	case ActionQuit:
		s.Section = lv.previousSection
		s.ledgerView = nil
		return HandleResult{Redraw: true}, true

	case ActionRedraw:
		s.Section = lv.previousSection
		s.ledgerView = nil
		return HandleResult{Reload: true}, true

	case ActionMoveUp:
		if lv.selectedIndex > 0 {
			lv.selectedIndex--
		}
		return HandleResult{Redraw: true}, true

	case ActionMoveDown:
		if len(lv.rows) > 0 && lv.selectedIndex < len(lv.rows)-1 {
			lv.selectedIndex++
		}
		return HandleResult{Redraw: true}, true

	case ActionPageUp:
		lv.selectedIndex -= 10
		if lv.selectedIndex < 0 {
			lv.selectedIndex = 0
		}
		return HandleResult{Redraw: true}, true

	case ActionPageDown:
		lv.selectedIndex += 10
		if len(lv.rows) > 0 && lv.selectedIndex > len(lv.rows)-1 {
			lv.selectedIndex = len(lv.rows) - 1
		}
		return HandleResult{Redraw: true}, true

	case ActionMoveTop:
		lv.selectedIndex = 0
		return HandleResult{Redraw: true}, true

	case ActionMoveBottom:
		if len(lv.rows) > 0 {
			lv.selectedIndex = len(lv.rows) - 1
		}
		return HandleResult{Redraw: true}, true

	case ActionConfirm:
		if len(lv.rows) > 0 && lv.selectedIndex < len(lv.rows) {
			code := lv.rows[lv.selectedIndex].AccountCode
			detail, ok := s.Data.LedgerReport.AccountDetail[code]
			if ok {
				lv.detail = &detail
				lv.mode = ledgerModeDetail
				lv.detailScroll = 0
			}
		}
		return HandleResult{Redraw: true}, true
	}

	// Left / BackTab → previous filter; Right / Tab → next filter.
	switch event.Key() { //nolint:exhaustive // only filter-cycling keys
	case tcell.KeyLeft, tcell.KeyBacktab:
		lv.filter = ledgerFilter((int(lv.filter) + ledgerFilterCount - 1) % ledgerFilterCount)
		s.computeLedgerRows()
		return HandleResult{Redraw: true}, true

	case tcell.KeyRight, tcell.KeyTab:
		lv.filter = ledgerFilter((int(lv.filter) + 1) % ledgerFilterCount)
		s.computeLedgerRows()
		return HandleResult{Redraw: true}, true
	}

	return HandleResult{}, true
}

func (s *Shell) handleLedgerDetailKey(event *tcell.EventKey, action Action) (HandleResult, bool) {
	lv := s.ledgerView

	// Export shortcuts (raw rune keys, not semantic actions).
	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case 'x':
			return HandleResult{Command: &Command{ID: "ledger.export.excel"}}, true
		case 'p':
			return HandleResult{Command: &Command{ID: "ledger.export.pdf"}}, true
		case 'c':
			return HandleResult{Command: &Command{ID: "ledger.export.csv"}}, true
		}
	}

	entryCount := 0
	if lv.detail != nil {
		entryCount = len(lv.detail.Entries)
	}

	switch action {
	case ActionQuit:
		lv.mode = ledgerModeList
		lv.detail = nil
		return HandleResult{Redraw: true}, true

	case ActionRedraw:
		s.Section = lv.previousSection
		s.ledgerView = nil
		return HandleResult{Reload: true}, true

	case ActionMoveUp:
		if lv.detailScroll > 0 {
			lv.detailScroll--
		}
		return HandleResult{Redraw: true}, true

	case ActionMoveDown:
		if entryCount > 0 && lv.detailScroll < entryCount-1 {
			lv.detailScroll++
		}
		return HandleResult{Redraw: true}, true

	case ActionPageUp:
		lv.detailScroll -= 10
		if lv.detailScroll < 0 {
			lv.detailScroll = 0
		}
		return HandleResult{Redraw: true}, true

	case ActionPageDown:
		lv.detailScroll += 10
		if entryCount > 0 && lv.detailScroll > entryCount-1 {
			lv.detailScroll = entryCount - 1
		}
		return HandleResult{Redraw: true}, true

	case ActionMoveTop:
		lv.detailScroll = 0
		return HandleResult{Redraw: true}, true

	case ActionMoveBottom:
		if entryCount > 0 {
			lv.detailScroll = entryCount - 1
		}
		return HandleResult{Redraw: true}, true
	}

	return HandleResult{}, true
}
