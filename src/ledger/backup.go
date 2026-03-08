package ledger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const backupTimestampLayout = "20060102-150405.000000000"

func createDatabaseBackup(databasePath string, backupDir string) (string, error) {
	if strings.TrimSpace(backupDir) == "" {
		return "", fmt.Errorf("backup directory path is required")
	}

	info, err := os.Stat(databasePath)
	if err != nil {
		return "", fmt.Errorf("stat database for backup: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("database path %q is a directory", databasePath)
	}

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	backupName := fmt.Sprintf(
		"%s.%s.bak",
		strings.TrimSuffix(filepath.Base(databasePath), filepath.Ext(databasePath)),
		time.Now().UTC().Format(backupTimestampLayout),
	)
	backupPath := filepath.Join(backupDir, backupName)

	source, err := os.Open(databasePath)
	if err != nil {
		return "", fmt.Errorf("open database for backup: %w", err)
	}
	defer source.Close()

	target, err := os.OpenFile(backupPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("create backup file: %w", err)
	}

	if _, err := io.Copy(target, source); err != nil {
		target.Close()
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("copy database backup: %w", err)
	}

	if err := target.Close(); err != nil {
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("close backup file: %w", err)
	}

	return backupPath, nil
}
