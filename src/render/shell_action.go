package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// HandleAction updates shell state for a semantic key action.
func (s *Shell) HandleAction(action Action) handleResult {
	if s == nil {
		return handleResult{}
	}

	if s.editor != nil {
		switch action {
		case ActionRedraw:
			s.editor = nil
			return handleResult{Reload: true}
		default:
			return handleResult{}
		}
	}
	if s.input != nil {
		return s.handleInputAction(action)
	}
	if s.glossary != nil {
		return s.handleGlossaryAction(action)
	}
	if s.compose != nil {
		switch action {
		case ActionNone, ActionConfirm, ActionHelp, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowSettings, ActionShowJournal, ActionShowQuests, ActionShowLoot, ActionShowAssets, ActionShowCodex,
			ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
			ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer,
			ActionEditTemplate, ActionExecuteTemplate,
			ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionSubmitCompose,
			ActionShowNotes, ActionSearch:
			return handleResult{}
		case ActionQuit:
			s.compose = nil
			return handleResult{Redraw: true}
		case ActionRedraw:
			s.compose = nil
			return handleResult{Reload: true}
		}
	}

	if s.confirm != nil {
		return s.handleConfirmAction(action)
	}

	switch action {
	case ActionNone, ActionConfirm, ActionSubmitCompose:
		return handleResult{}
	case ActionQuit:
		if s.Section == SectionSettings {
			s.Section = SectionDashboard
			return handleResult{Redraw: true}
		}
		return handleResult{Quit: true}
	case ActionRedraw:
		return handleResult{Reload: true}
	case ActionHelp:
		if s.toggleGlossary() {
			return handleResult{Redraw: true}
		}
	case ActionNextSection:
		if s.Section == SectionSettings {
			s.settingsTab = (s.settingsTab + 1) % len(settingsTabs)
			s.reconcileSelection(s.activeSettingsSection())
			return handleResult{Redraw: true}
		}
		s.Section = s.Section.Next()
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionPrevSection:
		if s.Section == SectionSettings {
			s.settingsTab = (s.settingsTab + len(settingsTabs) - 1) % len(settingsTabs)
			s.reconcileSelection(s.activeSettingsSection())
			return handleResult{Redraw: true}
		}
		s.Section = s.Section.Previous()
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowDashboard:
		s.Section = SectionDashboard
		return handleResult{Redraw: true}
	case ActionShowSettings:
		s.Section = SectionSettings
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowJournal:
		s.Section = SectionJournal
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowQuests:
		s.Section = SectionQuests
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowLoot:
		s.Section = SectionLoot
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowAssets:
		s.Section = SectionAssets
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowCodex:
		s.Section = SectionCodex
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionShowNotes:
		s.Section = SectionNotes
		s.reconcileSelection(s.Section)
		return handleResult{Redraw: true}
	case ActionMoveUp:
		if s.moveSelection(-1) {
			return handleResult{Redraw: true}
		}
	case ActionMoveDown:
		if s.moveSelection(1) {
			return handleResult{Redraw: true}
		}
	case ActionPageUp:
		if s.moveSelection(-s.pageSize()) {
			return handleResult{Redraw: true}
		}
	case ActionPageDown:
		if s.moveSelection(s.pageSize()) {
			return handleResult{Redraw: true}
		}
	case ActionMoveTop:
		if s.moveSelectionTo(0) {
			return handleResult{Redraw: true}
		}
	case ActionMoveBottom:
		if s.moveSelectionTo(1 << 30) {
			return handleResult{Redraw: true}
		}
	case ActionNewExpense:
		if s.openComposeForAction(ActionNewExpense) {
			return handleResult{Redraw: true}
		}
	case ActionNewIncome:
		if s.openComposeForAction(ActionNewIncome) {
			return handleResult{Redraw: true}
		}
	case ActionNewCustom:
		if s.openComposeForAction(ActionNewCustom) {
			return handleResult{Redraw: true}
		}
	case ActionSearch:
		if s.openSearch() {
			return handleResult{Redraw: true}
		}
	case ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer, ActionEditTemplate, ActionExecuteTemplate:
		if s.openAction(action) {
			return handleResult{Redraw: true}
		}
	}

	return handleResult{}
}

// HandleKeyEvent updates shell state for raw key input when the shell needs more
// than semantic action mapping, such as text entry inside a modal.
func (s *Shell) HandleKeyEvent(event *tcell.EventKey, keymap KeyMap) handleResult {
	if s == nil {
		return handleResult{}
	}

	action := keymap.Resolve(event)
	if s.editor != nil {
		if result, handled := s.handleEditorKeyEvent(event, action); handled {
			return result
		}
	}
	if s.codexPicker != nil {
		if result, handled := s.handleCodexPickerKeyEvent(event, action); handled {
			return result
		}
	}
	if s.search != nil {
		if result, handled := s.handleSearchKeyEvent(event, action); handled {
			return result
		}
	}
	if s.input != nil {
		if result, handled := s.handleInputKeyEvent(event, action); handled {
			return result
		}
	}
	if s.compose != nil {
		if result, handled := s.handleComposeKeyEvent(event, action); handled {
			return result
		}
	}

	return s.HandleAction(action)
}

