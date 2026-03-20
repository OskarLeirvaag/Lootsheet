package notes

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestDeleteNoteRemovesReferences(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	note, err := CreateNote(ctx, databasePath, campaignID, &CreateNoteInput{
		Title: "Session 5",
		Body:  "Party visited @[person/Mayor Elra] and discussed @[quest/Bridge Toll].",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	// Verify references were created.
	refCount := testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT COUNT(*) FROM entity_references WHERE source_type = 'note' AND source_id = '"+note.ID+"'")
	if refCount != "2" {
		t.Fatalf("expected 2 references after create, got %s", refCount)
	}

	// Delete the note.
	if err := DeleteNote(ctx, databasePath, campaignID, note.ID); err != nil {
		t.Fatalf("delete note: %v", err)
	}

	// Verify references were cleaned up.
	refCount = testutil.RunSQLiteQueryForTest(t, databasePath,
		"SELECT COUNT(*) FROM entity_references WHERE source_type = 'note' AND source_id = '"+note.ID+"'")
	if refCount != "0" {
		t.Fatalf("expected 0 references after delete, got %s", refCount)
	}
}
