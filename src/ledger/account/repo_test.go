package account

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestListAccountsReturnsSeededAccounts(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	accounts, err := ListAccounts(context.Background(), databasePath, campaignID)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	if len(accounts) != 16 {
		t.Fatalf("account count = %d, want 16", len(accounts))
	}

	if accounts[0].Code != "1000" || accounts[0].Name != "Party Cash" {
		t.Fatalf("first account = %+v, want Party Cash at code 1000", accounts[0])
	}
}

func TestCreateAccountInsertsNewAccount(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	account, err := CreateAccount(context.Background(), databasePath, campaignID, "5600", "Tavern Reparations", ledger.AccountTypeExpense)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	if account.Code != "5600" || account.Name != "Tavern Reparations" || account.Type != ledger.AccountTypeExpense || !account.Active {
		t.Fatalf("created account = %+v, want code=5600 name=Tavern Reparations type=expense active=true", account)
	}

	if account.ID == "" {
		t.Fatal("created account ID is empty")
	}

	accounts, err := ListAccounts(context.Background(), databasePath, campaignID)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	if len(accounts) != 17 {
		t.Fatalf("account count = %d, want 17", len(accounts))
	}
}

func TestCreateAccountRejectsDuplicateCode(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	_, err := CreateAccount(context.Background(), databasePath, campaignID, "1000", "Duplicate Cash", ledger.AccountTypeAsset)
	if err == nil {
		t.Fatal("expected create account with duplicate code to fail")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %q, want duplicate code error", err)
	}
}

func TestCreateAccountRejectsInvalidType(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	_, err := CreateAccount(context.Background(), databasePath, campaignID, "9999", "Bad Type", ledger.AccountType("bogus"))
	if err == nil {
		t.Fatal("expected create account with invalid type to fail")
	}

	if !strings.Contains(err.Error(), "invalid account type") {
		t.Fatalf("error = %q, want invalid type error", err)
	}
}

func TestCreateAccountRejectsEmptyCodeAndName(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	_, err := CreateAccount(context.Background(), databasePath, campaignID, "", "No Code", ledger.AccountTypeAsset)
	if err == nil {
		t.Fatal("expected create account with empty code to fail")
	}

	_, err = CreateAccount(context.Background(), databasePath, campaignID, "9999", "", ledger.AccountTypeAsset)
	if err == nil {
		t.Fatal("expected create account with empty name to fail")
	}
}

func TestRenameAccountChangesName(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	if err := RenameAccount(context.Background(), databasePath, campaignID, "1000", "Gold Hoard"); err != nil {
		t.Fatalf("rename account: %v", err)
	}

	accounts, err := ListAccounts(context.Background(), databasePath, campaignID)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	for _, account := range accounts {
		if account.Code == "1000" {
			if account.Name != "Gold Hoard" {
				t.Fatalf("renamed account name = %q, want Gold Hoard", account.Name)
			}
			return
		}
	}

	t.Fatal("account code 1000 not found after rename")
}

func TestRenameAccountRejectsNonexistentCode(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	err := RenameAccount(context.Background(), databasePath, campaignID, "9999", "Ghost Account")
	if err == nil {
		t.Fatal("expected rename of nonexistent account to fail")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does-not-exist error", err)
	}
}

func TestDeactivateAndActivateAccount(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	if err := DeactivateAccount(context.Background(), databasePath, campaignID, "1000"); err != nil {
		t.Fatalf("deactivate account: %v", err)
	}

	accounts, err := ListAccounts(context.Background(), databasePath, campaignID)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	for _, account := range accounts {
		if account.Code == "1000" {
			if account.Active {
				t.Fatal("expected account 1000 to be inactive after deactivation")
			}
			break
		}
	}

	// Reactivate
	if err := ActivateAccount(context.Background(), databasePath, campaignID, "1000"); err != nil {
		t.Fatalf("activate account: %v", err)
	}

	accounts, err = ListAccounts(context.Background(), databasePath, campaignID)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	for _, account := range accounts {
		if account.Code == "1000" {
			if !account.Active {
				t.Fatal("expected account 1000 to be active after reactivation")
			}
			return
		}
	}

	t.Fatal("account code 1000 not found after reactivation")
}

func TestDeactivateAccountRejectsNonexistentCode(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	err := DeactivateAccount(context.Background(), databasePath, campaignID, "9999")
	if err == nil {
		t.Fatal("expected deactivate of nonexistent account to fail")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does-not-exist error", err)
	}
}

func TestDeleteAccountWithoutPostings(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	_, err := CreateAccount(context.Background(), databasePath, campaignID, "9900", "Disposable Account", ledger.AccountTypeExpense)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	if err := DeleteAccount(context.Background(), databasePath, campaignID, "9900"); err != nil {
		t.Fatalf("delete account: %v", err)
	}

	accounts, err := ListAccounts(context.Background(), databasePath, campaignID)
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}

	for _, account := range accounts {
		if account.Code == "9900" {
			t.Fatal("expected account 9900 to be deleted")
		}
	}
}

func TestDeleteAccountRejectsNonexistentCode(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	err := DeleteAccount(context.Background(), databasePath, campaignID, "9999")
	if err == nil {
		t.Fatal("expected delete of nonexistent account to fail")
	}

	if !errors.Is(err, ledger.ErrAccountNotFound) {
		t.Fatalf("error = %v, want ErrAccountNotFound", err)
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does-not-exist message", err)
	}
}

func TestDeleteAccountRejectsEmptyCode(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	err := DeleteAccount(context.Background(), databasePath, campaignID, "")
	if err == nil {
		t.Fatal("expected delete with empty code to fail")
	}

	if !strings.Contains(err.Error(), "account code is required") {
		t.Fatalf("error = %q, want code-required error", err)
	}
}

func TestAccountCodeImmutable(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)

	accountsBefore, err := ListAccounts(context.Background(), databasePath, campaignID)
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

	if err := RenameAccount(context.Background(), databasePath, campaignID, "1000", "Gold Hoard"); err != nil {
		t.Fatalf("rename account: %v", err)
	}

	accountsAfter, err := ListAccounts(context.Background(), databasePath, campaignID)
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
