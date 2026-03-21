package render

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// HandleAction updates shell state for a semantic key action.
func (s *Shell) HandleAction(action Action) HandleResult { //nolint:revive // large action dispatch
	if s == nil {
		return HandleResult{}
	}

	if s.editor != nil {
		switch action {
		case ActionRedraw:
			s.editor = nil
			return HandleResult{Reload: true}
		default:
			return HandleResult{}
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
		case ActionNone, ActionConfirm, ActionHelp, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowSettings, ActionShowLedger, ActionShowJournal, ActionShowQuests, ActionShowLoot, ActionShowAssets, ActionShowCodex,
			ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
			ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer,
			ActionEditTemplate, ActionExecuteTemplate,
			ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionSubmitCompose,
			ActionShowNotes, ActionSearch:
			return HandleResult{}
		case ActionQuit:
			s.compose = nil
			return HandleResult{Redraw: true}
		case ActionRedraw:
			s.compose = nil
			return HandleResult{Reload: true}
		}
	}

	if s.confirm != nil {
		return s.handleConfirmAction(action)
	}

	switch action {
	case ActionNone, ActionSubmitCompose:
		return HandleResult{}
	case ActionConfirm:
		if s.Section == SectionSettings && s.activeSettingsSection() == settingsTabCampaigns {
			item := s.currentSelectedItem(settingsTabCampaigns)
			if item != nil {
				return HandleResult{Command: &Command{
					ID:      "campaign.switch",
					ItemKey: item.Key,
				}}
			}
		}
		return HandleResult{}
	case ActionQuit:
		if s.Section == SectionSettings {
			s.Section = SectionDashboard
			return HandleResult{Redraw: true}
		}
		s.quitConfirm = true
		return HandleResult{Redraw: true}
	case ActionRedraw:
		return HandleResult{Reload: true}
	case ActionHelp:
		if s.toggleGlossary() {
			return HandleResult{Redraw: true}
		}
	case ActionNextSection:
		if s.Section == SectionSettings {
			s.settingsTab = (s.settingsTab + 1) % len(settingsTabs)
			s.reconcileSelection(s.activeSettingsSection())
			return HandleResult{Redraw: true}
		}
		s.Section = s.Section.Next()
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionPrevSection:
		if s.Section == SectionSettings {
			s.settingsTab = (s.settingsTab + len(settingsTabs) - 1) % len(settingsTabs)
			s.reconcileSelection(s.activeSettingsSection())
			return HandleResult{Redraw: true}
		}
		s.Section = s.Section.Previous()
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowDashboard:
		s.Section = SectionDashboard
		return HandleResult{Redraw: true}
	case ActionShowSettings:
		s.Section = SectionSettings
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowLedger:
		s.Section = SectionLedger
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowJournal:
		s.Section = SectionJournal
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowQuests:
		s.Section = SectionQuests
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowLoot:
		s.Section = SectionLoot
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowAssets:
		s.Section = SectionAssets
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowCodex:
		s.Section = SectionCodex
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionShowNotes:
		s.Section = SectionNotes
		s.reconcileSelection(s.Section)
		return HandleResult{Redraw: true}
	case ActionMoveUp:
		if s.moveSelection(-1) {
			return HandleResult{Redraw: true}
		}
	case ActionMoveDown:
		if s.moveSelection(1) {
			return HandleResult{Redraw: true}
		}
	case ActionPageUp:
		if s.moveSelection(-s.pageSize()) {
			return HandleResult{Redraw: true}
		}
	case ActionPageDown:
		if s.moveSelection(s.pageSize()) {
			return HandleResult{Redraw: true}
		}
	case ActionMoveTop:
		if s.moveSelectionTo(0) {
			return HandleResult{Redraw: true}
		}
	case ActionMoveBottom:
		if s.moveSelectionTo(1 << 30) {
			return HandleResult{Redraw: true}
		}
	case ActionNewExpense:
		if s.openComposeForAction(ActionNewExpense) {
			return HandleResult{Redraw: true}
		}
	case ActionNewIncome:
		if s.openComposeForAction(ActionNewIncome) {
			return HandleResult{Redraw: true}
		}
	case ActionNewCustom:
		if s.openComposeForAction(ActionNewCustom) {
			return HandleResult{Redraw: true}
		}
	case ActionSearch:
		if s.openSearch() {
			return HandleResult{Redraw: true}
		}
	case ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer, ActionEditTemplate, ActionExecuteTemplate:
		if s.openAction(action) {
			return HandleResult{Redraw: true}
		}
	default:
	}

	return HandleResult{}
}

