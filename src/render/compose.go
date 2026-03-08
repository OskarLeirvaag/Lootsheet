package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type composeMode string

const (
	composeModeExpense composeMode = "expense"
	composeModeIncome  composeMode = "income"
	composeModeCustom  composeMode = "custom"
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
	FieldIndex      int
	Fields          map[string]string
	FieldErrors     map[string]string
	Lines           []composeLineState
	GeneralError    string
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
		Fields: map[string]string{
			"date":                catalog.DefaultDate,
			"description":         "",
			"amount":              "",
			"account_code":        "",
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
		Fields: map[string]string{
			"date":                catalog.DefaultDate,
			"description":         "",
			"amount":              "",
			"account_code":        "",
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
		Fields: map[string]string{
			"date":        "",
			"description": "",
		},
		FieldErrors: make(map[string]string),
		Lines: []composeLineState{
			{Side: "debit"},
			{Side: "credit"},
		},
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
	default:
		return false
	}
	return true
}

func (s *Shell) handleComposeKeyEvent(event *tcell.EventKey, action Action) (handleResult, bool) {
	if s.compose == nil || event == nil {
		return handleResult{}, false
	}

	switch event.Key() {
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
	case tcell.KeyCtrlN:
		if s.compose.Mode == composeModeCustom && len(s.compose.Lines) < 8 {
			s.compose.Lines = append(s.compose.Lines, composeLineState{Side: "debit"})
			s.compose.FieldIndex = s.composeFieldCount() - 4
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyCtrlD:
		if s.compose.Mode == composeModeCustom && len(s.compose.Lines) > 2 {
			lineIndex, column := s.composeCurrentLinePosition()
			if lineIndex >= 0 && lineIndex < len(s.compose.Lines) {
				s.compose.Lines = append(s.compose.Lines[:lineIndex], s.compose.Lines[lineIndex+1:]...)
				if len(s.compose.Lines) == 0 {
					s.compose.Lines = []composeLineState{{Side: "debit"}, {Side: "credit"}}
				}
				if lineIndex >= len(s.compose.Lines) {
					lineIndex = len(s.compose.Lines) - 1
				}
				s.compose.FieldIndex = 2 + lineIndex*4 + column
			}
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyRune:
		if s.compose.Mode == composeModeCustom {
			lineIndex, column := s.composeCurrentLinePosition()
			if lineIndex >= 0 && column == 0 && event.Rune() == ' ' {
				if s.compose.Lines[lineIndex].Side == "debit" {
					s.compose.Lines[lineIndex].Side = "credit"
				} else {
					s.compose.Lines[lineIndex].Side = "debit"
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
	case ActionSubmitCompose:
		if command, ok := s.composeCommand(); ok {
			return handleResult{Command: command}, true
		}
		return handleResult{Redraw: true}, true
	case ActionNone, ActionNextSection, ActionPrevSection, ActionShowDashboard, ActionShowAccounts, ActionShowJournal, ActionShowQuests, ActionShowLoot,
		ActionMoveUp, ActionMoveDown, ActionPageUp, ActionPageDown, ActionMoveTop, ActionMoveBottom,
		ActionToggle, ActionReverse, ActionCollect, ActionWriteOff, ActionRecognize, ActionSell,
		ActionNewExpense, ActionNewIncome, ActionNewCustom, ActionConfirm:
		return handleResult{}, false
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
			{ID: "account_code", Label: "Expense account", Placeholder: "5100"},
			{ID: "offset_account_code", Label: "Paid from", Placeholder: "1000"},
			{ID: "memo", Label: "Memo", Placeholder: "Quiver refill"},
		}
	case composeModeIncome:
		return []composeField{
			{ID: "date", Label: "Date", Placeholder: "YYYY-MM-DD"},
			{ID: "description", Label: "Description", Placeholder: "Goblin bounty"},
			{ID: "amount", Label: "Amount", Placeholder: "25GP"},
			{ID: "account_code", Label: "Income account", Placeholder: "4000"},
			{ID: "offset_account_code", Label: "Deposit to", Placeholder: "1000"},
			{ID: "memo", Label: "Memo", Placeholder: "Mayor payout"},
		}
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
	if s.compose.Mode != composeModeCustom {
		return len(s.composeFieldDefinitions())
	}
	return len(s.composeFieldDefinitions()) + len(s.compose.Lines)*4
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
		return "account_code"
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

func (s *Shell) composeAppendRune(value rune) {
	if s.compose == nil {
		return
	}
	fieldID := s.composeCurrentFieldID()
	switch fieldID {
	case "side":
		return
	case "date", "description", "amount", "account_code", "offset_account_code", "memo":
		s.compose.Fields[fieldID] += string(value)
		delete(s.compose.FieldErrors, fieldID)
	default:
		lineIndex, column := s.composeCurrentLinePosition()
		if lineIndex < 0 || lineIndex >= len(s.compose.Lines) {
			return
		}
		switch column {
		case 1:
			s.compose.Lines[lineIndex].AccountCode += string(value)
		case 2:
			s.compose.Lines[lineIndex].Amount += string(value)
		case 3:
			s.compose.Lines[lineIndex].Memo += string(value)
		}
		delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
	}
	s.compose.GeneralError = ""
}

func (s *Shell) composeBackspace() {
	if s.compose == nil {
		return
	}
	fieldID := s.composeCurrentFieldID()
	switch fieldID {
	case "side":
		return
	case "date", "description", "amount", "account_code", "offset_account_code", "memo":
		s.compose.Fields[fieldID] = trimLastRune(s.compose.Fields[fieldID])
		delete(s.compose.FieldErrors, fieldID)
	default:
		lineIndex, column := s.composeCurrentLinePosition()
		if lineIndex < 0 || lineIndex >= len(s.compose.Lines) {
			return
		}
		switch column {
		case 1:
			s.compose.Lines[lineIndex].AccountCode = trimLastRune(s.compose.Lines[lineIndex].AccountCode)
		case 2:
			s.compose.Lines[lineIndex].Amount = trimLastRune(s.compose.Lines[lineIndex].Amount)
		case 3:
			s.compose.Lines[lineIndex].Memo = trimLastRune(s.compose.Lines[lineIndex].Memo)
		}
		delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
	}
	s.compose.GeneralError = ""
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[:len(runes)-1])
}

func (s *Shell) composeClearCurrent() {
	if s.compose == nil {
		return
	}
	fieldID := s.composeCurrentFieldID()
	switch fieldID {
	case "side":
		return
	case "date", "description", "amount", "account_code", "offset_account_code", "memo":
		s.compose.Fields[fieldID] = ""
		delete(s.compose.FieldErrors, fieldID)
	default:
		lineIndex, column := s.composeCurrentLinePosition()
		if lineIndex < 0 || lineIndex >= len(s.compose.Lines) {
			return
		}
		switch column {
		case 1:
			s.compose.Lines[lineIndex].AccountCode = ""
		case 2:
			s.compose.Lines[lineIndex].Amount = ""
		case 3:
			s.compose.Lines[lineIndex].Memo = ""
		}
		delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
	}
	s.compose.GeneralError = ""
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
		lineIndex, column := s.composeCurrentLinePosition()
		delete(s.compose.FieldErrors, s.composeLineFieldKey(lineIndex, column))
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

	switch s.compose.Mode {
	case composeModeExpense:
		command.ID = "entry.expense.create"
	case composeModeIncome:
		command.ID = "entry.income.create"
	default:
		command.ID = "entry.custom.create"
	}

	for key, value := range s.compose.Fields {
		command.Fields[key] = strings.TrimSpace(value)
	}

	required := []string{"date", "description"}
	if s.compose.Mode != composeModeCustom {
		required = append(required, "amount", "account_code", "offset_account_code")
	}
	for _, key := range required {
		if strings.TrimSpace(command.Fields[key]) == "" {
			s.compose.FieldErrors[key] = "Required."
		}
	}

	if s.compose.Mode == composeModeCustom {
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

	left, right := rect.SplitVertical(maxInt(38, rect.W/2), 1)
	DrawPanel(buffer, left, theme, Panel{
		Title: s.composeTitle(),
		Lines: s.composeFormLines(),
	})
	DrawPanel(buffer, right, theme, Panel{
		Title: "Preview",
		Lines: s.composePreviewLines(),
	})
}

func (s *Shell) composeTitle() string {
	switch s.compose.Mode {
	case composeModeExpense:
		return "Guided Expense Entry"
	case composeModeIncome:
		return "Guided Income Entry"
	default:
		return "Custom Journal Entry"
	}
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

	if s.compose.Mode == composeModeCustom {
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
	}

	if strings.TrimSpace(s.compose.GeneralError) != "" {
		lines = append(lines, "", "Error: "+s.compose.GeneralError)
	}
	lines = append(lines, "", s.composeHelpText())
	return lines
}

func (s *Shell) composeHelpText() string {
	if s.compose.Mode == composeModeCustom {
		return "Tab move  Ctrl+S submit  Ctrl+N add line  Ctrl+D delete line  Space toggle side  Esc cancel"
	}
	return "Tab move  Ctrl+S submit  Esc cancel"
}

func (s *Shell) composePreviewLines() []string {
	if s.compose == nil {
		return nil
	}
	lines := []string{
		"Date: " + displayComposeValue(s.compose.Fields["date"], "YYYY-MM-DD"),
		"Description: " + displayComposeValue(s.compose.Fields["description"], "required"),
		"",
	}

	switch s.compose.Mode {
	case composeModeExpense:
		lines = append(lines,
			"Dr "+displayComposeValue(s.compose.Fields["account_code"], "expense account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"Cr "+displayComposeValue(s.compose.Fields["offset_account_code"], "funding account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"",
			"Expense accounts:",
		)
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.ExpenseAccounts)...)
		lines = append(lines, "", "Funding accounts:")
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.FundingAccounts)...)
	case composeModeIncome:
		lines = append(lines,
			"Dr "+displayComposeValue(s.compose.Fields["offset_account_code"], "deposit account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"Cr "+displayComposeValue(s.compose.Fields["account_code"], "income account")+" "+displayComposeValue(s.compose.Fields["amount"], "amount"),
			"",
			"Income accounts:",
		)
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.IncomeAccounts)...)
		lines = append(lines, "", "Deposit accounts:")
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.DepositAccounts)...)
	default:
		var debitCount, creditCount int
		lines = append(lines, "Lines:")
		for index := range s.compose.Lines {
			line := s.compose.Lines[index]
			lines = append(lines, fmt.Sprintf("L%d %s %s %s", index+1, displayComposeValue(line.Side, "side"), displayComposeValue(line.AccountCode, "account"), displayComposeValue(line.Amount, "amount")))
			if line.Side == "debit" {
				debitCount++
			}
			if line.Side == "credit" {
				creditCount++
			}
		}
		lines = append(lines, "", fmt.Sprintf("Line counts: %d debit / %d credit", debitCount, creditCount), "", "Active accounts:")
		lines = append(lines, accountOptionLines(s.Data.EntryCatalog.AllAccounts)...)
	}

	return lines
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
	lines := make([]string, 0, minInt(len(options), 10))
	for index := 0; index < len(options) && index < 10; index++ {
		lines = append(lines, fmt.Sprintf("%s %s (%s)", options[index].Code, options[index].Name, options[index].Type))
	}
	if len(options) > 10 {
		lines = append(lines, fmt.Sprintf("... %d more", len(options)-10))
	}
	return lines
}
