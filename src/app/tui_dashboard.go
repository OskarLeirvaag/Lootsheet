package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/account"
	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

const (
	tuiCommandAccountActivate   = "account.activate"
	tuiCommandAccountDeactivate = "account.deactivate"
	tuiCommandJournalReverse    = "journal.reverse"
	tuiCommandQuestCollectFull  = "quest.collect_full"
	tuiCommandQuestWriteOffFull = "quest.writeoff_full"
)

var tuiNow = time.Now

type tuiQuestRow struct {
	Record      quest.QuestRecord
	TotalPaid   int64
	Outstanding int64
	Collectible bool
	OffLedger   bool
}

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
				"Use arrows, Tab, or 1-5 to move between boxed screens.",
			},
		},
		Accounts: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Chart of accounts from %s.", databaseName),
				"Select an account to inspect it. `t` toggles active/inactive with confirmation.",
			},
			EmptyLines: []string{
				"No accounts found.",
				"The chart of accounts is empty in this database.",
			},
		},
		Journal: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Posted journal history from %s.", databaseName),
				"Select an entry to inspect it. `r` reverses the selected posted entry on its original date.",
			},
			EmptyLines: []string{
				"No journal entries yet.",
				"Posting stays in the CLI for now.",
			},
		},
		Quests: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Quest register from %s.", databaseName),
				"Select a quest to inspect it. `c` collects the full balance and `w` writes off the full balance using today's date.",
			},
			EmptyLines: []string{
				"No quests tracked yet.",
				"Quest actions appear when a quest has an outstanding collectible balance.",
			},
		},
		Loot: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Unrealized loot register from %s.", databaseName),
				"Select a loot item to inspect it. Appraisals stay off-ledger until recognized.",
			},
			EmptyLines: []string{
				"No loot tracked yet.",
				"Loot workflows stay read-only in this slice.",
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
	if err != nil {
		data.Dashboard.LedgerLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "ledger")
	} else {
		data.Dashboard.LedgerLines = summarizeLedger(trialBalance)
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

	lootRows, err := report.GetLootSummary(ctx, databasePath)
	if err != nil {
		data.Dashboard.LootLines = unavailablePanelLines(err)
		data.Loot = unavailableSectionData("Loot register unavailable.", err.Error())
		panelErrors = append(panelErrors, "loot")
	} else {
		data.Dashboard.LootLines = summarizeLoot(lootRows)
		data.Loot.SummaryLines = summarizeLoot(lootRows)
		data.Loot.Items = buildLootItems(lootRows)
	}

	if len(panelErrors) > 0 {
		data.Dashboard.HeaderLines[1] = "Some panels are unavailable: " + strings.Join(uniqueStrings(panelErrors), ", ") + "."
	}

	return data, nil
}

