package loot

import (
	"context"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestSaveAssetTemplate(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Tavern", "Town square", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
	}

	if err := SaveAssetTemplate(ctx, databasePath, campaignID, item.ID, lines); err != nil {
		t.Fatalf("save template: %v", err)
	}

	lineCount := strings.TrimSpace(testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT COUNT(*) FROM asset_template_lines WHERE loot_item_id = '"+item.ID+"';",
	))
	if lineCount != "2" {
		t.Fatalf("line count = %s, want 2", lineCount)
	}
	lineData := strings.TrimSpace(testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT side || '/' || account_code || '/' || sort_order FROM asset_template_lines WHERE loot_item_id = '"+item.ID+"' ORDER BY sort_order;",
	))
	expectedLines := "debit/1000/0\ncredit/4400/1"
	if lineData != expectedLines {
		t.Fatalf("template lines = %q, want %q", lineData, expectedLines)
	}
}

func TestSaveAssetTemplateRejectsLootItem(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Ruby", "Cave", 1, "", "", "loot")
	if err != nil {
		t.Fatalf("create loot: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
	}

	err = SaveAssetTemplate(ctx, databasePath, campaignID, item.ID, lines)
	if err == nil {
		t.Fatal("expected error for loot item")
	}
	if !strings.Contains(err.Error(), "not an asset") {
		t.Fatalf("error = %q, want 'not an asset'", err)
	}
}

func TestSaveAssetTemplateRejectsInvalidSide(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Mine", "Hills", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "invalid", AccountCode: "1000"},
	}

	err = SaveAssetTemplate(ctx, databasePath, campaignID, item.ID, lines)
	if err == nil {
		t.Fatal("expected error for invalid side")
	}
	if !strings.Contains(err.Error(), "side must be debit or credit") {
		t.Fatalf("error = %q, want side validation", err)
	}
}

func TestSaveAssetTemplateRejectsEmptyAccountCode(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Mine", "Hills", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	lines := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: ""},
	}

	err = SaveAssetTemplate(ctx, databasePath, campaignID, item.ID, lines)
	if err == nil {
		t.Fatal("expected error for empty account code")
	}
	if !strings.Contains(err.Error(), "account code is required") {
		t.Fatalf("error = %q, want account code validation", err)
	}
}

func TestSaveAssetTemplateReplace(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Rental House", "Market district", 1, "", "", "asset")
	if err != nil {
		t.Fatalf("create asset: %v", err)
	}

	original := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
	}
	if err := SaveAssetTemplate(ctx, databasePath, campaignID, item.ID, original); err != nil {
		t.Fatalf("save original template: %v", err)
	}

	replacement := []AssetTemplateLineRecord{
		{Side: "debit", AccountCode: "1000"},
		{Side: "credit", AccountCode: "4400"},
		{Side: "debit", AccountCode: "5600"},
		{Side: "credit", AccountCode: "1000"},
	}
	if err := SaveAssetTemplate(ctx, databasePath, campaignID, item.ID, replacement); err != nil {
		t.Fatalf("save replacement template: %v", err)
	}

	lineCount := strings.TrimSpace(testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT COUNT(*) FROM asset_template_lines WHERE loot_item_id = '"+item.ID+"';",
	))
	if lineCount != "4" {
		t.Fatalf("line count = %s, want 4 (replacement)", lineCount)
	}
	line2Account := strings.TrimSpace(testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT account_code FROM asset_template_lines WHERE loot_item_id = '"+item.ID+"' AND sort_order = 2;",
	))
	if line2Account != "5600" {
		t.Fatalf("line 2 account = %q, want 5600", line2Account)
	}
}
