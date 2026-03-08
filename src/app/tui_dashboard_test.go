package app

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func TestBuildTUIDashboardDataForUninitializedDatabase(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := t.TempDir() + "/lootsheet.db"
	data, err := buildTUIDashboardData(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("build dashboard data: %v", err)
	}

	if !strings.Contains(data.HeaderLines[0], "uninitialized") {
		t.Fatalf("header = %q, want uninitialized state", data.HeaderLines[0])
	}
	if !strings.Contains(data.HeaderLines[1], "lootsheet init") {
		t.Fatalf("header detail = %q, want init guidance", data.HeaderLines[1])
	}
}

func TestBuildTUIDashboardDataUsesReadOnlySummaries(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	posted, err := journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25, Memo: "Quiver refill"},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	if _, err := journal.ReverseJournalEntry(ctx, databasePath, posted.ID, "2026-03-09", "Correct duplicate"); err != nil {
		t.Fatalf("reverse journal entry: %v", err)
	}

	if _, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Goblin Bounty",
		Patron:             "Mayor Rowan",
		PromisedBaseReward: 2500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-08",
	}); err != nil {
		t.Fatalf("create quest: %v", err)
	}

	completedQuest, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Bridge Toll Cleanup",
		Patron:             "Road Warden",
		PromisedBaseReward: 1000,
		Status:             "accepted",
		AcceptedOn:         "2026-03-07",
	})
	if err != nil {
		t.Fatalf("create collectible quest: %v", err)
	}

	if err := quest.CompleteQuest(ctx, databasePath, completedQuest.ID, "2026-03-09"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	lootItem, err := loot.CreateLootItem(ctx, databasePath, "Silver Chalice", "Goblin den", 2, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}

	appraisal, err := loot.AppraiseLootItem(ctx, databasePath, lootItem.ID, 800, "Guild factor", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise loot item: %v", err)
	}

	if _, err := loot.RecognizeLootAppraisal(ctx, databasePath, appraisal.ID, "2026-03-09", "Recognize chalice"); err != nil {
		t.Fatalf("recognize loot appraisal: %v", err)
	}

	data, err := buildTUIDashboardData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build dashboard data: %v", err)
	}

	if !strings.Contains(strings.Join(data.AccountsLines, "\n"), "Accounts: 16 total") {
		t.Fatalf("accounts lines = %q", data.AccountsLines)
	}
	if !strings.Contains(strings.Join(data.JournalLines, "\n"), "Entries: 3 total") {
		t.Fatalf("journal lines = %q", data.JournalLines)
	}
	if !strings.Contains(strings.Join(data.LedgerLines, "\n"), "Status: BALANCED") {
		t.Fatalf("ledger lines = %q", data.LedgerLines)
	}
	if !strings.Contains(strings.Join(data.QuestLines, "\n"), "Promised quests: 1") {
		t.Fatalf("quest lines = %q", data.QuestLines)
	}
	if !strings.Contains(strings.Join(data.QuestLines, "\n"), "Receivables: 1") {
		t.Fatalf("quest lines = %q", data.QuestLines)
	}
	if !strings.Contains(strings.Join(data.LootLines, "\n"), "Tracked items: 1") {
		t.Fatalf("loot lines = %q", data.LootLines)
	}
}

func TestBuildTUIShellDataUsesReadOnlySectionRows(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	if _, err := journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 25},
		},
	}); err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	if _, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Goblin Bounty",
		Patron:             "Mayor Rowan",
		PromisedBaseReward: 2500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-08",
	}); err != nil {
		t.Fatalf("create quest: %v", err)
	}

	lootItem, err := loot.CreateLootItem(ctx, databasePath, "Silver Chalice", "Goblin den", 2, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}

	if _, err := loot.AppraiseLootItem(ctx, databasePath, lootItem.ID, 800, "Guild factor", "2026-03-08", ""); err != nil {
		t.Fatalf("appraise loot item: %v", err)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	if !strings.Contains(joinItemRows(data.Accounts.Items), "1000 asset") {
		t.Fatalf("account items = %#v", data.Accounts.Items)
	}
	if !strings.Contains(joinItemRows(data.Journal.Items), "Restock arrows") {
		t.Fatalf("journal items = %#v", data.Journal.Items)
	}
	if !strings.Contains(joinItemRows(data.Quests.Items), "Goblin Bounty") {
		t.Fatalf("quest items = %#v", data.Quests.Items)
	}
	if !strings.Contains(joinItemRows(data.Loot.Items), "Silver Chalice") {
		t.Fatalf("loot items = %#v", data.Loot.Items)
	}
}

