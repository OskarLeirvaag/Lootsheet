package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

const sqliteCommand = "sqlite3"

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

func EnsureSQLiteInitialized(ctx context.Context, databasePath string, assets config.InitAssets) (InitResult, error) {
	if err := ensureSQLiteAvailable(); err != nil {
		return InitResult{}, err
	}

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

	if err := runSQLiteScript(ctx, databasePath, buildInitScript(assets)); err != nil {
		return InitResult{}, err
	}

	return InitResult{
		Initialized: true,
		SeededCounts: SeededCounts{
			Accounts: len(assets.Accounts),
		},
	}, nil
}

func GetDatabaseStatus(ctx context.Context, databasePath string) (DatabaseStatus, error) {
	if err := ensureSQLiteAvailable(); err != nil {
		return DatabaseStatus{}, err
	}

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
	if err := ensureSQLiteAvailable(); err != nil {
		return nil, err
	}

	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return nil, err
	}

	output, err := runSQLiteQuery(
		ctx,
		databasePath,
		"SELECT id, code, name, type, active FROM accounts ORDER BY code, id;",
		"-separator",
		"\t",
	)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []AccountRecord{}, nil
	}

	accounts := make([]AccountRecord, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			return nil, fmt.Errorf("parse account row: expected 5 columns, got %d", len(fields))
		}

		accountType := service.AccountType(fields[3])
		if !accountType.Valid() {
			return nil, fmt.Errorf("parse account row: invalid account type %q", fields[3])
		}

		active, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, fmt.Errorf("parse account row active flag: %w", err)
		}

		accounts = append(accounts, AccountRecord{
			ID:     fields[0],
			Code:   fields[1],
			Name:   fields[2],
			Type:   accountType,
			Active: active == 1,
		})
	}

	return accounts, nil
}