func handleTUICommand(ctx context.Context, command render.Command, databasePath string, assets config.InitAssets) (render.ShellData, render.StatusMessage, error) {
	var message render.StatusMessage
	today := tuiToday()

	switch command.ID {
	case tuiCommandAccountActivate:
		if err := account.ActivateAccount(ctx, databasePath, command.ItemKey); err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Account %s activated.", command.ItemKey),
		}
	case tuiCommandAccountDeactivate:
		if err := account.DeactivateAccount(ctx, databasePath, command.ItemKey); err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}
		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Account %s deactivated.", command.ItemKey),
		}
	case tuiCommandJournalReverse:
		entries, err := journal.ListBrowseEntries(ctx, databasePath)
		if err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}

		entry, ok := findBrowseEntry(entries, command.ItemKey)
		if !ok {
			return render.ShellData{}, render.StatusMessage{}, fmt.Errorf("journal entry %q does not exist", command.ItemKey)
		}

		result, err := journal.ReverseJournalEntry(ctx, databasePath, command.ItemKey, entry.EntryDate, "")
		if err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Entry #%d reversed as entry #%d.", entry.EntryNumber, result.EntryNumber),
		}
	case tuiCommandQuestCollectFull:
		quests, err := loadTUIQuestRows(ctx, databasePath)
		if err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}

		questRow, ok := findTUIQuestRow(quests, command.ItemKey)
		if !ok {
			return render.ShellData{}, render.StatusMessage{}, fmt.Errorf("quest %q does not exist", command.ItemKey)
		}
		if !questRow.Collectible || questRow.Outstanding <= 0 {
			return render.ShellData{}, render.StatusMessage{}, fmt.Errorf("quest %q cannot be collected right now", command.ItemKey)
		}

		result, err := quest.CollectQuestPayment(ctx, databasePath, quest.CollectQuestPaymentInput{
			QuestID:     command.ItemKey,
			Amount:      questRow.Outstanding,
			Date:        today,
			Description: "",
		})
		if err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Collected %s for quest %q as entry #%d.", tools.FormatAmount(questRow.Outstanding), questRow.Record.Title, result.EntryNumber),
		}
	case tuiCommandQuestWriteOffFull:
		quests, err := loadTUIQuestRows(ctx, databasePath)
		if err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}

		questRow, ok := findTUIQuestRow(quests, command.ItemKey)
		if !ok {
			return render.ShellData{}, render.StatusMessage{}, fmt.Errorf("quest %q does not exist", command.ItemKey)
		}
		if !questRow.Collectible || questRow.Outstanding <= 0 {
			return render.ShellData{}, render.StatusMessage{}, fmt.Errorf("quest %q cannot be written off right now", command.ItemKey)
		}

		result, err := quest.WriteOffQuest(ctx, databasePath, quest.WriteOffQuestInput{
			QuestID:     command.ItemKey,
			Date:        today,
			Description: "",
		})
		if err != nil {
			return render.ShellData{}, render.StatusMessage{}, err
		}

		message = render.StatusMessage{
			Level: render.StatusSuccess,
			Text:  fmt.Sprintf("Wrote off %s for quest %q as entry #%d.", tools.FormatAmount(questRow.Outstanding), questRow.Record.Title, result.EntryNumber),
		}
	default:
		return render.ShellData{}, render.StatusMessage{}, fmt.Errorf("unsupported TUI command %q", command.ID)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		return render.ShellData{}, render.StatusMessage{}, err
	}

	return data, message, nil
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
			HeaderLines:   []string{stateLine, detail},
			AccountsLines: []string{"No account data loaded.", stateLine},
			JournalLines:  []string{"No journal data loaded.", stateLine},
			LedgerLines:   []string{"No ledger totals loaded.", stateLine},
			QuestLines:    []string{"No quest register data loaded.", stateLine},
			LootLines:     []string{"No loot register data loaded.", stateLine},
		},
		Accounts: unavailableSectionData(stateLine, detail),
		Journal:  unavailableSectionData(stateLine, detail),
		Quests:   unavailableSectionData(stateLine, detail),
		Loot:     unavailableSectionData(stateLine, detail),
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

func summarizeAccounts(accounts []ledger.AccountRecord) []string {
	counts := map[ledger.AccountType]int{
		ledger.AccountTypeAsset:     0,
		ledger.AccountTypeLiability: 0,
		ledger.AccountTypeEquity:    0,
		ledger.AccountTypeIncome:    0,
		ledger.AccountTypeExpense:   0,
	}

	active := 0
	for _, record := range accounts {
		counts[record.Type]++
		if record.Active {
			active++
		}
	}

	return []string{
		fmt.Sprintf("Accounts: %d total", len(accounts)),
		fmt.Sprintf("Active: %d  Inactive: %d", active, len(accounts)-active),
		fmt.Sprintf("A/L/E: %d / %d / %d", counts[ledger.AccountTypeAsset], counts[ledger.AccountTypeLiability], counts[ledger.AccountTypeEquity]),
		fmt.Sprintf("I/X: %d / %d", counts[ledger.AccountTypeIncome], counts[ledger.AccountTypeExpense]),
	}
}

