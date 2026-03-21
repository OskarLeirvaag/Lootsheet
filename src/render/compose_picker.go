package render

import "github.com/gdamore/tcell/v2"

const pickerTitleAccount = "Pick Account"

func (s *Shell) pickerOptionsForCurrentField() (string, []pickerOption) {
	if s.compose == nil {
		return "", nil
	}
	fieldID := s.composeCurrentFieldID()
	switch s.compose.Mode {
	case composeModeExpense:
		switch fieldID {
		case fieldAccountCode:
			return pickerTitleAccount, accountsToPickerOptions(s.Data.EntryCatalog.ExpenseAccounts)
		case "offset_account_code":
			return pickerTitleAccount, accountsToPickerOptions(s.Data.EntryCatalog.FundingAccounts)
		default:
		}
	case composeModeIncome:
		switch fieldID {
		case fieldAccountCode:
			return pickerTitleAccount, accountsToPickerOptions(s.Data.EntryCatalog.IncomeAccounts)
		case "offset_account_code":
			return pickerTitleAccount, accountsToPickerOptions(s.Data.EntryCatalog.DepositAccounts)
		default:
		}
	case composeModeCustom, composeModeAssetTemplate:
		if fieldID == fieldAccountCode {
			return pickerTitleAccount, accountsToPickerOptions(s.Data.EntryCatalog.AllAccounts)
		}
	case composeModeAccount:
		if fieldID == "code" || fieldID == "name" {
			return pickerTitleAccount, accountsToPickerOptions(s.Data.EntryCatalog.AllAccounts)
		}
	case composeModeQuest:
		if fieldID == "patron" {
			return "Pick Patron", s.codexPersonPickerOptions()
		}
	case composeModeLoot, composeModeAsset:
		switch fieldID {
		case "holder":
			return "Pick Holder", s.codexPersonPickerOptions()
		case "source":
			return "Pick Source", s.sourcePickerOptions()
		default:
		}
	case composeModeCodex, composeModeCodexType, composeModeCampaign, composeModeNotes:
	default:
	}
	return "", nil
}

func (s *Shell) openComposePicker() bool {
	title, options := s.pickerOptionsForCurrentField()
	if len(options) == 0 {
		return false
	}
	s.compose.picker = newPicker(title, options)
	return true
}

func (s *Shell) handleComposePickerKey(event *tcell.EventKey) (HandleResult, bool) {
	p := s.compose.picker
	if p == nil {
		return HandleResult{}, false
	}

	closed, value := handlePickerKey(p, event)
	if closed {
		if value != "" {
			s.composeEditCurrentField(func(_ string) string { return value })
		}
		s.compose.picker = nil
	}
	return HandleResult{Redraw: true}, true
}

func accountsToPickerOptions(accounts []AccountOption) []pickerOption {
	opts := make([]pickerOption, len(accounts))
	for i, a := range accounts {
		opts[i] = pickerOption{Value: a.Code, Label: a.Code + " " + a.Name, Kind: a.Type}
	}
	return opts
}

func (s *Shell) codexPersonPickerOptions() []pickerOption {
	opts := make([]pickerOption, 0, len(s.Data.Codex.Items))
	for _, item := range s.Data.Codex.Items {
		opts = append(opts, pickerOption{Value: item.DetailTitle, Label: item.DetailTitle, Kind: "person"})
	}
	return opts
}

func (s *Shell) sourcePickerOptions() []pickerOption {
	opts := make([]pickerOption, 0, len(s.Data.Quests.Items)+len(s.Data.Codex.Items))
	for _, item := range s.Data.Quests.Items {
		opts = append(opts, pickerOption{Value: item.DetailTitle, Label: item.DetailTitle, Kind: "quest"})
	}
	for _, item := range s.Data.Codex.Items {
		opts = append(opts, pickerOption{Value: item.DetailTitle, Label: item.DetailTitle, Kind: "person"})
	}
	return opts
}
