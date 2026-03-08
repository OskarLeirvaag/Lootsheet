package repo

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

func TestDeleteAccountWithoutPostings(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	// Create a fresh account with no postings.
	_, err = CreateAccount(context.Background(), databasePath, "9900", "Disposable Account", service.AccountTypeExpense)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	// Delete should succeed.
	if err := DeleteAccount(context.Background(), databasePath, "9900"); err != nil {
		t.Fatalf("delete account: %v", err)
	}

	// Verify the account is gone.
	accounts, err := ListAccounts(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	for _, account := range accounts {
		if account.Code == "9900" {
			t.Fatal("expected account 9900 to be deleted")
		}
	}
}

func TestDeleteAccountWithPostings(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	// Post a journal entry using account 1000 (Party Cash) and 5100 (Adventuring Supplies).
	_, err = PostJournalEntry(context.Background(), databasePath, service.JournalPostInput{
		EntryDate:   "2026-03-08",
		Description: "Restock arrows",
		Lines: []service.JournalLineInput{
			{AccountCode: "5100", DebitAmount: 25, Memo: "Quiver refill"},
			{AccountCode: "1000", CreditAmount: 25},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	// Attempt to delete account 1000 which now has postings.
	err = DeleteAccount(context.Background(), databasePath, "1000")
	if err == nil {
		t.Fatal("expected delete of account with postings to fail")
	}

	if !errors.Is(err, ErrAccountHasPostings) {
		t.Fatalf("error = %v, want ErrAccountHasPostings", err)
	}

	// Verify the account still exists.
	accounts, err := ListAccounts(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	found := false
	for _, account := range accounts {
		if account.Code == "1000" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected account 1000 to still exist after failed delete")
	}
}

func TestAccountCodeImmutable(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	// Get the original account details for code 1000.
	accountsBefore, err := ListAccounts(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	var originalID, originalCode string
	for _, account := range accountsBefore {
		if account.Code == "1000" {
			originalID = account.ID
			originalCode = account.Code
			break
		}
	}

	if originalID == "" {
		t.Fatal("account code 1000 not found")
	}

	// Rename the account — this should only change the name.
	if err := RenameAccount(context.Background(), databasePath, "1000", "Gold Hoard"); err != nil {
		t.Fatalf("rename account: %v", err)
	}

	// Verify the code and ID are unchanged.
	accountsAfter, err := ListAccounts(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	found := false
	for _, account := range accountsAfter {
		if account.ID == originalID {
			found = true

			if account.Code != originalCode {
				t.Fatalf("account code changed from %q to %q after rename", originalCode, account.Code)
			}

			if account.Name != "Gold Hoard" {
				t.Fatalf("account name = %q, want Gold Hoard", account.Name)
			}

			break
		}
	}

	if !found {
		t.Fatal("original account ID not found after rename")
	}
}

func TestDeleteAccountRejectsNonexistentCode(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	err = DeleteAccount(context.Background(), databasePath, "9999")
	if err == nil {
		t.Fatal("expected delete of nonexistent account to fail")
	}

	if !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("error = %v, want ErrAccountNotFound", err)
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does-not-exist message", err)
	}
}

func TestDeleteAccountRejectsEmptyCode(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "lootsheet.db")

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, assets); err != nil {
		t.Fatalf("initialize sqlite database: %v", err)
	}

	err = DeleteAccount(context.Background(), databasePath, "")
	if err == nil {
		t.Fatal("expected delete with empty code to fail")
	}

	if !strings.Contains(err.Error(), "account code is required") {
		t.Fatalf("error = %q, want code-required error", err)
	}
}