func summarizeJournal(summary journal.Summary) []string {
	if summary.TotalEntries == 0 {
		return []string{
			"Entries: 0 total",
			"Posted: 0",
			"Reversal entries: 0",
			"No journal activity yet.",
		}
	}

	return []string{
		fmt.Sprintf("Entries: %d total", summary.TotalEntries),
		fmt.Sprintf("Posted: %d", summary.PostedEntries),
		fmt.Sprintf("Reversal entries: %d", summary.ReversalEntries),
		fmt.Sprintf("Latest: #%d %s", summary.LatestEntryNumber, summary.LatestEntryDate),
	}
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

func summarizeQuests(promised []report.PromisedQuestRow, receivables []report.QuestReceivableRow) []string {
	var promisedValue int64
	for _, row := range promised {
		promisedValue += row.PromisedReward
	}

	var receivableValue int64
	for _, row := range receivables {
		receivableValue += row.Outstanding
	}

	return []string{
		fmt.Sprintf("Promised quests: %d", len(promised)),
		"Promised value: " + tools.FormatAmount(promisedValue),
		fmt.Sprintf("Receivables: %d", len(receivables)),
		"Outstanding: " + tools.FormatAmount(receivableValue),
	}
}

func summarizeLoot(rows []report.LootSummaryRow) []string {
	totalQuantity := 0
	recognized := 0
	var appraisedValue int64

	for _, row := range rows {
		totalQuantity += row.Quantity
		appraisedValue += row.LatestAppraisalValue
		if row.Status == ledger.LootStatusRecognized {
			recognized++
		}
	}

	return []string{
		fmt.Sprintf("Tracked items: %d", len(rows)),
		fmt.Sprintf("Recognized: %d", recognized),
		fmt.Sprintf("Total quantity: %d", totalQuantity),
		"Appraised value: " + tools.FormatAmount(appraisedValue),
	}
}

func buildAccountItems(accounts []ledger.AccountRecord) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(accounts))
	for _, record := range accounts {
		status := "inactive"
		action := &render.ItemActionData{
			Trigger:      render.ActionToggle,
			ID:           tuiCommandAccountActivate,
			Label:        "t activate",
			ConfirmTitle: fmt.Sprintf("Activate account %s?", record.Code),
			ConfirmLines: []string{
				record.Name,
				"This account will become available for new journal entries again.",
			},
		}
		if record.Active {
			status = "active"
			action = &render.ItemActionData{
				Trigger:      render.ActionToggle,
				ID:           tuiCommandAccountDeactivate,
				Label:        "t deactivate",
				ConfirmTitle: fmt.Sprintf("Deactivate account %s?", record.Code),
				ConfirmLines: []string{
					record.Name,
					"Inactive accounts stay in history but cannot be used in new journal entries.",
				},
			}
		}

		items = append(items, render.ListItemData{
			Key:         record.Code,
			Row:         fmt.Sprintf("%-4s %-9s %-8s %s", record.Code, string(record.Type), status, record.Name),
			DetailTitle: "Account " + record.Code,
			DetailLines: []string{
				"Name: " + record.Name,
				"Type: " + string(record.Type),
				"Status: " + status,
				"Code: " + record.Code + " (immutable)",
				"Used accounts may be marked inactive. Accounts with postings cannot be deleted.",
			},
			Actions: []render.ItemActionData{*action},
		})
	}

	return items
}

