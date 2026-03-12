package render

import (
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
)

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
	case composeModeCodexType:
		return []composeField{
			{ID: "id", Label: "ID", Placeholder: "deity"},
			{ID: "name", Label: "Name", Placeholder: "Deity"},
			{ID: "form_id", Label: "Form", Placeholder: "npc|player|settlement"},
		}
	case composeModeQuest:
		if s.compose.CommandID == "quest.update" {
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
			fields[i] = composeField(f)
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
		case composeModeCodexType:
			command.ID = "codex_type.create"
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
	case composeModeCodexType:
		required = []string{"id", "name", "form_id"}
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
