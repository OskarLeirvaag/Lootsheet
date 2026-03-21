package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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
	ss := sectionStyleFor(s.Section, theme)
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
		ss := sectionStyleFor(s.Section, theme)
		renderPicker(s.compose.picker, buffer, rect, theme, &ss)
	}
}

func (s *Shell) composeTitle() string {
	fallback := "Compose"
	switch s.compose.Mode {
	case composeModeExpense:
		return "Guided Expense Entry"
	case composeModeIncome:
		return "Guided Income Entry"
	case composeModeAccount:
		return "Add Account"
	case composeModeCodexType:
		return "Add Codex Type"
	case composeModeCampaign:
		return "New Campaign"
	case composeModeQuest:
		fallback = "Add Quest"
	case composeModeLoot:
		fallback = "Add Loot"
	case composeModeAsset:
		fallback = "Add Asset"
	case composeModeCustom:
		fallback = "Custom Journal Entry"
	case composeModeAssetTemplate:
		fallback = "Edit Entry Template"
	case composeModeCodex:
		fallback = "Add to Codex"
	case composeModeNotes:
		fallback = "Add Note"
	}

	if strings.TrimSpace(s.compose.Title) != "" {
		return s.compose.Title
	}
	return fallback
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
	if _, opts := s.pickerOptionsForCurrentField(); len(opts) > 0 {
		pickerHint = "  C-a pick"
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
	case composeModeCodexType:
		lines = append(lines,
			"ID: "+displayComposeValue(s.compose.Fields["id"], "required"),
			"Name: "+displayComposeValue(s.compose.Fields["name"], "required"),
			"Form: "+displayComposeValue(s.compose.Fields["form_id"], "required"),
			"",
			"Valid form templates:",
			"npc  player  settlement",
		)
	case composeModeCampaign:
		lines = append(lines,
			"Name: "+displayComposeValue(s.compose.Fields["name"], "required"),
			"",
			"A new campaign with seed accounts will be created.",
			"It will be set as the active campaign.",
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
			"Use @[type/name] in notes for cross-references:",
			"@[quest/Name], @[loot/Name], @[person/Name]",
		)
	case composeModeNotes:
		lines = append(lines,
			"Title: "+displayComposeValue(s.compose.Fields["title"], "required"),
			"Body: "+displayComposeValue(s.compose.Fields["body"], "optional"),
			"",
			"Use @[type/name] in body for cross-references:",
			"@[quest/Name], @[loot/Name], @[person/Name], @[note/Name]",
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
