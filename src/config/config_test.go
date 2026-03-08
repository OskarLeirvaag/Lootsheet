package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "data")
	backupDir := filepath.Join(tmpDir, "saves")
	exportDir := filepath.Join(tmpDir, "exports-out")

	t.Setenv(EnvConfigPath, configPath)
	t.Setenv(EnvDataDir, dataDir)
	t.Setenv(EnvDatabasePath, "party-ledger.db")
	t.Setenv(EnvBackupDir, backupDir)
	t.Setenv(EnvExportDir, exportDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Paths.ConfigFile != configPath {
		t.Fatalf("config file = %q, want %q", cfg.Paths.ConfigFile, configPath)
	}

	if cfg.Paths.DataDir != dataDir {
		t.Fatalf("data dir = %q, want %q", cfg.Paths.DataDir, dataDir)
	}

	wantDatabase := filepath.Join(dataDir, "party-ledger.db")
	if cfg.Paths.DatabasePath != wantDatabase {
		t.Fatalf("database path = %q, want %q", cfg.Paths.DatabasePath, wantDatabase)
	}

	if cfg.Paths.BackupDir != backupDir {
		t.Fatalf("backup dir = %q, want %q", cfg.Paths.BackupDir, backupDir)
	}

	if cfg.Paths.ExportDir != exportDir {
		t.Fatalf("export dir = %q, want %q", cfg.Paths.ExportDir, exportDir)
	}
}

func TestLoadMergesConfigFilePaths(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	configJSON := []byte(`{"paths":{"data_dir":"../party-data","database_path":"books/ledger.db","backup_dir":"../vault/backups","export_dir":"exports"}}`)
	if err := os.WriteFile(configPath, configJSON, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	t.Setenv(EnvConfigPath, configPath)
	t.Setenv(EnvDataDir, "")
	t.Setenv(EnvDatabasePath, "")
	t.Setenv(EnvBackupDir, "")
	t.Setenv(EnvExportDir, "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	wantDataDir := filepath.Join(tmpDir, "party-data")
	if cfg.Paths.DataDir != wantDataDir {
		t.Fatalf("data dir = %q, want %q", cfg.Paths.DataDir, wantDataDir)
	}

	wantDatabase := filepath.Join(wantDataDir, "books", "ledger.db")
	if cfg.Paths.DatabasePath != wantDatabase {
		t.Fatalf("database path = %q, want %q", cfg.Paths.DatabasePath, wantDatabase)
	}

	wantBackupDir := filepath.Join(tmpDir, "vault", "backups")
	if cfg.Paths.BackupDir != wantBackupDir {
		t.Fatalf("backup dir = %q, want %q", cfg.Paths.BackupDir, wantBackupDir)
	}

	wantExportDir := filepath.Join(wantDataDir, "exports")
	if cfg.Paths.ExportDir != wantExportDir {
		t.Fatalf("export dir = %q, want %q", cfg.Paths.ExportDir, wantExportDir)
	}
}

func TestLoadDerivesBackupAndExportDefaultsFromDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config", "config.json")
	dataDir := filepath.Join(tmpDir, "party-data")

	t.Setenv(EnvConfigPath, configPath)
	t.Setenv(EnvDataDir, dataDir)
	t.Setenv(EnvDatabasePath, "")
	t.Setenv(EnvBackupDir, "")
	t.Setenv(EnvExportDir, "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Paths.BackupDir != filepath.Join(dataDir, DefaultBackupDirName) {
		t.Fatalf("backup dir = %q, want default under data dir", cfg.Paths.BackupDir)
	}

	if cfg.Paths.ExportDir != filepath.Join(dataDir, DefaultExportDirName) {
		t.Fatalf("export dir = %q, want default under data dir", cfg.Paths.ExportDir)
	}
}
