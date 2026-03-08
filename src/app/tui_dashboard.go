package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/account"
	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

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
				"Read-only list. Account codes stay immutable; edits remain in the CLI for now.",
			},
		},
		Journal: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Posted journal history from %s.", databaseName),
				"Read-only browser. Corrections still happen by reversal or adjustment.",
			},
		},
		Quests: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Quest register from %s.", databaseName),
				"Promised rewards stay off-ledger until earned.",
			},
		},
		Loot: render.ListScreenData{
			HeaderLines: []string{
				fmt.Sprintf("Unrealized loot register from %s.", databaseName),
				"Appraisals stay off-ledger until explicitly recognized.",
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
		data.Accounts.RowLines = formatAccountRows(accounts)
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

	journalEntries, err := journal.ListEntries(ctx, databasePath)
	if err != nil {
		if len(data.Journal.SummaryLines) == 0 {
			data.Journal = unavailableSectionData("Journal unavailable.", err.Error())
		}
		data.Dashboard.JournalLines = unavailablePanelLines(err)
		data.Journal.RowLines = nil
		data.Journal.EmptyLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "journal")
	} else {
		data.Journal.RowLines = formatJournalRows(journalEntries)
	}

	trialBalance, err := report.GetTrialBalance(ctx, databasePath)
	if err != nil {
		data.Dashboard.LedgerLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "ledger")
	} else {
		data.Dashboard.LedgerLines = summarizeLedger(trialBalance)
	}

	promisedQuests, err := report.GetPromisedQuests(ctx, databasePath)
	if err != nil {
		data.Dashboard.QuestLines = unavailablePanelLines(err)
		data.Quests = unavailableSectionData("Quest register unavailable.", err.Error())
		panelErrors = append(panelErrors, "quests")
	} else {
		receivables, receivableErr := report.GetQuestReceivables(ctx, databasePath)
		if receivableErr != nil {
			data.Dashboard.QuestLines = unavailablePanelLines(receivableErr)
			data.Quests = unavailableSectionData("Quest register unavailable.", receivableErr.Error())
			panelErrors = append(panelErrors, "quests")
		} else {
			data.Dashboard.QuestLines = summarizeQuests(promisedQuests, receivables)
			data.Quests.SummaryLines = summarizeQuests(promisedQuests, receivables)
		}
	}

	quests, err := quest.ListQuests(ctx, databasePath)
	if err != nil {
		if len(data.Quests.SummaryLines) == 0 {
			data.Quests = unavailableSectionData("Quest register unavailable.", err.Error())
		}
		data.Quests.RowLines = nil
		data.Quests.EmptyLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "quests")
	} else {
		data.Quests.RowLines = formatQuestRows(quests)
	}

	lootRows, err := report.GetLootSummary(ctx, databasePath)
	if err != nil {
		data.Dashboard.LootLines = unavailablePanelLines(err)
		data.Loot = unavailableSectionData("Loot register unavailable.", err.Error())
		panelErrors = append(panelErrors, "loot")
	} else {
		data.Dashboard.LootLines = summarizeLoot(lootRows)
		data.Loot.SummaryLines = summarizeLoot(lootRows)
		data.Loot.RowLines = formatLootRows(lootRows)
	}

	if len(panelErrors) > 0 {
		data.Dashboard.HeaderLines[1] = "Some panels are unavailable: " + strings.Join(uniqueStrings(panelErrors), ", ") + "."
	}

	return data, nil
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

func formatAccountRows(accounts []ledger.AccountRecord) []string {
	rows := make([]string, 0, len(accounts))
	for _, record := range accounts {
		active := "inactive"
		if record.Active {
			active = "active"
		}

		rows = append(rows, fmt.Sprintf(
			"%-4s %-9s %-8s %s",
			record.Code,
			string(record.Type),
			active,
			record.Name,
		))
	}

	return rows
}

func formatJournalRows(entries []journal.EntryRecord) []string {
	rows := make([]string, 0, len(entries))
	for _, entry := range entries {
		status := string(entry.Status)
		if entry.ReversesEntryID != "" {
			status = "reversal"
		}

		rows = append(rows, fmt.Sprintf(
			"#%-4d %-10s %-8s %s",
			entry.EntryNumber,
			entry.EntryDate,
			status,
			entry.Description,
		))
	}

	return rows
}

func formatQuestRows(quests []quest.QuestRecord) []string {
	rows := make([]string, 0, len(quests))
	for index := range quests {
		record := &quests[index]
		patron := record.Patron
		if strings.TrimSpace(patron) == "" {
			patron = "No patron"
		}

		rows = append(rows, fmt.Sprintf(
			"%-12s %-14s %s (%s)",
			tools.FormatAmount(record.PromisedBaseReward),
			string(record.Status),
			record.Title,
			patron,
		))
	}

	return rows
}

func formatLootRows(rows []report.LootSummaryRow) []string {
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		appraised := "-"
		if row.LatestAppraisalValue > 0 {
			appraised = tools.FormatAmount(row.LatestAppraisalValue)
		}

		name := row.Name
		if strings.TrimSpace(row.Source) != "" {
			name = name + " (" + row.Source + ")"
		}

		lines = append(lines, fmt.Sprintf(
			"%-12s qty:%-3d %-11s %s",
			appraised,
			row.Quantity,
			string(row.Status),
			name,
		))
	}

	return lines
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
