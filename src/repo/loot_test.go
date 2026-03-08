package repo

import (
	"context"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

func TestCreateLootItem(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Ruby Gemstone", "Dragon Hoard", 1, "Bard", "Large and shiny")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}

	if item.ID == "" {
		t.Fatal("loot item ID is empty")
	}

	if item.Name != "Ruby Gemstone" {
		t.Fatalf("loot item name = %q, want Ruby Gemstone", item.Name)
	}

	if item.Status != service.LootStatusHeld {
		t.Fatalf("loot item status = %q, want held", item.Status)
	}

	if item.Source != "Dragon Hoard" {
		t.Fatalf("loot item source = %q, want Dragon Hoard", item.Source)
	}

	if item.Quantity != 1 {
		t.Fatalf("loot item quantity = %d, want 1", item.Quantity)
	}
}

func TestCreateLootItemRejectsEmptyName(t *testing.T) {
	databasePath := initTestDB(t)

	_, err := CreateLootItem(context.Background(), databasePath, "", "source", 1, "", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}

	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("error = %q, want name required", err)
	}
}

func TestListLootItems(t *testing.T) {
	databasePath := initTestDB(t)

	_, err := CreateLootItem(context.Background(), databasePath, "Sword", "", 1, "", "")
	if err != nil {
		t.Fatalf("create sword: %v", err)
	}

	_, err = CreateLootItem(context.Background(), databasePath, "Shield", "", 2, "", "")
	if err != nil {
		t.Fatalf("create shield: %v", err)
	}

	items, err := ListLootItems(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list loot items: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("loot item count = %d, want 2", len(items))
	}

	// Both should be held, sorted by name.
	if items[0].Name != "Shield" {
		t.Fatalf("first item = %q, want Shield (alphabetical)", items[0].Name)
	}

	if items[1].Name != "Sword" {
		t.Fatalf("second item = %q, want Sword", items[1].Name)
	}
}

func TestAppraiseLootItem(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Emerald", "Cave", 1, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}

	appraisal, err := AppraiseLootItem(context.Background(), databasePath, item.ID, 500, "Jeweler", "2026-03-08", "Fine quality")
	if err != nil {
		t.Fatalf("appraise loot item: %v", err)
	}

	if appraisal.ID == "" {
		t.Fatal("appraisal ID is empty")
	}

	if appraisal.AppraisedValue != 500 {
		t.Fatalf("appraised value = %d, want 500", appraisal.AppraisedValue)
	}

	if appraisal.Appraiser != "Jeweler" {
		t.Fatalf("appraiser = %q, want Jeweler", appraisal.Appraiser)
	}

	if appraisal.RecognizedEntryID != "" {
		t.Fatalf("recognized_entry_id should be empty (off-ledger), got %q", appraisal.RecognizedEntryID)
	}
}

func TestAppraiseRejectsNonHeldItem(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Diamond", "Mine", 1, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}

	// Appraise and recognize to move to 'recognized' status.
	appraisal, err := AppraiseLootItem(context.Background(), databasePath, item.ID, 1000, "", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}

	if _, err := RecognizeLootAppraisal(context.Background(), databasePath, appraisal.ID, "2026-03-08", ""); err != nil {
		t.Fatalf("recognize: %v", err)
	}

	// Try to appraise again — item is now 'recognized', not 'held'.
	_, err = AppraiseLootItem(context.Background(), databasePath, item.ID, 1200, "", "2026-03-09", "")
	if err == nil {
		t.Fatal("expected error appraising a recognized item")
	}

	if !strings.Contains(err.Error(), "cannot be appraised") {
		t.Fatalf("error = %q, want cannot be appraised", err)
	}
}

