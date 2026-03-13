package campaign_test

import (
	"context"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func loadAssets(t *testing.T) config.InitAssets {
	t.Helper()
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}
	return assets
}

func TestCreateAndListCampaign(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()
	assets := loadAssets(t)

	record, err := campaign.Create(ctx, databasePath, "Frostfall", assets.Accounts)
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}
	if record.Name != "Frostfall" {
		t.Fatalf("name = %q, want Frostfall", record.Name)
	}
	if record.ID == "" {
		t.Fatal("expected non-empty campaign ID")
	}

	campaigns, err := campaign.List(ctx, databasePath)
	if err != nil {
		t.Fatalf("list campaigns: %v", err)
	}
	// Default campaign from InitTestDB + newly created one.
	if len(campaigns) != 2 {
		t.Fatalf("campaign count = %d, want 2", len(campaigns))
	}

	// Verify seed accounts were created for the new campaign.
	accountCount := strings.TrimSpace(testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT COUNT(*) FROM accounts WHERE campaign_id = '"+record.ID+"'"))
	if accountCount == "0" {
		t.Fatal("expected seed accounts for new campaign")
	}
}

func TestCreateCampaignRejectsEmptyName(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()
	assets := loadAssets(t)

	_, err := campaign.Create(ctx, databasePath, "  ", assets.Accounts)
	if err == nil {
		t.Fatal("expected error for empty campaign name")
	}
	if !strings.Contains(err.Error(), "campaign name is required") {
		t.Fatalf("error = %q, want name validation", err)
	}
}

func TestRenameCampaign(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()
	assets := loadAssets(t)

	record, err := campaign.Create(ctx, databasePath, "Old Name", assets.Accounts)
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}

	renamed, err := campaign.Rename(ctx, databasePath, record.ID, "New Name")
	if err != nil {
		t.Fatalf("rename campaign: %v", err)
	}
	if renamed.Name != "New Name" {
		t.Fatalf("name = %q, want New Name", renamed.Name)
	}
}

func TestRenameCampaignRejectsEmptyName(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()

	_, err := campaign.Rename(ctx, databasePath, "some-id", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRenameCampaignRejectsUnknownID(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()

	_, err := campaign.Rename(ctx, databasePath, "nonexistent", "Anything")
	if err == nil {
		t.Fatal("expected error for unknown campaign ID")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does not exist", err)
	}
}

func TestDeleteEmptyCampaign(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()
	assets := loadAssets(t)

	record, err := campaign.Create(ctx, databasePath, "Disposable", assets.Accounts)
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}

	if err := campaign.Delete(ctx, databasePath, record.ID); err != nil {
		t.Fatalf("delete campaign: %v", err)
	}

	campaigns, err := campaign.List(ctx, databasePath)
	if err != nil {
		t.Fatalf("list campaigns: %v", err)
	}
	for _, c := range campaigns {
		if c.ID == record.ID {
			t.Fatal("deleted campaign still present in list")
		}
	}
}

func TestDeleteCampaignWithDataFails(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()
	assets := loadAssets(t)

	record, err := campaign.Create(ctx, databasePath, "Has Data", assets.Accounts)
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}

	// Insert a quest to block deletion.
	testutil.ExecSQLiteForTest(t, databasePath,
		"INSERT INTO quests (id, campaign_id, title, patron, description, promised_base_reward, partial_advance, status, created_at, updated_at) VALUES (?, ?, 'Test Quest', 'Patron', 'Desc', 100, 0, 'offered', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		"test-quest", record.ID)

	err = campaign.Delete(ctx, databasePath, record.ID)
	if err == nil {
		t.Fatal("expected error when deleting campaign with data")
	}
	if !strings.Contains(err.Error(), "cannot delete campaign") {
		t.Fatalf("error = %q, want cannot delete", err)
	}
}

func TestGetActiveAndSetActive(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()
	assets := loadAssets(t)

	// Default campaign is active after InitTestDB.
	active, err := campaign.GetActive(ctx, databasePath)
	if err != nil {
		t.Fatalf("get active campaign: %v", err)
	}
	if active.Name != "Default" {
		t.Fatalf("active name = %q, want Default", active.Name)
	}

	// Create and switch to a new campaign.
	newCampaign, err := campaign.Create(ctx, databasePath, "Second Campaign", assets.Accounts)
	if err != nil {
		t.Fatalf("create second campaign: %v", err)
	}

	if err := campaign.SetActive(ctx, databasePath, newCampaign.ID); err != nil {
		t.Fatalf("set active campaign: %v", err)
	}

	active, err = campaign.GetActive(ctx, databasePath)
	if err != nil {
		t.Fatalf("get active after switch: %v", err)
	}
	if active.ID != newCampaign.ID {
		t.Fatalf("active ID = %q, want %q", active.ID, newCampaign.ID)
	}
}

func TestSetActiveRejectsUnknownCampaign(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()

	err := campaign.SetActive(ctx, databasePath, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown campaign ID")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does not exist", err)
	}
}
