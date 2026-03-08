package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

type InitResult struct {
	Initialized  bool
	SeededCounts SeededCounts
}

type SeededCounts struct {
	Accounts int
}

type AccountRecord struct {
	ID     string
	Code   string
	Name   string
	Type   service.AccountType
	Active bool
}

type PostedJournalEntry struct {
	ID          string
	EntryNumber int
	EntryDate   string
	Description string
	LineCount   int
	DebitTotal  int64
	CreditTotal int64
}

type DatabaseStatus struct {
	Exists              bool
	Initialized         bool
	State               DatabaseLifecycleState
	UserTableCount      int
	SchemaVersion       string
	TargetSchemaVersion string
	AppliedMigrations   []AppliedMigration
	PendingMigrations   []PendingMigration
}

type AppliedMigration struct {
	Version   string
	Name      string
	AppliedAt string
}

type PendingMigration struct {
	Version string
	Name    string
}

type MigrationResult struct {
	Migrated          bool
	MetadataRepaired  bool
	FromSchemaVersion string
	ToSchemaVersion   string
	AppliedMigrations []PendingMigration
}

type DatabaseLifecycleState string

const (
	DatabaseStateUninitialized DatabaseLifecycleState = "uninitialized"
	DatabaseStateCurrent       DatabaseLifecycleState = "current"
	DatabaseStateUpgradeable   DatabaseLifecycleState = "upgradeable"
	DatabaseStateUnknown       DatabaseLifecycleState = "unknown"
)

func openDB(databasePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("set database pragma: %w", err)
		}
	}

	return db, nil
}

