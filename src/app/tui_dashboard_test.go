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

	data, status, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandJournalReverse,
		Section: render.SectionJournal,
		ItemKey: posted.ID,
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("reverse journal entry through tui command: %v", err)
	}
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

	data, status, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandQuestCollectFull,
		Section: render.SectionQuests,
		ItemKey: record.ID,
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("collect quest through tui command: %v", err)
	}
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

	data, status, err := handleTUICommand(ctx, render.Command{
		ID:      tuiCommandQuestWriteOffFull,
		Section: render.SectionQuests,
		ItemKey: record.ID,
	}, databasePath, assets)
	if err != nil {
		t.Fatalf("write off quest through tui command: %v", err)
	}
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

func joinItemRows(items []render.ListItemData) string {
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, item.Row)
	}

	return strings.Join(rows, "\n")
}
