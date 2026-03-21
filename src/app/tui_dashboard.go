package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/account"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/notes"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

const (
	tuiCommandAccountCreate       = "account.create"
	tuiCommandAccountRename       = "account.rename"
	tuiCommandAccountActivate     = "account.activate"
	tuiCommandAccountDeactivate   = "account.deactivate"
	tuiCommandAccountDelete       = "account.delete"
	tuiCommandJournalReverse      = "journal.reverse"
	tuiCommandCreateExpense       = "entry.expense.create"
	tuiCommandCreateIncome        = "entry.income.create"
	tuiCommandCreateCustom        = "entry.custom.create"
	tuiCommandQuestCreate         = "quest.create"
	tuiCommandQuestUpdate         = "quest.update"
	tuiCommandQuestCollectFull    = "quest.collect_full"
	tuiCommandQuestWriteOffFull   = "quest.writeoff_full"
	tuiCommandLootCreate          = "loot.create"
	tuiCommandLootUpdate          = "loot.update"
	tuiCommandLootAppraise        = "loot.appraise"
	tuiCommandLootRecognize       = "loot.recognize_latest"
	tuiCommandLootSell            = "loot.sell"
	tuiCommandLootTransferToAsset = "loot.transfer_to_asset"
	tuiCommandAssetCreate         = "asset.create"
	tuiCommandAssetUpdate         = "asset.update"
	tuiCommandAssetAppraise       = "asset.appraise"
	tuiCommandAssetRecognize      = "asset.recognize_latest"
	tuiCommandAssetTransferToLoot = "asset.transfer_to_loot"
	tuiCommandAssetTemplateSave   = "asset.template.save"
	tuiCommandCodexCreate         = "codex.create"
	tuiCommandCodexUpdate         = "codex.update"
	tuiCommandCodexDelete         = "codex.delete"
	tuiCommandNotesCreate         = "notes.create"
	tuiCommandNotesUpdate         = "notes.update"
	tuiCommandNotesDelete         = "notes.delete"
	tuiCommandCodexTypeCreate     = "codex_type.create"
	tuiCommandCodexTypeRename     = "codex_type.rename"
	tuiCommandCodexTypeDelete     = "codex_type.delete"
	tuiCommandCampaignCreate      = "campaign.create"
	tuiCommandCampaignRename      = "campaign.rename"
	tuiCommandCampaignSwitch      = "campaign.switch"
	tuiCommandCampaignDelete      = "campaign.delete"
)

var tuiNow = time.Now

func buildTUIShellData(ctx context.Context, loader TUIDataLoader) (render.ShellData, error) {
	status, err := loader.GetDatabaseStatus(ctx)
	if err != nil {
		return render.ErrorShellData("Database status unavailable.", err.Error()), nil
	}

	switch status.State {
	case ledger.DatabaseStateUninitialized:
		return unavailableShellData(&status, "Run `lootsheet init` before opening live dashboard summaries."), nil
	case ledger.DatabaseStateUpgradeable:
		return unavailableShellData(&status, "Run `lootsheet db migrate` before opening live dashboard summaries."), nil
	case ledger.DatabaseStateForeign, ledger.DatabaseStateDamaged:
		return unavailableShellData(&status, blankStatusDetail(status.Detail)), nil
	case ledger.DatabaseStateCurrent:
	default:
	}

	databaseName := loader.DatabaseName()
	campaignName := loader.CampaignName()

	var campaignOptions []render.CampaignOption
	campaignRecords, campaignListErr := loader.ListCampaigns(ctx)
	if campaignListErr == nil {
		campaignOptions = make([]render.CampaignOption, len(campaignRecords))
		for i, c := range campaignRecords {
			campaignOptions[i] = render.CampaignOption{ID: c.ID, Name: c.Name}
		}
	}

	data := render.ShellData{
		CampaignName: campaignName,
		Campaigns:    campaignOptions,
		Dashboard: render.DashboardData{
			HeaderLines: []string{
				fmt.Sprintf("Campaign: %s  |  Read-only snapshot from %s.", campaignName, databaseName),
				"Use arrows, Tab, or 1-7 to move between boxed screens. Use e/i/a for guided entry creation. @ opens settings.",
			},
			HoardLines: []string{
				"To share now: awaiting ledger snapshot.",
				"Unsold loot: awaiting register snapshot.",
			},
			QuickEntryLines: []string{
				"e  I have an expense",
				"i  I have income",
				"a  Add custom entry",
			},
		},
		Ledger: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("General ledger from %s.", databaseName),
				"Select an account to view its posting history.",
			},
			EmptyLines: []string{
				"No accounts found.",
				"The chart of accounts is empty in this database.",
			},
		},
		Journal: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Posted journal history from %s.", databaseName),
				"Select an entry to inspect it. `e`/`i` add guided entries and `r` reverses the selected posted entry.",
			},
			EmptyLines: []string{
				"No journal entries yet.",
				"Posting stays in the CLI for now.",
			},
		},
		Quests: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Quest register from %s.", databaseName),
				"Select a quest to inspect it. `a` adds, `u` edits, `c` collects the full balance, and `w` writes off the full balance.",
			},
			EmptyLines: []string{
				"No quests tracked yet.",
				"Quest actions appear when a quest has an outstanding collectible balance.",
			},
		},
		Loot: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Unrealized loot register from %s.", databaseName),
				"Select a loot item to inspect it. `a` adds, `u` edits, `n` recognizes the latest appraisal, and `s` sells recognized loot.",
			},
			EmptyLines: []string{
				"No loot tracked yet.",
				"Recognition appears for held items with a latest appraisal of at least 1 CP. Sale appears for recognized items.",
			},
		},
		Assets: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Party asset register from %s.", databaseName),
				"Select an asset to inspect it. `a` adds, `u` edits, `n` recognizes the latest appraisal, and `t` transfers to loot.",
			},
			EmptyLines: []string{
				"No assets tracked yet.",
				"Assets are high-value items the party keeps. Transfer to loot when ready to sell.",
			},
		},
		Codex: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Codex from %s.", databaseName),
				"Select an entry to inspect. `a` adds, `u` edits, `d` deletes. Use @type/name in notes for cross-references.",
			},
			EmptyLines: []string{
				"No codex entries yet.",
				"Add players, NPCs, and contacts here.",
			},
		},
		Notes: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Campaign notes from %s.", databaseName),
				"Select a note to inspect. `a` adds, `u` edits, `d` deletes. Use @type/name in body for cross-references.",
			},
			EmptyLines: []string{
				"No notes yet.",
				"Add campaign and session notes here.",
			},
		},
	}

	var panelErrors []string

	trialBalance, err := loader.GetTrialBalance(ctx)
	trialBalanceAvailable := false
	if err != nil {
		data.Dashboard.LedgerLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "ledger")
	} else {
		data.Dashboard.LedgerLines = summarizeLedger(trialBalance)
		trialBalanceAvailable = true
	}

	accounts, err := loader.ListAccounts(ctx)
	if err != nil {
		data.Dashboard.AccountsLines = unavailablePanelLines(err)
		data.Ledger = unavailableSectionData("Ledger unavailable.", err.Error())
		panelErrors = append(panelErrors, "accounts")
	} else {
		data.Dashboard.AccountsLines = summarizeSettingsAccounts(accounts)

		accountLedgers := make(map[string]journal.AccountLedgerReport, len(accounts))
		for _, acct := range accounts {
			if rpt, ledgerErr := loader.GetAccountLedger(ctx, acct.Code); ledgerErr == nil {
				accountLedgers[acct.Code] = rpt
			}
		}

		balanceMap := make(map[string]report.TrialBalanceRow, len(trialBalance.Accounts))
		if trialBalanceAvailable {
			for _, row := range trialBalance.Accounts {
				balanceMap[row.AccountCode] = row
			}
		}

		if trialBalanceAvailable {
			data.Ledger.SummaryLines = summarizeLedgerSection(trialBalance, len(accounts))
		} else {
			data.Ledger.SummaryLines = summarizeSettingsAccounts(accounts)
		}
		data.Ledger.ListHeaderRow = fmt.Sprintf("%-4s %-9s %10s %10s %10s  %s", "CODE", "TYPE", "DEBITS", "CREDITS", "BALANCE", "NAME")
		data.Ledger.Items = buildLedgerItems(accounts, accountLedgers, balanceMap)
		data.EntryCatalog = buildEntryCatalog(accounts, tuiToday())
	}

	journalSummary, err := loader.GetJournalSummary(ctx)
	if err != nil {
		data.Dashboard.JournalLines = unavailablePanelLines(err)
		data.Journal = unavailableSectionData("Journal unavailable.", err.Error())
		panelErrors = append(panelErrors, "journal")
	} else {
		data.Dashboard.JournalLines = summarizeJournal(journalSummary)
		data.Journal.SummaryLines = summarizeJournal(journalSummary)
	}

	journalEntries, err := loader.ListBrowseJournalEntries(ctx)
	if err != nil {
		if len(data.Journal.SummaryLines) == 0 {
			data.Journal = unavailableSectionData("Journal unavailable.", err.Error())
		}
		data.Dashboard.JournalLines = unavailablePanelLines(err)
		data.Journal.Items = nil
		data.Journal.EmptyLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "journal")
	} else {
		data.Journal.ListHeaderRow = fmt.Sprintf("#%-4s %-10s %-8s %s", "NUM", "DATE", "STATUS", "DESCRIPTION")
		data.Journal.Items = buildJournalItems(journalEntries)
	}

	panelErrors = loadShellQuestData(ctx, loader, &data, panelErrors)

	lootRows, lootSummaryAvailable, panelErrors := loadShellLootData(ctx, loader, &data, panelErrors)
	panelErrors = loadShellAssetData(ctx, loader, &data, panelErrors)

	codexTypes, panelErrors := loadShellCodexData(ctx, loader, &data, panelErrors)

	// Build Settings tabs from accounts, codex types, and campaigns.
	buildShellSettingsData(&data, databaseName, accounts, codexTypes, campaignRecords, campaignListErr, loader)

	panelErrors = loadShellNotesData(ctx, loader, &data, panelErrors)

	// Load and append entity cross-reference links.
	panelErrors = appendEntityReferenceLinks(ctx, loader, &data, panelErrors)

	if trialBalanceAvailable {
		data.Dashboard.HoardLines = summarizeShareableGold(trialBalance, lootRows, lootSummaryAvailable)
	} else {
		data.Dashboard.HoardLines = []string{
			"To share now: unavailable",
			"Ledger snapshot is unavailable.",
		}
	}

	if len(panelErrors) > 0 {
		data.Dashboard.HeaderLines[1] = "Some panels are unavailable: " + strings.Join(uniqueStrings(panelErrors), ", ") + "."
	}

	return data, nil
}