func TestBuildTUIShellDataAddsAccountToggleAction(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	data, err := buildTUIShellData(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	if len(data.Accounts.Items) == 0 {
		t.Fatal("expected account items")
	}

	item := data.Accounts.Items[0]
	if len(item.Actions) != 1 {
		t.Fatalf("account item actions = %#v, want single toggle action", item.Actions)
	}
	action := item.Actions[0]
	if action.ID != tuiCommandAccountDeactivate {
		t.Fatalf("action id = %q, want %q", action.ID, tuiCommandAccountDeactivate)
	}
	if action.Trigger != render.ActionToggle {
		t.Fatalf("action trigger = %q, want %q", action.Trigger, render.ActionToggle)
	}
	if !strings.Contains(action.Label, "deactivate") {
		t.Fatalf("action label = %q, want deactivate", action.Label)
	}
}

func TestBuildTUIShellDataAddsJournalReverseActionAndLineDetail(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	posted, err := journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25, Memo: "Quiver refill"},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	found := false
	for _, item := range data.Journal.Items {
		if item.Key != posted.ID {
			continue
		}
		found = true
		if len(item.Actions) != 1 {
			t.Fatalf("journal item actions = %#v, want single reverse action", item.Actions)
		}
		action := item.Actions[0]
		if action.Trigger != render.ActionReverse {
			t.Fatalf("journal action trigger = %q, want %q", action.Trigger, render.ActionReverse)
		}
		if action.ID != tuiCommandJournalReverse {
			t.Fatalf("journal action id = %q, want %q", action.ID, tuiCommandJournalReverse)
		}
		detail := strings.Join(item.DetailLines, "\n")
		for _, token := range []string{"Lines:", "5100 Arrows & Ammunition DR 2 SP 5 CP (Quiver refill)", "1000 Party Cash CR 2 SP 5 CP"} {
			if !strings.Contains(detail, token) {
				t.Fatalf("journal detail missing %q:\n%s", token, detail)
			}
		}
		break
	}
	if !found {
		t.Fatalf("expected posted journal item %q", posted.ID)
	}
}

func TestBuildTUIShellDataOmitsJournalReverseActionForReversedOriginal(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	posted, err := journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	reversal, err := journal.ReverseJournalEntry(ctx, databasePath, posted.ID, "2026-03-08", "")
	if err != nil {
		t.Fatalf("reverse journal entry: %v", err)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	foundOriginal := false
	foundReversal := false
	for _, item := range data.Journal.Items {
		switch item.Key {
		case posted.ID:
			foundOriginal = true
			if len(item.Actions) != 0 {
				t.Fatalf("reversed original should not expose reverse action: %#v", item)
			}
			detail := strings.Join(item.DetailLines, "\n")
			if !strings.Contains(detail, "Reversed by: entry #"+strconv.Itoa(reversal.EntryNumber)) {
				t.Fatalf("reversed original detail missing reversal linkage:\n%s", detail)
			}
		case reversal.ID:
			foundReversal = true
			if len(item.Actions) != 1 || item.Actions[0].ID != tuiCommandJournalReverse {
				t.Fatalf("reversal entry should still expose reverse action: %#v", item)
			}
			detail := strings.Join(item.DetailLines, "\n")
			if !strings.Contains(detail, "Reverses: entry #"+strconv.Itoa(posted.EntryNumber)) {
				t.Fatalf("reversal detail missing original linkage:\n%s", detail)
			}
		}
	}

	if !foundOriginal || !foundReversal {
		t.Fatalf("expected original and reversal entries in journal items")
	}
}

func TestHandleTUICommandTogglesAccountState(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	result, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandAccountDeactivate,
		Section: render.SectionAccounts,
		ItemKey: "1000",
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("deactivate account through tui command: %v", err)
	}
	data := result.Data
	status := result.Status
	if status.Level != render.StatusSuccess {
		t.Fatalf("status level = %q, want %q", status.Level, render.StatusSuccess)
	}
	if !strings.Contains(status.Text, "deactivated") {
		t.Fatalf("status text = %q, want deactivated", status.Text)
	}

	found := false
	for _, item := range data.Accounts.Items {
		if item.Key != "1000" {
			continue
		}
		found = true
		if len(item.Actions) != 1 || item.Actions[0].ID != tuiCommandAccountActivate {
			t.Fatalf("account item after deactivate = %#v, want activate action", item)
		}
		break
	}
	if !found {
		t.Fatal("expected account 1000 in refreshed shell data")
	}
}

