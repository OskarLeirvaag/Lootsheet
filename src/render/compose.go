package render

import (
	"strings"
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
	composeModeCodexType     composeMode = "codex_type"
	composeModeCampaign      composeMode = "campaign"
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
	picker          *pickerState
	CodexFormID     string
	CodexTypeID     string
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

func newCampaignCompose(previous Section) *composeState {
	return &composeState{
		Mode:            composeModeCampaign,
		PreviousSection: previous,
		CommandID:       "campaign.create",
		Fields: map[string]string{
			"name": "",
		},
		FieldErrors: make(map[string]string),
	}
}

func newCodexTypeCompose(previous Section) *composeState {
	return &composeState{
		Mode:            composeModeCodexType,
		PreviousSection: previous,
		CommandID:       "codex_type.create",
		Fields: map[string]string{
			"id":      "",
			"name":    "",
			"form_id": "",
		},
		FieldErrors: make(map[string]string),
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
	case composeModeCodexType:
		s.compose = newCodexTypeCompose(s.Section)
	case composeModeCampaign:
		s.compose = newCampaignCompose(s.Section)
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
		case SectionSettings:
			switch s.activeSettingsSection() {
			case settingsTabCodexTypes:
				return s.openCompose(composeModeCodexType)
			case settingsTabCampaigns:
				return s.openCompose(composeModeCampaign)
			default:
				return s.openCompose(composeModeAccount)
			}
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