func loadShellQuestData(ctx context.Context, loader TUIDataLoader, data *render.ShellData, panelErrors []string) []string {
	promisedQuests, err := loader.GetPromisedQuests(ctx)
	var receivables []report.QuestReceivableRow
	questSummaryAvailable := false
	if err != nil {
		data.Dashboard.QuestLines = unavailablePanelLines(err)
		data.Quests = unavailableSectionData("Quest register unavailable.", err.Error())
		panelErrors = append(panelErrors, "quests")
	} else {
		var receivableErr error
		receivables, receivableErr = loader.GetQuestReceivables(ctx)
		if receivableErr != nil {
			data.Dashboard.QuestLines = unavailablePanelLines(receivableErr)
			data.Quests = unavailableSectionData("Quest register unavailable.", receivableErr.Error())
			panelErrors = append(panelErrors, "quests")
		} else {
			writeOffCandidates, _ := loader.GetWriteOffCandidates(ctx)
			data.Dashboard.QuestLines = summarizeQuests(promisedQuests, receivables, writeOffCandidates)
			data.Quests.SummaryLines = summarizeQuests(promisedQuests, receivables, writeOffCandidates)
			questSummaryAvailable = true
		}
	}

	if questSummaryAvailable {
		questRows, questErr := loadTUIQuestRows(ctx, loader)
		if questErr != nil {
			if len(data.Quests.SummaryLines) == 0 {
				data.Quests = unavailableSectionData("Quest register unavailable.", questErr.Error())
			}
			data.Quests.Items = nil
			data.Quests.EmptyLines = unavailablePanelLines(questErr)
			panelErrors = append(panelErrors, "quests")
		} else {
			data.Quests.ListHeaderRow = fmt.Sprintf("%-12s %-14s %-12s %s", "REWARD", "STATUS", "OUTSTANDING", "TITLE")
			data.Quests.Items = buildQuestItems(questRows, tuiToday())
		}
	}

	return panelErrors
}

func loadShellLootData(ctx context.Context, loader TUIDataLoader, data *render.ShellData, panelErrors []string) ([]report.LootSummaryRow, bool, []string) {
	lootRows, err := loader.GetLootSummary(ctx, "loot")
	lootSummaryAvailable := false
	if err != nil {
		data.Dashboard.LootLines = unavailablePanelLines(err)
		data.Loot = unavailableSectionData("Loot register unavailable.", err.Error())
		panelErrors = append(panelErrors, "loot")
	} else {
		data.Dashboard.LootLines = summarizeItemRegister(lootRows, "items")
		data.Loot.SummaryLines = summarizeItemRegister(lootRows, "items")
		lootSummaryAvailable = true
	}

	if lootSummaryAvailable {
		browseItems, browseErr := loader.ListBrowseLootItems(ctx, "loot")
		if browseErr != nil {
			if len(data.Loot.SummaryLines) == 0 {
				data.Loot = unavailableSectionData("Loot register unavailable.", browseErr.Error())
			}
			data.Loot.Items = nil
			data.Loot.EmptyLines = unavailablePanelLines(browseErr)
			panelErrors = append(panelErrors, "loot")
		} else {
			data.Loot.ListHeaderRow = fmt.Sprintf("%-12s %-7s %-11s %s", "VALUE", "QTY", "STATUS", "NAME")
			data.Loot.Items = buildLootItems(browseItems, tuiToday())
		}
	}

	return lootRows, lootSummaryAvailable, panelErrors
}

