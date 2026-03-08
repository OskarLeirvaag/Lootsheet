package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
)

func TestRunAccountLedgerShowsTransactions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	// Post entry 1: Dr 5100:25, Cr 1000:25
	err := Run(context.Background(), []string{
		"journal", "post",
		"--date", "2026-03-08",
		"--description", "Restock arrows",
		"--debit", "5100:25:Quiver refill",
		"--credit", "1000:25",
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run journal post 1: %v", err)
	}

	// Post entry 2: Dr 1000:100, Cr 4000:100
	err = Run(context.Background(), []string{
		"journal", "post",
		"--date", "2026-03-08",
		"--description", "Quest reward earned",
		"--debit", "1000:100:Goblin bounty",
		"--credit", "4000:100",
	}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run journal post 2: %v", err)
	}

	var stdout bytes.Buffer
	err = Run(context.Background(), []string{
		"account", "ledger",
		"--code", "1000",
	}, &stdout)
	if err != nil {
		t.Fatalf("run account ledger: %v", err)
	}

	output := stdout.String()

	if !strings.Contains(output, "Account: 1000 Party Cash (asset)") {
		t.Fatalf("ledger output missing account header: %q", output)
	}

	if !strings.Contains(output, "DATE") || !strings.Contains(output, "DESCRIPTION") || !strings.Contains(output, "BALANCE") {
		t.Fatalf("ledger output missing column headers: %q", output)
	}

	if !strings.Contains(output, "Restock arrows") {
		t.Fatalf("ledger output missing entry 1: %q", output)
	}

	if !strings.Contains(output, "Quest reward earned") {
		t.Fatalf("ledger output missing entry 2: %q", output)
	}

	if !strings.Contains(output, "Balance: 75") {
		t.Fatalf("ledger output missing final balance: %q", output)
	}
}

func TestRunAccountLedgerEmptyAccount(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	var stdout bytes.Buffer
	err := Run(context.Background(), []string{
		"account", "ledger",
		"--code", "1000",
	}, &stdout)
	if err != nil {
		t.Fatalf("run account ledger: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Account: 1000 Party Cash (asset)") {
		t.Fatalf("ledger output missing account header: %q", output)
	}

	if !strings.Contains(output, "No transactions.") {
		t.Fatalf("ledger output missing empty message: %q", output)
	}
}

func TestRunAccountLedgerMissingCode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	err := Run(context.Background(), []string{
		"account", "ledger",
	}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing --code flag")
	}

	if !strings.Contains(err.Error(), "--code is required") {
		t.Fatalf("error = %q, want --code required error", err)
	}
}

func TestRunAccountLedgerNonexistentAccount(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv(config.EnvConfigPath, configPath)
	t.Setenv(config.EnvDataDir, dataDir)
	t.Setenv(config.EnvDatabasePath, "ledger.db")

	var initStdout bytes.Buffer
	if err := Run(context.Background(), []string{"init"}, &initStdout); err != nil {
		t.Fatalf("run init: %v", err)
	}

	err := Run(context.Background(), []string{
		"account", "ledger",
		"--code", "9999",
	}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for nonexistent account")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does-not-exist error", err)
	}
}
