package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
)

const (
	sideDebit        = "debit"
	sideCredit       = "credit"
	fieldAccountCode = "account_code"
)

type composeMode string

const (
	composeModeExpense       composeMode = "expense"
	composeModeIncome        composeMode = "income"
	composeModeCustom        composeMode = "custom"
	composeModeAccount       composeMode = "account"
	composeModeQuest         composeMode = "quest"
	composeModeLoot          composeMode = "loot"
	composeModeAsset         composeMode = "asset"
	composeModeAssetTemplate composeMode = "asset_template"
	composeModeCodex         composeMode = "codex"
	composeModeNotes         composeMode = "notes"
)

type composeLineState struct {
	Side        string
	AccountCode string
	Amount      string
	Memo        string
}

type composeState struct {
	Mode            composeMode
	PreviousSection Section
	CommandID       string
	ItemKey         string
	Title           string
	FieldIndex      int
	Fields          map[string]string
	FieldErrors     map[string]string
	Lines           []composeLineState
	GeneralError    string
	picker          *accountPickerState
	CodexFormID     string
	CodexTypeID     string
}

type accountPickerState struct {
	Options       []AccountOption
	Query         string
	Filtered      []AccountOption
	SelectedIndex int
	Scroll        int
}

type composeField struct {
	ID          string
	Label       string
	Placeholder string
}

func newExpenseCompose(previous Section, catalog *EntryCatalog) *composeState {
	if catalog == nil {
		catalog = &EntryCatalog{}
	}
	return &composeState{
		Mode:            composeModeExpense,
		PreviousSection: previous,
		CommandID:       "entry.expense.create",
		Fields: map[string]string{
			"date":                catalog.DefaultDate,
			"description":         "",
			"amount":              "",
			fieldAccountCode:      "",
			"offset_account_code": defaultAccountCode(catalog.FundingAccounts, "1000"),
			"memo":                "",
		},
		FieldErrors: make(map[string]string),
	}
}

func newIncomeCompose(previous Section, catalog *EntryCatalog) *composeState {
	if catalog == nil {
		catalog = &EntryCatalog{}
	}
	return &composeState{
		Mode:            composeModeIncome,
		PreviousSection: previous,
		CommandID:       "entry.income.create",
		Fields: map[string]string{
			"date":                catalog.DefaultDate,
			"description":         "",
			"amount":              "",
			fieldAccountCode:      "",
			"offset_account_code": defaultAccountCode(catalog.DepositAccounts, "1000"),
			"memo":                "",
		},
		FieldErrors: make(map[string]string),
	}
}

func newCustomCompose(previous Section) *composeState {
	return &composeState{
		Mode:            composeModeCustom,
		PreviousSection: previous,
		CommandID:       "entry.custom.create",
		Fields: map[string]string{
			"date":        "",
			"description": "",
		},
		FieldErrors: make(map[string]string),
		Lines: []composeLineState{
			{Side: sideDebit},
			{Side: sideCredit},
		},
	}
}

func newAccountCompose(previous Section) *composeState {
	return &composeState{
		Mode:            composeModeAccount,
		PreviousSection: previous,
		CommandID:       "account.create",
		Fields: map[string]string{
			"code":         "",
			"name":         "",
			"account_type": "",
		},
		FieldErrors: make(map[string]string),
	}
}

func newQuestCompose(previous Section, _ *EntryCatalog) *composeState {
	return &composeState{
		Mode:            composeModeQuest,
		PreviousSection: previous,
		CommandID:       "quest.create",
		Fields: map[string]string{
			"title":       "",
			"patron":      "",
			"description": "",
			"reward":      "0",
			"advance":     "0",
			"bonus":       "",
			"notes":       "",
			"status":      "offered",
			"accepted_on": "",
		},
		FieldErrors: make(map[string]string),
	}
}

func newLootCompose(previous Section) *composeState {
	return &composeState{
		Mode:            composeModeLoot,
		PreviousSection: previous,
		CommandID:       "loot.create",
		Fields: map[string]string{
			"name":     "",
			"source":   "",
			"quantity": "1",
			"holder":   "",
			"notes":    "",
		},
		FieldErrors: make(map[string]string),
	}
}

func newEditCompose(base *composeState, commandID, itemKey, title string, fields map[string]string) *composeState {
	base.CommandID = commandID
	base.ItemKey = strings.TrimSpace(itemKey)
	base.Title = title
	for key, value := range fields {
		base.Fields[key] = strings.TrimSpace(value)
	}
	return base
}

func newQuestEditCompose(previous Section, itemKey string, fields map[string]string, catalog *EntryCatalog) *composeState {
	return newEditCompose(newQuestCompose(previous, catalog), "quest.update", itemKey, "Edit Quest", fields)
}

func newAssetCompose(previous Section) *composeState {
	return &composeState{
		Mode:            composeModeAsset,
		PreviousSection: previous,
		CommandID:       "asset.create",
		Fields: map[string]string{
			"name":     "",
			"source":   "",
			"quantity": "1",
			"holder":   "",
			"notes":    "",
		},
		FieldErrors: make(map[string]string),
	}
}