func loadShellAssetData(ctx context.Context, loader TUIDataLoader, data *render.ShellData, panelErrors []string) []string {
	assetRows, assetSummaryErr := loader.GetLootSummary(ctx, "asset")
	assetSummaryAvailable := false
	if assetSummaryErr != nil {
		data.Dashboard.AssetLines = unavailablePanelLines(assetSummaryErr)
		data.Assets = unavailableSectionData("Asset register unavailable.", assetSummaryErr.Error())
		panelErrors = append(panelErrors, "assets")
	} else {
		data.Dashboard.AssetLines = summarizeItemRegister(assetRows, "assets")
		data.Assets.SummaryLines = summarizeItemRegister(assetRows, "assets")
		assetSummaryAvailable = true
	}

	if assetSummaryAvailable {
		assetBrowseItems, assetBrowseErr := loader.ListBrowseLootItems(ctx, "asset")
		if assetBrowseErr != nil {
			if len(data.Assets.SummaryLines) == 0 {
				data.Assets = unavailableSectionData("Asset register unavailable.", assetBrowseErr.Error())
			}
			data.Assets.Items = nil
			data.Assets.EmptyLines = unavailablePanelLines(assetBrowseErr)
			panelErrors = append(panelErrors, "assets")
		} else {
			data.Assets.ListHeaderRow = fmt.Sprintf("%-12s %-11s %-4s %-12s %s", "VALUE", "STATUS", "TPL", "HOLDER", "NAME")
			data.Assets.Items = buildAssetItems(assetBrowseItems, tuiToday())
		}
	}

	return panelErrors
}

func loadShellCodexData(ctx context.Context, loader TUIDataLoader, data *render.ShellData, panelErrors []string) ([]codex.CodexType, []string) {
	var allCodexRefs map[string][]refs.EntityReference

	codexEntries, codexErr := loader.ListCodexEntries(ctx)
	if codexErr != nil {
		data.Codex = unavailableSectionData("Codex unavailable.", codexErr.Error())
		panelErrors = append(panelErrors, "codex")
	} else {
		data.Codex.SummaryLines = summarizeCodex(codexEntries)
		data.Codex.ListHeaderRow = fmt.Sprintf("%-8s %-14s %s", "TYPE", "SECONDARY", "NAME")

		codexRefs, refsErr := loader.ListAllCodexReferences(ctx)
		if refsErr != nil {
			panelErrors = append(panelErrors, "codex-refs")
		} else {
			allCodexRefs = codexRefs
		}

		data.Codex.Items = buildCodexItems(codexEntries, allCodexRefs)
	}

	// Load codex types for the picker.
	codexTypes, codexTypesErr := loader.ListCodexTypes(ctx)
	if codexTypesErr == nil {
		data.CodexTypes = make([]render.CodexTypeOption, len(codexTypes))
		for i, ct := range codexTypes {
			data.CodexTypes[i] = render.CodexTypeOption{
				ID:     ct.ID,
				Name:   ct.Name,
				FormID: ct.FormID,
			}
		}
	}

	return codexTypes, panelErrors
}

func loadShellNotesData(ctx context.Context, loader TUIDataLoader, data *render.ShellData, panelErrors []string) []string {
	var allNotesRefs map[string][]refs.EntityReference

	noteRecords, notesErr := loader.ListNotes(ctx)
	if notesErr != nil {
		data.Notes = unavailableSectionData("Notes unavailable.", notesErr.Error())
		panelErrors = append(panelErrors, "notes")
	} else {
		data.Notes.SummaryLines = summarizeNotes(noteRecords)
		data.Notes.ListHeaderRow = fmt.Sprintf("%-11s %s", "UPDATED", "TITLE")

		noteRefs, noteRefsErr := loader.ListAllNotesReferences(ctx)
		if noteRefsErr != nil {
			panelErrors = append(panelErrors, "notes-refs")
		} else {
			allNotesRefs = noteRefs
		}

		data.Notes.Items = buildNotesItems(noteRecords, allNotesRefs)
	}

	return panelErrors
}

func buildShellSettingsData(
	data *render.ShellData,
	databaseName string,
	accounts []ledger.AccountRecord,
	codexTypes []codex.CodexType,
	campaignRecords []campaign.Record,
	campaignListErr error,
	loader TUIDataLoader,
) {
	data.SettingsAccounts = render.ListScreenData{
		HeaderLines: []string{
			fmt.Sprintf("Accounts from %s.", databaseName),
			"Chart of accounts. `a` adds, `u` renames, `d` deletes, `t` toggles active/inactive.",
		},
		SummaryLines:  summarizeSettingsAccounts(accounts),
		ListHeaderRow: fmt.Sprintf("%-4s %-9s %-8s %s", "CODE", "TYPE", "STATUS", "NAME"),
		Items:         buildSettingsAccountItems(accounts),
		EmptyLines: []string{
			"No accounts to display.",
			"Create an account with `a`.",
		},
	}
	data.SettingsCodexTypes = render.ListScreenData{
		HeaderLines: []string{
			fmt.Sprintf("Codex types from %s.", databaseName),
			"Entry categories. `a` adds, `u` renames, `d` deletes.",
		},
		SummaryLines:  summarizeSettingsCodexTypes(codexTypes),
		ListHeaderRow: fmt.Sprintf("%-12s %-12s %s", "ID", "FORM", "NAME"),
		Items:         buildSettingsCodexTypeItems(codexTypes),
		EmptyLines: []string{
			"No codex types to display.",
			"Create a codex type with `a`.",
		},
	}
	if campaignListErr == nil {
		activeCampaignID := loader.CampaignID()
		data.SettingsCampaigns = render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Campaigns from %s.", databaseName),
				"Manage campaigns. `a` adds, `u` renames, `d` deletes. Enter switches.",
			},
			SummaryLines:  summarizeSettingsCampaigns(campaignRecords, activeCampaignID),
			ListHeaderRow: fmt.Sprintf("%-40s %s", "NAME", "STATUS"),
			Items:         buildSettingsCampaignItems(campaignRecords, activeCampaignID),
			EmptyLines: []string{
				"No campaigns to display.",
				"Create a campaign with `a`.",
			},
		}
	}
}

func appendEntityReferenceLinks(ctx context.Context, loader TUIDataLoader, data *render.ShellData, panelErrors []string) []string {
	allRefsByTarget, entityRefsErr := loader.ListAllEntityReferences(ctx)
	if entityRefsErr != nil {
		panelErrors = append(panelErrors, "entity-refs")
	}

	if allRefsByTarget != nil {
		appendItemLinks(data.Quests.Items, allRefsByTarget, "quest")
		appendItemLinks(data.Loot.Items, allRefsByTarget, "loot")
		appendItemLinks(data.Assets.Items, allRefsByTarget, "asset")
		appendItemLinks(data.Codex.Items, allRefsByTarget, "person")
		appendItemLinks(data.Notes.Items, allRefsByTarget, "note")
	}

	return panelErrors
}

func appendItemLinks(items []render.ListItemData, allRefsByTarget map[string][]refs.EntityReference, targetType string) {
	for i, item := range items {
		linkLines := buildLinkedFromLines(allRefsByTarget, targetType, item.DetailTitle)
		if len(linkLines) > 0 {
			items[i].DetailLines = append(items[i].DetailLines, linkLines...)
		}
	}
}

