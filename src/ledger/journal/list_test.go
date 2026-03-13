package journal

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestListEntriesOrdersNewestFirst(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()

	first, err := PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post first journal entry: %v", err)
	}

	if _, err := PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-09",
		Description: "Quest reward earned",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1100", DebitAmount: 100},
			{AccountCode: "4000", CreditAmount: 100},
		},
	}); err != nil {
		t.Fatalf("post second journal entry: %v", err)
	}

	reversal, err := ReverseJournalEntry(ctx, databasePath, first.ID, "2026-03-10", "Correct duplicate")
	if err != nil {
		t.Fatalf("reverse first journal entry: %v", err)
	}

	entries, err := ListEntries(ctx, databasePath)
	if err != nil {
		t.Fatalf("list journal entries: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("entry count = %d, want 3", len(entries))
	}

	if entries[0].EntryNumber != reversal.EntryNumber {
		t.Fatalf("newest entry number = %d, want %d", entries[0].EntryNumber, reversal.EntryNumber)
	}
	if entries[0].ReversesEntryID != first.ID {
		t.Fatalf("reversal entry reverses = %q, want %q", entries[0].ReversesEntryID, first.ID)
	}
	if entries[1].EntryNumber != 2 {
		t.Fatalf("second entry number = %d, want 2", entries[1].EntryNumber)
	}
	if entries[2].EntryNumber != 1 {
		t.Fatalf("oldest entry number = %d, want 1", entries[2].EntryNumber)
	}
	if entries[2].Status != ledger.JournalEntryStatusReversed {
		t.Fatalf("original entry status = %q, want %q", entries[2].Status, ledger.JournalEntryStatusReversed)
	}
}