func (s *Shell) handleConfirmAction(action Action) handleResult {
	switch action {
	case ActionQuit:
		s.confirm = nil
		return handleResult{Redraw: true}
	case ActionConfirm:
		command := s.pendingCommand()
		s.confirm = nil
		if command == nil {
			return handleResult{Redraw: true}
		}
		return handleResult{Command: command}
	case ActionRedraw:
		s.confirm = nil
		return handleResult{Reload: true}
	default:
		return handleResult{}
	}
}

func (s *Shell) handleInputAction(action Action) handleResult {
	switch action {
	case ActionNone, ActionHelp, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowSettings, ActionShowJournal, ActionShowQuests, ActionShowLoot, ActionShowAssets, ActionShowCodex, ActionShowNotes,
		ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
		ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer,
		ActionEditTemplate, ActionExecuteTemplate,
		ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionSubmitCompose, ActionSearch:
		return handleResult{}
	case ActionQuit:
		s.input = nil
		return handleResult{Redraw: true}
	case ActionRedraw:
		s.input = nil
		return handleResult{Reload: true}
	case ActionConfirm:
		return handleResult{}
	}

	return handleResult{}
}

func (s *Shell) handleGlossaryAction(action Action) handleResult {
	switch action {
	case ActionQuit, ActionHelp:
		s.glossary = nil
		return handleResult{Redraw: true}
	case ActionRedraw:
		s.glossary = nil
		return handleResult{Reload: true}
	default:
		return handleResult{}
	}
}

func (s *Shell) toggleGlossary() bool {
	if s == nil {
		return false
	}
	if s.glossary != nil {
		s.glossary = nil
		return true
	}

	s.glossary = &glossaryState{
		Title: s.glossaryTitle(),
		Lines: s.glossaryLines(),
	}
	return true
}

func (s *Shell) handleInputKeyEvent(event *tcell.EventKey, action Action) (handleResult, bool) {
	if s.input == nil || event == nil {
		return handleResult{}, false
	}

	switch action {
	case ActionQuit, ActionRedraw:
		return s.handleInputAction(action), true
	case ActionConfirm:
		if strings.TrimSpace(s.input.Value) == "" {
			msg := s.input.RequiredMessage
			if msg == "" {
				msg = "Value is required."
			}
			s.input.ErrorText = msg
			return handleResult{Redraw: true}, true
		}

		command := s.pendingCommand()
		if command == nil {
			return handleResult{Redraw: true}, true
		}

		return handleResult{Command: command}, true
	default:
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		runes := []rune(s.input.Value)
		if len(runes) == 0 {
			return handleResult{}, true
		}
		s.input.Value = string(runes[:len(runes)-1])
		s.input.ErrorText = ""
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlU:
		if s.input.Value == "" && s.input.ErrorText == "" {
			return handleResult{}, true
		}
		s.input.Value = ""
		s.input.ErrorText = ""
		return handleResult{Redraw: true}, true
	case tcell.KeyRune:
		s.input.Value += string(event.Rune())
		s.input.ErrorText = ""
		return handleResult{Redraw: true}, true
	default:
		return handleResult{}, true
	}
}

// ApplyInputError updates the open input modal with a validation message.
func (s *Shell) ApplyInputError(message string) {
	if s == nil {
		return
	}
	if s.compose != nil {
		s.applyComposeInputError(message)
		return
	}
	if s.input == nil {
		return
	}
	s.input.ErrorText = strings.TrimSpace(message)
}

func (s *Shell) openAction(trigger Action) bool {
	item := s.currentSelectedItem(s.listSection())
	if item == nil || len(item.Actions) == 0 {
		return false
	}

	for index := range item.Actions {
		action := item.Actions[index]
		if action.Trigger != trigger {
			continue
		}

		switch action.Mode {
		case ItemActionModeCompose:
			return s.openComposeFromAction(item.Key, &action)
		case ItemActionModeInput:
			s.input = &inputState{
				Section:         s.Section,
				ItemKey:         item.Key,
				Action:          action,
				Title:           action.InputTitle,
				Prompt:          action.InputPrompt,
				RequiredMessage: action.InputRequired,
				Placeholder:     action.Placeholder,
				HelpLines:       append([]string{}, action.InputHelp...),
			}
		default:
			s.confirm = &confirmState{
				Section: s.Section,
				ItemKey: item.Key,
				Action:  action,
			}
		}
		return true
	}

	return false
}

func (s *Shell) pendingCommand() *Command {
	if s.compose != nil {
		command, ok := s.composeCommand()
		if !ok {
			return nil
		}
		return command
	}
	if s.confirm == nil {
		if s.input == nil {
			return nil
		}

		command := &Command{
			ID:      s.input.Action.ID,
			Section: s.input.Section,
			ItemKey: s.input.ItemKey,
		}
		if strings.TrimSpace(s.input.Value) != "" {
			command.Fields = map[string]string{
				"amount": s.input.Value,
			}
		}
		return command
	}

	command := &Command{
		ID:      s.confirm.Action.ID,
		Section: s.confirm.Section,
		ItemKey: s.confirm.ItemKey,
	}

	return command
}
