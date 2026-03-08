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
