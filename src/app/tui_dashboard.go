package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/account"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/notes"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/quest"
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
	}

	databaseName := loader.DatabaseName()
	data := render.ShellData{
		Dashboard: render.DashboardData{
			HeaderLines: []string{
				fmt.Sprintf("Read-only snapshot from %s.", databaseName),
				"Use arrows, Tab, or 1-8 to move between boxed screens. Use e/i/a for guided entry creation.",
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
		Accounts: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Chart of accounts from %s.", databaseName),
				"Select an account to inspect it. `a` adds, `d` removes, and `t` toggles active/inactive.",
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

	accounts, err := loader.ListAccounts(ctx)
	if err != nil {
		data.Dashboard.AccountsLines = unavailablePanelLines(err)
		data.Accounts = unavailableSectionData("Accounts unavailable.", err.Error())
		panelErrors = append(panelErrors, "accounts")
	} else {
		data.Dashboard.AccountsLines = summarizeAccounts(accounts)
		data.Accounts.SummaryLines = summarizeAccounts(accounts)
		data.Accounts.ListHeaderRow = fmt.Sprintf("%-4s %-9s %-8s %s", "CODE", "TYPE", "STATUS", "NAME")
		data.Accounts.Items = buildAccountItems(accounts)
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

	trialBalance, err := loader.GetTrialBalance(ctx)
	trialBalanceAvailable := false
	if err != nil {
		data.Dashboard.LedgerLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "ledger")
	} else {
		data.Dashboard.LedgerLines = summarizeLedger(trialBalance)
		trialBalanceAvailable = true
	}

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
			data.Dashboard.QuestLines = summarizeQuests(promisedQuests, receivables)
			data.Quests.SummaryLines = summarizeQuests(promisedQuests, receivables)
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

	// Load codex entries, references, and types.
	var allCodexEntries []codex.CodexEntry
	var allCodexRefs map[string][]codex.Reference

	codexEntries, codexErr := loader.ListCodexEntries(ctx)
	if codexErr != nil {
		data.Codex = unavailableSectionData("Codex unavailable.", codexErr.Error())
		panelErrors = append(panelErrors, "codex")
	} else {
		allCodexEntries = codexEntries
		data.Codex.SummaryLines = summarizeCodex(codexEntries)
		data.Codex.ListHeaderRow = fmt.Sprintf("%-8s %-14s %s", "TYPE", "SECONDARY", "NAME")

		refs, refsErr := loader.ListAllCodexReferences(ctx)
		if refsErr != nil {
			panelErrors = append(panelErrors, "codex-refs")
		} else {
			allCodexRefs = refs
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

	// Load notes and references.
	var allNotes []notes.NoteRecord
	var allNotesRefs map[string][]notes.ReferenceRecord

	noteRecords, notesErr := loader.ListNotes(ctx)
	if notesErr != nil {
		data.Notes = unavailableSectionData("Notes unavailable.", notesErr.Error())
		panelErrors = append(panelErrors, "notes")
	} else {
		allNotes = noteRecords
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

	// Append "Mentioned by:" lines to quest detail views.
	if allCodexRefs != nil && allCodexEntries != nil {
		for i, item := range data.Quests.Items {
			mentionLines := buildMentionedByLines(allCodexRefs, allCodexEntries, "quest", item.DetailTitle)
			if len(mentionLines) > 0 {
				data.Quests.Items[i].DetailLines = append(data.Quests.Items[i].DetailLines, mentionLines...)
			}
		}
		for i, item := range data.Loot.Items {
			mentionLines := buildMentionedByLines(allCodexRefs, allCodexEntries, "loot", item.DetailTitle)
			if len(mentionLines) > 0 {
				data.Loot.Items[i].DetailLines = append(data.Loot.Items[i].DetailLines, mentionLines...)
			}
		}
		for i, item := range data.Assets.Items {
			mentionLines := buildMentionedByLines(allCodexRefs, allCodexEntries, "asset", item.DetailTitle)
			if len(mentionLines) > 0 {
				data.Assets.Items[i].DetailLines = append(data.Assets.Items[i].DetailLines, mentionLines...)
			}
		}
	}

	// Append "Referenced in:" lines from notes to quest/loot/asset/people detail views.
	if allNotesRefs != nil && allNotes != nil {
		for i, item := range data.Quests.Items {
			refLines := buildNoteReferencedInLines(allNotesRefs, allNotes, "quest", item.DetailTitle)
			if len(refLines) > 0 {
				data.Quests.Items[i].DetailLines = append(data.Quests.Items[i].DetailLines, refLines...)
			}
		}
		for i, item := range data.Loot.Items {
			refLines := buildNoteReferencedInLines(allNotesRefs, allNotes, "loot", item.DetailTitle)
			if len(refLines) > 0 {
				data.Loot.Items[i].DetailLines = append(data.Loot.Items[i].DetailLines, refLines...)
			}
		}
		for i, item := range data.Assets.Items {
			refLines := buildNoteReferencedInLines(allNotesRefs, allNotes, "asset", item.DetailTitle)
			if len(refLines) > 0 {
				data.Assets.Items[i].DetailLines = append(data.Assets.Items[i].DetailLines, refLines...)
			}
		}
		for i, item := range data.Codex.Items {
			refLines := buildNoteReferencedInLines(allNotesRefs, allNotes, "person", item.DetailTitle)
			if len(refLines) > 0 {
				data.Codex.Items[i].DetailLines = append(data.Codex.Items[i].DetailLines, refLines...)
			}
		}
	}

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

func handleTUICommand(ctx context.Context, command render.Command, databasePath string, loader TUIDataLoader) (render.CommandResult, error) {
	var message render.StatusMessage
	var navigateTo render.Section
	var selectItemKey string
	today := tuiToday()

	switch command.ID {
	case tuiCommandAccountCreate:
		accountType := ledger.AccountType(strings.TrimSpace(command.Fields["account_type"]))
		result, err := account.CreateAccount(
			ctx,
			databasePath,
			strings.TrimSpace(command.Fields["code"]),
			strings.TrimSpace(command.Fields["name"]),
			accountType,
		)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Created account %s.", result.Code),
		}
		navigateTo = render.SectionAccounts
		selectItemKey = result.Code
	case tuiCommandAccountRename:
		newName := strings.TrimSpace(command.Fields["name"])
		if err := account.RenameAccount(ctx, databasePath, command.ItemKey, newName); err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Renamed account %s to %q.", command.ItemKey, newName),
		}
		selectItemKey = command.ItemKey
	case tuiCommandAccountActivate:
		if err := account.ActivateAccount(ctx, databasePath, command.ItemKey); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Account %s activated.", command.ItemKey),
		}
	case tuiCommandAccountDeactivate:
		if err := account.DeactivateAccount(ctx, databasePath, command.ItemKey); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Account %s deactivated.", command.ItemKey),
		}
	case tuiCommandAccountDelete:
		if err := account.DeleteAccount(ctx, databasePath, command.ItemKey); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Removed account %s.", command.ItemKey),
		}
	case tuiCommandJournalReverse:
		entries, err := loader.ListBrowseJournalEntries(ctx)
		if err != nil {
			return render.CommandResult{}, err
		}

		entry, ok := findBrowseEntry(entries, command.ItemKey)
		if !ok {
			return render.CommandResult{}, fmt.Errorf("journal entry %q does not exist", command.ItemKey)
		}

		result, err := journal.ReverseJournalEntry(ctx, databasePath, command.ItemKey, entry.EntryDate, "")
		if err != nil {
			return render.CommandResult{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Entry #%d reversed as entry #%d.", entry.EntryNumber, result.EntryNumber),
		}
	case tuiCommandCreateExpense:
		result, err := handleEntryCommand(ctx, command, databasePath, today, "expense",
			func(date, desc, acct, offset string, amt int64, memo string) (ledger.PostedJournalEntry, error) {
				return journal.PostExpenseEntry(ctx, databasePath, &journal.ExpenseEntryInput{
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
		result, err := handleEntryCommand(ctx, command, databasePath, today, "income",
			func(date, desc, acct, offset string, amt int64, memo string) (ledger.PostedJournalEntry, error) {
				return journal.PostIncomeEntry(ctx, databasePath, &journal.IncomeEntryInput{
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

		result, err := journal.PostJournalEntry(ctx, databasePath, input)
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
		result, err := handleQuestCreateOrUpdate(ctx, command, databasePath, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandQuestUpdate:
		result, err := handleQuestCreateOrUpdate(ctx, command, databasePath, false)
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

		result, err := quest.CollectQuestPayment(ctx, databasePath, quest.CollectQuestPaymentInput{
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

		result, err := quest.WriteOffQuest(ctx, databasePath, quest.WriteOffQuestInput{
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
	case tuiCommandLootCreate:
		result, err := handleItemCreateOrUpdate(ctx, command, databasePath, "loot", "loot item", render.SectionLoot, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandLootUpdate:
		result, err := handleItemCreateOrUpdate(ctx, command, databasePath, "loot", "loot item", render.SectionLoot, false)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandLootAppraise:
		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return render.CommandResult{}, render.InputError{Message: "Appraised value is required."}
		}
		amount, err := currency.ParseAmount(amountText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}
		if amount < 0 {
			return render.CommandResult{}, render.InputError{Message: "Appraised value must be non-negative."}
		}
		if _, err := loot.AppraiseLootItem(ctx, databasePath, command.ItemKey, amount, "", today, ""); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Appraised loot item at %s.", currency.FormatAmount(amount)),
		}
	case tuiCommandLootRecognize:
		items, err := loader.ListBrowseLootItems(ctx, "loot")
		if err != nil {
			return render.CommandResult{}, err
		}

		item, ok := findBrowseLootItem(items, command.ItemKey)
		if !ok {
			return render.CommandResult{}, fmt.Errorf("loot item %q does not exist", command.ItemKey)
		}
		if !lootRecognizable(&item) {
			return render.CommandResult{}, fmt.Errorf("loot item %q cannot be recognized right now", command.ItemKey)
		}

		result, err := loot.RecognizeLootAppraisal(ctx, databasePath, item.LatestAppraisal.ID, today, "")
		if err != nil {
			return render.CommandResult{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Recognized loot item %q as entry #%d.", item.Name, result.EntryNumber),
		}
	case tuiCommandLootSell:
		items, err := loader.ListBrowseLootItems(ctx, "loot")
		if err != nil {
			return render.CommandResult{}, err
		}

		item, ok := findBrowseLootItem(items, command.ItemKey)
		if !ok {
			return render.CommandResult{}, fmt.Errorf("loot item %q does not exist", command.ItemKey)
		}
		if !lootSellable(&item) {
			return render.CommandResult{}, fmt.Errorf("loot item %q cannot be sold right now", command.ItemKey)
		}

		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return render.CommandResult{}, render.InputError{Message: "Sale amount is required."}
		}

		amount, err := currency.ParseAmount(amountText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}
		if amount <= 0 {
			return render.CommandResult{}, render.InputError{Message: "Sale amount must be positive."}
		}

		result, err := loot.SellLootItem(ctx, databasePath, command.ItemKey, amount, today, "")
		if err != nil {
			return render.CommandResult{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Sold loot item %q as entry #%d.", item.Name, result.EntryNumber),
		}
	case tuiCommandLootTransferToAsset:
		if err := loot.TransferItemType(ctx, databasePath, command.ItemKey, "asset"); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  "Transferred item to asset register.",
		}
		navigateTo = render.SectionAssets
		selectItemKey = command.ItemKey
	case tuiCommandAssetCreate:
		result, err := handleItemCreateOrUpdate(ctx, command, databasePath, "asset", "asset", render.SectionAssets, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandAssetUpdate:
		result, err := handleItemCreateOrUpdate(ctx, command, databasePath, "asset", "asset", render.SectionAssets, false)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandAssetAppraise:
		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return render.CommandResult{}, render.InputError{Message: "Appraised value is required."}
		}
		amount, err := currency.ParseAmount(amountText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}
		if amount < 0 {
			return render.CommandResult{}, render.InputError{Message: "Appraised value must be non-negative."}
		}
		if _, err := loot.AppraiseLootItem(ctx, databasePath, command.ItemKey, amount, "", today, ""); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Appraised asset at %s.", currency.FormatAmount(amount)),
		}
	case tuiCommandAssetRecognize:
		items, err := loader.ListBrowseLootItems(ctx, "asset")
		if err != nil {
			return render.CommandResult{}, err
		}

		item, ok := findBrowseLootItem(items, command.ItemKey)
		if !ok {
			return render.CommandResult{}, fmt.Errorf("asset %q does not exist", command.ItemKey)
		}
		if !lootRecognizable(&item) {
			return render.CommandResult{}, fmt.Errorf("asset %q cannot be recognized right now", command.ItemKey)
		}

		result, err := loot.RecognizeLootAppraisal(ctx, databasePath, item.LatestAppraisal.ID, today, "")
		if err != nil {
			return render.CommandResult{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Recognized asset %q as entry #%d.", item.Name, result.EntryNumber),
		}
	case tuiCommandAssetTransferToLoot:
		if err := loot.TransferItemType(ctx, databasePath, command.ItemKey, "loot"); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  "Transferred item to loot register.",
		}
		navigateTo = render.SectionLoot
		selectItemKey = command.ItemKey
	case tuiCommandAssetTemplateSave:
		lines := make([]loot.AssetTemplateLineRecord, 0, len(command.Lines))
		for _, cl := range command.Lines {
			lines = append(lines, loot.AssetTemplateLineRecord{
				Side:        strings.TrimSpace(cl.Side),
				AccountCode: strings.TrimSpace(cl.AccountCode),
				Amount:      strings.TrimSpace(cl.Amount),
			})
		}
		if err := loot.SaveAssetTemplate(ctx, databasePath, command.ItemKey, lines); err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  "Saved entry template.",
		}
		navigateTo = render.SectionAssets
		selectItemKey = command.ItemKey
	case tuiCommandCodexCreate:
		result, err := handleCodexCreateOrUpdate(ctx, command, databasePath, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandCodexUpdate:
		result, err := handleCodexCreateOrUpdate(ctx, command, databasePath, false)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandCodexDelete:
		if err := codex.DeleteEntry(ctx, databasePath, command.ItemKey); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Deleted codex entry %q.", command.ItemKey),
		}
	case tuiCommandNotesCreate:
		result, err := handleNotesCreateOrUpdate(ctx, command, databasePath, true)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandNotesUpdate:
		result, err := handleNotesCreateOrUpdate(ctx, command, databasePath, false)
		if err != nil {
			return render.CommandResult{}, err
		}
		message = result.message
		navigateTo = result.navigateTo
		selectItemKey = result.selectItemKey
	case tuiCommandNotesDelete:
		if err := notes.DeleteNote(ctx, databasePath, command.ItemKey); err != nil {
			return render.CommandResult{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Deleted note %q.", command.ItemKey),
		}
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

func buildTUIDashboardData(ctx context.Context, loader TUIDataLoader) (render.DashboardData, error) {
	data, err := buildTUIShellData(ctx, loader)
	if err != nil {
		return render.ErrorDashboardData("Dashboard data unavailable.", err.Error()), nil
	}

	return data.Dashboard, nil
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
		Accounts: unavailableSectionData(stateLine, detail),
		Journal:  unavailableSectionData(stateLine, detail),
		Quests:   unavailableSectionData(stateLine, detail),
		Loot:     unavailableSectionData(stateLine, detail),
		Assets:   unavailableSectionData(stateLine, detail),
		Codex:   unavailableSectionData(stateLine, detail),
		Notes:    unavailableSectionData(stateLine, detail),
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
		return ledger.JournalPostInput{}, fmt.Errorf("description is required")
	}
	if len(command.Lines) < 2 {
		return ledger.JournalPostInput{}, fmt.Errorf("custom entry must contain at least 2 lines")
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
	ctx context.Context,
	command render.Command,
	databasePath string,
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

// handleQuestCreateOrUpdate extracts the shared quest create and update TUI
// command logic. When create is true it calls CreateQuest; otherwise it calls
// UpdateQuest.
func handleQuestCreateOrUpdate(
	ctx context.Context,
	command render.Command,
	databasePath string,
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
	notes := strings.TrimSpace(command.Fields["notes"])
	acceptedOn := strings.TrimSpace(command.Fields["accepted_on"])

	var resultTitle, resultID string
	if create {
		result, createErr := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
			Title:              title,
			Patron:             patron,
			Description:        description,
			PromisedBaseReward: reward,
			PartialAdvance:     advance,
			BonusConditions:    bonus,
			Notes:              notes,
			Status:             strings.TrimSpace(command.Fields["status"]),
			AcceptedOn:         acceptedOn,
		})
		if createErr != nil {
			return tuiCommandResult{}, render.InputError{Message: createErr.Error()}
		}
		resultTitle = result.Title
		resultID = result.ID
	} else {
		result, updateErr := quest.UpdateQuest(ctx, databasePath, command.ItemKey, &quest.UpdateQuestInput{
			Title:              title,
			Patron:             patron,
			Description:        description,
			PromisedBaseReward: reward,
			PartialAdvance:     advance,
			BonusConditions:    bonus,
			Notes:              notes,
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
		result, updateErr := loot.UpdateLootItem(ctx, databasePath, command.ItemKey, &loot.UpdateLootItemInput{
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
	create bool,
) (tuiCommandResult, error) {
	typeID := strings.TrimSpace(command.Fields["_type_id"])
	name := strings.TrimSpace(command.Fields["name"])
	title := strings.TrimSpace(command.Fields["title"])
	location := strings.TrimSpace(command.Fields["location"])
	faction := strings.TrimSpace(command.Fields["faction"])
	disposition := strings.TrimSpace(command.Fields["disposition"])
	class := strings.TrimSpace(command.Fields["class"])
	race := strings.TrimSpace(command.Fields["race"])
	background := strings.TrimSpace(command.Fields["background"])
	description := strings.TrimSpace(command.Fields["description"])
	notes := strings.TrimSpace(command.Fields["notes"])

	var resultName, resultID string
	if create {
		result, err := codex.CreateEntry(ctx, databasePath, &codex.CreateInput{
			TypeID:      typeID,
			Name:        name,
			Title:       title,
			Location:    location,
			Faction:     faction,
			Disposition: disposition,
			Class:       class,
			Race:        race,
			Background:  background,
			Description: description,
			Notes:       notes,
		})
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		resultName = result.Name
		resultID = result.ID
	} else {
		result, err := codex.UpdateEntry(ctx, databasePath, command.ItemKey, &codex.UpdateInput{
			TypeID:      typeID,
			Name:        name,
			Title:       title,
			Location:    location,
			Faction:     faction,
			Disposition: disposition,
			Class:       class,
			Race:        race,
			Background:  background,
			Description: description,
			Notes:       notes,
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
	create bool,
) (tuiCommandResult, error) {
	title := strings.TrimSpace(command.Fields["title"])
	body := strings.TrimSpace(command.Fields["body"])

	var resultTitle, resultID string
	if create {
		result, err := notes.CreateNote(ctx, databasePath, &notes.CreateNoteInput{
			Title: title,
			Body:  body,
		})
		if err != nil {
			return tuiCommandResult{}, render.InputError{Message: err.Error()}
		}
		resultTitle = result.Title
		resultID = result.ID
	} else {
		result, err := notes.UpdateNote(ctx, databasePath, command.ItemKey, &notes.UpdateNoteInput{
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
