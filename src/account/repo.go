package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// ListAccounts returns all accounts ordered by code.
func ListAccounts(ctx context.Context, databasePath string) ([]ledger.AccountRecord, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return nil, err
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SELECT id, code, name, type, active FROM accounts ORDER BY code, id")
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()

	accounts := []ledger.AccountRecord{}
	for rows.Next() {
		var r ledger.AccountRecord
		var accountType string
		var active int

		if err := rows.Scan(&r.ID, &r.Code, &r.Name, &accountType, &active); err != nil {
			return nil, fmt.Errorf("scan account row: %w", err)
		}

		r.Type = ledger.AccountType(accountType)
		if !r.Type.Valid() {
			return nil, fmt.Errorf("scan account row: invalid account type %q", accountType)
		}

		r.Active = active == 1
		accounts = append(accounts, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate account rows: %w", err)
	}

	return accounts, nil
}

// CreateAccount inserts a new account with a generated UUID.
// Code must be unique. The account defaults to active=true.
func CreateAccount(ctx context.Context, databasePath string, code string, name string, accountType ledger.AccountType) (ledger.AccountRecord, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return ledger.AccountRecord{}, err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return ledger.AccountRecord{}, fmt.Errorf("account code is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return ledger.AccountRecord{}, fmt.Errorf("account name is required")
	}

	if !accountType.Valid() {
		return ledger.AccountRecord{}, fmt.Errorf("invalid account type %q", accountType)
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return ledger.AccountRecord{}, err
	}
	defer db.Close()

	id := uuid.NewString()

	if _, err := db.ExecContext(ctx,
		"INSERT INTO accounts (id, code, name, type, active) VALUES (?, ?, ?, ?, 1)",
		id, code, name, string(accountType),
	); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ledger.AccountRecord{}, fmt.Errorf("account code %q already exists", code)
		}
		return ledger.AccountRecord{}, fmt.Errorf("insert account: %w", err)
	}

	return ledger.AccountRecord{
		ID:     id,
		Code:   code,
		Name:   name,
		Type:   accountType,
		Active: true,
	}, nil
}

// RenameAccount updates the name of an existing account identified by code.
// Account IDs are immutable; only the name changes.
func RenameAccount(ctx context.Context, databasePath string, code string, newName string) error {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("account code is required")
	}

	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("account name is required")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.ExecContext(ctx,
		"UPDATE accounts SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE code = ?",
		newName, code,
	)
	if err != nil {
		return fmt.Errorf("rename account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rename result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account code %q does not exist", code)
	}

	return nil
}

// DeactivateAccount marks an account as inactive.
// Inactive accounts cannot be used in new journal entries.
func DeactivateAccount(ctx context.Context, databasePath string, code string) error {
	return setAccountActive(ctx, databasePath, code, false)
}

// ActivateAccount marks an account as active.
func ActivateAccount(ctx context.Context, databasePath string, code string) error {
	return setAccountActive(ctx, databasePath, code, true)
}

func setAccountActive(ctx context.Context, databasePath string, code string, active bool) error {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("account code is required")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	activeVal := 0
	if active {
		activeVal = 1
	}

	result, err := db.ExecContext(ctx,
		"UPDATE accounts SET active = ?, updated_at = CURRENT_TIMESTAMP WHERE code = ?",
		activeVal, code,
	)
	if err != nil {
		return fmt.Errorf("update account active status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account code %q does not exist", code)
	}

	return nil
}

// DeleteAccount deletes an account identified by code, but only if no
// journal_lines reference it. Returns ErrAccountHasPostings if the account
// has any postings, and ErrAccountNotFound if the code does not exist.
func DeleteAccount(ctx context.Context, databasePath string, code string) error {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("account code is required")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Look up the account by code.
	var accountID string
	if err := db.QueryRowContext(ctx,
		"SELECT id FROM accounts WHERE code = ?", code,
	).Scan(&accountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("account code %q does not exist: %w", code, ledger.ErrAccountNotFound)
		}
		return fmt.Errorf("query account by code: %w", err)
	}

	// Check if any journal lines reference this account.
	var postingCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM journal_lines WHERE account_id = ?", accountID,
	).Scan(&postingCount); err != nil {
		return fmt.Errorf("count account postings: %w", err)
	}

	if postingCount > 0 {
		return ledger.ErrAccountHasPostings
	}

	// No postings — safe to delete.
	result, err := db.ExecContext(ctx, "DELETE FROM accounts WHERE id = ?", accountID)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check delete result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account code %q was not deleted", code)
	}

	return nil
}