func handleTUICommand(ctx context.Context, command render.Command, databasePath string, loader TUIDataLoader) (render.CommandResult, error) { //nolint:revive // cyclomatic: command dispatch switch
	var message render.StatusMessage
	var navigateTo render.Section
	var selectItemKey string
	today := tuiToday()
	campaignID := loader.CampaignID()

	switch command.ID {
	case tuiCommandAccountCreate, tuiCommandAccountRename, tuiCommandAccountActivate,
		tuiCommandAccountDeactivate, tuiCommandAccountDelete:
		result, err := handleAccountCommand(ctx, command, databasePath, campaignID)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandJournalReverse:
		entries, err := loader.ListBrowseJournalEntries(ctx)
		if err != nil {
			return render.CommandResult{}, err
		}

		entry, ok := findBrowseEntry(entries, command.ItemKey)
		if !ok {
			return render.CommandResult{}, fmt.Errorf("journal entry %q does not exist", command.ItemKey)
		}

		result, err := journal.ReverseJournalEntry(ctx, databasePath, campaignID, command.ItemKey, entry.EntryDate, "")
		if err != nil {
			return render.CommandResult{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Entry #%d reversed as entry #%d.", entry.EntryNumber, result.EntryNumber),
		}
	case tuiCommandCreateExpense:
		result, err := handleEntryCommand(command, today, "expense",
			func(date, desc, acct, offset string, amt int64, memo string) (ledger.PostedJournalEntry, error) {
				return journal.PostExpenseEntry(ctx, databasePath, campaignID, &journal.ExpenseEntryInput{
					Date:               date,
					Description:        desc,
					ExpenseAccountCode: acct,
					FundingAccountCode: offset,
					Amount:             amt,
					Memo:               memo,
				})
			},
		)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandCreateIncome:
		result, err := handleEntryCommand(command, today, "income",
			func(date, desc, acct, offset string, amt int64, memo string) (ledger.PostedJournalEntry, error) {
				return journal.PostIncomeEntry(ctx, databasePath, campaignID, &journal.IncomeEntryInput{
					Date:               date,
					Description:        desc,
					IncomeAccountCode:  acct,
					DepositAccountCode: offset,
					Amount:             amt,
					Memo:               memo,
				})
			},
		)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandCreateCustom:
		input, err := buildTUIJournalPostInput(command, today)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}

		result, err := journal.PostJournalEntry(ctx, databasePath, campaignID, input)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Recorded custom entry as journal entry #%d.", result.EntryNumber),
		}
		navigateTo = render.SectionJournal
		selectItemKey = result.ID
	case tuiCommandQuestCreate:
		result, err := handleQuestCreateOrUpdate(ctx, command, databasePath, campaignID, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandQuestUpdate:
		result, err := handleQuestCreateOrUpdate(ctx, command, databasePath, campaignID, false)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandQuestCollectFull:
		quests, err := loadTUIQuestRows(ctx, loader)
		if err != nil {
			return render.CommandResult{}, err
		}

		questRow, ok := findTUIQuestRow(quests, command.ItemKey)
		if !ok {
			return render.CommandResult{}, fmt.Errorf("quest %q does not exist", command.ItemKey)
		}
		if !questRow.Collectible || questRow.Outstanding <= 0 {
			return render.CommandResult{}, fmt.Errorf("quest %q cannot be collected right now", command.ItemKey)
		}

		result, err := quest.CollectQuestPayment(ctx, databasePath, campaignID, quest.CollectQuestPaymentInput{
			QuestID:     command.ItemKey,
			Amount:      questRow.Outstanding,
			Date:        today,
			Description: "",
		})
		if err != nil {
			return render.CommandResult{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Collected %s for quest %q as entry #%d.", currency.FormatAmount(questRow.Outstanding), questRow.Record.Title, result.EntryNumber),
		}
	case tuiCommandQuestWriteOffFull:
		quests, err := loadTUIQuestRows(ctx, loader)
		if err != nil {
			return render.CommandResult{}, err
		}

		questRow, ok := findTUIQuestRow(quests, command.ItemKey)
		if !ok {
			return render.CommandResult{}, fmt.Errorf("quest %q does not exist", command.ItemKey)
		}
		if !questRow.Collectible || questRow.Outstanding <= 0 {
			return render.CommandResult{}, fmt.Errorf("quest %q cannot be written off right now", command.ItemKey)
		}

		result, err := quest.WriteOffQuest(ctx, databasePath, campaignID, quest.WriteOffQuestInput{
			QuestID:     command.ItemKey,
			Date:        today,
			Description: "",
		})
		if err != nil {
			return render.CommandResult{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Wrote off %s for quest %q as entry #%d.", currency.FormatAmount(questRow.Outstanding), questRow.Record.Title, result.EntryNumber),
		}
	case tuiCommandLootCreate, tuiCommandLootUpdate, tuiCommandLootAppraise,
		tuiCommandLootRecognize, tuiCommandLootSell, tuiCommandLootTransferToAsset:
		result, err := handleLootCommand(ctx, command, databasePath, campaignID, today, loader)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandAssetCreate, tuiCommandAssetUpdate, tuiCommandAssetAppraise,
		tuiCommandAssetRecognize, tuiCommandAssetTransferToLoot, tuiCommandAssetTemplateSave:
		result, err := handleAssetCommand(ctx, command, databasePath, campaignID, today, loader)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandCodexCreate:
		result, err := handleCodexCreateOrUpdate(ctx, command, databasePath, campaignID, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandCodexUpdate:
		result, err := handleCodexCreateOrUpdate(ctx, command, databasePath, campaignID, false)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandCodexDelete:
		if err := codex.DeleteEntry(ctx, databasePath, campaignID, command.ItemKey); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Deleted codex entry %q.", command.ItemKey),
		}
	case tuiCommandNotesCreate:
		result, err := handleNotesCreateOrUpdate(ctx, command, databasePath, campaignID, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandNotesUpdate:
		result, err := handleNotesCreateOrUpdate(ctx, command, databasePath, campaignID, false)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandNotesDelete:
		if err := notes.DeleteNote(ctx, databasePath, campaignID, command.ItemKey); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Deleted note %q.", command.ItemKey),
		}
	case tuiCommandCodexTypeCreate:
		id := strings.TrimSpace(command.Fields["id"])
		name := strings.TrimSpace(command.Fields["name"])
		formID := strings.TrimSpace(command.Fields["form_id"])
		result, err := codex.CreateType(ctx, databasePath, id, name, formID)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Created codex type %q.", result.Name),
		}
		navigateTo = render.SectionSettings
		selectItemKey = "codex_type:" + result.ID
	case tuiCommandCodexTypeRename:
		newName := strings.TrimSpace(command.Fields["name"])
		// command.ItemKey is "codex_type:<id>", strip prefix.
		typeID := strings.TrimPrefix(command.ItemKey, "codex_type:")
		if err := codex.RenameType(ctx, databasePath, typeID, newName); err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Renamed codex type %q to %q.", typeID, newName),
		}
		selectItemKey = command.ItemKey
	case tuiCommandCodexTypeDelete:
		typeID := strings.TrimPrefix(command.ItemKey, "codex_type:")
		if err := codex.DeleteType(ctx, databasePath, typeID); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Deleted codex type %q.", typeID),
		}
	case tuiCommandCampaignCreate, tuiCommandCampaignRename, tuiCommandCampaignSwitch, tuiCommandCampaignDelete:
		return handleCampaignCommand(ctx, command, databasePath, loader)

	default:
		return render.CommandResult{}, fmt.Errorf("unsupported TUI command %q", command.ID)
	}

	data, err := buildTUIShellData(ctx, loader)
	if err != nil {
		return render.CommandResult{}, err
	}

	return render.CommandResult{
		Data:          data,
		Status:        message,
		NavigateTo:    navigateTo,
		SelectItemKey: selectItemKey,
	}, nil
}

