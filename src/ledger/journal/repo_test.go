package journal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

func TestPostJournalEntryCreatesPostedEntryAndLines(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	result, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	if result.EntryNumber != 1 {
		t.Fatalf("entry number = %d, want 1", result.EntryNumber)
	}

	if result.LineCount != 2 || result.DebitTotal != 25 || result.CreditTotal != 25 {
		t.Fatalf("result = %+v, want 2 lines and 25/25 totals", result)
	}

	entryRow := strings.TrimSpace(ledger.RunSQLiteQueryForTest(
		t,
		databasePath,
		"SELECT status || '\t' || entry_date || '\t' || description || '\t' || posted_at FROM journal_entries;",
	))
	fields := strings.Split(entryRow, "\t")
	if len(fields) != 4 {
		t.Fatalf("entry row columns = %d, want 4", len(fields))
	}

	if fields[0] != "posted" || fields[1] != "2026-03-08" || fields[2] != "Restock arrows" {
		t.Fatalf("entry row = %q, want posted journal entry", entryRow)
	}

	if fields[3] == "" {
		t.Fatalf("posted_at is empty in row %q", entryRow)
	}

	lineCount := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_lines;"))
	if lineCount != "2" {
		t.Fatalf("journal line count = %q, want 2", lineCount)
	}
}

func TestPostJournalEntryRejectsUnbalancedInput(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	_, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Broken entry",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25},
			{AccountCode: "1000", CreditAmount: 20},
		},
	})
	if err == nil {
		t.Fatal("expected post journal entry to fail")
	}

	if !strings.Contains(err.Error(), "journal entry is not balanced") {
		t.Fatalf("error = %q, want balance error", err)
	}

	entryCount := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_entries;"))
	if entryCount != "0" {
		t.Fatalf("journal entry count = %q, want 0", entryCount)
	}
}

func TestUpdatePostedJournalEntryReturnsImmutabilityError(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	err = UpdateJournalEntry(context.Background(), databasePath, posted.ID, "Tampered description", "2026-03-09")
	if err == nil {
		t.Fatal("expected update of posted entry to fail")
	}

	if !errors.Is(err, ledger.ErrImmutableEntry) {
		t.Fatalf("error = %v, want ErrImmutableEntry", err)
	}

	descRow := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath,
		fmt.Sprintf("SELECT description FROM journal_entries WHERE id = '%s';", posted.ID),
	))
	if descRow != "Restock arrows" {
		t.Fatalf("description = %q, want original value", descRow)
	}
}

func TestDeletePostedJournalEntryReturnsImmutabilityError(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	err = DeleteJournalEntry(context.Background(), databasePath, posted.ID)
	if err == nil {
		t.Fatal("expected delete of posted entry to fail")
	}

	if !errors.Is(err, ledger.ErrImmutableEntry) {
		t.Fatalf("error = %v, want ErrImmutableEntry", err)
	}

	entryCount := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_entries;"))
	if entryCount != "1" {
		t.Fatalf("journal entry count = %q, want 1", entryCount)
	}
}

func TestCheckJournalEntryMutableReturnsImmutabilityErrorForPosted(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	err = CheckJournalEntryMutable(context.Background(), databasePath, posted.ID)
	if !errors.Is(err, ledger.ErrImmutableEntry) {
		t.Fatalf("error = %v, want ErrImmutableEntry", err)
	}
}

func TestCheckJournalEntryMutableReturnsImmutabilityErrorForReversed(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	ledger.RunSQLiteScriptForTest(t, databasePath,
		fmt.Sprintf("UPDATE journal_entries SET status = 'reversed' WHERE id = '%s';", posted.ID),
	)

	err = CheckJournalEntryMutable(context.Background(), databasePath, posted.ID)
	if !errors.Is(err, ledger.ErrImmutableEntry) {
		t.Fatalf("error = %v, want ErrImmutableEntry", err)
	}
}

func TestUpdateJournalLineOnPostedEntryReturnsImmutabilityError(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	lineID := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath,
		fmt.Sprintf("SELECT id FROM journal_lines WHERE journal_entry_id = '%s' LIMIT 1;", posted.ID),
	))

	err = UpdateJournalLine(context.Background(), databasePath, lineID, "Tampered memo", 999, 0)
	if !errors.Is(err, ledger.ErrImmutableEntry) {
		t.Fatalf("error = %v, want ErrImmutableEntry", err)
	}
}