func TestHandleTUICommandReversesJournalEntryOnOriginalDate(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	posted, err := journal.PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	result, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandJournalReverse,
		Section: render.SectionJournal,
		ItemKey: posted.ID,
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("reverse journal entry through tui command: %v", err)
	}
	data := result.Data
	status := result.Status
	if status.Level != render.StatusSuccess {
		t.Fatalf("status level = %q, want %q", status.Level, render.StatusSuccess)
	}
	if !strings.Contains(status.Text, "Entry #1 reversed as entry #2.") {
		t.Fatalf("status text = %q, want reversal summary", status.Text)
	}

	entries, err := journal.ListBrowseEntries(ctx, databasePath)
	if err != nil {
		t.Fatalf("list browse entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if entries[0].EntryDate != "2026-03-08" {
		t.Fatalf("reversal entry date = %q, want original date 2026-03-08", entries[0].EntryDate)
	}
	if entries[0].Description != "Reversal of entry #1" {
		t.Fatalf("reversal description = %q, want default description", entries[0].Description)
	}

	found := false
	for _, item := range data.Journal.Items {
		if item.Key != posted.ID {
			continue
		}
		found = true
		if len(item.Actions) != 0 {
			t.Fatalf("reversed original should not expose action after refresh: %#v", item)
		}
		if !strings.Contains(item.Row, "reversed") {
			t.Fatalf("refreshed row = %q, want reversed status", item.Row)
		}
		break
	}
	if !found {
		t.Fatal("expected original journal item in refreshed shell data")
	}
}

