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
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

func buildTUIDashboardData(ctx context.Context, databasePath string, assets config.InitAssets) (render.DashboardData, error) {
	status, err := ledger.GetDatabaseStatusWithAssets(ctx, databasePath, assets)
	if err != nil {
		return render.ErrorDashboardData("Database status unavailable.", err.Error()), nil
	}

	switch status.State {
	case ledger.DatabaseStateUninitialized:
		return unavailableDashboardData(&status, "Run `lootsheet init` before opening live dashboard summaries."), nil
	case ledger.DatabaseStateUpgradeable:
		return unavailableDashboardData(&status, "Run `lootsheet db migrate` before opening live dashboard summaries."), nil
	case ledger.DatabaseStateForeign, ledger.DatabaseStateDamaged:
		return unavailableDashboardData(&status, blankStatusDetail(status.Detail)), nil
	case ledger.DatabaseStateCurrent:
	}

	data := render.DashboardData{
		HeaderLines: []string{
			fmt.Sprintf("Read-only snapshot from %s.", filepath.Base(databasePath)),
			"Ctrl+L refreshes dashboard totals without leaving the TUI.",
		},
	}

	var panelErrors []string

	accounts, err := account.ListAccounts(ctx, databasePath)
	if err != nil {
		data.AccountsLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "accounts")
	} else {
		data.AccountsLines = summarizeAccounts(accounts)
	}

	journalSummary, err := journal.GetSummary(ctx, databasePath)
	if err != nil {
		data.JournalLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "journal")
	} else {
		data.JournalLines = summarizeJournal(journalSummary)
	}

	trialBalance, err := report.GetTrialBalance(ctx, databasePath)
	if err != nil {
		data.LedgerLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "ledger")
	} else {
		data.LedgerLines = summarizeLedger(trialBalance)
	}

	promisedQuests, err := report.GetPromisedQuests(ctx, databasePath)
	if err != nil {
		data.QuestLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "quests")
	} else {
		receivables, receivableErr := report.GetQuestReceivables(ctx, databasePath)
		if receivableErr != nil {
			data.QuestLines = unavailablePanelLines(receivableErr)
			panelErrors = append(panelErrors, "quests")
		} else {
			data.QuestLines = summarizeQuests(promisedQuests, receivables)
		}
	}

	lootRows, err := report.GetLootSummary(ctx, databasePath)
	if err != nil {
		data.LootLines = unavailablePanelLines(err)
		panelErrors = append(panelErrors, "loot")
	} else {
		data.LootLines = summarizeLoot(lootRows)
	}

	if len(panelErrors) > 0 {
		data.HeaderLines[1] = "Some panels are unavailable: " + strings.Join(panelErrors, ", ") + "."
	}

	return data, nil
}

func unavailableDashboardData(status *ledger.DatabaseStatus, detail string) render.DashboardData {
	if status == nil {
		return render.ErrorDashboardData("Database status unavailable.", detail)
	}

	stateLine := fmt.Sprintf("Database state: %s.", status.State)
	if detail == "" {
		detail = "Dashboard data is not available for this database state."
	}

	return render.DashboardData{
		HeaderLines:   []string{stateLine, detail},
		AccountsLines: []string{"No account data loaded.", stateLine},
		JournalLines:  []string{"No journal data loaded.", stateLine},
		LedgerLines:   []string{"No ledger totals loaded.", stateLine},
		QuestLines:    []string{"No quest register data loaded.", stateLine},
		LootLines:     []string{"No loot register data loaded.", stateLine},
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

func blankStatusDetail(detail string) string {
	if strings.TrimSpace(detail) == "" {
		return "Dashboard data is not available for this database state."
	}
	return detail
}
