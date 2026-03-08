package app

import (
	"context"
	"strings"
	"testing"

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
	if item.PrimaryAction == nil {
		t.Fatalf("account item missing primary action: %#v", item)
	}
	if item.PrimaryAction.ID != tuiCommandAccountDeactivate {
		t.Fatalf("primary action id = %q, want %q", item.PrimaryAction.ID, tuiCommandAccountDeactivate)
	}
	if !strings.Contains(item.PrimaryAction.Label, "deactivate") {
		t.Fatalf("primary action label = %q, want deactivate", item.PrimaryAction.Label)
	}
}

func TestHandleTUICommandTogglesAccountState(t *testing.T) {
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	data, status, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandAccountDeactivate,
		Section: render.SectionAccounts,
		ItemKey: "1000",
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("deactivate account through tui command: %v", err)
	}
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
		if item.PrimaryAction == nil || item.PrimaryAction.ID != tuiCommandAccountActivate {
			t.Fatalf("account item after deactivate = %#v, want activate action", item)
		}
		break
	}
	if !found {
		t.Fatal("expected account 1000 in refreshed shell data")
	}
}

func joinItemRows(items []render.ListItemData) string {
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, item.Row)
	}

	return strings.Join(rows, "\n")
}