// HandleKeyEvent updates shell state for raw key input when the shell needs more
// than semantic action mapping, such as text entry inside a modal.
func (s *Shell) HandleKeyEvent(event *tcell.EventKey, keymap KeyMap) HandleResult {
	if s == nil {
		return HandleResult{}
	}

	if s.disconnected {
		return HandleResult{Quit: true}
	}

	action := keymap.Resolve(event)

	if s.quitConfirm {
		switch action {
		case ActionConfirm:
			return HandleResult{Quit: true}
		case ActionQuit:
			s.quitConfirm = false
			return HandleResult{Redraw: true}
		default:
			return HandleResult{}
		}
	}
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

func (s *Shell) handleConfirmAction(action Action) HandleResult {
	switch action {
	case ActionQuit:
		s.confirm = nil
		return HandleResult{Redraw: true}
	case ActionConfirm:
		command := s.pendingCommand()
		s.confirm = nil
		if command == nil {
			return HandleResult{Redraw: true}
		}
		return HandleResult{Command: command}
	case ActionRedraw:
		s.confirm = nil
		return HandleResult{Reload: true}
	default:
		return HandleResult{}
	}
}

func (s *Shell) handleInputAction(action Action) HandleResult {
	switch action {
	case ActionNone, ActionHelp, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowSettings, ActionShowLedger, ActionShowJournal, ActionShowQuests, ActionShowLoot, ActionShowAssets, ActionShowCodex, ActionShowNotes,
		ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
		ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer,
		ActionEditTemplate, ActionExecuteTemplate,
		ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionSubmitCompose, ActionSearch:
		return HandleResult{}
	case ActionQuit:
		s.input = nil
		return HandleResult{Redraw: true}
	case ActionRedraw:
		s.input = nil
		return HandleResult{Reload: true}
	case ActionConfirm:
		return HandleResult{}
	}

	return HandleResult{}
}

func (s *Shell) handleGlossaryAction(action Action) HandleResult {
	switch action {
	case ActionQuit, ActionHelp:
		s.glossary = nil
		return HandleResult{Redraw: true}
	case ActionRedraw:
		s.glossary = nil
		return HandleResult{Reload: true}
	default:
		return HandleResult{}
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

func (s *Shell) handleInputKeyEvent(event *tcell.EventKey, action Action) (HandleResult, bool) {
	if s.input == nil || event == nil {
		return HandleResult{}, false
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
			return HandleResult{Redraw: true}, true
		}

		command := s.pendingCommand()
		if command == nil {
			return HandleResult{Redraw: true}, true
		}

		return HandleResult{Command: command}, true
	default:
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		runes := []rune(s.input.Value)
		if len(runes) == 0 {
			return HandleResult{}, true
		}
		s.input.Value = string(runes[:len(runes)-1])
		s.input.ErrorText = ""
		return HandleResult{Redraw: true}, true
	case tcell.KeyCtrlU:
		if s.input.Value == "" && s.input.ErrorText == "" {
			return HandleResult{}, true
		}
		s.input.Value = ""
		s.input.ErrorText = ""
		return HandleResult{Redraw: true}, true
	case tcell.KeyRune:
		s.input.Value += string(event.Rune())
		s.input.ErrorText = ""
		return HandleResult{Redraw: true}, true
	default:
		return HandleResult{}, true
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