func newAssetEditCompose(previous Section, itemKey string, fields map[string]string) *composeState {
	return newEditCompose(newAssetCompose(previous), "asset.update", itemKey, "Edit Asset", fields)
}

func newLootEditCompose(previous Section, itemKey string, fields map[string]string) *composeState {
	return newEditCompose(newLootCompose(previous), "loot.update", itemKey, "Edit Loot", fields)
}

func newAssetTemplateCompose(previous Section, itemKey string, lines []composeLineState) *composeState {
	if len(lines) == 0 {
		lines = []composeLineState{
			{Side: sideDebit},
			{Side: sideCredit},
		}
	}
	return &composeState{
		Mode:            composeModeAssetTemplate,
		PreviousSection: previous,
		CommandID:       "asset.template.save",
		ItemKey:         itemKey,
		Title:           "Edit Entry Template",
		Fields:          make(map[string]string),
		FieldErrors:     make(map[string]string),
		Lines:           lines,
	}
}

func newCodexCompose(formID, typeID, name string, previous Section) *composeState {
	form, _ := LookupCodexForm(formID)
	fields := make(map[string]string, len(form.Fields))
	for _, f := range form.Fields {
		fields[f.ID] = ""
	}
	fields["name"] = name
	return &composeState{
		Mode:            composeModeCodex,
		PreviousSection: previous,
		CommandID:       "codex.create",
		Fields:          fields,
		FieldErrors:     make(map[string]string),
		CodexFormID:     formID,
		CodexTypeID:     typeID,
	}
}

func newCodexEditCompose(previous Section, itemKey string, fields map[string]string) *composeState {
	formID := fields["_form_id"]
	typeID := fields["_type_id"]
	if formID == "" {
		formID = "npc"
	}
	if typeID == "" {
		typeID = "npc"
	}
	delete(fields, "_form_id")
	delete(fields, "_type_id")

	base := newCodexCompose(formID, typeID, "", previous)
	base.CommandID = "codex.update"
	base.ItemKey = strings.TrimSpace(itemKey)
	base.Title = "Edit Entry"
	for key, value := range fields {
		base.Fields[key] = strings.TrimSpace(value)
	}
	return base
}

func newCustomComposeFromTemplate(previous Section, catalog *EntryCatalog, assetName string, lines []composeLineState) *composeState {
	if catalog == nil {
		catalog = &EntryCatalog{}
	}
	return &composeState{
		Mode:            composeModeCustom,
		PreviousSection: previous,
		CommandID:       "entry.custom.create",
		Title:           assetName,
		Fields: map[string]string{
			"date":        catalog.DefaultDate,
			"description": assetName,
		},
		FieldErrors: make(map[string]string),
		Lines:       lines,
	}
}

func defaultAccountCode(options []AccountOption, preferred string) string {
	preferred = strings.TrimSpace(preferred)
	for index := range options {
		if options[index].Code == preferred {
			return preferred
		}
	}
	return ""
}

func (s *Shell) openCompose(mode composeMode) bool {
	if s == nil {
		return false
	}
	switch mode {
	case composeModeExpense:
		s.compose = newExpenseCompose(s.Section, &s.Data.EntryCatalog)
	case composeModeIncome:
		s.compose = newIncomeCompose(s.Section, &s.Data.EntryCatalog)
	case composeModeCustom:
		compose := newCustomCompose(s.Section)
		compose.Fields["date"] = s.Data.EntryCatalog.DefaultDate
		s.compose = compose
	case composeModeAccount:
		s.compose = newAccountCompose(s.Section)
	case composeModeQuest:
		s.compose = newQuestCompose(s.Section, &s.Data.EntryCatalog)
	case composeModeLoot:
		s.compose = newLootCompose(s.Section)
	case composeModeAsset:
		s.compose = newAssetCompose(s.Section)
	case composeModeCodex:
		return false
	case composeModeNotes:
		s.openEditor()
		return true
	default:
		return false
	}
	return true
}

func (s *Shell) openComposeFromAction(itemKey string, action *ItemActionData) bool {
	if s == nil {
		return false
	}
	if action == nil {
		return false
	}

	switch strings.TrimSpace(action.ComposeMode) {
	case "quest":
		s.compose = newQuestEditCompose(s.Section, itemKey, action.ComposeFields, &s.Data.EntryCatalog)
	case "loot":
		s.compose = newLootEditCompose(s.Section, itemKey, action.ComposeFields)
	case "asset":
		s.compose = newAssetEditCompose(s.Section, itemKey, action.ComposeFields)
	case "asset_template":
		lines := commandLinesToComposeLines(action.ComposeLines)
		s.compose = newAssetTemplateCompose(s.Section, itemKey, lines)
	case "codex":
		s.compose = newCodexEditCompose(s.Section, itemKey, action.ComposeFields)
	case "notes":
		s.openEditorFromAction(itemKey, action)
		return true
	case "custom_from_template":
		lines := commandLinesToComposeLines(action.ComposeLines)
		s.compose = newCustomComposeFromTemplate(s.Section, &s.Data.EntryCatalog, strings.TrimSpace(action.ComposeTitle), lines)
	default:
		return false
	}

	if strings.TrimSpace(action.ComposeTitle) != "" {
		s.compose.Title = strings.TrimSpace(action.ComposeTitle)
	}
	return true
}

