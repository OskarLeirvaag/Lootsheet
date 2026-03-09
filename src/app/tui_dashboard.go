package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/account"
	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

const (
	tuiCommandAccountCreate     = "account.create"
	tuiCommandAccountActivate   = "account.activate"
	tuiCommandAccountDeactivate = "account.deactivate"
	tuiCommandAccountDelete     = "account.delete"
	tuiCommandJournalReverse    = "journal.reverse"
	tuiCommandCreateExpense     = "entry.expense.create"
	tuiCommandCreateIncome      = "entry.income.create"
	tuiCommandCreateCustom      = "entry.custom.create"
	tuiCommandQuestCreate       = "quest.create"
	tuiCommandQuestUpdate       = "quest.update"
	tuiCommandQuestCollectFull  = "quest.collect_full"
	tuiCommandQuestWriteOffFull = "quest.writeoff_full"
	tuiCommandLootCreate             = "loot.create"
	tuiCommandLootUpdate             = "loot.update"
	tuiCommandLootRecognize          = "loot.recognize_latest"
	tuiCommandLootSell               = "loot.sell"
	tuiCommandLootTransferToAsset    = "loot.transfer_to_asset"
	tuiCommandAssetCreate            = "asset.create"
	tuiCommandAssetUpdate            = "asset.update"
	tuiCommandAssetRecognize         = "asset.recognize_latest"
	tuiCommandAssetTransferToLoot    = "asset.transfer_to_loot"
)

var tuiNow = time.Now

