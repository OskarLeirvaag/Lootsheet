package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (s *Shell) pickerAccountsForCurrentField() []AccountOption {
	if s.compose == nil {
		return nil
	}
	fieldID := s.composeCurrentFieldID()
	switch s.compose.Mode {
	case composeModeExpense:
		switch fieldID {
		case fieldAccountCode:
			return s.Data.EntryCatalog.ExpenseAccounts
		case "offset_account_code":
			return s.Data.EntryCatalog.FundingAccounts
		}
	case composeModeIncome:
		switch fieldID {
		case fieldAccountCode:
			return s.Data.EntryCatalog.IncomeAccounts
		case "offset_account_code":
			return s.Data.EntryCatalog.DepositAccounts
		}
	case composeModeCustom, composeModeAssetTemplate:
		if fieldID == fieldAccountCode {
			return s.Data.EntryCatalog.AllAccounts
		}
	case composeModeAccount:
		if fieldID == "code" || fieldID == "name" {
			return s.Data.EntryCatalog.AllAccounts
		}
	case composeModeQuest, composeModeLoot, composeModeAsset, composeModeCodex, composeModeCodexType, composeModeNotes:
	}
	return nil
}

func (s *Shell) openAccountPicker() bool {
	options := s.pickerAccountsForCurrentField()
	if len(options) == 0 {
		return false
	}
	s.compose.picker = &accountPickerState{
		Options:  options,
		Filtered: options,
	}
	return true
}

func (s *Shell) pickerRefilter() {
	p := s.compose.picker
	if p == nil {
		return
	}
	if p.Query == "" {
		p.Filtered = p.Options
	} else {
		query := strings.ToLower(p.Query)
		filtered := make([]AccountOption, 0, len(p.Options))
		for _, opt := range p.Options {
			if strings.Contains(strings.ToLower(opt.Code), query) ||
				strings.Contains(strings.ToLower(opt.Name), query) ||
				strings.Contains(strings.ToLower(opt.Type), query) {
				filtered = append(filtered, opt)
			}
		}
		p.Filtered = filtered
	}
	if p.SelectedIndex >= len(p.Filtered) {
		p.SelectedIndex = max(0, len(p.Filtered)-1)
	}
	p.Scroll = 0
}

func (s *Shell) pickerApplySelection(code string) {
	s.composeEditCurrentField(func(_ string) string { return code })
}

func (s *Shell) handlePickerKeyEvent(event *tcell.EventKey) (handleResult, bool) {
	p := s.compose.picker
	if p == nil {
		return handleResult{}, false
	}

	switch event.Key() {
	case tcell.KeyEsc:
		s.compose.picker = nil
		return handleResult{Redraw: true}, true
	case tcell.KeyEnter:
		if len(p.Filtered) > 0 && p.SelectedIndex < len(p.Filtered) {
			s.pickerApplySelection(p.Filtered[p.SelectedIndex].Code)
		}
		s.compose.picker = nil
		return handleResult{Redraw: true}, true
	case tcell.KeyUp:
		if p.SelectedIndex > 0 {
			p.SelectedIndex--
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyDown:
		if p.SelectedIndex < len(p.Filtered)-1 {
			p.SelectedIndex++
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(p.Query) > 0 {
			p.Query = trimLastRune(p.Query)
			s.pickerRefilter()
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlU:
		p.Query = ""
		s.pickerRefilter()
		return handleResult{Redraw: true}, true
	case tcell.KeyRune:
		r := event.Rune()
		if r == 'q' && p.Query == "" {
			s.compose.picker = nil
			return handleResult{Redraw: true}, true
		}
		p.Query += string(r)
		s.pickerRefilter()
		return handleResult{Redraw: true}, true
	default:
		return handleResult{}, true
	}
}