func unavailableShellData(status *ledger.DatabaseStatus, detail string) render.ShellData {
	if status == nil {
		return render.ErrorShellData("Database status unavailable.", detail)
	}

	stateLine := fmt.Sprintf("Database state: %s.", status.State)
	if detail == "" {
		detail = "TUI data is not available for this database state."
	}

	return render.ShellData{
		Dashboard: render.DashboardData{
			HeaderLines:     []string{stateLine, detail},
			AccountsLines:   []string{"No account data loaded.", stateLine},
			JournalLines:    []string{"No journal data loaded.", stateLine},
			QuickEntryLines: []string{"Quick entry unavailable.", stateLine},
			LedgerLines:     []string{"No ledger totals loaded.", stateLine},
			QuestLines:      []string{"No quest register data loaded.", stateLine},
			LootLines:       []string{"No loot register data loaded.", stateLine},
			AssetLines:      []string{"No asset register data loaded.", stateLine},
		},
		Ledger:  unavailableSectionData(stateLine, detail),
		Journal: unavailableSectionData(stateLine, detail),
		Quests:  unavailableSectionData(stateLine, detail),
		Loot:    unavailableSectionData(stateLine, detail),
		Assets:  unavailableSectionData(stateLine, detail),
		Codex:   unavailableSectionData(stateLine, detail),
		Notes:   unavailableSectionData(stateLine, detail),
	}
}

func unavailableSectionData(summary string, detail string) render.ListScreenData {
	return render.ListScreenData{
		HeaderLines:  []string{summary, detail},
		SummaryLines: []string{"Data unavailable.", detail},
		EmptyLines:   []string{"No rows loaded.", detail},
	}
}

func unavailablePanelLines(err error) []string {
	return []string{
		"Data unavailable.",
		err.Error(),
	}
}

func buildEntryCatalog(accounts []ledger.AccountRecord, today string) render.EntryCatalog {
	catalog := render.EntryCatalog{
		DefaultDate: today,
	}

	for index := range accounts {
		record := accounts[index]
		if !record.Active {
			continue
		}

		option := render.AccountOption{
			Code: record.Code,
			Name: record.Name,
			Type: string(record.Type),
		}
		catalog.AllAccounts = append(catalog.AllAccounts, option)
		switch record.Type {
		case ledger.AccountTypeExpense:
			catalog.ExpenseAccounts = append(catalog.ExpenseAccounts, option)
		case ledger.AccountTypeIncome:
			catalog.IncomeAccounts = append(catalog.IncomeAccounts, option)
		case ledger.AccountTypeAsset:
			catalog.DepositAccounts = append(catalog.DepositAccounts, option)
			catalog.FundingAccounts = append(catalog.FundingAccounts, option)
		case ledger.AccountTypeLiability:
			catalog.FundingAccounts = append(catalog.FundingAccounts, option)
		case ledger.AccountTypeEquity:
		default:
		}
	}

	return catalog
}

func buildTUIJournalPostInput(command render.Command, today string) (ledger.JournalPostInput, error) {
	entryDate := strings.TrimSpace(command.Fields["date"])
	if entryDate == "" {
		entryDate = today
	}
	description := strings.TrimSpace(command.Fields["description"])
	if description == "" {
		return ledger.JournalPostInput{}, errors.New("description is required")
	}
	if len(command.Lines) < 2 {
		return ledger.JournalPostInput{}, errors.New("custom entry must contain at least 2 lines")
	}

	lines := make([]ledger.JournalLineInput, 0, len(command.Lines))
	for index := range command.Lines {
		line := command.Lines[index]
		accountCode := strings.TrimSpace(line.AccountCode)
		if accountCode == "" {
			return ledger.JournalPostInput{}, fmt.Errorf("line %d account code is required", index+1)
		}
		amountText := strings.TrimSpace(line.Amount)
		if amountText == "" {
			return ledger.JournalPostInput{}, fmt.Errorf("line %d amount is required", index+1)
		}

		amount, err := currency.ParseAmount(amountText)
		if err != nil {
			return ledger.JournalPostInput{}, fmt.Errorf("line %d amount %q is invalid", index+1, amountText)
		}
		if amount <= 0 {
			return ledger.JournalPostInput{}, fmt.Errorf("line %d amount must be positive", index+1)
		}

		journalLine := ledger.JournalLineInput{
			AccountCode: accountCode,
			Memo:        strings.TrimSpace(line.Memo),
		}
		switch strings.TrimSpace(line.Side) {
		case "debit":
			journalLine.DebitAmount = amount
		case "credit":
			journalLine.CreditAmount = amount
		default:
			return ledger.JournalPostInput{}, fmt.Errorf("line %d side must be debit or credit", index+1)
		}

		lines = append(lines, journalLine)
	}

	return ledger.JournalPostInput{
		EntryDate:   entryDate,
		Description: description,
		Lines:       lines,
	}, nil
}

func defaultDate(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func summarizeLedger(trialBalance report.TrialBalanceReport) []string {
	status := "BALANCED"
	if !trialBalance.Balanced {
		status = "UNBALANCED"
	}

	return []string{
		fmt.Sprintf("Posted accounts: %d", len(trialBalance.Accounts)),
		"Debits: " + currency.FormatAmount(trialBalance.TotalDebits),
		"Credits: " + currency.FormatAmount(trialBalance.TotalCredits),
		"Status: " + status,
	}
}

func summarizeShareableGold(trialBalance report.TrialBalanceReport, lootRows []report.LootSummaryRow, lootAvailable bool) []string {
	cashBalance := int64(0)
	cashFound := false
	for _, row := range trialBalance.Accounts {
		if row.AccountCode == "1000" {
			cashBalance = row.Balance
			cashFound = true
			break
		}
		if row.AccountType == ledger.AccountTypeAsset && strings.EqualFold(row.AccountName, "Party Cash") {
			cashBalance = row.Balance
			cashFound = true
			break
		}
	}

	shareLine := "To share now: unknown"
	if cashFound {
		shareLine = "To share now: " + currency.FormatAmount(cashBalance)
	}

	if !lootAvailable {
		return []string{
			shareLine,
			"Unsold loot: unavailable",
		}
	}

	recognizedLoot := int64(0)
	for _, row := range lootRows {
		if row.Status == ledger.LootStatusRecognized {
			recognizedLoot += row.LatestAppraisalValue
		}
	}

	return []string{
		shareLine,
		"Unsold loot: " + currency.FormatAmount(recognizedLoot),
	}
}

// tuiCommandResult bundles the fields that most TUI command handlers need to
// propagate back to the main switch.
type tuiCommandResult struct {
	message       render.StatusMessage
	navigateTo    render.Section
	selectItemKey string
}

// handleEntryCommand extracts the shared expense/income TUI command logic.
func handleEntryCommand(
	command render.Command,
	today string,
	label string,
	poster func(date, desc, acct, offset string, amt int64, memo string) (ledger.PostedJournalEntry, error),
) (tuiCommandResult, error) {
	amountText := strings.TrimSpace(command.Fields["amount"])
	if amountText == "" {
		return tuiCommandResult{}, render.InputError{Message: "Amount is required."}
	}
	amount, err := currency.ParseAmount(amountText)
	if err != nil {
		return tuiCommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
	}

	result, err := poster(
		defaultDate(strings.TrimSpace(command.Fields["date"]), today),
		strings.TrimSpace(command.Fields["description"]),
		strings.TrimSpace(command.Fields["account_code"]),
		strings.TrimSpace(command.Fields["offset_account_code"]),
		amount,
		strings.TrimSpace(command.Fields["memo"]),
	)
	if err != nil {
		return tuiCommandResult{}, render.InputError{Message: err.Error()}
	}

	return tuiCommandResult{
		message: render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Recorded %s as journal entry #%d.", label, result.EntryNumber),
		},
		navigateTo:    render.SectionJournal,
		selectItemKey: result.ID,
	}, nil
}