func PostJournalEntry(ctx context.Context, databasePath string, input service.JournalPostInput) (PostedJournalEntry, error) {
	if err := ensureSQLiteAvailable(); err != nil {
		return PostedJournalEntry{}, err
	}

	if err := ensureInitializedDatabase(ctx, databasePath); err != nil {
		return PostedJournalEntry{}, err
	}

	validated, err := service.ValidateJournalPostInput(input)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	accountIDsByCode, err := resolveActiveAccountIDsByCode(ctx, databasePath, validated.Lines)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	entryNumber, err := nextJournalEntryNumber(ctx, databasePath)
	if err != nil {
		return PostedJournalEntry{}, err
	}

	entryID := uuid.NewString()
	if err := runSQLiteScript(ctx, databasePath, buildPostJournalEntryScript(entryID, entryNumber, validated, accountIDsByCode)); err != nil {
		return PostedJournalEntry{}, err
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

	userTableCountOutput, err := runSQLiteQuery(
		ctx,
		databasePath,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%';",
	)
	if err != nil {
		return databaseState{}, err
	}

	userTableCount, err := strconv.Atoi(strings.TrimSpace(userTableCountOutput))
	if err != nil {
		return databaseState{}, fmt.Errorf("parse user table count: %w", err)
	}

	appliedMigrations, err := loadAppliedMigrations(ctx, databasePath)
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

	schemaVersion, err := runSQLiteQuery(
		ctx,
		databasePath,
		"SELECT value FROM settings WHERE key = 'schema_version';",
	)
	if err != nil {
		if strings.Contains(err.Error(), "no such table: settings") {
			return databaseState{Exists: true, UserTableCount: userTableCount}, nil
		}
		return databaseState{}, err
	}

	return databaseState{
		Exists:             true,
		UserTableCount:     userTableCount,
		SchemaVersion:      strings.TrimSpace(schemaVersion),
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

func loadAppliedMigrations(ctx context.Context, databasePath string) ([]AppliedMigration, error) {
	output, err := runSQLiteQuery(
		ctx,
		databasePath,
		"SELECT version, name, applied_at FROM schema_migrations ORDER BY CAST(version AS INTEGER), name;",
		"-separator",
		"\t",
	)
	if err != nil {
		return nil, err
	}

	trimmedOutput := strings.TrimSpace(output)
	if trimmedOutput == "" {
		return []AppliedMigration{}, nil
	}

	migrations := make([]AppliedMigration, 0)
	for _, line := range strings.Split(trimmedOutput, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			return nil, fmt.Errorf("parse applied migration row: expected 3 columns, got %d", len(fields))
		}

		migrations = append(migrations, AppliedMigration{
			Version:   fields[0],
			Name:      fields[1],
			AppliedAt: fields[2],
		})
	}

	return migrations, nil
}

func resolveActiveAccountIDsByCode(
	ctx context.Context,
	databasePath string,
	lines []service.JournalLineInput,
) (map[string]string, error) {
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

	query := "SELECT code, id, active FROM accounts WHERE code IN (" + sqlStringList(accountCodes) + ");"
	output, err := runSQLiteQuery(ctx, databasePath, query, "-separator", "\t")
	if err != nil {
		return nil, err
	}

	records := map[string]accountLookupRecord{}
	trimmedOutput := strings.TrimSpace(output)
	if trimmedOutput != "" {
		for _, line := range strings.Split(trimmedOutput, "\n") {
			fields := strings.Split(line, "\t")
			if len(fields) != 3 {
				return nil, fmt.Errorf("parse account lookup row: expected 3 columns, got %d", len(fields))
			}

			active, err := strconv.Atoi(fields[2])
			if err != nil {
				return nil, fmt.Errorf("parse account lookup active flag: %w", err)
			}

			records[fields[0]] = accountLookupRecord{
				ID:     fields[1],
				Active: active == 1,
			}
		}
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

func nextJournalEntryNumber(ctx context.Context, databasePath string) (int, error) {
	output, err := runSQLiteQuery(
		ctx,
		databasePath,
		"SELECT COALESCE(MAX(entry_number), 0) + 1 FROM journal_entries;",
	)
	if err != nil {
		return 0, err
	}

	entryNumber, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("parse next journal entry number: %w", err)
	}

	return entryNumber, nil
}

func ensureSQLiteAvailable() error {
	if _, err := exec.LookPath(sqliteCommand); err != nil {
		return fmt.Errorf("sqlite3 command not available: %w", err)
	}

	return nil
}

func runSQLiteQuery(ctx context.Context, databasePath string, sql string, extraArgs ...string) (string, error) {
	args := []string{"-batch", "-noheader"}
	args = append(args, extraArgs...)
	args = append(args, databasePath, sql)

	command := exec.CommandContext(ctx, sqliteCommand, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("run sqlite query: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return string(output), nil
}

func runSQLiteScript(ctx context.Context, databasePath string, sql string) error {
	command := exec.CommandContext(ctx, sqliteCommand, "-batch", "-bail", databasePath)
	command.Stdin = strings.NewReader(sql)

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run sqlite script: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func buildInitScript(assets config.InitAssets) string {
	statements := make([]string, 0, len(assets.Migrations)*2+len(assets.Accounts)+2)

	for _, migration := range assets.Migrations {
		statements = append(statements, migration.SQL)
	}

	for _, migration := range assets.Migrations {
		statements = append(statements, buildInsertStatement(
			"schema_migrations",
			[]string{"version", "name"},
			[]string{sqlString(migration.Version), sqlString(migration.Name)},
		))
	}

	statements = append(statements,
		buildInsertStatement("settings", []string{"key", "value"}, []string{sqlString("schema_version"), sqlString(assets.SchemaVersion)}),
		buildInsertStatement("settings", []string{"key", "value"}, []string{sqlString("initialized_at"), sqlCurrentTimestamp}),
	)

	for _, account := range assets.Accounts {
		statements = append(statements, buildInsertStatement(
			"accounts",
			[]string{"id", "code", "name", "type", "active"},
			[]string{
				sqlString(account.ID),
				sqlString(account.Code),
				sqlString(account.Name),
				sqlString(account.Type),
				sqlBool(account.Active),
			},
		))
	}

	return buildTransactionScript(statements...)
}

func buildPostJournalEntryScript(
	entryID string,
	entryNumber int,
	input service.ValidatedJournalPost,
	accountIDsByCode map[string]string,
) string {
	statements := make([]string, 0, len(input.Lines)+1)

	statements = append(statements, buildInsertStatement(
		"journal_entries",
		[]string{"id", "entry_number", "status", "entry_date", "description", "posted_at"},
		[]string{
			sqlString(entryID),
			sqlInt(entryNumber),
			sqlString("posted"),
			sqlString(input.EntryDate),
			sqlString(input.Description),
			sqlCurrentTimestamp,
		},
	))

	for index, line := range input.Lines {
		statements = append(statements, buildInsertStatement(
			"journal_lines",
			[]string{"id", "journal_entry_id", "line_number", "account_id", "memo", "debit_amount", "credit_amount"},
			[]string{
				sqlString(uuid.NewString()),
				sqlString(entryID),
				sqlInt(index + 1),
				sqlString(accountIDsByCode[line.AccountCode]),
				sqlString(line.Memo),
				sqlInt64(line.DebitAmount),
				sqlInt64(line.CreditAmount),
			},
		))
	}

	return buildTransactionScript(statements...)
}
