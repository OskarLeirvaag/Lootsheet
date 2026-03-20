// Package proto provides protobuf message definitions and bidirectional
// conversion between Go model types and their proto counterparts.
package proto

import (
	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// ---------- Go → Proto (server-side) ----------

// ShellDataToProto converts a model.ShellData to its proto representation.
func ShellDataToProto(d *model.ShellData) *ShellDataProto {
	return &ShellDataProto{
		Dashboard:          dashboardDataToProto(&d.Dashboard),
		Accounts:           listScreenDataToProto(&d.Accounts),
		Journal:            listScreenDataToProto(&d.Journal),
		Quests:             listScreenDataToProto(&d.Quests),
		Loot:               listScreenDataToProto(&d.Loot),
		Assets:             listScreenDataToProto(&d.Assets),
		Codex:              listScreenDataToProto(&d.Codex),
		Notes:              listScreenDataToProto(&d.Notes),
		SettingsAccounts:   listScreenDataToProto(&d.SettingsAccounts),
		SettingsCodexTypes: listScreenDataToProto(&d.SettingsCodexTypes),
		EntryCatalog:       entryCatalogToProto(&d.EntryCatalog),
		CodexTypes:         codexTypeOptionsToProto(d.CodexTypes),
		CampaignName:       d.CampaignName,
		Campaigns:          campaignOptionsToProto(d.Campaigns),
	}
}

// CommandResultToProto converts a model.CommandResult to its proto representation.
func CommandResultToProto(r *model.CommandResult) *CommandResultProto {
	return &CommandResultProto{
		Data:          ShellDataToProto(&r.Data),
		Status:        statusMessageToProto(r.Status),
		NavigateTo:    int32(r.NavigateTo), //nolint:gosec // Section values are small constants
		SelectItemKey: r.SelectItemKey,
	}
}

// CommandToProto converts a model.Command to its proto representation.
func CommandToProto(c model.Command) *CommandProto {
	lines := make([]*CommandLineProto, len(c.Lines))
	for i := range c.Lines {
		lines[i] = commandLineToProto(&c.Lines[i])
	}
	return &CommandProto{
		Id:      c.ID,
		Section: int32(c.Section), //nolint:gosec // Section values are small constants
		ItemKey: c.ItemKey,
		Fields:  c.Fields,
		Lines:   lines,
	}
}

// ListItemsToProto converts a slice of model.ListItemData to proto.
func ListItemsToProto(items []model.ListItemData) []*ListItemDataProto {
	out := make([]*ListItemDataProto, len(items))
	for i := range items {
		out[i] = listItemDataToProto(&items[i])
	}
	return out
}

// CampaignOptionsToProto converts campaign options to proto.
func CampaignOptionsToProto(opts []model.CampaignOption) []*CampaignOptionProto {
	return campaignOptionsToProto(opts)
}

// ---------- Proto → Go (client-side) ----------

// ShellDataFromProto converts a proto ShellData to the Go model.
func ShellDataFromProto(p *ShellDataProto) model.ShellData {
	if p == nil {
		return model.ShellData{}
	}
	return model.ShellData{
		Dashboard:          dashboardDataFromProto(p.Dashboard),
		Accounts:           listScreenDataFromProto(p.Accounts),
		Journal:            listScreenDataFromProto(p.Journal),
		Quests:             listScreenDataFromProto(p.Quests),
		Loot:               listScreenDataFromProto(p.Loot),
		Assets:             listScreenDataFromProto(p.Assets),
		Codex:              listScreenDataFromProto(p.Codex),
		Notes:              listScreenDataFromProto(p.Notes),
		SettingsAccounts:   listScreenDataFromProto(p.SettingsAccounts),
		SettingsCodexTypes: listScreenDataFromProto(p.SettingsCodexTypes),
		EntryCatalog:       entryCatalogFromProto(p.EntryCatalog),
		CodexTypes:         codexTypeOptionsFromProto(p.CodexTypes),
		CampaignName:       p.CampaignName,
		Campaigns:          campaignOptionsFromProto(p.Campaigns),
	}
}

// CommandResultFromProto converts a proto CommandResult to the Go model.
func CommandResultFromProto(p *CommandResultProto) model.CommandResult {
	if p == nil {
		return model.CommandResult{}
	}
	return model.CommandResult{
		Data:          ShellDataFromProto(p.Data),
		Status:        statusMessageFromProto(p.Status),
		NavigateTo:    model.Section(p.NavigateTo),
		SelectItemKey: p.SelectItemKey,
	}
}

// CommandFromProto converts a proto Command to the Go model.
func CommandFromProto(p *CommandProto) model.Command {
	if p == nil {
		return model.Command{}
	}
	lines := make([]model.CommandLine, len(p.Lines))
	for i, l := range p.Lines {
		lines[i] = commandLineFromProto(l)
	}
	return model.Command{
		ID:      p.Id,
		Section: model.Section(p.Section),
		ItemKey: p.ItemKey,
		Fields:  p.Fields,
		Lines:   lines,
	}
}

// ListItemsFromProto converts a slice of proto ListItemData to the Go model.
func ListItemsFromProto(items []*ListItemDataProto) []model.ListItemData {
	out := make([]model.ListItemData, len(items))
	for i, item := range items {
		out[i] = listItemDataFromProto(item)
	}
	return out
}

// CampaignOptionsFromProto converts proto campaign options to Go model.
func CampaignOptionsFromProto(opts []*CampaignOptionProto) []model.CampaignOption {
	return campaignOptionsFromProto(opts)
}

// ---------- Internal helpers: Go → Proto ----------

func dashboardDataToProto(d *model.DashboardData) *DashboardDataProto {
	return &DashboardDataProto{
		HeaderLines:     d.HeaderLines,
		AccountsLines:   d.AccountsLines,
		JournalLines:    d.JournalLines,
		HoardLines:      d.HoardLines,
		QuickEntryLines: d.QuickEntryLines,
		LedgerLines:     d.LedgerLines,
		QuestLines:      d.QuestLines,
		LootLines:       d.LootLines,
		AssetLines:      d.AssetLines,
	}
}

func listScreenDataToProto(d *model.ListScreenData) *ListScreenDataProto {
	items := make([]*ListItemDataProto, len(d.Items))
	for i := range d.Items {
		items[i] = listItemDataToProto(&d.Items[i])
	}
	return &ListScreenDataProto{
		HeaderLines:   d.HeaderLines,
		SummaryLines:  d.SummaryLines,
		ListHeaderRow: d.ListHeaderRow,
		Items:         items,
		EmptyLines:    d.EmptyLines,
	}
}

func listItemDataToProto(d *model.ListItemData) *ListItemDataProto {
	actions := make([]*ItemActionDataProto, len(d.Actions))
	for i := range d.Actions {
		actions[i] = itemActionDataToProto(&d.Actions[i])
	}
	return &ListItemDataProto{
		Key:         d.Key,
		Row:         d.Row,
		DetailTitle: d.DetailTitle,
		DetailLines: d.DetailLines,
		DetailBody:  d.DetailBody,
		Actions:     actions,
	}
}

func itemActionDataToProto(a *model.ItemActionData) *ItemActionDataProto {
	composeLines := make([]*CommandLineProto, len(a.ComposeLines))
	for i := range a.ComposeLines {
		composeLines[i] = commandLineToProto(&a.ComposeLines[i])
	}
	return &ItemActionDataProto{
		Trigger:       string(a.Trigger),
		Id:            a.ID,
		Label:         a.Label,
		Mode:          string(a.Mode),
		ConfirmTitle:  a.ConfirmTitle,
		ConfirmLines:  a.ConfirmLines,
		InputTitle:    a.InputTitle,
		InputPrompt:   a.InputPrompt,
		InputRequired: a.InputRequired,
		InputHelp:     a.InputHelp,
		Placeholder:   a.Placeholder,
		ComposeMode:   a.ComposeMode,
		ComposeTitle:  a.ComposeTitle,
		ComposeFields: a.ComposeFields,
		ComposeLines:  composeLines,
	}
}

func entryCatalogToProto(c *model.EntryCatalog) *EntryCatalogProto {
	return &EntryCatalogProto{
		DefaultDate:     c.DefaultDate,
		ExpenseAccounts: accountOptionsToProto(c.ExpenseAccounts),
		IncomeAccounts:  accountOptionsToProto(c.IncomeAccounts),
		FundingAccounts: accountOptionsToProto(c.FundingAccounts),
		DepositAccounts: accountOptionsToProto(c.DepositAccounts),
		AllAccounts:     accountOptionsToProto(c.AllAccounts),
	}
}

func accountOptionsToProto(opts []model.AccountOption) []*AccountOptionProto {
	out := make([]*AccountOptionProto, len(opts))
	for i, o := range opts {
		out[i] = &AccountOptionProto{Code: o.Code, Name: o.Name, Type: o.Type}
	}
	return out
}

func campaignOptionsToProto(opts []model.CampaignOption) []*CampaignOptionProto {
	out := make([]*CampaignOptionProto, len(opts))
	for i, o := range opts {
		out[i] = &CampaignOptionProto{Id: o.ID, Name: o.Name}
	}
	return out
}

func codexTypeOptionsToProto(opts []model.CodexTypeOption) []*CodexTypeOptionProto {
	out := make([]*CodexTypeOptionProto, len(opts))
	for i, o := range opts {
		out[i] = &CodexTypeOptionProto{Id: o.ID, Name: o.Name, FormId: o.FormID}
	}
	return out
}

func commandLineToProto(l *model.CommandLine) *CommandLineProto {
	return &CommandLineProto{
		Side:        l.Side,
		AccountCode: l.AccountCode,
		Amount:      l.Amount,
		Memo:        l.Memo,
	}
}

func statusMessageToProto(s model.StatusMessage) *StatusMessageProto {
	return &StatusMessageProto{
		Level: string(s.Level),
		Text:  s.Text,
	}
}

// ---------- Internal helpers: Proto → Go ----------

func dashboardDataFromProto(p *DashboardDataProto) model.DashboardData {
	if p == nil {
		return model.DashboardData{}
	}
	return model.DashboardData{
		HeaderLines:     p.HeaderLines,
		AccountsLines:   p.AccountsLines,
		JournalLines:    p.JournalLines,
		HoardLines:      p.HoardLines,
		QuickEntryLines: p.QuickEntryLines,
		LedgerLines:     p.LedgerLines,
		QuestLines:      p.QuestLines,
		LootLines:       p.LootLines,
		AssetLines:      p.AssetLines,
	}
}

func listScreenDataFromProto(p *ListScreenDataProto) model.ListScreenData {
	if p == nil {
		return model.ListScreenData{}
	}
	items := make([]model.ListItemData, len(p.Items))
	for i, item := range p.Items {
		items[i] = listItemDataFromProto(item)
	}
	return model.ListScreenData{
		HeaderLines:   p.HeaderLines,
		SummaryLines:  p.SummaryLines,
		ListHeaderRow: p.ListHeaderRow,
		Items:         items,
		EmptyLines:    p.EmptyLines,
	}
}

func listItemDataFromProto(p *ListItemDataProto) model.ListItemData {
	if p == nil {
		return model.ListItemData{}
	}
	actions := make([]model.ItemActionData, len(p.Actions))
	for i, a := range p.Actions {
		actions[i] = itemActionDataFromProto(a)
	}
	return model.ListItemData{
		Key:         p.Key,
		Row:         p.Row,
		DetailTitle: p.DetailTitle,
		DetailLines: p.DetailLines,
		DetailBody:  p.DetailBody,
		Actions:     actions,
	}
}

func itemActionDataFromProto(p *ItemActionDataProto) model.ItemActionData {
	if p == nil {
		return model.ItemActionData{}
	}
	composeLines := make([]model.CommandLine, len(p.ComposeLines))
	for i, l := range p.ComposeLines {
		composeLines[i] = commandLineFromProto(l)
	}
	return model.ItemActionData{
		Trigger:       model.Action(p.Trigger),
		ID:            p.Id,
		Label:         p.Label,
		Mode:          model.ItemActionMode(p.Mode),
		ConfirmTitle:  p.ConfirmTitle,
		ConfirmLines:  p.ConfirmLines,
		InputTitle:    p.InputTitle,
		InputPrompt:   p.InputPrompt,
		InputRequired: p.InputRequired,
		InputHelp:     p.InputHelp,
		Placeholder:   p.Placeholder,
		ComposeMode:   p.ComposeMode,
		ComposeTitle:  p.ComposeTitle,
		ComposeFields: p.ComposeFields,
		ComposeLines:  composeLines,
	}
}

func entryCatalogFromProto(p *EntryCatalogProto) model.EntryCatalog {
	if p == nil {
		return model.EntryCatalog{}
	}
	return model.EntryCatalog{
		DefaultDate:     p.DefaultDate,
		ExpenseAccounts: accountOptionsFromProto(p.ExpenseAccounts),
		IncomeAccounts:  accountOptionsFromProto(p.IncomeAccounts),
		FundingAccounts: accountOptionsFromProto(p.FundingAccounts),
		DepositAccounts: accountOptionsFromProto(p.DepositAccounts),
		AllAccounts:     accountOptionsFromProto(p.AllAccounts),
	}
}

func accountOptionsFromProto(opts []*AccountOptionProto) []model.AccountOption {
	out := make([]model.AccountOption, len(opts))
	for i, o := range opts {
		out[i] = model.AccountOption{Code: o.Code, Name: o.Name, Type: o.Type}
	}
	return out
}

func campaignOptionsFromProto(opts []*CampaignOptionProto) []model.CampaignOption {
	out := make([]model.CampaignOption, len(opts))
	for i, o := range opts {
		out[i] = model.CampaignOption{ID: o.Id, Name: o.Name}
	}
	return out
}

func codexTypeOptionsFromProto(opts []*CodexTypeOptionProto) []model.CodexTypeOption {
	out := make([]model.CodexTypeOption, len(opts))
	for i, o := range opts {
		out[i] = model.CodexTypeOption{ID: o.Id, Name: o.Name, FormID: o.FormId}
	}
	return out
}

func commandLineFromProto(p *CommandLineProto) model.CommandLine {
	if p == nil {
		return model.CommandLine{}
	}
	return model.CommandLine{
		Side:        p.Side,
		AccountCode: p.AccountCode,
		Amount:      p.Amount,
		Memo:        p.Memo,
	}
}

func statusMessageFromProto(p *StatusMessageProto) model.StatusMessage {
	if p == nil {
		return model.StatusMessage{}
	}
	return model.StatusMessage{
		Level: model.StatusLevel(p.Level),
		Text:  p.Text,
	}
}