func commandLinesToComposeLines(lines []CommandLine) []composeLineState {
	if len(lines) == 0 {
		return nil
	}
	result := make([]composeLineState, len(lines))
	for i, line := range lines {
		result[i] = composeLineState{
			Side:        strings.TrimSpace(line.Side),
			AccountCode: strings.TrimSpace(line.AccountCode),
			Amount:      strings.TrimSpace(line.Amount),
			Memo:        strings.TrimSpace(line.Memo),
		}
	}
	return result
}

func (s *Shell) openComposeForAction(action Action) bool {
	switch action {
	case ActionNewExpense:
		if s.Section != SectionDashboard && s.Section != SectionJournal {
			return false
		}
		return s.openCompose(composeModeExpense)
	case ActionNewIncome:
		if s.Section != SectionDashboard && s.Section != SectionJournal {
			return false
		}
		return s.openCompose(composeModeIncome)
	case ActionNewCustom:
		switch s.Section {
		case SectionAccounts:
			return s.openCompose(composeModeAccount)
		case SectionQuests:
			return s.openCompose(composeModeQuest)
		case SectionLoot:
			return s.openCompose(composeModeLoot)
		case SectionAssets:
			return s.openCompose(composeModeAsset)
		case SectionCodex:
			return s.openCodexPicker()
		case SectionNotes:
			return s.openCompose(composeModeNotes)
		case SectionJournal:
			return false
		default:
			return s.openCompose(composeModeCustom)
		}
	default:
		return false
	}
}

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
		if s.compose.Mode == composeModeCustom {
			lineIndex, column := s.composeCurrentLinePosition()
			if lineIndex >= 0 && column == 0 && event.Rune() == ' ' {
				if s.compose.Lines[lineIndex].Side == sideDebit {
					s.compose.Lines[lineIndex].Side = sideCredit
				} else {
					s.compose.Lines[lineIndex].Side = sideDebit
				}
				s.clearErrorForCurrentField()
				return handleResult{Redraw: true}, true
			}
		}
		if s.compose.Mode == composeModeAssetTemplate {
			lineIndex, column := s.composeCurrentTemplateLinePosition()
			if lineIndex >= 0 && column == 0 && event.Rune() == ' ' {
				if s.compose.Lines[lineIndex].Side == sideDebit {
					s.compose.Lines[lineIndex].Side = sideCredit
				} else {
					s.compose.Lines[lineIndex].Side = sideDebit
				}
				s.clearErrorForCurrentField()
				return handleResult{Redraw: true}, true
			}
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
	case ActionNone, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowAccounts, ActionShowJournal, ActionShowQuests, ActionShowLoot, ActionShowAssets, ActionShowCodex, ActionShowNotes,
		ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
		ActionEdit, ActionDelete, ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionAppraise, ActionRecognize, ActionSell, ActionTransfer,
		ActionEditTemplate, ActionExecuteTemplate,
		ActionNewExpense, ActionNewIncome, ActionNewCustom:
		return handleResult{}, false
	case ActionHelp:
		return handleResult{}, true
	}

	return handleResult{}, true
}