func EnsureSQLiteInitialized(ctx context.Context, databasePath string, assets config.InitAssets) (InitResult, error) {
	state, err := inspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return InitResult{}, err
	}

	switch {
	case state.SchemaVersion != "":
		return InitResult{}, nil
	case state.UserTableCount > 0:
		return InitResult{}, fmt.Errorf("database %q already has tables but is missing LootSheet init metadata", databasePath)
	}

	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return InitResult{}, fmt.Errorf("create database directory: %w", err)
	}

	db, err := openDB(databasePath)
	if err != nil {
		return InitResult{}, err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return InitResult{}, fmt.Errorf("begin init transaction: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range assets.Migrations {
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			return InitResult{}, fmt.Errorf("execute init migration %s: %w", migration.Name, err)
		}
	}

	for _, migration := range assets.Migrations {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version, migration.Name,
		); err != nil {
			return InitResult{}, fmt.Errorf("record init migration %s: %w", migration.Name, err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO settings (key, value) VALUES (?, ?)",
		"schema_version", assets.SchemaVersion,
	); err != nil {
		return InitResult{}, fmt.Errorf("record schema version: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO settings (key, value) VALUES (?, CURRENT_TIMESTAMP)",
		"initialized_at",
	); err != nil {
		return InitResult{}, fmt.Errorf("record initialization timestamp: %w", err)
	}

	for _, account := range assets.Accounts {
		active := 0
		if account.Active {
			active = 1
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO accounts (id, code, name, type, active) VALUES (?, ?, ?, ?, ?)",
			account.ID, account.Code, account.Name, account.Type, active,
		); err != nil {
			return InitResult{}, fmt.Errorf("seed account %s: %w", account.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return InitResult{}, fmt.Errorf("commit init transaction: %w", err)
	}

	return InitResult{
		Initialized: true,
		SeededCounts: SeededCounts{
			Accounts: len(assets.Accounts),
		},
	}, nil
}

func GetDatabaseStatus(ctx context.Context, databasePath string) (DatabaseStatus, error) {
	state, err := inspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return DatabaseStatus{}, err
	}

	return DatabaseStatus{
		Exists:            state.Exists,
		Initialized:       state.SchemaVersion != "",
		UserTableCount:    state.UserTableCount,
		SchemaVersion:     state.SchemaVersion,
		AppliedMigrations: state.AppliedMigrations,
	}, nil
}

func ListAccounts(ctx context.Context, databasePath string) ([]AccountRecord, error) {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return nil, err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, "SELECT id, code, name, type, active FROM accounts ORDER BY code, id")
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()

	accounts := []AccountRecord{}
	for rows.Next() {
		var r AccountRecord
		var accountType string
		var active int

		if err := rows.Scan(&r.ID, &r.Code, &r.Name, &accountType, &active); err != nil {
			return nil, fmt.Errorf("scan account row: %w", err)
		}

		r.Type = service.AccountType(accountType)
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

func PostJournalEntry(ctx context.Context, databasePath string, input service.JournalPostInput) (PostedJournalEntry, error) {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return PostedJournalEntry{}, err
	}

	validated, err := service.ValidateJournalPostInput(input)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return PostedJournalEntry{}, err
	}
	defer db.Close()

	accountIDsByCode, err := resolveActiveAccountIDsByCode(ctx, db, validated.Lines)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	entryNumber, err := nextJournalEntryNumber(ctx, db)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	entryID := uuid.NewString()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return PostedJournalEntry{}, fmt.Errorf("begin journal post transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries (id, entry_number, status, entry_date, description, posted_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		entryID, entryNumber, "posted", validated.EntryDate, validated.Description,
	); err != nil {
		return PostedJournalEntry{}, fmt.Errorf("insert journal entry: %w", err)
	}

	for index, line := range validated.Lines {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
			uuid.NewString(), entryID, index+1, accountIDsByCode[line.AccountCode], line.Memo, line.DebitAmount, line.CreditAmount,
		); err != nil {
			return PostedJournalEntry{}, fmt.Errorf("insert journal line %d: %w", index+1, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return PostedJournalEntry{}, fmt.Errorf("commit journal post transaction: %w", err)
	}

	return PostedJournalEntry{
		ID:          entryID,
		EntryNumber: entryNumber,
		EntryDate:   validated.EntryDate,
		Description: validated.Description,
		LineCount:   len(validated.Lines),
		DebitTotal:  validated.Totals.DebitAmount,
		CreditTotal: validated.Totals.CreditAmount,
	}, nil
}

type databaseState struct {
	Exists             bool
	UserTableCount     int
	SchemaVersion      string
	AppliedMigrations  []AppliedMigration
	UsesLegacyMetadata bool
}

type accountLookupRecord struct {
	ID     string
	Active bool
}

func inspectSQLiteDatabase(ctx context.Context, databasePath string) (databaseState, error) {
	info, err := os.Stat(databasePath)
	if errors.Is(err, os.ErrNotExist) {
		return databaseState{}, nil
	}
	if err != nil {
		return databaseState{}, fmt.Errorf("inspect database path: %w", err)
	}
	if info.IsDir() {
		return databaseState{}, fmt.Errorf("database path %q is a directory", databasePath)
	}

	db, err := openDB(databasePath)
	if err != nil {
		return databaseState{}, err
	}
	defer db.Close()

	var userTableCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'",
	).Scan(&userTableCount); err != nil {
		return databaseState{}, fmt.Errorf("count user tables: %w", err)
	}

	appliedMigrations, err := loadAppliedMigrations(ctx, db)
	if err == nil {
		schemaVersion := ""
		if len(appliedMigrations) > 0 {
			schemaVersion = appliedMigrations[len(appliedMigrations)-1].Version
		}

		return databaseState{
			Exists:            true,
			UserTableCount:    userTableCount,
			SchemaVersion:     schemaVersion,
			AppliedMigrations: appliedMigrations,
		}, nil
	}

	if !strings.Contains(err.Error(), "no such table: schema_migrations") {
		return databaseState{}, err
	}

	var schemaVersion string
	queryErr := db.QueryRowContext(ctx,
		"SELECT value FROM settings WHERE key = 'schema_version'",
	).Scan(&schemaVersion)
	if queryErr != nil {
		if strings.Contains(queryErr.Error(), "no such table: settings") {
			return databaseState{Exists: true, UserTableCount: userTableCount}, nil
		}
		return databaseState{}, queryErr
	}

	return databaseState{
		Exists:             true,
		UserTableCount:     userTableCount,
		SchemaVersion:      schemaVersion,
		UsesLegacyMetadata: true,
	}, nil
}

func ensureInitializedDatabase(ctx context.Context, databasePath string) error {
	state, err := inspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return err
	}

	if state.SchemaVersion == "" {
		return fmt.Errorf("database %q is not initialized; run `lootsheet init`", databasePath)
	}

	return nil
}