func TestBuildTUIShellDataIncludesEntryCatalog(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	data, err := buildTUIShellData(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	if data.EntryCatalog.DefaultDate == "" {
		t.Fatal("expected entry catalog default date")
	}
	if len(data.EntryCatalog.ExpenseAccounts) == 0 || len(data.EntryCatalog.IncomeAccounts) == 0 || len(data.EntryCatalog.DepositAccounts) == 0 {
		t.Fatalf("entry catalog missing expected account classes: %#v", data.EntryCatalog)
	}
	if data.Dashboard.QuickEntryLines[0] != "e  I have an expense" {
		t.Fatalf("quick entry lines = %#v, want expense launcher", data.Dashboard.QuickEntryLines)
	}
}

func TestHandleTUICommandCreatesExpenseAndNavigatesToJournal(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	result, err := handleTUICommand(ctx, render.Command{
		ID: tuiCommandCreateExpense,
		Fields: map[string]string{
			"date":                "2026-03-10",
			"description":         "Restock arrows",
			"amount":              "25",
			"account_code":        "5100",
			"offset_account_code": "1000",
			"memo":                "Quiver refill",
		},
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("create expense through tui command: %v", err)
	}
	if result.NavigateTo != render.SectionJournal {
		t.Fatalf("navigate section = %v, want journal", result.NavigateTo)
	}
	if result.SelectItemKey == "" {
		t.Fatal("expected selected journal item key after expense create")
	}
	if !strings.Contains(result.Status.Text, "Recorded expense as journal entry #1.") {
		t.Fatalf("status text = %q, want expense summary", result.Status.Text)
	}
}

func TestHandleTUICommandCreatesIncomeAndNavigatesToJournal(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	result, err := handleTUICommand(ctx, render.Command{
		ID: tuiCommandCreateIncome,
		Fields: map[string]string{
			"date":                "2026-03-10",
			"description":         "Goblin bounty",
			"amount":              "1gp",
			"account_code":        "4000",
			"offset_account_code": "1000",
			"memo":                "Mayor payout",
		},
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("create income through tui command: %v", err)
	}
	if result.NavigateTo != render.SectionJournal {
		t.Fatalf("navigate section = %v, want journal", result.NavigateTo)
	}
	if result.SelectItemKey == "" {
		t.Fatal("expected selected journal item key after income create")
	}
	if !strings.Contains(result.Status.Text, "Recorded income as journal entry #1.") {
		t.Fatalf("status text = %q, want income summary", result.Status.Text)
	}
}

func TestHandleTUICommandCreatesCustomEntryAndNavigatesToJournal(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	result, err := handleTUICommand(ctx, render.Command{
		ID: tuiCommandCreateCustom,
		Fields: map[string]string{
			"date":        "2026-03-10",
			"description": "Gear transfer",
		},
		Lines: []render.CommandLine{
			{Side: "debit", AccountCode: "1300", Amount: "500"},
			{Side: "credit", AccountCode: "1000", Amount: "500"},
		},
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("create custom through tui command: %v", err)
	}
	if result.NavigateTo != render.SectionJournal {
		t.Fatalf("navigate section = %v, want journal", result.NavigateTo)
	}
	if result.SelectItemKey == "" {
		t.Fatal("expected selected journal item key after custom create")
	}
	if !strings.Contains(result.Status.Text, "Recorded custom entry as journal entry #1.") {
		t.Fatalf("status text = %q, want custom summary", result.Status.Text)
	}
}

func TestBuildTUIShellDataAddsQuestActionsAndBalanceDetail(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	record, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Goblin Bounty",
		Patron:             "Mayor Rowan",
		PromisedBaseReward: 2500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-08",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}
	if err := quest.CompleteQuest(ctx, databasePath, record.ID, "2026-03-09"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	found := false
	for _, item := range data.Quests.Items {
		if item.Key != record.ID {
			continue
		}
		found = true
		if len(item.Actions) != 2 {
			t.Fatalf("quest item actions = %#v, want collect and write off", item.Actions)
		}
		if item.Actions[0].Trigger != render.ActionCollect || item.Actions[0].ID != tuiCommandQuestCollectFull {
			t.Fatalf("quest collect action = %#v", item.Actions[0])
		}
		if item.Actions[1].Trigger != render.ActionWriteOff || item.Actions[1].ID != tuiCommandQuestWriteOffFull {
			t.Fatalf("quest write-off action = %#v", item.Actions[1])
		}
		detail := strings.Join(item.DetailLines, "\n")
		for _, token := range []string{"Outstanding: 2 PP 5 GP", "Collected so far: 0 CP", "Accounting state: collectible but unpaid"} {
			if !strings.Contains(detail, token) {
				t.Fatalf("quest detail missing %q:\n%s", token, detail)
			}
		}
		if !strings.Contains(strings.Join(item.Actions[0].ConfirmLines, "\n"), "Collection date: 2026-03-10") {
			t.Fatalf("collect confirm lines = %#v", item.Actions[0].ConfirmLines)
		}
		break
	}
	if !found {
		t.Fatalf("expected quest item %q", record.ID)
	}
}

func TestHandleTUICommandCollectsQuestOnTodayDate(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	record, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Goblin Bounty",
		Patron:             "Mayor Rowan",
		PromisedBaseReward: 2500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-08",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}
	if err := quest.CompleteQuest(ctx, databasePath, record.ID, "2026-03-09"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	result, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandQuestCollectFull,
		Section: render.SectionQuests,
		ItemKey: record.ID,
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("collect quest through tui command: %v", err)
	}
	data := result.Data
	status := result.Status
	if status.Level != render.StatusSuccess {
		t.Fatalf("status level = %q, want %q", status.Level, render.StatusSuccess)
	}
	if !strings.Contains(status.Text, "Collected 2 PP 5 GP") {
		t.Fatalf("status text = %q, want collected amount", status.Text)
	}

	entries, err := journal.ListBrowseEntries(ctx, databasePath)
	if err != nil {
		t.Fatalf("list browse entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	if entries[0].EntryDate != "2026-03-10" {
		t.Fatalf("collection entry date = %q, want 2026-03-10", entries[0].EntryDate)
	}
	if entries[0].Description != "Quest payment: Goblin Bounty" {
		t.Fatalf("collection description = %q, want default description", entries[0].Description)
	}

	found := false
	for _, item := range data.Quests.Items {
		if item.Key != record.ID {
			continue
		}
		found = true
		if len(item.Actions) != 0 {
			t.Fatalf("paid quest should not expose actions after refresh: %#v", item)
		}
		detail := strings.Join(item.DetailLines, "\n")
		for _, token := range []string{"Status: paid", "Outstanding: 0 CP", "Collected so far: 2 PP 5 GP"} {
			if !strings.Contains(detail, token) {
				t.Fatalf("paid quest detail missing %q:\n%s", token, detail)
			}
		}
		break
	}
	if !found {
		t.Fatal("expected refreshed quest item after collection")
	}
}

func TestHandleTUICommandWritesOffQuestOnTodayDate(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	record, err := quest.CreateQuest(ctx, databasePath, &quest.CreateQuestInput{
		Title:              "Bridge Toll Cleanup",
		Patron:             "Road Warden",
		PromisedBaseReward: 1000,
		Status:             "accepted",
		AcceptedOn:         "2026-03-08",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}
	if err := quest.CompleteQuest(ctx, databasePath, record.ID, "2026-03-09"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	result, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandQuestWriteOffFull,
		Section: render.SectionQuests,
		ItemKey: record.ID,
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("write off quest through tui command: %v", err)
	}
	data := result.Data
	status := result.Status
	if status.Level != render.StatusSuccess {
		t.Fatalf("status level = %q, want %q", status.Level, render.StatusSuccess)
	}
	if !strings.Contains(status.Text, "Wrote off 1 PP") {
		t.Fatalf("status text = %q, want write-off amount", status.Text)
	}

	entries, err := journal.ListBrowseEntries(ctx, databasePath)
	if err != nil {
		t.Fatalf("list browse entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	if entries[0].EntryDate != "2026-03-10" {
		t.Fatalf("write-off entry date = %q, want 2026-03-10", entries[0].EntryDate)
	}
	if entries[0].Description != "Quest write-off: Bridge Toll Cleanup" {
		t.Fatalf("write-off description = %q, want default description", entries[0].Description)
	}

	found := false
	for _, item := range data.Quests.Items {
		if item.Key != record.ID {
			continue
		}
		found = true
		if len(item.Actions) != 0 {
			t.Fatalf("defaulted quest should not expose actions after refresh: %#v", item)
		}
		if !strings.Contains(strings.Join(item.DetailLines, "\n"), "Status: defaulted") {
			t.Fatalf("refreshed quest detail = %#v, want defaulted", item.DetailLines)
		}
		break
	}
	if !found {
		t.Fatal("expected refreshed quest item after write-off")
	}
}

func TestBuildTUIShellDataAddsLootRecognizeActionFromLatestAppraisal(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := loot.CreateLootItem(ctx, databasePath, "Gold Necklace", "Merchant", 1, "Bard", "Wrapped in velvet")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}
	if _, err := loot.AppraiseLootItem(ctx, databasePath, item.ID, 600, "Guild factor", "2026-03-08", "Initial pass"); err != nil {
		t.Fatalf("first appraisal: %v", err)
	}
	latest, err := loot.AppraiseLootItem(ctx, databasePath, item.ID, 750, "Master jeweler", "2026-03-09", "Better lighting")
	if err != nil {
		t.Fatalf("second appraisal: %v", err)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	found := false
	for _, row := range data.Loot.Items {
		if row.Key != item.ID {
			continue
		}
		found = true
		if len(row.Actions) != 1 {
			t.Fatalf("loot row actions = %#v, want single recognize action", row.Actions)
		}
		action := row.Actions[0]
		if action.Trigger != render.ActionRecognize {
			t.Fatalf("loot action trigger = %q, want %q", action.Trigger, render.ActionRecognize)
		}
		if action.ID != tuiCommandLootRecognize {
			t.Fatalf("loot action id = %q, want %q", action.ID, tuiCommandLootRecognize)
		}
		detail := strings.Join(row.DetailLines, "\n")
		for _, token := range []string{
			"Latest appraisal: 7 GP 5 SP",
			"Appraised on: 2026-03-09",
			"Appraiser: Master jeweler",
			"Appraisals tracked: 2",
			"Holder: Bard",
			"Item notes: Wrapped in velvet",
			"Accounting state: appraised but off-ledger",
		} {
			if !strings.Contains(detail, token) {
				t.Fatalf("loot detail missing %q:\n%s", token, detail)
			}
		}
		confirm := strings.Join(action.ConfirmLines, "\n")
		for _, token := range []string{
			"Recognition date: 2026-03-10",
			"This uses the latest of 2 appraisals.",
			latest.ID,
		} {
			if !strings.Contains(confirm, token) {
				t.Fatalf("loot confirm lines missing %q:\n%s", token, confirm)
			}
		}
		break
	}
	if !found {
		t.Fatalf("expected loot row for %q", item.ID)
	}
}

func TestBuildTUIShellDataAddsLootSellActionForRecognizedItems(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := loot.CreateLootItem(ctx, databasePath, "Gold Necklace", "Merchant", 1, "Bard", "Wrapped in velvet")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}
	appraisal, err := loot.AppraiseLootItem(ctx, databasePath, item.ID, 750, "Master jeweler", "2026-03-09", "Better lighting")
	if err != nil {
		t.Fatalf("appraise item: %v", err)
	}
	if _, err := loot.RecognizeLootAppraisal(ctx, databasePath, appraisal.ID, "2026-03-10", ""); err != nil {
		t.Fatalf("recognize item: %v", err)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	found := false
	for _, row := range data.Loot.Items {
		if row.Key != item.ID {
			continue
		}
		found = true
		if len(row.Actions) != 1 {
			t.Fatalf("recognized loot row actions = %#v, want single sell action", row.Actions)
		}
		action := row.Actions[0]
		if action.Trigger != render.ActionSell {
			t.Fatalf("loot action trigger = %q, want %q", action.Trigger, render.ActionSell)
		}
		if action.ID != tuiCommandLootSell {
			t.Fatalf("loot action id = %q, want %q", action.ID, tuiCommandLootSell)
		}
		if action.Mode != render.ItemActionModeInput {
			t.Fatalf("loot action mode = %q, want input", action.Mode)
		}
		if action.Placeholder != "7 GP 5 SP" {
			t.Fatalf("loot action placeholder = %q, want 7 GP 5 SP", action.Placeholder)
		}
		detail := strings.Join(row.DetailLines, "\n")
		for _, token := range []string{
			"Status: recognized",
			"Recognized value: 7 GP 5 SP",
			"Sale state: sellable from recognized basis",
		} {
			if !strings.Contains(detail, token) {
				t.Fatalf("loot sell detail missing %q:\n%s", token, detail)
			}
		}
		help := strings.Join(action.InputHelp, "\n")
		for _, token := range []string{
			"Sale date: 2026-03-10",
			"Recognized value: 7 GP 5 SP",
			"Enter sale proceeds in GP/SP/CP format.",
		} {
			if !strings.Contains(help, token) {
				t.Fatalf("loot sale input help missing %q:\n%s", token, help)
			}
		}
		break
	}
	if !found {
		t.Fatalf("expected recognized loot row for %q", item.ID)
	}
}

func TestBuildTUIShellDataOmitsLootRecognizeWithoutPositiveLatestAppraisal(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	noAppraisal, err := loot.CreateLootItem(ctx, databasePath, "Unknown Relic", "Ruins", 1, "", "")
	if err != nil {
		t.Fatalf("create no-appraisal item: %v", err)
	}

	zeroAppraisal, err := loot.CreateLootItem(ctx, databasePath, "Worthless Trinket", "Roadside", 1, "", "")
	if err != nil {
		t.Fatalf("create zero-appraisal item: %v", err)
	}
	if _, err := loot.AppraiseLootItem(ctx, databasePath, zeroAppraisal.ID, 0, "", "2026-03-08", "Unknown value"); err != nil {
		t.Fatalf("zero appraisal: %v", err)
	}

	data, err := buildTUIShellData(ctx, databasePath, assets)
	if err != nil {
		t.Fatalf("build shell data: %v", err)
	}

	foundNoAppraisal := false
	foundZero := false
	for _, row := range data.Loot.Items {
		switch row.Key {
		case noAppraisal.ID:
			foundNoAppraisal = true
			if len(row.Actions) != 0 {
				t.Fatalf("no-appraisal row should not expose actions: %#v", row)
			}
			if !strings.Contains(strings.Join(row.DetailLines, "\n"), "Latest appraisal: Unknown / none") {
				t.Fatalf("no-appraisal detail = %#v", row.DetailLines)
			}
		case zeroAppraisal.ID:
			foundZero = true
			if len(row.Actions) != 0 {
				t.Fatalf("zero-appraisal row should not expose actions: %#v", row)
			}
			if !strings.Contains(strings.Join(row.DetailLines, "\n"), "Latest appraisal: 0 CP") {
				t.Fatalf("zero-appraisal detail = %#v", row.DetailLines)
			}
		}
	}

	if !foundNoAppraisal || !foundZero {
		t.Fatalf("expected both non-actionable loot rows")
	}
}

func TestHandleTUICommandRecognizesLootOnTodayDate(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := loot.CreateLootItem(ctx, databasePath, "Gold Necklace", "Merchant", 1, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}
	if _, err := loot.AppraiseLootItem(ctx, databasePath, item.ID, 600, "Guild factor", "2026-03-08", ""); err != nil {
		t.Fatalf("first appraisal: %v", err)
	}
	latest, err := loot.AppraiseLootItem(ctx, databasePath, item.ID, 750, "Master jeweler", "2026-03-09", "")
	if err != nil {
		t.Fatalf("second appraisal: %v", err)
	}

	result, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandLootRecognize,
		Section: render.SectionLoot,
		ItemKey: item.ID,
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("recognize loot through tui command: %v", err)
	}
	data := result.Data
	status := result.Status
	if status.Level != render.StatusSuccess {
		t.Fatalf("status level = %q, want %q", status.Level, render.StatusSuccess)
	}
	if !strings.Contains(status.Text, "Recognized loot item \"Gold Necklace\" as entry #1.") {
		t.Fatalf("status text = %q, want recognition summary", status.Text)
	}

	entries, err := journal.ListBrowseEntries(ctx, databasePath)
	if err != nil {
		t.Fatalf("list browse entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	if entries[0].EntryDate != "2026-03-10" {
		t.Fatalf("recognition entry date = %q, want 2026-03-10", entries[0].EntryDate)
	}
	if entries[0].Description != "Recognize loot appraisal: "+latest.ID {
		t.Fatalf("recognition description = %q, want default appraisal description", entries[0].Description)
	}

	found := false
	for _, row := range data.Loot.Items {
		if row.Key != item.ID {
			continue
		}
		found = true
		if len(row.Actions) != 1 || row.Actions[0].ID != tuiCommandLootSell {
			t.Fatalf("recognized loot should expose sell action after refresh: %#v", row)
		}
		detail := strings.Join(row.DetailLines, "\n")
		for _, token := range []string{"Status: recognized", "Accounting state: on-ledger recognized inventory", "Latest appraisal: 7 GP 5 SP", "Recognized value: 7 GP 5 SP"} {
			if !strings.Contains(detail, token) {
				t.Fatalf("recognized loot detail missing %q:\n%s", token, detail)
			}
		}
		break
	}
	if !found {
		t.Fatal("expected recognized loot row after refresh")
	}
}

func TestHandleTUICommandRejectsInvalidLootSaleAmountAsInputError(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := loot.CreateLootItem(ctx, databasePath, "Gold Necklace", "Merchant", 1, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}
	appraisal, err := loot.AppraiseLootItem(ctx, databasePath, item.ID, 750, "Master jeweler", "2026-03-09", "")
	if err != nil {
		t.Fatalf("appraise loot item: %v", err)
	}
	if _, err := loot.RecognizeLootAppraisal(ctx, databasePath, appraisal.ID, "2026-03-10", ""); err != nil {
		t.Fatalf("recognize loot: %v", err)
	}

	_, err = handleTUICommand(ctx, render.Command{
		ID:      tuiCommandLootSell,
		Section: render.SectionLoot,
		ItemKey: item.ID,
		Fields: map[string]string{
			"amount": "banana",
		},
	}, databasePath, assets)
	if err == nil {
		t.Fatal("expected invalid sale amount error")
	}

	inputErr, ok := err.(render.InputError)
	if !ok {
		t.Fatalf("error type = %T, want render.InputError", err)
	}
	if !strings.Contains(inputErr.Error(), `Invalid amount "banana".`) {
		t.Fatalf("input error = %q, want invalid amount message", inputErr.Error())
	}
}

func TestHandleTUICommandSellsLootOnTodayDate(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	originalNow := tuiNow
	tuiNow = func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }
	defer func() { tuiNow = originalNow }()

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := loot.CreateLootItem(ctx, databasePath, "Gold Necklace", "Merchant", 1, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}
	appraisal, err := loot.AppraiseLootItem(ctx, databasePath, item.ID, 750, "Master jeweler", "2026-03-09", "")
	if err != nil {
		t.Fatalf("appraise loot item: %v", err)
	}
	if _, err := loot.RecognizeLootAppraisal(ctx, databasePath, appraisal.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("recognize loot: %v", err)
	}

	result, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandLootSell,
		Section: render.SectionLoot,
		ItemKey: item.ID,
		Fields: map[string]string{
			"amount": "8 gp",
		},
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("sell loot through tui command: %v", err)
	}
	data := result.Data
	status := result.Status
	if status.Level != render.StatusSuccess {
		t.Fatalf("status level = %q, want %q", status.Level, render.StatusSuccess)
	}
	if !strings.Contains(status.Text, `Sold loot item "Gold Necklace" as entry #2.`) {
		t.Fatalf("status text = %q, want sale summary", status.Text)
	}

	entries, err := journal.ListBrowseEntries(ctx, databasePath)
	if err != nil {
		t.Fatalf("list browse entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if entries[0].EntryDate != "2026-03-10" {
		t.Fatalf("sale entry date = %q, want 2026-03-10", entries[0].EntryDate)
	}
	if entries[0].Description != "Sale of loot item: "+item.ID {
		t.Fatalf("sale description = %q, want default sale description", entries[0].Description)
	}

	for _, row := range data.Loot.Items {
		if row.Key == item.ID {
			t.Fatalf("sold loot item should be absent from refreshed loot screen: %#v", row)
		}
	}
}

func joinItemRows(items []render.ListItemData) string {
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, item.Row)
	}

	return strings.Join(rows, "\n")
}