func handleAccountCommand(ctx context.Context, command render.Command, databasePath, campaignID string) (tuiCommandResult, error) {
	switch command.ID {
	case tuiCommandAccountCreate:
		accountType := ledger.AccountType(strings.TrimSpace(command.Fields["account_type"]))
		result, err := account.CreateAccount(
			ctx,
			databasePath,
			campaignID,
			strings.TrimSpace(command.Fields["code"]),
			strings.TrimSpace(command.Fields["name"]),
			accountType,
		)
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		return tuiCommandResult{
			message:       render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Created account %s.", result.Code)},
			navigateTo:    render.SectionSettings,
			selectItemKey: result.Code,
		}, nil
	case tuiCommandAccountRename:
		newName := strings.TrimSpace(command.Fields["name"])
		if err := account.RenameAccount(ctx, databasePath, campaignID, command.ItemKey, newName); err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		return tuiCommandResult{
			message:       render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Renamed account %s to %q.", command.ItemKey, newName)},
			selectItemKey: command.ItemKey,
		}, nil
	case tuiCommandAccountActivate:
		if err := account.ActivateAccount(ctx, databasePath, campaignID, command.ItemKey); err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Account %s activated.", command.ItemKey)},
		}, nil
	case tuiCommandAccountDeactivate:
		if err := account.DeactivateAccount(ctx, databasePath, campaignID, command.ItemKey); err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Account %s deactivated.", command.ItemKey)},
		}, nil
	case tuiCommandAccountDelete:
		if err := account.DeleteAccount(ctx, databasePath, campaignID, command.ItemKey); err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Removed account %s.", command.ItemKey)},
		}, nil
	default:
		return tuiCommandResult{}, fmt.Errorf("unsupported account command %q", command.ID)
	}
}

func handleLootCommand(ctx context.Context, command render.Command, databasePath, campaignID, today string, loader TUIDataLoader) (tuiCommandResult, error) { //nolint:revive // cyclomatic: loot command dispatch switch
	switch command.ID {
	case tuiCommandLootCreate:
		return handleItemCreateOrUpdate(ctx, command, databasePath, campaignID, "loot", "loot item", render.SectionLoot, true)
	case tuiCommandLootUpdate:
		return handleItemCreateOrUpdate(ctx, command, databasePath, campaignID, "loot", "loot item", render.SectionLoot, false)
	case tuiCommandLootAppraise:
		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return tuiCommandResult{}, render.InputError{Message: "Appraised value is required."}
		}
		amount, err := currency.ParseAmount(amountText)
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}
		if amount < 0 {
			return tuiCommandResult{}, render.InputError{Message: "Appraised value must be non-negative."}
		}
		if _, err := loot.AppraiseLootItem(ctx, databasePath, campaignID, command.ItemKey, amount, "", today, ""); err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Appraised loot item at %s.", currency.FormatAmount(amount))},
		}, nil
	case tuiCommandLootRecognize:
		items, err := loader.ListBrowseLootItems(ctx, "loot")
		if err != nil {
			return tuiCommandResult{}, err
		}
		item, ok := findBrowseLootItem(items, command.ItemKey)
		if !ok {
			return tuiCommandResult{}, fmt.Errorf("loot item %q does not exist", command.ItemKey)
		}
		if !lootRecognizable(&item) {
			return tuiCommandResult{}, fmt.Errorf("loot item %q cannot be recognized right now", command.ItemKey)
		}
		result, err := loot.RecognizeLootAppraisal(ctx, databasePath, campaignID, item.LatestAppraisal.ID, today, "")
		if err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Recognized loot item %q as entry #%d.", item.Name, result.EntryNumber)},
		}, nil
	case tuiCommandLootSell:
		items, err := loader.ListBrowseLootItems(ctx, "loot")
		if err != nil {
			return tuiCommandResult{}, err
		}
		item, ok := findBrowseLootItem(items, command.ItemKey)
		if !ok {
			return tuiCommandResult{}, fmt.Errorf("loot item %q does not exist", command.ItemKey)
		}
		if !lootSellable(&item) {
			return tuiCommandResult{}, fmt.Errorf("loot item %q cannot be sold right now", command.ItemKey)
		}
		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return tuiCommandResult{}, render.InputError{Message: "Sale amount is required."}
		}
		amount, err := currency.ParseAmount(amountText)
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}
		if amount <= 0 {
			return tuiCommandResult{}, render.InputError{Message: "Sale amount must be positive."}
		}
		result, err := loot.SellLootItem(ctx, databasePath, campaignID, command.ItemKey, amount, today, "")
		if err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Sold loot item %q as entry #%d.", item.Name, result.EntryNumber)},
		}, nil
	case tuiCommandLootTransferToAsset:
		if err := loot.TransferItemType(ctx, databasePath, campaignID, command.ItemKey, "asset"); err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message:       render.StatusMessage{Level: render.StatusSuccess, Text: "Transferred item to asset register."},
			navigateTo:    render.SectionAssets,
			selectItemKey: command.ItemKey,
		}, nil
	default:
		return tuiCommandResult{}, fmt.Errorf("unsupported loot command %q", command.ID)
	}
}