func loadAppliedMigrations(ctx context.Context, db *sql.DB) ([]AppliedMigration, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT version, name, applied_at FROM schema_migrations ORDER BY CAST(version AS INTEGER), name",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := []AppliedMigration{}
	for rows.Next() {
		var m AppliedMigration
		if err := rows.Scan(&m.Version, &m.Name, &m.AppliedAt); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		migrations = append(migrations, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return migrations, nil
}

func resolveActiveAccountIDsByCode(ctx context.Context, db *sql.DB, lines []service.JournalLineInput) (map[string]string, error) {
	accountCodes := make([]string, 0, len(lines))
	seenCodes := make(map[string]struct{}, len(lines))

	for _, line := range lines {
		if _, seen := seenCodes[line.AccountCode]; seen {
			continue
		}

		seenCodes[line.AccountCode] = struct{}{}
		accountCodes = append(accountCodes, line.AccountCode)
	}

	slices.Sort(accountCodes)

	placeholders := make([]string, len(accountCodes))
	args := make([]any, len(accountCodes))
	for i, code := range accountCodes {
		placeholders[i] = "?"
		args[i] = code
	}

	query := "SELECT code, id, active FROM accounts WHERE code IN (" + strings.Join(placeholders, ", ") + ")"
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query account codes: %w", err)
	}
	defer rows.Close()

	records := map[string]accountLookupRecord{}
	for rows.Next() {
		var code string
		var r accountLookupRecord
		var active int

		if err := rows.Scan(&code, &r.ID, &active); err != nil {
			return nil, fmt.Errorf("scan account lookup row: %w", err)
		}

		r.Active = active == 1
		records[code] = r
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate account lookup rows: %w", err)
	}

	resolved := make(map[string]string, len(accountCodes))
	for _, code := range accountCodes {
		record, ok := records[code]
		if !ok {
			return nil, fmt.Errorf("account code %q does not exist", code)
		}

		if !record.Active {
			return nil, fmt.Errorf("account code %q is inactive", code)
		}

		resolved[code] = record.ID
	}

	return resolved, nil
}

// ErrImmutableEntry is returned when an operation attempts to modify or delete
// a journal entry that has been posted or reversed. Corrections must use
// reversal, adjustment, or reclassification.
var ErrImmutableEntry = fmt.Errorf("posted or reversed journal entries are immutable; use reversal or adjustment to correct")

// getJournalEntryStatus returns the status of a journal entry by ID.
// Returns sql.ErrNoRows if the entry does not exist.
func getJournalEntryStatus(ctx context.Context, db *sql.DB, entryID string) (service.JournalEntryStatus, error) {
	var status string
	if err := db.QueryRowContext(ctx,
		"SELECT status FROM journal_entries WHERE id = ?", entryID,
	).Scan(&status); err != nil {
		return "", fmt.Errorf("query journal entry status: %w", err)
	}

	s := service.JournalEntryStatus(status)
	if !s.Valid() {
		return "", fmt.Errorf("journal entry %s has invalid status %q", entryID, status)
	}

	return s, nil
}

// CheckJournalEntryMutable verifies that a journal entry may be modified.
// Returns ErrImmutableEntry if the entry is posted or reversed.
func CheckJournalEntryMutable(ctx context.Context, databasePath string, entryID string) error {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	status, err := getJournalEntryStatus(ctx, db, entryID)
	if err != nil {
		return err
	}

	if status.Immutable() {
		return ErrImmutableEntry
	}

	return nil
}

// UpdateJournalEntry updates the description and/or entry_date of a journal entry.
// Returns ErrImmutableEntry if the entry is posted or reversed.
func UpdateJournalEntry(ctx context.Context, databasePath string, entryID string, description string, entryDate string) error {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	status, err := getJournalEntryStatus(ctx, db, entryID)
	if err != nil {
		return err
	}

	if status.Immutable() {
		return ErrImmutableEntry
	}

	if _, err := db.ExecContext(ctx,
		"UPDATE journal_entries SET description = ?, entry_date = ? WHERE id = ?",
		description, entryDate, entryID,
	); err != nil {
		return fmt.Errorf("update journal entry: %w", err)
	}

	return nil
}

// DeleteJournalEntry deletes a journal entry and its lines.
// Returns ErrImmutableEntry if the entry is posted or reversed.
func DeleteJournalEntry(ctx context.Context, databasePath string, entryID string) error {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	status, err := getJournalEntryStatus(ctx, db, entryID)
	if err != nil {
		return err
	}

	if status.Immutable() {
		return ErrImmutableEntry
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM journal_lines WHERE journal_entry_id = ?", entryID); err != nil {
		return fmt.Errorf("delete journal lines: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM journal_entries WHERE id = ?", entryID); err != nil {
		return fmt.Errorf("delete journal entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete transaction: %w", err)
	}

	return nil
}

// DeleteJournalLine deletes a single journal line.
// Returns ErrImmutableEntry if the parent entry is posted or reversed.
func DeleteJournalLine(ctx context.Context, databasePath string, lineID string) error {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	var entryID string
	if err := db.QueryRowContext(ctx,
		"SELECT journal_entry_id FROM journal_lines WHERE id = ?", lineID,
	).Scan(&entryID); err != nil {
		return fmt.Errorf("query journal line parent entry: %w", err)
	}

	status, err := getJournalEntryStatus(ctx, db, entryID)
	if err != nil {
		return err
	}

	if status.Immutable() {
		return ErrImmutableEntry
	}

	if _, err := db.ExecContext(ctx, "DELETE FROM journal_lines WHERE id = ?", lineID); err != nil {
		return fmt.Errorf("delete journal line: %w", err)
	}

	return nil
}

// UpdateJournalLine updates the amounts or memo of a single journal line.
// Returns ErrImmutableEntry if the parent entry is posted or reversed.
func UpdateJournalLine(ctx context.Context, databasePath string, lineID string, memo string, debitAmount int64, creditAmount int64) error {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	db, err := openDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	var entryID string
	if err := db.QueryRowContext(ctx,
		"SELECT journal_entry_id FROM journal_lines WHERE id = ?", lineID,
	).Scan(&entryID); err != nil {
		return fmt.Errorf("query journal line parent entry: %w", err)
	}

	status, err := getJournalEntryStatus(ctx, db, entryID)
	if err != nil {
		return err
	}

	if status.Immutable() {
		return ErrImmutableEntry
	}

	if _, err := db.ExecContext(ctx,
		"UPDATE journal_lines SET memo = ?, debit_amount = ?, credit_amount = ? WHERE id = ?",
		memo, debitAmount, creditAmount, lineID,
	); err != nil {
		return fmt.Errorf("update journal line: %w", err)
	}

	return nil
}

// CreateAccount inserts a new account with a generated UUID.
// Code must be unique. The account defaults to active=true.
func CreateAccount(ctx context.Context, databasePath string, code string, name string, accountType service.AccountType) (AccountRecord, error) {
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return AccountRecord{}, err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return AccountRecord{}, fmt.Errorf("account code is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return AccountRecord{}, fmt.Errorf("account name is required")
	}

	if !accountType.Valid() {
		return AccountRecord{}, fmt.Errorf("invalid account type %q", accountType)
	}

	db, err := openDB(databasePath)
	if err != nil {
		return AccountRecord{}, err
	}
	defer db.Close()

	id := uuid.NewString()

	if _, err := db.ExecContext(ctx,
		"INSERT INTO accounts (id, code, name, type, active) VALUES (?, ?, ?, ?, 1)",
		id, code, name, string(accountType),
	); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return AccountRecord{}, fmt.Errorf("account code %q already exists", code)
		}
		return AccountRecord{}, fmt.Errorf("insert account: %w", err)
	}

	return AccountRecord{
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
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
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

	db, err := openDB(databasePath)
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
	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("account code is required")
	}

	db, err := openDB(databasePath)
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

func nextJournalEntryNumber(ctx context.Context, db *sql.DB) (int, error) {
	var entryNumber int

	if err := db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(entry_number), 0) + 1 FROM journal_entries",
	).Scan(&entryNumber); err != nil {
		return 0, fmt.Errorf("query next journal entry number: %w", err)
	}

	return entryNumber, nil
}