func TestDeleteJournalLineOnPostedEntryReturnsImmutabilityError(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	lineID := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath,
		fmt.Sprintf("SELECT id FROM journal_lines WHERE journal_entry_id = '%s' LIMIT 1;", posted.ID),
	))

	err = DeleteJournalLine(context.Background(), databasePath, lineID)
	if !errors.Is(err, ledger.ErrImmutableEntry) {
		t.Fatalf("error = %v, want ErrImmutableEntry", err)
	}

	lineCount := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_lines;"))
	if lineCount != "2" {
		t.Fatalf("journal line count = %q, want 2", lineCount)
	}
}

func TestReverseJournalEntryCreatesReversalAndMarksOriginalReversed(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	reversal, err := ReverseJournalEntry(context.Background(), databasePath, posted.ID, "2026-03-09", "")
	if err != nil {
		t.Fatalf("reverse journal entry: %v", err)
	}

	if reversal.EntryNumber != 2 {
		t.Fatalf("reversal entry number = %d, want 2", reversal.EntryNumber)
	}

	if reversal.EntryDate != "2026-03-09" {
		t.Fatalf("reversal entry date = %q, want 2026-03-09", reversal.EntryDate)
	}

	if reversal.Description != "Reversal of entry #1" {
		t.Fatalf("reversal description = %q, want default description", reversal.Description)
	}

	if reversal.LineCount != 2 {
		t.Fatalf("reversal line count = %d, want 2", reversal.LineCount)
	}

	if reversal.DebitTotal != 25 || reversal.CreditTotal != 25 {
		t.Fatalf("reversal totals = %d/%d, want 25/25", reversal.DebitTotal, reversal.CreditTotal)
	}

	reversalLines := strings.TrimSpace(ledger.RunSQLiteQueryForTest(
		t,
		databasePath,
		fmt.Sprintf(
			"SELECT debit_amount || ',' || credit_amount FROM journal_lines WHERE journal_entry_id = '%s' ORDER BY line_number;",
			reversal.ID,
		),
	))
	if reversalLines != "0,25\n25,0" {
		t.Fatalf("reversal lines = %q, want swapped amounts", reversalLines)
	}

	reversesEntryID := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath,
		fmt.Sprintf("SELECT reverses_entry_id FROM journal_entries WHERE id = '%s';", reversal.ID),
	))
	if reversesEntryID != posted.ID {
		t.Fatalf("reverses_entry_id = %q, want %q", reversesEntryID, posted.ID)
	}

	originalStatus := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath,
		fmt.Sprintf("SELECT status FROM journal_entries WHERE id = '%s';", posted.ID),
	))
	if originalStatus != "reversed" {
		t.Fatalf("original status = %q, want reversed", originalStatus)
	}

	reversedAt := strings.TrimSpace(ledger.RunSQLiteQueryForTest(t, databasePath,
		fmt.Sprintf("SELECT COALESCE(reversed_at, '') FROM journal_entries WHERE id = '%s';", posted.ID),
	))
	if reversedAt == "" {
		t.Fatal("original entry reversed_at is empty")
	}
}

func TestReverseAlreadyReversedEntryFails(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	posted, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
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

	if _, err := ReverseJournalEntry(context.Background(), databasePath, posted.ID, "2026-03-09", ""); err != nil {
		t.Fatalf("first reversal: %v", err)
	}

	_, err = ReverseJournalEntry(context.Background(), databasePath, posted.ID, "2026-03-10", "")
	if err == nil {
		t.Fatal("expected reversing an already-reversed entry to fail")
	}

	if !errors.Is(err, ledger.ErrEntryNotReversible) {
		t.Fatalf("error = %v, want ErrEntryNotReversible", err)
	}
}

func TestReverseNonexistentEntryFails(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	_, err := ReverseJournalEntry(context.Background(), databasePath, "nonexistent-id", "2026-03-09", "")
	if err == nil {
		t.Fatal("expected reversing a nonexistent entry to fail")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does-not-exist error", err)
	}
}

func TestDeactivatedAccountRejectsJournalPosting(t *testing.T) {
	databasePath := ledger.InitTestDB(t)

	// Deactivate account 1000 directly via SQL
	ledger.RunSQLiteScriptForTest(t, databasePath,
		"UPDATE accounts SET active = 0 WHERE code = '1000';",
	)

	_, err := PostJournalEntry(context.Background(), databasePath, ledger.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Should fail",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 10},
			{AccountCode: "1000", CreditAmount: 10},
		},
	})
	if err == nil {
		t.Fatal("expected journal post to inactive account to fail")
	}

	if !strings.Contains(err.Error(), "inactive") {
		t.Fatalf("error = %q, want inactive error", err)
	}
}