func buildTUIShellData(ctx context.Context, databasePath string, assets config.InitAssets) (render.ShellData, error) {
	status, err := ledger.GetDatabaseStatusWithAssets(ctx, databasePath, assets)
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

	databaseName := filepath.Base(databasePath)
	data := render.ShellData{
		Dashboard: render.DashboardData{
			HeaderLines: []string{
				fmt.Sprintf("Read-only snapshot from %s.", databaseName),
				"Use arrows, Tab, or 1-6 to move between boxed screens. Use e/i/a for guided entry creation.",
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
	}

	var panelErrors []string

	accounts, err := account.ListAccounts(ctx, databasePath)
	if err != nil {
		data.Dashboard.AccountsLines = unavailablePanelLines(err)
		data.Accounts = unavailableSectionData("Accounts unavailable.", err.Error())
		panelErrors = append(panelErrors, "accounts")
	} else {
		data.Dashboard.AccountsLines = summarizeAccounts(accounts)
		data.Accounts.SummaryLines = summarizeAccounts(accounts)
		data.Accounts.Items = buildAccountItems(accounts)
		data.EntryCatalog = buildEntryCatalog(accounts, tuiToday())
	}

	journalSummary, err := journal.GetSummary(ctx, databasePath)
	if err != nil {
		data.Dashboard.JournalLines = unavailablePanelLines(err)
		data.Journal = unavailableSectionData("Journal unavailable.", err.Error())
		panelErrors = append(panelErrors, "journal")
	} else {
		data.Dashboard.JournalLines = summarizeJournal(journalSummary)
		data.Journal.SummaryLines = summarizeJournal(journalSummary)
	}

	journalEntries, err := journal.ListBrowseEntries(ctx, databasePath)
	if err != nil {
		if len(data.Journal.SummaryLines) == 0 {
			data.Journal = unavailableSectionData("Journal unavailable.", err.Error())
		}
		data.Dashboard.JournalLines = unavailablePanelLines(err)
		data.Journal.Items = nil
		data.Journal.EmptyLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "journal")
	} else {
		data.Journal.Items = buildJournalItems(journalEntries)
	}

	trialBalance, err := report.GetTrialBalance(ctx, databasePath)
	trialBalanceAvailable := false
	if err != nil {
		data.Dashboard.LedgerLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "ledger")
	} else {
		data.Dashboard.LedgerLines = summarizeLedger(trialBalance)
		trialBalanceAvailable = true
	}

	promisedQuests, err := report.GetPromisedQuests(ctx, databasePath)
	var receivables []report.QuestReceivableRow
	questSummaryAvailable := false
	if err != nil {
		data.Dashboard.QuestLines = unavailablePanelLines(err)
		data.Quests = unavailableSectionData("Quest register unavailable.", err.Error())
		panelErrors = append(panelErrors, "quests")
	} else {
		var receivableErr error
		receivables, receivableErr = report.GetQuestReceivables(ctx, databasePath)
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
		questRows, questErr := loadTUIQuestRows(ctx, databasePath)
		if questErr != nil {
			if len(data.Quests.SummaryLines) == 0 {
				data.Quests = unavailableSectionData("Quest register unavailable.", questErr.Error())
			}
			data.Quests.Items = nil
			data.Quests.EmptyLines = unavailablePanelLines(questErr)
			panelErrors = append(panelErrors, "quests")
		} else {
			data.Quests.Items = buildQuestItems(questRows, tuiToday())
		}
	}

	lootRows, err := report.GetLootSummary(ctx, databasePath, "loot")
	lootSummaryAvailable := false
	if err != nil {
		data.Dashboard.LootLines = unavailablePanelLines(err)
		data.Loot = unavailableSectionData("Loot register unavailable.", err.Error())
		panelErrors = append(panelErrors, "loot")
	} else {
		data.Dashboard.LootLines = summarizeLoot(lootRows)
		data.Loot.SummaryLines = summarizeLoot(lootRows)
		lootSummaryAvailable = true
	}

	if lootSummaryAvailable {
		browseItems, browseErr := loot.ListBrowseItems(ctx, databasePath, "loot")
		if browseErr != nil {
			if len(data.Loot.SummaryLines) == 0 {
				data.Loot = unavailableSectionData("Loot register unavailable.", browseErr.Error())
			}
			data.Loot.Items = nil
			data.Loot.EmptyLines = unavailablePanelLines(browseErr)
			panelErrors = append(panelErrors, "loot")
		} else {
			data.Loot.Items = buildLootItems(browseItems, tuiToday())
		}
	}

	assetRows, assetSummaryErr := report.GetLootSummary(ctx, databasePath, "asset")
	assetSummaryAvailable := false
	if assetSummaryErr != nil {
		data.Dashboard.AssetLines = unavailablePanelLines(assetSummaryErr)
		data.Assets = unavailableSectionData("Asset register unavailable.", assetSummaryErr.Error())
		panelErrors = append(panelErrors, "assets")
	} else {
		data.Dashboard.AssetLines = summarizeAssets(assetRows)
		data.Assets.SummaryLines = summarizeAssets(assetRows)
		assetSummaryAvailable = true
	}

	if assetSummaryAvailable {
		assetBrowseItems, assetBrowseErr := loot.ListBrowseItems(ctx, databasePath, "asset")
		if assetBrowseErr != nil {
			if len(data.Assets.SummaryLines) == 0 {
				data.Assets = unavailableSectionData("Asset register unavailable.", assetBrowseErr.Error())
			}
			data.Assets.Items = nil
			data.Assets.EmptyLines = unavailablePanelLines(assetBrowseErr)
			panelErrors = append(panelErrors, "assets")
		} else {
			data.Assets.Items = buildAssetItems(assetBrowseItems, tuiToday())
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

func handleTUICommand(ctx context.Context, command render.Command, databasePath string, assets config.InitAssets) (render.CommandResult, error) {
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
		entries, err := journal.ListBrowseEntries(ctx, databasePath)
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
		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return render.CommandResult{}, render.InputError{Message: "Amount is required."}
		}
		amount, err := tools.ParseAmount(amountText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}

		result, err := journal.PostExpenseEntry(ctx, databasePath, &journal.ExpenseEntryInput{
			Date:               defaultDate(strings.TrimSpace(command.Fields["date"]), today),
			Description:        strings.TrimSpace(command.Fields["description"]),
			ExpenseAccountCode: strings.TrimSpace(command.Fields["account_code"]),
			FundingAccountCode: strings.TrimSpace(command.Fields["offset_account_code"]),
			Amount:             amount,
			Memo:               strings.TrimSpace(command.Fields["memo"]),
		})
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Recorded expense as journal entry #%d.", result.EntryNumber),
		}
		navigateTo = render.SectionJournal
		selectItemKey = result.ID
	case tuiCommandCreateIncome:
		amountText := strings.TrimSpace(command.Fields["amount"])
		if amountText == "" {
			return render.CommandResult{}, render.InputError{Message: "Amount is required."}
		}
		amount, err := tools.ParseAmount(amountText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid amount %q.", amountText)}
		}

		result, err := journal.PostIncomeEntry(ctx, databasePath, &journal.IncomeEntryInput{
			Date:               defaultDate(strings.TrimSpace(command.Fields["date"]), today),
			Description:        strings.TrimSpace(command.Fields["description"]),
			IncomeAccountCode:  strings.TrimSpace(command.Fields["account_code"]),
			DepositAccountCode: strings.TrimSpace(command.Fields["offset_account_code"]),
			Amount:             amount,
			Memo:               strings.TrimSpace(command.Fields["memo"]),
		})
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Recorded income as journal entry #%d.", result.EntryNumber),
		}
		navigateTo = render.SectionJournal
		selectItemKey = result.ID
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
		rewardText := strings.TrimSpace(command.Fields["reward"])
		if rewardText == "" {
			rewardText = "0"
		}
		reward, err := tools.ParseAmount(rewardText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid reward %q.", rewardText)}
		}
		advanceText := strings.TrimSpace(command.Fields["advance"])
		if advanceText == "" {
			advanceText = "0"
		}
		advance, err := tools.ParseAmount(advanceText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid advance %q.", advanceText)}
		}
		result, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
			Title:              strings.TrimSpace(command.Fields["title"]),
			Patron:             strings.TrimSpace(command.Fields["patron"]),
			Description:        strings.TrimSpace(command.Fields["description"]),
			PromisedBaseReward: reward,
			PartialAdvance:     advance,
			BonusConditions:    strings.TrimSpace(command.Fields["bonus"]),
			Notes:              strings.TrimSpace(command.Fields["notes"]),
			Status:             strings.TrimSpace(command.Fields["status"]),
			AcceptedOn:         strings.TrimSpace(command.Fields["accepted_on"]),
		})
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Created quest %q.", result.Title),
		}
		navigateTo = render.SectionQuests
		selectItemKey = result.ID
	case tuiCommandQuestUpdate:
		rewardText := strings.TrimSpace(command.Fields["reward"])
		if rewardText == "" {
			rewardText = "0"
		}
		reward, err := tools.ParseAmount(rewardText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid reward %q.", rewardText)}
		}
		advanceText := strings.TrimSpace(command.Fields["advance"])
		if advanceText == "" {
			advanceText = "0"
		}
		advance, err := tools.ParseAmount(advanceText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid advance %q.", advanceText)}
		}
		result, err := quest.UpdateQuest(ctx, databasePath, command.ItemKey, &quest.UpdateQuestInput{
			Title:              strings.TrimSpace(command.Fields["title"]),
			Patron:             strings.TrimSpace(command.Fields["patron"]),
			Description:        strings.TrimSpace(command.Fields["description"]),
			PromisedBaseReward: reward,
			PartialAdvance:     advance,
			BonusConditions:    strings.TrimSpace(command.Fields["bonus"]),
			Notes:              strings.TrimSpace(command.Fields["notes"]),
			AcceptedOn:         strings.TrimSpace(command.Fields["accepted_on"]),
		})
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Updated quest %q.", result.Title),
		}
		navigateTo = render.SectionQuests
		selectItemKey = result.ID
	case tuiCommandQuestCollectFull:
		quests, err := loadTUIQuestRows(ctx, databasePath)
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
			Text:  fmt.Sprintf("Collected %s for quest %q as entry #%d.", tools.FormatAmount(questRow.Outstanding), questRow.Record.Title, result.EntryNumber),
		}
	case tuiCommandQuestWriteOffFull:
		quests, err := loadTUIQuestRows(ctx, databasePath)
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
			Text:  fmt.Sprintf("Wrote off %s for quest %q as entry #%d.", tools.FormatAmount(questRow.Outstanding), questRow.Record.Title, result.EntryNumber),
		}
	case tuiCommandLootCreate:
		quantityText := strings.TrimSpace(command.Fields["quantity"])
		if quantityText == "" {
			quantityText = "1"
		}
		quantity, err := strconv.Atoi(quantityText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid quantity %q.", quantityText)}
		}
		if quantity <= 0 {
			return render.CommandResult{}, render.InputError{Message: "Quantity must be positive."}
		}
		result, err := loot.CreateLootItem(
			ctx,
			databasePath,
			strings.TrimSpace(command.Fields["name"]),
			strings.TrimSpace(command.Fields["source"]),
			quantity,
			strings.TrimSpace(command.Fields["holder"]),
			strings.TrimSpace(command.Fields["notes"]),
			"loot",
		)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Created loot item %q.", result.Name),
		}
		navigateTo = render.SectionLoot
		selectItemKey = result.ID
	case tuiCommandLootUpdate:
		quantityText := strings.TrimSpace(command.Fields["quantity"])
		if quantityText == "" {
			quantityText = "1"
		}
		quantity, err := strconv.Atoi(quantityText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid quantity %q.", quantityText)}
		}
		if quantity <= 0 {
			return render.CommandResult{}, render.InputError{Message: "Quantity must be positive."}
		}
		result, err := loot.UpdateLootItem(ctx, databasePath, command.ItemKey, &loot.UpdateLootItemInput{
			Name:     strings.TrimSpace(command.Fields["name"]),
			Source:   strings.TrimSpace(command.Fields["source"]),
			Quantity: quantity,
			Holder:   strings.TrimSpace(command.Fields["holder"]),
			Notes:    strings.TrimSpace(command.Fields["notes"]),
		})
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Updated loot item %q.", result.Name),
		}
		navigateTo = render.SectionLoot
		selectItemKey = result.ID
	case tuiCommandLootRecognize:
		items, err := loot.ListBrowseItems(ctx, databasePath, "loot")
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
		items, err := loot.ListBrowseItems(ctx, databasePath, "loot")
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

		amount, err := tools.ParseAmount(amountText)
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
			Text:  fmt.Sprintf("Transferred item to asset register."),
		}
		navigateTo = render.SectionAssets
		selectItemKey = command.ItemKey
	case tuiCommandAssetCreate:
		quantityText := strings.TrimSpace(command.Fields["quantity"])
		if quantityText == "" {
			quantityText = "1"
		}
		quantity, err := strconv.Atoi(quantityText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid quantity %q.", quantityText)}
		}
		if quantity <= 0 {
			return render.CommandResult{}, render.InputError{Message: "Quantity must be positive."}
		}
		result, err := loot.CreateLootItem(
			ctx,
			databasePath,
			strings.TrimSpace(command.Fields["name"]),
			strings.TrimSpace(command.Fields["source"]),
			quantity,
			strings.TrimSpace(command.Fields["holder"]),
			strings.TrimSpace(command.Fields["notes"]),
			"asset",
		)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Created asset %q.", result.Name),
		}
		navigateTo = render.SectionAssets
		selectItemKey = result.ID
	case tuiCommandAssetUpdate:
		quantityText := strings.TrimSpace(command.Fields["quantity"])
		if quantityText == "" {
			quantityText = "1"
		}
		quantity, err := strconv.Atoi(quantityText)
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: fmt.Sprintf("Invalid quantity %q.", quantityText)}
		}
		if quantity <= 0 {
			return render.CommandResult{}, render.InputError{Message: "Quantity must be positive."}
		}
		result, err := loot.UpdateLootItem(ctx, databasePath, command.ItemKey, &loot.UpdateLootItemInput{
			Name:     strings.TrimSpace(command.Fields["name"]),
			Source:   strings.TrimSpace(command.Fields["source"]),
			Quantity: quantity,
			Holder:   strings.TrimSpace(command.Fields["holder"]),
			Notes:    strings.TrimSpace(command.Fields["notes"]),
		})
		if err != nil {
			return render.CommandResult{}, render.InputError{Message: err.Error()}
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Updated asset %q.", result.Name),
		}
		navigateTo = render.SectionAssets
		selectItemKey = result.ID
	case tuiCommandAssetRecognize:
		items, err := loot.ListBrowseItems(ctx, databasePath, "asset")
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
			Text:  fmt.Sprintf("Transferred item to loot register."),
		}
		navigateTo = render.SectionLoot
		selectItemKey = command.ItemKey
	default:
		return render.CommandResult{}, fmt.Errorf("unsupported TUI command %q", command.ID)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
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

func buildTUIDashboardData(ctx context.Context, databasePath string, assets config.InitAssets) (render.DashboardData, error) {
	data, err := buildTUIShellData(ctx, databasePath, assets)
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

		amount, err := tools.ParseAmount(amountText)
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
		"Debits: " + tools.FormatAmount(trialBalance.TotalDebits),
		"Credits: " + tools.FormatAmount(trialBalance.TotalCredits),
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
		shareLine = "To share now: " + tools.FormatAmount(cashBalance)
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
		"Unsold loot: " + tools.FormatAmount(recognizedLoot),
	}
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
