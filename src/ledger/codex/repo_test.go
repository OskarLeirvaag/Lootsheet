package codex

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestDeleteEntryRemovesReferences(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	entry, err := CreateEntry(ctx, databasePath, campaignID, &CreateInput{
		Name:  "Garrick",
		Notes: "Met near @[quest/Dragon Slaying] and @[person/Elra].",
	})
	if err != nil {
		t.Fatalf("create codex entry: %v", err)
	}

	// Verify references were created.
	refCount := testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT COUNT(*) FROM entity_references WHERE source_type = 'codex' AND source_id = '"+entry.ID+"'")
	if refCount != "2" {
		t.Fatalf("expected 2 references after create, got %s", refCount)
	}

	// Delete the entry.
	if err := DeleteEntry(ctx, databasePath, campaignID, entry.ID); err != nil {
		t.Fatalf("delete codex entry: %v", err)
	}

	// Verify references were cleaned up.
	refCount = testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT COUNT(*) FROM entity_references WHERE source_type = 'codex' AND source_id = '"+entry.ID+"'")
	if refCount != "0" {
		t.Fatalf("expected 0 references after delete, got %s", refCount)
	}
}
