package loot

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestListBrowseItemsIncludesLatestAppraisalMetadata(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Gold Necklace", "Merchant", 1, "Bard", "Wrapped in velvet", "loot")
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	if _, err := AppraiseLootItem(ctx, databasePath, campaignID, item.ID, 600, "Guild factor", "2026-03-08", "First pass"); err != nil {
		t.Fatalf("first appraisal: %v", err)
	}
	latest, err := AppraiseLootItem(ctx, databasePath, campaignID, item.ID, 750, "Master jeweler", "2026-03-09", "Better lighting")
	if err != nil {
		t.Fatalf("second appraisal: %v", err)
	}

	rows, err := ListBrowseItems(ctx, databasePath, campaignID, "loot")
	if err != nil {
		t.Fatalf("list browse items: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("row count = %d, want 1", len(rows))
	}

	row := rows[0]
	if row.ID != item.ID {
		t.Fatalf("item id = %q, want %q", row.ID, item.ID)
	}
	if row.AppraisalCount != 2 {
		t.Fatalf("appraisal count = %d, want 2", row.AppraisalCount)
	}
	if row.LatestAppraisal == nil {
		t.Fatal("expected latest appraisal")
	}
	if row.LatestAppraisal.ID != latest.ID {
		t.Fatalf("latest appraisal id = %q, want %q", row.LatestAppraisal.ID, latest.ID)
	}
	if row.LatestAppraisal.AppraisedValue != 750 {
		t.Fatalf("latest appraisal value = %d, want 750", row.LatestAppraisal.AppraisedValue)
	}
	if row.LatestAppraisal.Appraiser != "Master jeweler" {
		t.Fatalf("latest appraiser = %q, want Master jeweler", row.LatestAppraisal.Appraiser)
	}
	if row.Holder != "Bard" {
		t.Fatalf("holder = %q, want Bard", row.Holder)
	}
	if row.Notes != "Wrapped in velvet" {
		t.Fatalf("notes = %q, want Wrapped in velvet", row.Notes)
	}
}

func TestListBrowseItemsIncludesRecognizedItemsAndRecognizedEntryLinkage(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Silver Chalice", "Goblin den", 1, "", "", "loot")
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	appraisal, err := AppraiseLootItem(ctx, databasePath, campaignID, item.ID, 800, "Guild factor", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise item: %v", err)
	}

	entry, err := RecognizeLootAppraisal(ctx, databasePath, campaignID, appraisal.ID, "2026-03-09", "")
	if err != nil {
		t.Fatalf("recognize appraisal: %v", err)
	}

	rows, err := ListBrowseItems(ctx, databasePath, campaignID, "loot")
	if err != nil {
		t.Fatalf("list browse items: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("row count = %d, want 1", len(rows))
	}

	row := rows[0]
	if row.Status != ledger.LootStatusRecognized {
		t.Fatalf("status = %q, want recognized", row.Status)
	}
	if row.LatestAppraisal == nil {
		t.Fatal("expected latest appraisal")
	}
	if row.LatestAppraisal.RecognizedEntryID != entry.ID {
		t.Fatalf("recognized entry id = %q, want %q", row.LatestAppraisal.RecognizedEntryID, entry.ID)
	}
	if !row.HasRecognizedAppraisal {
		t.Fatal("expected recognized appraisal flag")
	}
	if row.RecognizedAppraisalValue != 800 {
		t.Fatalf("recognized appraisal value = %d, want 800", row.RecognizedAppraisalValue)
	}
}

func TestListBrowseItemsExcludesSoldItems(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	item, err := CreateLootItem(ctx, databasePath, campaignID, "Ruby", "Cave", 1, "", "", "loot")
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	appraisal, err := AppraiseLootItem(ctx, databasePath, campaignID, item.ID, 500, "", "2026-03-08", "")
	if err != nil {
		t.Fatalf("appraise item: %v", err)
	}
	if _, err := RecognizeLootAppraisal(ctx, databasePath, campaignID, appraisal.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("recognize appraisal: %v", err)
	}
	if _, err := SellLootItem(ctx, databasePath, campaignID, item.ID, 500, "2026-03-10", ""); err != nil {
		t.Fatalf("sell item: %v", err)
	}

	rows, err := ListBrowseItems(ctx, databasePath, campaignID, "loot")
	if err != nil {
		t.Fatalf("list browse items: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("row count = %d, want 0", len(rows))
	}
}