func (s *Shell) composeFieldDefinitions() []composeField {
	if s.compose == nil {
		return nil
	}

	switch s.compose.Mode {
	case composeModeExpense:
		return []composeField{
			{ID: "date", Label: "Date", Placeholder: "YYYY-MM-DD"},
			{ID: "description", Label: "Description", Placeholder: "Restock arrows"},
			{ID: "amount", Label: "Amount", Placeholder: "2SP5CP"},
			{ID: fieldAccountCode, Label: "Expense account", Placeholder: "5100"},
			{ID: "offset_account_code", Label: "Paid from", Placeholder: "1000"},
			{ID: "memo", Label: "Memo", Placeholder: "Quiver refill"},
		}
	case composeModeIncome:
		return []composeField{
			{ID: "date", Label: "Date", Placeholder: "YYYY-MM-DD"},
			{ID: "description", Label: "Description", Placeholder: "Goblin bounty"},
			{ID: "amount", Label: "Amount", Placeholder: "25GP"},
			{ID: fieldAccountCode, Label: "Income account", Placeholder: "4000"},
			{ID: "offset_account_code", Label: "Deposit to", Placeholder: "1000"},
			{ID: "memo", Label: "Memo", Placeholder: "Mayor payout"},
		}
	case composeModeAccount:
		return []composeField{
			{ID: "code", Label: "Code", Placeholder: "5600"},
			{ID: "name", Label: "Name", Placeholder: "Tavern Reparations"},
			{ID: "account_type", Label: "Type", Placeholder: "asset|liability|equity|income|expense"},
		}
	case composeModeQuest:
		if s.compose != nil && s.compose.CommandID == "quest.update" {
			return []composeField{
				{ID: "title", Label: "Title", Placeholder: "Clear the Goblin Cave"},
				{ID: "patron", Label: "Patron", Placeholder: "Mayor Rowan"},
				{ID: "description", Label: "Description", Placeholder: "Optional quest notes"},
				{ID: "reward", Label: "Reward", Placeholder: "25GP"},
				{ID: "advance", Label: "Advance", Placeholder: "0"},
				{ID: "bonus", Label: "Bonus", Placeholder: "Optional bonus terms"},
				{ID: "notes", Label: "Notes", Placeholder: "Optional register notes"},
				{ID: "accepted_on", Label: "Accepted on", Placeholder: "YYYY-MM-DD"},
			}
		}
		return []composeField{
			{ID: "title", Label: "Title", Placeholder: "Clear the Goblin Cave"},
			{ID: "patron", Label: "Patron", Placeholder: "Mayor Rowan"},
			{ID: "description", Label: "Description", Placeholder: "Optional quest notes"},
			{ID: "reward", Label: "Reward", Placeholder: "25GP"},
			{ID: "advance", Label: "Advance", Placeholder: "0"},
			{ID: "bonus", Label: "Bonus", Placeholder: "Optional bonus terms"},
			{ID: "notes", Label: "Notes", Placeholder: "Optional register notes"},
			{ID: "status", Label: "Status", Placeholder: "offered|accepted"},
			{ID: "accepted_on", Label: "Accepted on", Placeholder: "YYYY-MM-DD"},
		}
	case composeModeLoot:
		return []composeField{
			{ID: "name", Label: "Name", Placeholder: "Silver Chalice"},
			{ID: "source", Label: "Source", Placeholder: "Goblin den"},
			{ID: "quantity", Label: "Quantity", Placeholder: "1"},
			{ID: "holder", Label: "Holder", Placeholder: "Bard"},
			{ID: "notes", Label: "Notes", Placeholder: "Optional item notes"},
		}
	case composeModeAsset:
		return []composeField{
			{ID: "name", Label: "Name", Placeholder: "Staff of the Magi"},
			{ID: "source", Label: "Source", Placeholder: "Ancient tomb"},
			{ID: "quantity", Label: "Quantity", Placeholder: "1"},
			{ID: "holder", Label: "Holder", Placeholder: "Wizard"},
			{ID: "notes", Label: "Notes", Placeholder: "Optional item notes"},
		}
	case composeModeCodex:
		form, ok := LookupCodexForm(s.compose.CodexFormID)
		if !ok {
			return nil
		}
		fields := make([]composeField, len(form.Fields))
		for i, f := range form.Fields {
			fields[i] = composeField{ID: f.ID, Label: f.Label, Placeholder: f.Placeholder}
		}
		return fields
	case composeModeNotes:
		return []composeField{
			{ID: "title", Label: "Title", Placeholder: "Session 5"},
			{ID: "body", Label: "Body", Placeholder: "Met @person/Mayor Elra near @quest/Clear the Watchtower"},
		}
	case composeModeAssetTemplate:
		return nil
	default:
		return []composeField{
			{ID: "date", Label: "Date", Placeholder: "YYYY-MM-DD"},
			{ID: "description", Label: "Description", Placeholder: "Custom journal entry"},
		}
	}
}

