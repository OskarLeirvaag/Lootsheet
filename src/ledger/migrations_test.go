package ledger

import (
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

func loadInitAssetsForLedgerTest() (config.InitAssets, error) {
	return config.LoadInitAssets()
}
