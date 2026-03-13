package journal

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestListBrowseEntriesReturnsLinesAndReversalLinkage(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	first, err := PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25, Memo: "Quiver refill"},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post first journal entry: %v", err)
	}

	second, err := PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-09",
		Description: "Quest reward earned",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1100", DebitAmount: 100},
			{AccountCode: "4000", CreditAmount: 100},
		},
	})
	if err != nil {
		t.Fatalf("post second journal entry: %v", err)
	}

	reversal, err := ReverseJournalEntry(ctx, databasePath, campaignID, first.ID, "2026-03-08", "")
	if err != nil {
		t.Fatalf("reverse first journal entry: %v", err)
	}

	entries, err := ListBrowseEntries(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("list browse entries: %v", err)
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
	if entries[0].ReversesEntryNumber != first.EntryNumber {
		t.Fatalf("reversal entry reverses number = %d, want %d", entries[0].ReversesEntryNumber, first.EntryNumber)
	}
	if len(entries[0].Lines) != 2 {
		t.Fatalf("reversal line count = %d, want 2", len(entries[0].Lines))
	}
	if entries[0].Lines[0].AccountCode != "5100" || entries[0].Lines[0].AccountName != "Arrows & Ammunition" {
		t.Fatalf("reversal first line = %#v, want arrows expense", entries[0].Lines[0])
	}
	if entries[0].Lines[0].CreditAmount != 25 {
		t.Fatalf("reversal first line credit = %d, want 25", entries[0].Lines[0].CreditAmount)
	}

	if entries[1].EntryNumber != second.EntryNumber {
		t.Fatalf("second entry number = %d, want %d", entries[1].EntryNumber, second.EntryNumber)
	}
	if len(entries[1].Lines) != 2 {
		t.Fatalf("second line count = %d, want 2", len(entries[1].Lines))
	}

	if entries[2].EntryNumber != first.EntryNumber {
		t.Fatalf("oldest entry number = %d, want %d", entries[2].EntryNumber, first.EntryNumber)
	}
	if entries[2].ReversedByEntryID != reversal.ID {
		t.Fatalf("original reversed by = %q, want %q", entries[2].ReversedByEntryID, reversal.ID)
	}
	if entries[2].ReversedByEntryNumber != reversal.EntryNumber {
		t.Fatalf("original reversed by number = %d, want %d", entries[2].ReversedByEntryNumber, reversal.EntryNumber)
	}
	if len(entries[2].Lines) != 2 {
		t.Fatalf("original line count = %d, want 2", len(entries[2].Lines))
	}
	if entries[2].Lines[0].Memo != "Quiver refill" {
		t.Fatalf("original first line memo = %q, want Quiver refill", entries[2].Lines[0].Memo)
	}
	if entries[2].Lines[0].DebitAmount != 25 {
		t.Fatalf("original first line debit = %d, want 25", entries[2].Lines[0].DebitAmount)
	}
}
