package repo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

func ListAccounts(ctx context.Context, databasePath string) ([]AccountRecord, error) {
	if err := ensureSQLiteAvailable(); err != nil {
		return nil, err
	}

	state, err := inspectSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return nil, err
	}

	if state.SchemaVersion == "" {
		return nil, fmt.Errorf("database %q is not initialized; run `lootsheet init`", databasePath)
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

type databaseState struct {
	UserTableCount int
	SchemaVersion  string
}

func inspectSQLiteDatabase(ctx context.Context, databasePath string) (databaseState, error) {
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

	schemaVersion, err := runSQLiteQuery(
		ctx,
		databasePath,
		"SELECT value FROM settings WHERE key = 'schema_version';",
	)
	if err != nil {
		if strings.Contains(err.Error(), "no such table: settings") {
			return databaseState{UserTableCount: userTableCount}, nil
		}
		return databaseState{}, err
	}

	return databaseState{
		UserTableCount: userTableCount,
		SchemaVersion:  strings.TrimSpace(schemaVersion),
	}, nil
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
	command := exec.CommandContext(ctx, sqliteCommand, databasePath)
	command.Stdin = strings.NewReader(sql)

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run sqlite script: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func buildInitScript(assets config.InitAssets) string {
	var builder strings.Builder

	builder.WriteString("BEGIN;\n")
	builder.WriteString(assets.SchemaSQL)
	builder.WriteString("\n")
	builder.WriteString("INSERT INTO settings (key, value) VALUES ('schema_version', '1');\n")
	builder.WriteString("INSERT INTO settings (key, value) VALUES ('initialized_at', CURRENT_TIMESTAMP);\n")

	for _, account := range assets.Accounts {
		activeValue := "0"
		if account.Active {
			activeValue = "1"
		}

		builder.WriteString("INSERT INTO accounts (id, code, name, type, active) VALUES (")
		builder.WriteString(quoteSQLString(account.ID))
		builder.WriteString(", ")
		builder.WriteString(quoteSQLString(account.Code))
		builder.WriteString(", ")
		builder.WriteString(quoteSQLString(account.Name))
		builder.WriteString(", ")
		builder.WriteString(quoteSQLString(account.Type))
		builder.WriteString(", ")
		builder.WriteString(activeValue)
		builder.WriteString(");\n")
	}

	builder.WriteString("COMMIT;\n")

	return builder.String()
}

func quoteSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
