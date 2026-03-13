package ledger_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestGetDatabaseStatusWithAssetsForeignWhenMetadataMissing(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "foreign.db")

	testutil.RunSQLiteScriptForTest(t, databasePath, `CREATE TABLE camp_log (id TEXT PRIMARY KEY, note TEXT NOT NULL);`)

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := ledger.GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.State != ledger.DatabaseStateForeign {
		t.Fatalf("state = %q, want %q", status.State, ledger.DatabaseStateForeign)
	}

	if !strings.Contains(status.Detail, "missing LootSheet migration metadata") {
		t.Fatalf("detail = %q, want missing metadata detail", status.Detail)
	}
}

func TestGetDatabaseStatusWithAssetsForeignWhenSchemaVersionUnknown(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	testutil.RunSQLiteScriptForTest(t, databasePath, `
		INSERT INTO schema_migrations (version, name) VALUES ('99', '099_future.sql');
		UPDATE settings SET value = '99' WHERE key = 'schema_version';
	`)

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := ledger.GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.State != ledger.DatabaseStateForeign {
		t.Fatalf("state = %q, want %q", status.State, ledger.DatabaseStateForeign)
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

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	status, err := ledger.GetDatabaseStatusWithAssets(context.Background(), databasePath, assets)
	if err != nil {
		t.Fatalf("get database status: %v", err)
	}

	if status.State != ledger.DatabaseStateDamaged {
		t.Fatalf("state = %q, want %q", status.State, ledger.DatabaseStateDamaged)
	}

	if !strings.Contains(status.Detail, "valid SQLite database") {
		t.Fatalf("detail = %q, want damaged detail", status.Detail)
	}
}

func TestMigrateSQLiteDatabaseCreatesBackupBeforeApplyingMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	databasePath := filepath.Join(tmpDir, "ledger.db")
	backupDir := filepath.Join(tmpDir, "backups")

	fullAssets, legacyAssets := testutil.LoadMigrationAssetsForTest(t)
	if _, err := ledger.EnsureSQLiteInitialized(context.Background(), databasePath, legacyAssets); err != nil {
		t.Fatalf("initialize legacy db: %v", err)
	}

	beforeBytes, err := os.ReadFile(databasePath)
	if err != nil {
		t.Fatalf("read database before migration: %v", err)
	}

	result, err := ledger.MigrateSQLiteDatabase(context.Background(), databasePath, backupDir, fullAssets)
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