func (s *Shell) composeFieldCount() int {
	if s.compose == nil {
		return 0
	}
	switch s.compose.Mode {
	case composeModeCustom:
		return len(s.composeFieldDefinitions()) + len(s.compose.Lines)*4
	case composeModeAssetTemplate:
		return len(s.compose.Lines) * 3
	default:
		return len(s.composeFieldDefinitions())
	}
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

	switch fieldID {
	case "side":
		return
	case "date", "description", "amount", fieldAccountCode, "offset_account_code", "memo",
		"code", "name", "account_type", "title", "patron", "reward", "advance", "bonus", "notes", "status", "accepted_on",
		"source", "quantity", "holder",
		"location", "faction", "disposition", "party_member",
		"body":
		s.compose.Fields[fieldID] = op(s.compose.Fields[fieldID])
		delete(s.compose.FieldErrors, fieldID)
	default:
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

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[:len(runes)-1])
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
	case composeModeQuest, composeModeLoot, composeModeAsset, composeModeCodex, composeModeNotes:
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

func (s *Shell) composeBalanceSummary() (string, bool, bool) {
	if s.compose == nil {
		return "", false, false
	}
	if s.compose.Mode != composeModeCustom && s.compose.Mode != composeModeAssetTemplate {
		return "", false, false
	}

	var debitTotal, creditTotal int64
	allFilled := true
	for _, line := range s.compose.Lines {
		amt := strings.TrimSpace(line.Amount)
		if amt == "" {
			allFilled = false
			continue
		}
		parsed, err := currency.ParseAmount(amt)
		if err != nil {
			allFilled = false
			continue
		}
		if line.Side == sideDebit {
			debitTotal += parsed
		} else {
			creditTotal += parsed
		}
	}

	balanced := allFilled && debitTotal == creditTotal && len(s.compose.Lines) >= 2
	summary := fmt.Sprintf("Dr %s / Cr %s", currency.FormatAmount(debitTotal), currency.FormatAmount(creditTotal))
	if !allFilled {
		summary += "  (incomplete)"
	} else if balanced {
		summary += "  BALANCED"
	} else {
		summary += "  UNBALANCED"
	}
	return summary, balanced, allFilled
}

func (s *Shell) composeCommand() (*Command, bool) {
	if s.compose == nil {
		return nil, false
	}
	s.compose.FieldErrors = make(map[string]string)
	s.compose.GeneralError = ""

	command := &Command{
		Section: s.Section,
		Fields:  make(map[string]string),
	}

	command.ID = strings.TrimSpace(s.compose.CommandID)
	if command.ID == "" {
		switch s.compose.Mode {
		case composeModeExpense:
			command.ID = "entry.expense.create"
		case composeModeIncome:
			command.ID = "entry.income.create"
		case composeModeAccount:
			command.ID = "account.create"
		case composeModeQuest:
			command.ID = "quest.create"
		case composeModeLoot:
			command.ID = "loot.create"
		case composeModeAsset:
			command.ID = "asset.create"
		case composeModeCodex:
			command.ID = "codex.create"
		case composeModeNotes:
			command.ID = "notes.create"
		default:
			command.ID = "entry.custom.create"
		}
	}
	command.ItemKey = strings.TrimSpace(s.compose.ItemKey)

	for key, value := range s.compose.Fields {
		command.Fields[key] = strings.TrimSpace(value)
	}

	if s.compose.Mode == composeModeCodex {
		command.Fields["_type_id"] = s.compose.CodexTypeID
		command.Fields["_form_id"] = s.compose.CodexFormID
	}

	required := []string{"date", "description"}
	switch s.compose.Mode {
	case composeModeExpense, composeModeIncome:
		required = append(required, "amount", fieldAccountCode, "offset_account_code")
	case composeModeAccount:
		required = []string{"code", "name", "account_type"}
	case composeModeQuest:
		required = []string{"title", "reward", "advance"}
		if strings.EqualFold(command.Fields["status"], "accepted") {
			required = append(required, "accepted_on")
		}
	case composeModeLoot:
		required = []string{"name", "quantity"}
	case composeModeAsset:
		required = []string{"name", "quantity"}
	case composeModeCodex:
		required = []string{"name"}
	case composeModeNotes:
		required = []string{"title"}
	case composeModeCustom:
	case composeModeAssetTemplate:
		required = nil
	}
	for _, key := range required {
		if strings.TrimSpace(command.Fields[key]) == "" {
			s.compose.FieldErrors[key] = "Required."
		}
	}

	switch s.compose.Mode {
	case composeModeCustom:
		if len(s.compose.Lines) < 2 {
			s.compose.GeneralError = "Custom entry must contain at least 2 lines."
		}
		for index := range s.compose.Lines {
			line := s.compose.Lines[index]
			if strings.TrimSpace(line.AccountCode) == "" {
				s.compose.FieldErrors[s.composeLineFieldKey(index, 1)] = "Required."
			}
			if strings.TrimSpace(line.Amount) == "" {
				s.compose.FieldErrors[s.composeLineFieldKey(index, 2)] = "Required."
			}
			command.Lines = append(command.Lines, CommandLine{
				Side:        strings.TrimSpace(line.Side),
				AccountCode: strings.TrimSpace(line.AccountCode),
				Amount:      strings.TrimSpace(line.Amount),
				Memo:        strings.TrimSpace(line.Memo),
			})
		}
	case composeModeAssetTemplate:
		if len(s.compose.Lines) < 2 {
			s.compose.GeneralError = "Template must contain at least 2 lines."
		}
		for index := range s.compose.Lines {
			line := s.compose.Lines[index]
			if strings.TrimSpace(line.AccountCode) == "" {
				s.compose.FieldErrors[s.composeLineFieldKey(index, 1)] = "Required."
			}
			command.Lines = append(command.Lines, CommandLine{
				Side:        strings.TrimSpace(line.Side),
				AccountCode: strings.TrimSpace(line.AccountCode),
				Amount:      strings.TrimSpace(line.Amount),
			})
		}
	default:
	}

	if len(s.compose.FieldErrors) > 0 || strings.TrimSpace(s.compose.GeneralError) != "" {
		return nil, false
	}

	return command, true
}

func (s *Shell) applyComposeInputError(message string) {
	if s == nil || s.compose == nil {
		return
	}
	s.compose.GeneralError = strings.TrimSpace(message)
}

func (s *Shell) renderCompose(buffer *Buffer, rect Rect, theme *Theme) {
	if s.compose == nil || rect.Empty() {
		return
	}

	leftWidth := max(38, rect.W/2)
	left, right := rect.SplitVertical(leftWidth, 1)
	gapX := left.X + left.W
	if gapX < right.X {
		buffer.FillRect(Rect{X: gapX, Y: rect.Y, W: right.X - gapX, H: rect.H}, ' ', theme.Panel)
	}
	ss := s.Section.Style(theme)
	DrawPanel(buffer, left, theme, ss.Panel(s.composeTitle(), s.composeFormLines()))
	DrawPanel(buffer, right, theme, ss.Panel("Preview", s.composePreviewLines()))

	// Overlay balance summary line with color.
	if balText, balanced, allFilled := s.composeBalanceSummary(); balText != "" {
		content := right.Inset(1)
		if !content.Empty() {
			var style tcell.Style
			switch {
			case !allFilled:
				style = theme.Muted
			case balanced:
				style = theme.StatusOK
			default:
				style = theme.StatusError
			}
			buffer.WriteString(content.X, content.Y, style, clipText(balText, content.W))
		}
	}

	if s.compose.picker != nil {
		s.renderAccountPicker(buffer, rect, theme)
	}
}

func (s *Shell) renderAccountPicker(buffer *Buffer, rect Rect, theme *Theme) {
	p := s.compose.picker
	if p == nil || rect.Empty() {
		return
	}

	maxVisible := 10
	viewH := min(maxVisible, max(1, len(p.Filtered)))
	totalH := viewH + 6 // border(2) + search(1) + gap(1) + viewH + help(1) + gap(1)
	width := clampInt(rect.W/2, 36, 56)

	modalRect := Rect{
		X: rect.X + (rect.W-width)/2,
		Y: rect.Y + (rect.H-totalH)/2,
		W: width,
		H: totalH,
	}
	modalRect = modalRect.Intersect(rect)
	if modalRect.Empty() {
		return
	}

	accent := s.sectionStyle(theme)
	DrawPanel(buffer, modalRect, theme, Panel{
		Title:       "Pick Account",
		BorderStyle: &accent,
		TitleStyle:  &accent,
		Texture:     PanelTextureNone,
	})

	content := panelContentRect(modalRect, buffer.Bounds())
	if content.Empty() {
		return
	}

	y := content.Y
	searchText := "/ " + p.Query + "_"
	buffer.WriteString(content.X, y, theme.Text, clipText(searchText, content.W))
	y++

	// Adjust scroll to keep selection visible.
	if p.SelectedIndex < p.Scroll {
		p.Scroll = p.SelectedIndex
	}
	if p.SelectedIndex >= p.Scroll+viewH {
		p.Scroll = p.SelectedIndex - viewH + 1
	}
	maxScroll := max(0, len(p.Filtered)-viewH)
	p.Scroll = clampInt(p.Scroll, 0, maxScroll)

	if len(p.Filtered) == 0 {
		buffer.WriteString(content.X, y, theme.Muted, clipText("  No matching accounts.", content.W))
	} else {
		for row := 0; row < viewH && p.Scroll+row < len(p.Filtered); row++ {
			idx := p.Scroll + row
			opt := p.Filtered[idx]
			lineRect := Rect{X: content.X, Y: y + row, W: content.W, H: 1}
			style := theme.Text
			prefix := "  "
			if idx == p.SelectedIndex {
				buffer.FillRect(lineRect, ' ', theme.SelectedRow)
				style = theme.SelectedRow
				prefix = "> "
			}
			label := fmt.Sprintf("%s%s %s (%s)", prefix, opt.Code, opt.Name, opt.Type)
			buffer.WriteString(content.X, y+row, style, clipText(label, content.W))
		}
	}

	helpY := content.Y + content.H - 1
	buffer.WriteString(content.X, helpY, theme.Muted, clipText("↑↓ select  Enter pick  Esc cancel", content.W))
}

func (s *Shell) composeTitle() string {
	switch s.compose.Mode {
	case composeModeExpense:
		return "Guided Expense Entry"
	case composeModeIncome:
		return "Guided Income Entry"
	case composeModeAccount:
		return "Add Account"
	case composeModeQuest:
		if strings.TrimSpace(s.compose.Title) != "" {
			return s.compose.Title
		}
		return "Add Quest"
	case composeModeLoot:
		if strings.TrimSpace(s.compose.Title) != "" {
			return s.compose.Title
		}
		return "Add Loot"
	case composeModeAsset:
		if strings.TrimSpace(s.compose.Title) != "" {
			return s.compose.Title
		}
		return "Add Asset"
	case composeModeCustom:
		if strings.TrimSpace(s.compose.Title) != "" {
			return s.compose.Title
		}
		return "Custom Journal Entry"
	case composeModeAssetTemplate:
		if strings.TrimSpace(s.compose.Title) != "" {
			return s.compose.Title
		}
		return "Edit Entry Template"
	case composeModeCodex:
		if strings.TrimSpace(s.compose.Title) != "" {
			return s.compose.Title
		}
		return "Add to Codex"
	case composeModeNotes:
		if strings.TrimSpace(s.compose.Title) != "" {
			return s.compose.Title
		}
		return "Add Note"
	}

	return "Compose"
}

func (s *Shell) composeFormLines() []string {
	lines := make([]string, 0, 24)
	fields := s.composeFieldDefinitions()
	for index := range fields {
		field := fields[index]
		value := strings.TrimSpace(s.compose.Fields[field.ID])
		if value == "" {
			value = "[" + field.Placeholder + "]"
		}
		prefix := "  "
		if s.compose.FieldIndex == index {
			prefix = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s: %s", prefix, field.Label, value))
		if errText := strings.TrimSpace(s.compose.FieldErrors[field.ID]); errText != "" {
			lines = append(lines, "   Error: "+errText)
		}
	}

	switch s.compose.Mode {
	case composeModeCustom:
		lines = append(lines, "", "Lines:")
		for index := range s.compose.Lines {
			line := s.compose.Lines[index]
			for column, label := range []string{"Side", "Account", "Amount", "Memo"} {
				prefix := "  "
				if s.compose.FieldIndex == len(fields)+index*4+column {
					prefix = "> "
				}
				value := []string{line.Side, line.AccountCode, line.Amount, line.Memo}[column]
				if strings.TrimSpace(value) == "" {
					switch column {
					case 0:
						value = "[debit|credit]"
					case 1:
						value = "[account code]"
					case 2:
						value = "[amount]"
					default:
						value = "[memo]"
					}
				}
				lines = append(lines, fmt.Sprintf("%sL%d %s: %s", prefix, index+1, label, value))
				if errText := strings.TrimSpace(s.compose.FieldErrors[s.composeLineFieldKey(index, column)]); errText != "" {
					lines = append(lines, "   Error: "+errText)
				}
			}
		}
	case composeModeAssetTemplate:
		lines = append(lines, "Lines:")
		for index := range s.compose.Lines {
			line := s.compose.Lines[index]
			for column, label := range []string{"Side", "Account", "Amount"} {
				prefix := "  "
				if s.compose.FieldIndex == index*3+column {
					prefix = "> "
				}
				value := []string{sideHint(line.Side), line.AccountCode, line.Amount}[column]
				if strings.TrimSpace(value) == "" {
					switch column {
					case 0:
						value = "[debit|credit]"
					case 1:
						value = "[account code]"
					default:
						value = "[amount]"
					}
				}
				lines = append(lines, fmt.Sprintf("%sL%d %s: %s", prefix, index+1, label, value))
				if errText := strings.TrimSpace(s.compose.FieldErrors[s.composeLineFieldKey(index, column)]); errText != "" {
					lines = append(lines, "   Error: "+errText)
				}
			}
		}
	default:
	}

	if strings.TrimSpace(s.compose.GeneralError) != "" {
		lines = append(lines, "", "Error: "+s.compose.GeneralError)
	}
	lines = append(lines, "", s.composeHelpText())
	return lines
}

func (s *Shell) composeHelpText() string {
	if s.compose.picker != nil {
		return "Type to filter  ↑↓ select  Enter pick  Esc cancel"
	}
	pickerHint := ""
	if s.pickerAccountsForCurrentField() != nil {
		pickerHint = "  Ctrl+A pick account"
	}
	switch s.compose.Mode {
	case composeModeCustom, composeModeAssetTemplate:
		return "Tab/arrows move  Enter submit  Ctrl+N add  Ctrl+D del  Space toggle" + pickerHint + "  Esc cancel"
	default:
		return "Tab/arrows move  Enter submit" + pickerHint + "  Esc cancel"
	}
}

func (s *Shell) composePreviewLines() []string {
	if s.compose == nil {
		return nil
	}
	lines := []string{}

	switch s.compose.Mode {
	case composeModeExpense:
		lines = append(lines,
			"Date: "+displayComposeValue(s.compose.Fields["date"], "YYYY-MM-DD"),
			"Description: "+displayComposeValue(s.compose.Fields["description"], "required"),
			"Dr "+displayComposeValue(s.compose.Fields[fieldAccountCode], "expense account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"Cr "+displayComposeValue(s.compose.Fields["offset_account_code"], "funding account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"",
			"Expense accounts:",
		)
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.ExpenseAccounts)...)
		lines = append(lines, "", "Funding accounts:")
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.FundingAccounts)...)
	case composeModeIncome:
		lines = append(lines,
			"Date: "+displayComposeValue(s.compose.Fields["date"], "YYYY-MM-DD"),
			"Description: "+displayComposeValue(s.compose.Fields["description"], "required"),
			"Dr "+displayComposeValue(s.compose.Fields["offset_account_code"], "deposit account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"Cr "+displayComposeValue(s.compose.Fields[fieldAccountCode], "income account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"",
			"Income accounts:",
		)
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.IncomeAccounts)...)
		lines = append(lines, "", "Deposit accounts:")
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.DepositAccounts)...)
	case composeModeAccount:
		lines = append(lines,
			"Code: "+displayComposeValue(s.compose.Fields["code"], "required"),
			"Name: "+displayComposeValue(s.compose.Fields["name"], "required"),
			"Type: "+displayComposeValue(s.compose.Fields["account_type"], "required"),
			"",
			"Valid account types:",
			"asset  liability  equity  income  expense",
		)
	case composeModeQuest:
		statusLabel := "Status: " + displayComposeValue(s.compose.Fields["status"], "offered")
		statusHelpA := "Quest statuses:"
		statusHelpB := "offered  accepted"
		if s.compose.CommandID == "quest.update" {
			statusLabel = "Current status: " + displayComposeValue(s.compose.Fields["status"], "unknown")
			statusHelpA = "Quest lifecycle state is not edited here."
			statusHelpB = "Reward, advance, and accepted date may only change while the quest is offered or accepted."
		}
		lines = append(lines,
			"Title: "+displayComposeValue(s.compose.Fields["title"], "required"),
			"Patron: "+displayComposeValue(s.compose.Fields["patron"], "optional"),
			"Reward: "+displayComposeValue(s.compose.Fields["reward"], "0"),
			"Advance: "+displayComposeValue(s.compose.Fields["advance"], "0"),
			"Notes: "+displayComposeValue(s.compose.Fields["notes"], "optional"),
			statusLabel,
			"Accepted on: "+displayComposeValue(s.compose.Fields["accepted_on"], "YYYY-MM-DD"),
			"",
			statusHelpA,
			statusHelpB,
		)
	case composeModeLoot:
		lines = append(lines,
			"Name: "+displayComposeValue(s.compose.Fields["name"], "required"),
			"Source: "+displayComposeValue(s.compose.Fields["source"], "optional"),
			"Quantity: "+displayComposeValue(s.compose.Fields["quantity"], "1"),
			"Holder: "+displayComposeValue(s.compose.Fields["holder"], "optional"),
			"Notes: "+displayComposeValue(s.compose.Fields["notes"], "optional"),
			"",
			"Loot is created as held and off-ledger.",
			"No value is entered here. Appraise it later.",
		)
	case composeModeAsset:
		lines = append(lines,
			"Name: "+displayComposeValue(s.compose.Fields["name"], "required"),
			"Source: "+displayComposeValue(s.compose.Fields["source"], "optional"),
			"Quantity: "+displayComposeValue(s.compose.Fields["quantity"], "1"),
			"Holder: "+displayComposeValue(s.compose.Fields["holder"], "optional"),
			"Notes: "+displayComposeValue(s.compose.Fields["notes"], "optional"),
			"",
			"Asset is created as held and off-ledger.",
			"Transfer to loot register when ready to sell.",
		)
	case composeModeCodex:
		form, ok := LookupCodexForm(s.compose.CodexFormID)
		if ok {
			for _, f := range form.Fields {
				req := "optional"
				if f.ID == "name" {
					req = "required"
				}
				lines = append(lines, f.Label+": "+displayComposeValue(s.compose.Fields[f.ID], req))
			}
		}
		lines = append(lines,
			"",
			"Use @type/name in notes for cross-references:",
			"@quest/Name, @loot/Name, @asset/Name, @person/Name",
		)
	case composeModeNotes:
		lines = append(lines,
			"Title: "+displayComposeValue(s.compose.Fields["title"], "required"),
			"Body: "+displayComposeValue(s.compose.Fields["body"], "optional"),
			"",
			"Use @type/name in body for cross-references:",
			"@quest/Name, @loot/Name, @asset/Name, @person/Name, @note/Name",
		)
	case composeModeAssetTemplate:
		if balText, _, _ := s.composeBalanceSummary(); balText != "" {
			lines = append(lines, balText, "")
		}
		lines = append(lines, "Template structure:", "")
		var debitCount, creditCount int
		for index := range s.compose.Lines {
			line := s.compose.Lines[index]
			sideLabel := "Dr (to)"
			if line.Side == sideCredit {
				sideLabel = "Cr (from)"
			}
			amtPart := ""
			if strings.TrimSpace(line.Amount) != "" {
				amtPart = " " + line.Amount
			}
			lines = append(lines, fmt.Sprintf("L%d %s %s%s", index+1, sideLabel, displayComposeValue(line.AccountCode, "account"), amtPart))
			if line.Side == sideDebit {
				debitCount++
			} else {
				creditCount++
			}
		}
		lines = append(lines,
			"", fmt.Sprintf("Line counts: %d debit / %d credit", debitCount, creditCount),
			"", "Dr (to)   = money flows into this account",
			"Cr (from) = money flows out of this account",
			"Amounts are optional; blank amounts are entered at execution.",
			"", "Active accounts:")
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.AllAccounts)...)
	default:
		if balText, _, _ := s.composeBalanceSummary(); balText != "" {
			lines = append(lines, balText, "")
		}
		lines = append(lines,
			"Date: "+displayComposeValue(s.compose.Fields["date"], "YYYY-MM-DD"),
			"Description: "+displayComposeValue(s.compose.Fields["description"], "required"),
			"",
		)
		var debitCount, creditCount int
		lines = append(lines, "Lines:")
		for index := range s.compose.Lines {
			line := s.compose.Lines[index]
			lines = append(lines, fmt.Sprintf("L%d %s %s %s", index+1, displayComposeValue(line.Side, "side"), displayComposeValue(line.AccountCode, "account"), displayComposeValue(line.Amount, "amount")))
			if line.Side == sideDebit {
				debitCount++
			}
			if line.Side == sideCredit {
				creditCount++
			}
		}
		lines = append(lines, "", fmt.Sprintf("Line counts: %d debit / %d credit", debitCount, creditCount), "", "Active accounts:")
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.AllAccounts)...)
	}

	return lines
}

func sideHint(side string) string {
	if side == sideDebit {
		return "debit (to)"
	}
	return "credit (from)"
}

func displayComposeValue(value string, placeholder string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "[" + placeholder + "]"
	}
	return value
}

func accountOptionLines(options []AccountOption) []string {
	if len(options) == 0 {
		return []string{"No matching active accounts."}
	}
	lines := make([]string, 0, min(len(options), 10))
	for index := 0; index < len(options) && index < 10; index++ {
		lines = append(lines, fmt.Sprintf("%s %s (%s)", options[index].Code, options[index].Name, options[index].Type))
	}
	if len(options) > 10 {
		lines = append(lines, fmt.Sprintf("... %d more", len(options)-10))
	}
	return lines
}