func handleAssetCommand(ctx context.Context, command render.Command, databasePath, campaignID, today string, loader TUIDataLoader) (tuiCommandResult, error) {
	switch command.ID {
	case tuiCommandAssetCreate:
		return handleItemCreateOrUpdate(ctx, command, databasePath, campaignID, "asset", "asset", render.SectionAssets, true)
	case tuiCommandAssetUpdate:
		return handleItemCreateOrUpdate(ctx, command, databasePath, campaignID, "asset", "asset", render.SectionAssets, false)
	case tuiCommandAssetAppraise:
		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return tuiCommandResult{}, render.InputError{Message: "Appraised value is required."}
		}
		amount, err := currency.ParseAmount(amountText)
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}
		if amount < 0 {
			return tuiCommandResult{}, render.InputError{Message: "Appraised value must be non-negative."}
		}
		if _, err := loot.AppraiseLootItem(ctx, databasePath, campaignID, command.ItemKey, amount, "", today, ""); err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Appraised asset at %s.", currency.FormatAmount(amount))},
		}, nil
	case tuiCommandAssetRecognize:
		items, err := loader.ListBrowseLootItems(ctx, "asset")
		if err != nil {
			return tuiCommandResult{}, err
		}
		item, ok := findBrowseLootItem(items, command.ItemKey)
		if !ok {
			return tuiCommandResult{}, fmt.Errorf("asset %q does not exist", command.ItemKey)
		}
		if !lootRecognizable(&item) {
			return tuiCommandResult{}, fmt.Errorf("asset %q cannot be recognized right now", command.ItemKey)
		}
		result, err := loot.RecognizeLootAppraisal(ctx, databasePath, campaignID, item.LatestAppraisal.ID, today, "")
		if err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Recognized asset %q as entry #%d.", item.Name, result.EntryNumber)},
		}, nil
	case tuiCommandAssetTransferToLoot:
		if err := loot.TransferItemType(ctx, databasePath, campaignID, command.ItemKey, "loot"); err != nil {
			return tuiCommandResult{}, err
		}
		return tuiCommandResult{
			message:       render.StatusMessage{Level: render.StatusSuccess, Text: "Transferred item to loot register."},
			navigateTo:    render.SectionLoot,
			selectItemKey: command.ItemKey,
		}, nil
	case tuiCommandAssetTemplateSave:
		lines := make([]loot.AssetTemplateLineRecord, 0, len(command.Lines))
		for _, cl := range command.Lines {
			lines = append(lines, loot.AssetTemplateLineRecord{
				Side:        strings.TrimSpace(cl.Side),
				AccountCode: strings.TrimSpace(cl.AccountCode),
				Amount:      strings.TrimSpace(cl.Amount),
			})
		}
		if err := loot.SaveAssetTemplate(ctx, databasePath, campaignID, command.ItemKey, lines); err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		return tuiCommandResult{
			message:       render.StatusMessage{Level: render.StatusSuccess, Text: "Saved entry template."},
			navigateTo:    render.SectionAssets,
			selectItemKey: command.ItemKey,
		}, nil
	default:
		return tuiCommandResult{}, fmt.Errorf("unsupported asset command %q", command.ID)
	}
}

func handleCampaignCommand(ctx context.Context, command render.Command, databasePath string, loader TUIDataLoader) (render.CommandResult, error) { //nolint:revive // cognitive-complexity: campaign command dispatch
	switch command.ID {
	case tuiCommandCampaignCreate:
		name := command.Fields["name"]
		record, err := campaign.Create(ctx, databasePath, name, loader.SeedAccounts())
		if err != nil {
			return render.CommandResult{}, err
		}
		if err := campaign.SetActive(ctx, databasePath, record.ID); err != nil {
			return render.CommandResult{}, err
		}
		loader.SetCampaign(record.ID, record.Name)
		data, loadErr := buildTUIShellData(ctx, loader)
		if loadErr != nil {
			return render.CommandResult{}, loadErr
		}
		return render.CommandResult{
			Data:   data,
			Status: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Campaign %q created.", record.Name)},
		}, nil

	case tuiCommandCampaignRename:
		name := command.Fields["amount"]
		if name == "" {
			name = command.Fields["name"]
		}
		renameCampaignID := command.ItemKey
		if renameCampaignID == "" {
			renameCampaignID = loader.CampaignID()
		}
		record, err := campaign.Rename(ctx, databasePath, renameCampaignID, name)
		if err != nil {
			return render.CommandResult{}, err
		}
		if record.ID == loader.CampaignID() {
			loader.SetCampaign(record.ID, record.Name)
		}
		data, loadErr := buildTUIShellData(ctx, loader)
		if loadErr != nil {
			return render.CommandResult{}, loadErr
		}
		return render.CommandResult{
			Data:   data,
			Status: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Campaign renamed to %q.", record.Name)},
		}, nil

	case tuiCommandCampaignSwitch:
		campaignSwitchID := command.ItemKey
		if err := campaign.SetActive(ctx, databasePath, campaignSwitchID); err != nil {
			return render.CommandResult{}, err
		}
		campaigns, err := campaign.List(ctx, databasePath)
		if err != nil {
			return render.CommandResult{}, err
		}
		var selectedName string
		for _, c := range campaigns {
			if c.ID == campaignSwitchID {
				selectedName = c.Name
				break
			}
		}
		loader.SetCampaign(campaignSwitchID, selectedName)
		data, loadErr := buildTUIShellData(ctx, loader)
		if loadErr != nil {
			return render.CommandResult{}, loadErr
		}
		return render.CommandResult{
			Data:   data,
			Status: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Switched to campaign %q.", selectedName)},
		}, nil

	case tuiCommandCampaignDelete:
		deleteID := command.ItemKey
		if deleteID == loader.CampaignID() {
			return render.CommandResult{}, errors.New("cannot delete the active campaign")
		}
		if err := campaign.Delete(ctx, databasePath, deleteID); err != nil {
			return render.CommandResult{}, err
		}
		data, loadErr := buildTUIShellData(ctx, loader)
		if loadErr != nil {
			return render.CommandResult{}, loadErr
		}
		return render.CommandResult{
			Data:   data,
			Status: render.StatusMessage{Level: render.StatusSuccess, Text: "Campaign deleted."},
		}, nil

	default:
		return render.CommandResult{}, fmt.Errorf("unsupported campaign command %q", command.ID)
	}
}

// handleQuestCreateOrUpdate extracts the shared quest create and update TUI
// command logic. When create is true it calls CreateQuest; otherwise it calls
// UpdateQuest.
func handleQuestCreateOrUpdate(
	ctx context.Context,
	command render.Command,
	databasePath string,
	campaignID string,
	create bool,
) (tuiCommandResult, error) {
	rewardText := strings.TrimSpace(command.Fields["reward"])
	if rewardText == "" {
		rewardText = "0"
	}
	reward, err := currency.ParseAmount(rewardText)
	if err != nil {
		return tuiCommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid reward %q.", rewardText)}
	}
	advanceText := strings.TrimSpace(command.Fields["advance"])
	if advanceText == "" {
		advanceText = "0"
	}
	advance, err := currency.ParseAmount(advanceText)
	if err != nil {
		return tuiCommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid advance %q.", advanceText)}
	}

	title := strings.TrimSpace(command.Fields["title"])
	patron := strings.TrimSpace(command.Fields["patron"])
	description := strings.TrimSpace(command.Fields["description"])
	bonus := strings.TrimSpace(command.Fields["bonus"])
	questNotes := strings.TrimSpace(command.Fields["notes"])
	acceptedOn := strings.TrimSpace(command.Fields["accepted_on"])

	var resultTitle, resultID string
	if create {
		result, createErr := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
			Title:              title,
			Patron:             patron,
			Description:        description,
			PromisedBaseReward: reward,
			PartialAdvance:     advance,
			BonusConditions:    bonus,
			Notes:              questNotes,
			Status:             strings.TrimSpace(command.Fields["status"]),
			AcceptedOn:         acceptedOn,
		})
		if createErr != nil {
			return tuiCommandResult{}, render.InputError{Message: createErr.Error()}
		}
		resultTitle = result.Title
		resultID = result.ID
	} else {
		result, updateErr := quest.UpdateQuest(ctx, databasePath, campaignID, command.ItemKey, &quest.UpdateQuestInput{
			Title:              title,
			Patron:             patron,
			Description:        description,
			PromisedBaseReward: reward,
			PartialAdvance:     advance,
			BonusConditions:    bonus,
			Notes:              questNotes,
			AcceptedOn:         acceptedOn,
		})
		if updateErr != nil {
			return tuiCommandResult{}, render.InputError{Message: updateErr.Error()}
		}
		resultTitle = result.Title
		resultID = result.ID
	}

	verb := "Created"
	if !create {
		verb = "Updated"
	}

	return tuiCommandResult{
		message: render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("%s quest %q.", verb, resultTitle),
		},
		navigateTo:    render.SectionQuests,
		selectItemKey: resultID,
	}, nil
}