func buildJournalItems(entries []journal.BrowseEntryRecord) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(entries))
	for index := range entries {
		entry := &entries[index]
		rowStatus := string(entry.Status)
		detailLines := []string{
			"Date: " + entry.EntryDate,
			"Status: " + string(entry.Status),
			"Description: " + entry.Description,
		}
		if entry.ReversesEntryID != "" {
			rowStatus = "reversal"
			detailLines = append(detailLines, fmt.Sprintf("Reverses: entry #%d", entry.ReversesEntryNumber))
		}
		if entry.ReversedByEntryID != "" {
			detailLines = append(detailLines, fmt.Sprintf("Reversed by: entry #%d", entry.ReversedByEntryNumber))
		}
		if entry.Status == ledger.JournalEntryStatusReversed {
			detailLines = append(detailLines, "This entry has been reversed and remains in the audit trail.")
		}
		detailLines = append(detailLines, "", "Lines:")
		if len(entry.Lines) == 0 {
			detailLines = append(detailLines, "No journal lines loaded.")
		}
		for _, line := range entry.Lines {
			detailLines = append(detailLines, formatJournalDetailLine(line))
		}

		var actions []render.ItemActionData
		if entry.Status == ledger.JournalEntryStatusPosted {
			actions = []render.ItemActionData{{
				Trigger:      render.ActionReverse,
				ID:           tuiCommandJournalReverse,
				Label:        "r reverse",
				ConfirmTitle: fmt.Sprintf("Reverse entry #%d?", entry.EntryNumber),
				ConfirmLines: []string{
					entry.Description,
					"Original date: " + entry.EntryDate,
					"Reversal date: " + entry.EntryDate,
					"A new posted reversing entry will be created.",
					fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Reversal of entry #%d", entry.EntryNumber)),
				},
			}}
		}

		items = append(items, render.ListItemData{
			Key:         entry.ID,
			Row:         fmt.Sprintf("#%-4d %-10s %-8s %s", entry.EntryNumber, entry.EntryDate, rowStatus, entry.Description),
			DetailTitle: fmt.Sprintf("Entry #%d", entry.EntryNumber),
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

func buildQuestItems(quests []tuiQuestRow, today string) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(quests))
	for index := range quests {
		row := &quests[index]
		record := &row.Record
		patron := record.Patron
		if strings.TrimSpace(patron) == "" {
			patron = "No patron"
		}

		detailLines := []string{
			"Patron: " + patron,
			"Status: " + string(record.Status),
			"Promised reward: " + tools.FormatAmount(record.PromisedBaseReward),
			"Outstanding: " + tools.FormatAmount(row.Outstanding),
			"Collected so far: " + tools.FormatAmount(row.TotalPaid),
			"Accounting state: " + questAccountingState(row),
		}
		if record.PartialAdvance > 0 {
			detailLines = append(detailLines, "Partial advance: "+tools.FormatAmount(record.PartialAdvance))
		}
		if record.AcceptedOn != "" {
			detailLines = append(detailLines, "Accepted on: "+record.AcceptedOn)
		}
		if record.CompletedOn != "" {
			detailLines = append(detailLines, "Completed on: "+record.CompletedOn)
		}
		if record.ClosedOn != "" {
			detailLines = append(detailLines, "Closed on: "+record.ClosedOn)
		}
		if strings.TrimSpace(record.BonusConditions) != "" {
			detailLines = append(detailLines, "Bonus conditions: "+record.BonusConditions)
		}
		if strings.TrimSpace(record.Notes) != "" {
			detailLines = append(detailLines, "Notes: "+record.Notes)
		}

		var actions []render.ItemActionData
		if row.Collectible {
			actions = []render.ItemActionData{
				{
					Trigger:      render.ActionCollect,
					ID:           tuiCommandQuestCollectFull,
					Label:        "c collect",
					ConfirmTitle: fmt.Sprintf("Collect full payment for %q?", record.Title),
					ConfirmLines: []string{
						"Outstanding: " + tools.FormatAmount(row.Outstanding),
						"Collection date: " + today,
						"This collects the full remaining receivable.",
						fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Quest payment: %s", record.Title)),
					},
				},
				{
					Trigger:      render.ActionWriteOff,
					ID:           tuiCommandQuestWriteOffFull,
					Label:        "w write off",
					ConfirmTitle: fmt.Sprintf("Write off %q?", record.Title),
					ConfirmLines: []string{
						"Outstanding: " + tools.FormatAmount(row.Outstanding),
						"Write-off date: " + today,
						"This records the remaining balance as a failed patron loss.",
						fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Quest write-off: %s", record.Title)),
					},
				},
			}
		}

		items = append(items, render.ListItemData{
			Key:         record.ID,
			Row:         fmt.Sprintf("%-12s %-14s %-12s %s (%s)", tools.FormatAmount(record.PromisedBaseReward), string(record.Status), questOutstandingLabel(row.Outstanding), record.Title, patron),
			DetailTitle: record.Title,
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

func buildLootItems(rows []report.LootSummaryRow) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(rows))
	for _, row := range rows {
		appraised := "No appraisal"
		if row.LatestAppraisalValue > 0 {
			appraised = tools.FormatAmount(row.LatestAppraisalValue)
		}

		name := row.Name
		if strings.TrimSpace(row.Source) != "" {
			name = name + " (" + row.Source + ")"
		}

		detailLines := []string{
			"Status: " + string(row.Status),
			fmt.Sprintf("Quantity: %d", row.Quantity),
			"Latest appraisal: " + appraised,
		}
		if row.AppraisedAt != "" {
			detailLines = append(detailLines, "Appraised on: "+row.AppraisedAt)
		}
		if strings.TrimSpace(row.Source) != "" {
			detailLines = append(detailLines, "Source: "+row.Source)
		}

		items = append(items, render.ListItemData{
			Key:         row.ItemID,
			Row:         fmt.Sprintf("%-12s qty:%-3d %-11s %s", appraisedValueLabel(row.LatestAppraisalValue), row.Quantity, string(row.Status), name),
			DetailTitle: row.Name,
			DetailLines: detailLines,
		})
	}

	return items
}

func questAccountingState(row *tuiQuestRow) string {
	if row == nil {
		return ""
	}

	switch row.Record.Status {
	case ledger.QuestStatusOffered, ledger.QuestStatusAccepted:
		return "off-ledger promise"
	case ledger.QuestStatusCompleted, ledger.QuestStatusCollectible:
		return "collectible but unpaid"
	case ledger.QuestStatusPartiallyPaid:
		return "partially collected receivable"
	case ledger.QuestStatusPaid:
		return "fully collected"
	case ledger.QuestStatusDefaulted:
		return "written off"
	case ledger.QuestStatusVoided:
		return "closed without collection"
	default:
		return string(row.Record.Status)
	}
}

func questOutstandingLabel(value int64) string {
	if value <= 0 {
		return "-"
	}

	return tools.FormatAmount(value) + " due"
}

func appraisedValueLabel(value int64) string {
	if value <= 0 {
		return "-"
	}
	return tools.FormatAmount(value)
}

func loadTUIQuestRows(ctx context.Context, databasePath string) ([]tuiQuestRow, error) {
	quests, err := quest.ListQuests(ctx, databasePath)
	if err != nil {
		return nil, err
	}

	receivables, err := report.GetQuestReceivables(ctx, databasePath)
	if err != nil {
		return nil, err
	}

	receivablesByQuestID := make(map[string]report.QuestReceivableRow, len(receivables))
	for _, row := range receivables {
		receivablesByQuestID[row.QuestID] = row
	}

	rows := make([]tuiQuestRow, 0, len(quests))
	for index := range quests {
		record := quests[index]
		receivable, ok := receivablesByQuestID[record.ID]
		row := tuiQuestRow{
			Record:    record,
			OffLedger: record.Status == ledger.QuestStatusOffered || record.Status == ledger.QuestStatusAccepted,
		}
		if ok {
			row.TotalPaid = receivable.TotalPaid
			row.Outstanding = receivable.Outstanding
		} else if record.Status == ledger.QuestStatusPaid {
			row.TotalPaid = record.PromisedBaseReward
		}

		row.Collectible = row.Outstanding > 0 && questCollectibleStatus(record.Status)
		rows = append(rows, row)
	}

	return rows, nil
}

func questCollectibleStatus(status ledger.QuestStatus) bool {
	switch status {
	case ledger.QuestStatusCompleted, ledger.QuestStatusCollectible, ledger.QuestStatusPartiallyPaid:
		return true
	default:
		return false
	}
}

func tuiToday() string {
	return tuiNow().Format("2006-01-02")
}

func formatJournalDetailLine(line journal.BrowseEntryLine) string {
	side := "CR"
	amount := line.CreditAmount
	if line.DebitAmount > 0 {
		side = "DR"
		amount = line.DebitAmount
	}

	text := fmt.Sprintf("%s %s %s %s", line.AccountCode, line.AccountName, side, tools.FormatAmount(amount))
	if strings.TrimSpace(line.Memo) == "" {
		return text
	}

	return text + " (" + line.Memo + ")"
}

func findBrowseEntry(entries []journal.BrowseEntryRecord, entryID string) (journal.BrowseEntryRecord, bool) {
	for index := range entries {
		entry := entries[index]
		if entry.ID == entryID {
			return entry, true
		}
	}

	return journal.BrowseEntryRecord{}, false
}

func findTUIQuestRow(rows []tuiQuestRow, questID string) (tuiQuestRow, bool) {
	for index := range rows {
		if rows[index].Record.ID == questID {
			return rows[index], true
		}
	}

	return tuiQuestRow{}, false
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
