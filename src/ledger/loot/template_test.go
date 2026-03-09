package loot

import (
	"context"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

func TestSaveAssetTemplate(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, "Tavern", "Town square", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
	}

	if err := SaveAssetTemplate(ctx, databasePath, item.ID, lines); err != nil {
		t.Fatalf("save template: %v", err)
	}

	saved, err := ListAssetTemplateLines(ctx, databasePath, item.ID)
	if err != nil {
		t.Fatalf("list template lines: %v", err)
	}
	if len(saved) != 2 {
		t.Fatalf("line count = %d, want 2", len(saved))
	}
	if saved[0].Side != "debit" || saved[0].AccountCode != "1000" {
		t.Fatalf("line 0 = %s/%s, want debit/1000", saved[0].Side, saved[0].AccountCode)
	}
	if saved[1].Side != "credit" || saved[1].AccountCode != "4400" {
		t.Fatalf("line 1 = %s/%s, want credit/4400", saved[1].Side, saved[1].AccountCode)
	}
	if saved[0].SortOrder != 0 || saved[1].SortOrder != 1 {
		t.Fatalf("sort orders = %d/%d, want 0/1", saved[0].SortOrder, saved[1].SortOrder)
	}
}

func TestSaveAssetTemplateRejectsLootItem(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, "Ruby", "Cave", 1, "", "", "loot")
	if err != nil {
		t.Fatalf("create loot: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
	}

	err = SaveAssetTemplate(ctx, databasePath, item.ID, lines)
	if err == nil {
		t.Fatal("expected error for loot item")
	}
	if !strings.Contains(err.Error(), "not an asset") {
		t.Fatalf("error = %q, want 'not an asset'", err)
	}
}

func TestSaveAssetTemplateRejectsInvalidSide(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, "Mine", "Hills", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "invalid", AccountCode: "1000"},
	}

	err = SaveAssetTemplate(ctx, databasePath, item.ID, lines)
	if err == nil {
		t.Fatal("expected error for invalid side")
	}
	if !strings.Contains(err.Error(), "side must be debit or credit") {
		t.Fatalf("error = %q, want side validation", err)
	}
}

func TestSaveAssetTemplateRejectsEmptyAccountCode(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, "Mine", "Hills", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: ""},
	}

	err = SaveAssetTemplate(ctx, databasePath, item.ID, lines)
	if err == nil {
		t.Fatal("expected error for empty account code")
	}
	if !strings.Contains(err.Error(), "account code is required") {
		t.Fatalf("error = %q, want account code validation", err)
	}
}

func TestListAssetTemplateLines(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, "Tavern", "Town", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
		{Side: "debit", AccountCode: "5600"},
		{Side: "credit", AccountCode: "1000"},
	}

	if err := SaveAssetTemplate(ctx, databasePath, item.ID, lines); err != nil {
		t.Fatalf("save template: %v", err)
	}

	saved, err := ListAssetTemplateLines(ctx, databasePath, item.ID)
	if err != nil {
		t.Fatalf("list template lines: %v", err)
	}
	if len(saved) != 4 {
		t.Fatalf("line count = %d, want 4", len(saved))
	}
	for i, line := range saved {
		if line.SortOrder != i {
			t.Fatalf("line %d sort_order = %d, want %d", i, line.SortOrder, i)
		}
		if line.LootItemID != item.ID {
			t.Fatalf("line %d item id = %q, want %q", i, line.LootItemID, item.ID)
		}
	}
}

func TestDeleteAssetTemplate(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, "Tavern", "Town", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
	}

	if err := SaveAssetTemplate(ctx, databasePath, item.ID, lines); err != nil {
		t.Fatalf("save template: %v", err)
	}

	if err := DeleteAssetTemplate(ctx, databasePath, item.ID); err != nil {
		t.Fatalf("delete template: %v", err)
	}

	saved, err := ListAssetTemplateLines(ctx, databasePath, item.ID)
	if err != nil {
		t.Fatalf("list template lines: %v", err)
	}
	if len(saved) != 0 {
		t.Fatalf("line count = %d, want 0", len(saved))
	}
}

func TestSaveAssetTemplateReplace(t *testing.T) {
	databasePath := ledger.InitTestDB(t)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, "Rental House", "Market district", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	original := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
	}
	if err := SaveAssetTemplate(ctx, databasePath, item.ID, original); err != nil {
		t.Fatalf("save original template: %v", err)
	}

	replacement := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
		{Side: "debit", AccountCode: "5600"},
		{Side: "credit", AccountCode: "1000"},
	}
	if err := SaveAssetTemplate(ctx, databasePath, item.ID, replacement); err != nil {
		t.Fatalf("save replacement template: %v", err)
	}

	saved, err := ListAssetTemplateLines(ctx, databasePath, item.ID)
	if err != nil {
		t.Fatalf("list template lines: %v", err)
	}
	if len(saved) != 4 {
		t.Fatalf("line count = %d, want 4 (replacement)", len(saved))
	}
	if saved[2].AccountCode != "5600" {
		t.Fatalf("line 2 account = %q, want 5600", saved[2].AccountCode)
	}
}
