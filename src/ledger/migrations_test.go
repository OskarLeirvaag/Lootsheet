package ledger

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
)

func TestGetDatabaseStatusWithAssetsForeignWhenMetadataMissing(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "foreign.db")

	RunSQLiteScriptForTest(t, databasePath, `CREATE TABLE camp_log (id TEXT PRIMARY KEY, note TEXT NOT NULL);`)

	assets, err := loadInitAssetsForLedgerTest()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.State != DatabaseStateForeign {
		t.Fatalf("state = %q, want %q", status.State, DatabaseStateForeign)
	}

	if !strings.Contains(status.Detail, "missing LootSheet migration metadata") {
		t.Fatalf("detail = %q, want missing metadata detail", status.Detail)
	}
}

func TestGetDatabaseStatusWithAssetsForeignWhenSchemaVersionUnknown(t *testing.T) {
	databasePath := InitTestDB(t)

	RunSQLiteScriptForTest(t, databasePath, `
		INSERT INTO schema_migrations (version, name) VALUES ('99', '099_future.sql');
		UPDATE settings SET value = '99' WHERE key = 'schema_version';
	`)

	assets, err := loadInitAssetsForLedgerTest()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.State != DatabaseStateForeign {
		t.Fatalf("state = %q, want %q", status.State, DatabaseStateForeign)
	}

	if !strings.Contains(status.Detail, `schema version "99" is not recognized`) {
		t.Fatalf("detail = %q, want unknown schema detail", status.Detail)
	}
}

func TestGetDatabaseStatusWithAssetsDamagedForNonSQLiteFile(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "damaged.db")

	if err := os.WriteFile(databasePath, []byte("definitely not sqlite"), 0o600); err != nil {
		t.Fatalf("write damaged db file: %v", err)
	}

	assets, err := loadInitAssetsForLedgerTest()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.State != DatabaseStateDamaged {
		t.Fatalf("state = %q, want %q", status.State, DatabaseStateDamaged)
	}

	if !strings.Contains(status.Detail, "valid SQLite database") {
		t.Fatalf("detail = %q, want damaged detail", status.Detail)
	}
}

func TestMigrateSQLiteDatabaseCreatesBackupBeforeApplyingMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "ledger.db")
	backupDir := filepath.Join(tmpDir, "backups")

	fullAssets, legacyAssets := LoadMigrationAssetsForTest(t)
	if _, err := EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
		t.Fatalf("initialize legacy db: %v", err)
	}

	beforeBytes, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatalf("read database before migration: %v", err)
	}

	result, err := MigrateSQLiteDatabase(context.Background(), databasePath, backupDir, fullAssets)
	if err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	if !result.Migrated {
		t.Fatal("expected migration to run")
	}

	if result.BackupPath == "" {
		t.Fatal("expected backup path to be recorded")
	}

	if filepath.Dir(result.BackupPath) != backupDir {
		t.Fatalf("backup path = %q, want inside %q", result.BackupPath, backupDir)
	}

	backupBytes, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("read backup file: %v", err)
	}

	if !bytes.Equal(backupBytes, beforeBytes) {
		t.Fatal("backup file contents do not match the pre-migration database")
	}
}

func loadInitAssetsForLedgerTest() (config.InitAssets, error) {
	return config.LoadInitAssets()
}