// handleItemCreateOrUpdate extracts the shared loot/asset create and update
// TUI command logic. When create is true it calls CreateLootItem; otherwise it
// calls UpdateLootItem.
func handleItemCreateOrUpdate(
	ctx context.Context,
	command render.Command,
	databasePath string,
	campaignID string,
	itemType string,
	itemLabel string,
	section render.Section,
	create bool,
) (tuiCommandResult, error) {
	quantityText := strings.TrimSpace(command.Fields["quantity"])
	if quantityText == "" {
		quantityText = "1"
	}
	quantity, err := strconv.Atoi(quantityText)
	if err != nil {
		return tuiCommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid quantity %q.", quantityText)}
	}
	if quantity <= 0 {
		return tuiCommandResult{}, render.InputError{Message: "Quantity must be positive."}
	}

	var name, id string
	if create {
		result, createErr := loot.CreateLootItem(
			ctx,
			databasePath,
			campaignID,
			strings.TrimSpace(command.Fields["name"]),
			strings.TrimSpace(command.Fields["source"]),
			quantity,
			strings.TrimSpace(command.Fields["holder"]),
			strings.TrimSpace(command.Fields["notes"]),
			itemType,
		)
		if createErr != nil {
			return tuiCommandResult{}, render.InputError{Message: createErr.Error()}
		}
		name = result.Name
		id = result.ID
	} else {
		result, updateErr := loot.UpdateLootItem(ctx, databasePath, campaignID, command.ItemKey, &loot.UpdateLootItemInput{
			Name:     strings.TrimSpace(command.Fields["name"]),
			Source:   strings.TrimSpace(command.Fields["source"]),
			Quantity: quantity,
			Holder:   strings.TrimSpace(command.Fields["holder"]),
			Notes:    strings.TrimSpace(command.Fields["notes"]),
		})
		if updateErr != nil {
			return tuiCommandResult{}, render.InputError{Message: updateErr.Error()}
		}
		name = result.Name
		id = result.ID
	}

	verb := "Created"
	if !create {
		verb = "Updated"
	}

	return tuiCommandResult{
		message: render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("%s %s %q.", verb, itemLabel, name),
		},
		navigateTo:    section,
		selectItemKey: id,
	}, nil
}

func tuiToday() string {
	return tuiNow().Format("2006-01-02")
}

func blankStatusDetail(detail string) string {
	if strings.TrimSpace(detail) == "" {
		return "TUI data is not available for this database state."
	}
	return detail
}

// handleCodexCreateOrUpdate extracts the shared codex create and update TUI
// command logic. When create is true it calls CreateEntry; otherwise UpdateEntry.
func handleCodexCreateOrUpdate(
	ctx context.Context,
	command render.Command,
	databasePath string,
	campaignID string,
	create bool,
) (tuiCommandResult, error) {
	typeID := strings.TrimSpace(command.Fields["_type_id"])
	name := strings.TrimSpace(command.Fields["name"])
	title := strings.TrimSpace(command.Fields["title"])
	location := strings.TrimSpace(command.Fields["location"])
	faction := strings.TrimSpace(command.Fields["faction"])
	disposition := strings.TrimSpace(command.Fields["disposition"])
	playerName := strings.TrimSpace(command.Fields["player_name"])
	class := strings.TrimSpace(command.Fields["class"])
	race := strings.TrimSpace(command.Fields["race"])
	background := strings.TrimSpace(command.Fields["background"])
	description := strings.TrimSpace(command.Fields["description"])
	codexNotes := strings.TrimSpace(command.Fields["notes"])

	var resultName, resultID string
	if create {
		result, err := codex.CreateEntry(ctx, databasePath, campaignID, &codex.CreateInput{
			TypeID:      typeID,
			Name:        name,
			Title:       title,
			Location:    location,
			Faction:     faction,
			Disposition: disposition,
			PlayerName:  playerName,
			Class:       class,
			Race:        race,
			Background:  background,
			Description: description,
			Notes:       codexNotes,
		})
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		resultName = result.Name
		resultID = result.ID
	} else {
		result, err := codex.UpdateEntry(ctx, databasePath, campaignID, command.ItemKey, &codex.UpdateInput{
			TypeID:      typeID,
			Name:        name,
			Title:       title,
			Location:    location,
			Faction:     faction,
			Disposition: disposition,
			PlayerName:  playerName,
			Class:       class,
			Race:        race,
			Background:  background,
			Description: description,
			Notes:       codexNotes,
		})
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		resultName = result.Name
		resultID = result.ID
	}

	verb := "Created"
	if !create {
		verb = "Updated"
	}

	return tuiCommandResult{
		message: render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("%s codex entry %q.", verb, resultName),
		},
		navigateTo:    render.SectionCodex,
		selectItemKey: resultID,
	}, nil
}

// handleNotesCreateOrUpdate extracts the shared notes create and update TUI
// command logic. When create is true it calls CreateNote; otherwise UpdateNote.
func handleNotesCreateOrUpdate(
	ctx context.Context,
	command render.Command,
	databasePath string,
	campaignID string,
	create bool,
) (tuiCommandResult, error) {
	title := strings.TrimSpace(command.Fields["title"])
	body := strings.TrimSpace(command.Fields["body"])

	var resultTitle, resultID string
	if create {
		result, err := notes.CreateNote(ctx, databasePath, campaignID, &notes.CreateNoteInput{
			Title: title,
			Body:  body,
		})
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		resultTitle = result.Title
		resultID = result.ID
	} else {
		result, err := notes.UpdateNote(ctx, databasePath, campaignID, command.ItemKey, &notes.UpdateNoteInput{
			Title: title,
			Body:  body,
		})
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		resultTitle = result.Title
		resultID = result.ID
	}

	verb := "Created"
	if !create {
		verb = "Updated"
	}

	return tuiCommandResult{
		message: render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("%s note %q.", verb, resultTitle),
		},
		navigateTo:    render.SectionNotes,
		selectItemKey: resultID,
	}, nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}
