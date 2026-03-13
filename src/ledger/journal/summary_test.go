package journal

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestGetSummaryEmptyJournal(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	summary, err := GetSummary(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("get journal summary: %v", err)
	}

	if summary.TotalEntries != 0 {
		t.Fatalf("total entries = %d, want 0", summary.TotalEntries)
	}
	if summary.LatestEntryNumber != 0 {
		t.Fatalf("latest entry number = %d, want 0", summary.LatestEntryNumber)
	}
}

func TestGetSummaryTracksPostedAndReversalEntries(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()

	posted, err := PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
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

	if _, err := ReverseJournalEntry(ctx, databasePath, posted.ID, "2026-03-09", "Correct duplicate"); err != nil {
		t.Fatalf("reverse journal entry: %v", err)
	}

	summary, err := GetSummary(ctx, databasePath)
	if err != nil {
		t.Fatalf("get journal summary: %v", err)
	}

	if summary.TotalEntries != 2 {
		t.Fatalf("total entries = %d, want 2", summary.TotalEntries)
	}
	if summary.PostedEntries != 1 {
		t.Fatalf("posted entries = %d, want 1", summary.PostedEntries)
	}
	if summary.ReversedEntries != 1 {
		t.Fatalf("reversed entries = %d, want 1", summary.ReversedEntries)
	}
	if summary.ReversalEntries != 1 {
		t.Fatalf("reversal entries = %d, want 1", summary.ReversalEntries)
	}
	if summary.LatestEntryNumber != 2 {
		t.Fatalf("latest entry number = %d, want 2", summary.LatestEntryNumber)
	}
	if summary.LatestEntryDate != "2026-03-09" {
		t.Fatalf("latest entry date = %q, want 2026-03-09", summary.LatestEntryDate)
	}
	if summary.LatestDescription != "Correct duplicate" {
		t.Fatalf("latest description = %q, want custom description", summary.LatestDescription)
	}
}