func TestRecognizeLootAppraisal(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Gold Necklace", "Merchant", 1, "", "")
	if err != nil {
		t.Fatalf("create loot item: %v", err)
	}

	appraisal, err := AppraiseLootItem(context.Background(), databasePath, item.ID, 750, "Goldsmith", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}

	entry, err := RecognizeLootAppraisal(context.Background(), databasePath, appraisal.ID, "2026-03-09", "Recognize gold necklace")
	if err != nil {
		t.Fatalf("recognize: %v", err)
	}

	if entry.EntryNumber < 1 {
		t.Fatalf("entry number = %d, want >= 1", entry.EntryNumber)
	}

	if entry.DebitTotal != 750 || entry.CreditTotal != 750 {
		t.Fatalf("entry totals = %d/%d, want 750/750", entry.DebitTotal, entry.CreditTotal)
	}

	// Verify item status changed to recognized.
	items, err := ListLootItems(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}

	if items[0].Status != service.LootStatusRecognized {
		t.Fatalf("item status = %q, want recognized", items[0].Status)
	}

	// Verify journal entry was created.
	lineCount := strings.TrimSpace(runSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_lines;"))
	if lineCount != "2" {
		t.Fatalf("journal line count = %q, want 2", lineCount)
	}
}

func TestRecognizeAlreadyRecognizedAppraisal(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Silver Ring", "Dungeon", 1, "", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	appraisal, err := AppraiseLootItem(context.Background(), databasePath, item.ID, 200, "", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}

	if _, err := RecognizeLootAppraisal(context.Background(), databasePath, appraisal.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("first recognize: %v", err)
	}

	// Try to recognize again.
	_, err = RecognizeLootAppraisal(context.Background(), databasePath, appraisal.ID, "2026-03-10", "")
	if err == nil {
		t.Fatal("expected error recognizing an already-recognized appraisal")
	}

	if !strings.Contains(err.Error(), "already recognized") {
		t.Fatalf("error = %q, want already recognized", err)
	}
}

func TestSellLootItemAtAppraisalValue(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Ruby", "Cave", 1, "", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	appraisal, err := AppraiseLootItem(context.Background(), databasePath, item.ID, 500, "", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}

	if _, err := RecognizeLootAppraisal(context.Background(), databasePath, appraisal.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("recognize: %v", err)
	}

	entry, err := SellLootItem(context.Background(), databasePath, item.ID, 500, "2026-03-10", "Sell ruby at appraised value")
	if err != nil {
		t.Fatalf("sell: %v", err)
	}

	// Exact match: 2 lines (Dr Cash, Cr Inventory), no gain/loss.
	if entry.LineCount != 2 {
		t.Fatalf("line count = %d, want 2", entry.LineCount)
	}

	if entry.DebitTotal != 500 || entry.CreditTotal != 500 {
		t.Fatalf("entry totals = %d/%d, want 500/500", entry.DebitTotal, entry.CreditTotal)
	}

	// Verify item status.
	items, err := ListLootItems(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}

	if items[0].Status != service.LootStatusSold {
		t.Fatalf("item status = %q, want sold", items[0].Status)
	}
}

func TestSellLootItemBelowAppraisalValue(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Damaged Gem", "Ruins", 1, "", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	appraisal, err := AppraiseLootItem(context.Background(), databasePath, item.ID, 500, "", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}

	if _, err := RecognizeLootAppraisal(context.Background(), databasePath, appraisal.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("recognize: %v", err)
	}

	entry, err := SellLootItem(context.Background(), databasePath, item.ID, 300, "2026-03-10", "Sell below appraisal")
	if err != nil {
		t.Fatalf("sell: %v", err)
	}

	// 3 lines: Dr Cash 300, Dr Loss 200, Cr Inventory 500.
	if entry.LineCount != 3 {
		t.Fatalf("line count = %d, want 3", entry.LineCount)
	}

	if entry.DebitTotal != 500 || entry.CreditTotal != 500 {
		t.Fatalf("entry totals = %d/%d, want 500/500", entry.DebitTotal, entry.CreditTotal)
	}
}

func TestSellLootItemAboveAppraisalValue(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Rare Pearl", "Ocean", 1, "", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	appraisal, err := AppraiseLootItem(context.Background(), databasePath, item.ID, 500, "", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise: %v", err)
	}

	if _, err := RecognizeLootAppraisal(context.Background(), databasePath, appraisal.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("recognize: %v", err)
	}

	entry, err := SellLootItem(context.Background(), databasePath, item.ID, 700, "2026-03-10", "Sell above appraisal")
	if err != nil {
		t.Fatalf("sell: %v", err)
	}

	// 3 lines: Dr Cash 700, Cr Inventory 500, Cr Gain 200.
	if entry.LineCount != 3 {
		t.Fatalf("line count = %d, want 3", entry.LineCount)
	}

	if entry.DebitTotal != 700 || entry.CreditTotal != 700 {
		t.Fatalf("entry totals = %d/%d, want 700/700", entry.DebitTotal, entry.CreditTotal)
	}
}

func TestSellLootItemRejectsHeldItem(t *testing.T) {
	databasePath := initTestDB(t)

	item, err := CreateLootItem(context.Background(), databasePath, "Unsold Gem", "Cave", 1, "", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = SellLootItem(context.Background(), databasePath, item.ID, 100, "2026-03-10", "")
	if err == nil {
		t.Fatal("expected error selling a held item")
	}

	if !strings.Contains(err.Error(), "cannot be sold") {
		t.Fatalf("error = %q, want cannot be sold", err)
	}
}

func TestSellLootItemRejectsNonexistentItem(t *testing.T) {
	databasePath := initTestDB(t)

	_, err := SellLootItem(context.Background(), databasePath, "nonexistent-id", 100, "2026-03-10", "")
	if err == nil {
		t.Fatal("expected error for nonexistent item")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does not exist", err)
	}
}

func TestRecognizeNonexistentAppraisal(t *testing.T) {
	databasePath := initTestDB(t)

	_, err := RecognizeLootAppraisal(context.Background(), databasePath, "nonexistent-id", "2026-03-09", "")
	if err == nil {
		t.Fatal("expected error for nonexistent appraisal")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does not exist", err)
	}
}
