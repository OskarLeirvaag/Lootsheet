package render

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

func (s *Shell) handleComposeKeyEvent(event *tcell.EventKey, action Action) (handleResult, bool) {
	if s.compose == nil || event == nil {
		return handleResult{}, false
	}

	if s.compose.picker != nil {
		return s.handlePickerKeyEvent(event)
	}

	switch event.Key() {
	case tcell.KeyUp, tcell.KeyLeft:
		s.composeAdvance(-1)
		return handleResult{Redraw: true}, true
	case tcell.KeyDown, tcell.KeyRight:
		s.composeAdvance(1)
		return handleResult{Redraw: true}, true
	case tcell.KeyTab:
		s.composeAdvance(1)
		return handleResult{Redraw: true}, true
	case tcell.KeyBacktab:
		s.composeAdvance(-1)
		return handleResult{Redraw: true}, true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		s.composeBackspace()
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlU:
		s.composeClearCurrent()
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlA:
		if s.openAccountPicker() {
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyCtrlN:
		switch s.compose.Mode {
		case composeModeCustom:
			if len(s.compose.Lines) < 8 {
				s.compose.Lines = append(s.compose.Lines, composeLineState{Side: sideDebit})
				s.compose.FieldIndex = s.composeFieldCount() - 4
			}
		case composeModeAssetTemplate:
			if len(s.compose.Lines) < 8 {
				s.compose.Lines = append(s.compose.Lines, composeLineState{Side: sideDebit})
				s.compose.FieldIndex = s.composeFieldCount() - 3
			}
		default:
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlD:
		switch s.compose.Mode {
		case composeModeCustom:
			if len(s.compose.Lines) > 2 {
				lineIndex, column := s.composeCurrentLinePosition()
				if lineIndex >= 0 && lineIndex < len(s.compose.Lines) {
					s.compose.Lines = append(s.compose.Lines[:lineIndex], s.compose.Lines[lineIndex+1:]...)
					if len(s.compose.Lines) == 0 {
						s.compose.Lines = []composeLineState{{Side: sideDebit}, {Side: sideCredit}}
					}
					if lineIndex >= len(s.compose.Lines) {
						lineIndex = len(s.compose.Lines) - 1
					}
					s.compose.FieldIndex = 2 + lineIndex*4 + column
				}
			}
		case composeModeAssetTemplate:
			if len(s.compose.Lines) > 2 {
				lineIndex, column := s.composeCurrentTemplateLinePosition()
				if lineIndex >= 0 && lineIndex < len(s.compose.Lines) {
					s.compose.Lines = append(s.compose.Lines[:lineIndex], s.compose.Lines[lineIndex+1:]...)
					if len(s.compose.Lines) == 0 {
						s.compose.Lines = []composeLineState{{Side: sideDebit}, {Side: sideCredit}}
					}
					if lineIndex >= len(s.compose.Lines) {
						lineIndex = len(s.compose.Lines) - 1
					}
					if column > 2 {
						column = 2
					}
					s.compose.FieldIndex = lineIndex*3 + column
				}
			}
		default:
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyRune:
		if toggled := s.composeToggleSide(event.Rune()); toggled {
			return handleResult{Redraw: true}, true
		}

		s.composeAppendRune(event.Rune())
		return handleResult{Redraw: true}, true
	default:
	}

	switch action {
	case ActionQuit:
		s.compose = nil
		return handleResult{Redraw: true}, true
	case ActionRedraw:
		s.compose = nil
		return handleResult{Reload: true}, true
	case ActionSubmitCompose, ActionConfirm:
		if command, ok := s.composeCommand(); ok {
			return handleResult{Command: command}, true
		}
		return handleResult{Redraw: true}, true
	case ActionNone, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowSettings, ActionShowJournal, ActionShowQuests, ActionShowLoot, ActionShowAssets, ActionShowCodex, ActionShowNotes,
		ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
		ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer,
		ActionEditTemplate, ActionExecuteTemplate,
		ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionSearch:
		return handleResult{}, false
	case ActionHelp:
		return handleResult{}, true
	}

	return handleResult{}, true
}

func (s *Shell) composeAdvance(delta int) {
	if s.compose == nil {
		return
	}
	count := s.composeFieldCount()
	if count == 0 {
		return
	}
	index := s.compose.FieldIndex + delta
	for index < 0 {
		index += count
	}
	s.compose.FieldIndex = index % count
	s.compose.GeneralError = ""
}

func (s *Shell) composeCurrentFieldID() string {
	if s.compose == nil {
		return ""
	}
	if s.compose.Mode == composeModeAssetTemplate {
		_, column := s.composeCurrentTemplateLinePosition()
		switch column {
		case 0:
			return "side"
		case 1:
			return fieldAccountCode
		case 2:
			return "amount"
		default:
			return ""
		}
	}
	fields := s.composeFieldDefinitions()
	if s.compose.FieldIndex < len(fields) {
		return fields[s.compose.FieldIndex].ID
	}
	if s.compose.Mode != composeModeCustom {
		return ""
	}
	_, column := s.composeCurrentLinePosition()
	switch column {
	case 0:
		return "side"
	case 1:
		return fieldAccountCode
	case 2:
		return "amount"
	case 3:
		return "memo"
	default:
		return ""
	}
}

func (s *Shell) composeCurrentLinePosition() (int, int) {
	if s.compose == nil || s.compose.Mode != composeModeCustom {
		return -1, -1
	}
	offset := s.compose.FieldIndex - len(s.composeFieldDefinitions())
	if offset < 0 {
		return -1, -1
	}
	return offset / 4, offset % 4
}

func (s *Shell) composeCurrentTemplateLinePosition() (int, int) {
	if s.compose == nil || s.compose.Mode != composeModeAssetTemplate {
		return -1, -1
	}
	return s.compose.FieldIndex / 3, s.compose.FieldIndex % 3
}

type fieldOp func(string) string

func (s *Shell) composeEditCurrentField(op fieldOp) {
	if s.compose == nil {
		return
	}
	fieldID := s.composeCurrentFieldID()

	if s.compose.Mode == composeModeAssetTemplate {
		lineIndex, column := s.composeCurrentTemplateLinePosition()
		if lineIndex < 0 || lineIndex >= len(s.compose.Lines) {
			return
		}
		switch column {
		case 0:
			// side — toggled with space, not typed
			return
		case 1:
			s.compose.Lines[lineIndex].AccountCode = op(s.compose.Lines[lineIndex].AccountCode)
		case 2:
			s.compose.Lines[lineIndex].Amount = op(s.compose.Lines[lineIndex].Amount)
		}
		delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
		s.compose.GeneralError = ""
		return
	}

	if fieldID == "side" {
		return
	}
	if s.composeIsDefinedField(fieldID) {
		s.compose.Fields[fieldID] = op(s.compose.Fields[fieldID])
		delete(s.compose.FieldErrors, fieldID)
		s.compose.GeneralError = ""
		return
	}
	{
		lineIndex, column := s.composeCurrentLinePosition()
		if lineIndex < 0 || lineIndex >= len(s.compose.Lines) {
			return
		}
		switch column {
		case 1:
			s.compose.Lines[lineIndex].AccountCode = op(s.compose.Lines[lineIndex].AccountCode)
		case 2:
			s.compose.Lines[lineIndex].Amount = op(s.compose.Lines[lineIndex].Amount)
		case 3:
			s.compose.Lines[lineIndex].Memo = op(s.compose.Lines[lineIndex].Memo)
		}
		delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
	}
	s.compose.GeneralError = ""
}

func (s *Shell) composeAppendRune(r rune) {
	s.composeEditCurrentField(func(v string) string { return v + string(r) })
}

func (s *Shell) composeBackspace() {
	s.composeEditCurrentField(trimLastRune)
}

func (s *Shell) composeClearCurrent() {
	s.composeEditCurrentField(func(_ string) string { return "" })
}

func (s *Shell) clearErrorForCurrentField() {
	if s.compose == nil {
		return
	}
	fieldID := s.composeCurrentFieldID()
	if fieldID == "" {
		return
	}
	if fieldID == "side" {
		if s.compose.Mode == composeModeAssetTemplate {
			lineIndex, column := s.composeCurrentTemplateLinePosition()
			delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
		} else {
			lineIndex, column := s.composeCurrentLinePosition()
			delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
		}
		return
	}
	delete(s.compose.FieldErrors, fieldID)
}

func (s *Shell) composeLineFieldKey(index int, column int) string {
	switch column {
	case 0:
		return fmt.Sprintf("line_%d_side", index)
	case 1:
		return fmt.Sprintf("line_%d_account", index)
	case 2:
		return fmt.Sprintf("line_%d_amount", index)
	default:
		return fmt.Sprintf("line_%d_memo", index)
	}
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[:len(runes)-1])
}

// composeToggleSide handles space-to-toggle side for custom and asset-template
// compose modes. Returns true if the toggle was applied.
func (s *Shell) composeToggleSide(r rune) bool {
	if r != ' ' || s.compose == nil {
		return false
	}

	var lineIndex, column int
	switch s.compose.Mode {
	case composeModeCustom:
		lineIndex, column = s.composeCurrentLinePosition()
	case composeModeAssetTemplate:
		lineIndex, column = s.composeCurrentTemplateLinePosition()
	default:
		return false
	}

	if lineIndex < 0 || column != 0 {
		return false
	}

	if s.compose.Lines[lineIndex].Side == sideDebit {
		s.compose.Lines[lineIndex].Side = sideCredit
	} else {
		s.compose.Lines[lineIndex].Side = sideDebit
	}
	s.clearErrorForCurrentField()
	return true
}

// composeIsDefinedField reports whether fieldID is among the current compose
// form's field definitions (i.e. not a line-level field).
func (s *Shell) composeIsDefinedField(fieldID string) bool {
	for _, f := range s.composeFieldDefinitions() {
		if f.ID == fieldID {
			return true
		}
	}
	return false
}
