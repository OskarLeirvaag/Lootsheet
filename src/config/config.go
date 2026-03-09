// Package config handles configuration loading, resolution, and validation
// for the LootSheet CLI. Configuration is resolved with the following
// precedence: defaults, then config file, then environment variables.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// AppName is the application identifier used for directory naming.
	AppName = "lootsheet"
	// DefaultDatabase is the default SQLite database filename.
	DefaultDatabase = "lootsheet.db"
	// DefaultBackupDirName is the default backup directory inside the data directory.
	DefaultBackupDirName = "backups"
	// DefaultExportDirName is the default export directory inside the data directory.
	DefaultExportDirName = "exports"
	// EnvConfigPath is the environment variable that overrides the config file location.
	EnvConfigPath = "LOOTSHEET_CONFIG"
	// EnvDataDir is the environment variable that overrides the data directory.
	EnvDataDir = "LOOTSHEET_DATA_DIR"
	// EnvDatabasePath is the environment variable that overrides the database file path.
	EnvDatabasePath = "LOOTSHEET_DATABASE_PATH"
	// EnvBackupDir is the environment variable that overrides the backup directory.
	EnvBackupDir = "LOOTSHEET_BACKUP_DIR"
	// EnvExportDir is the environment variable that overrides the export directory.
	EnvExportDir      = "LOOTSHEET_EXPORT_DIR"
	defaultConfigFile = "config.json"
	dirPerm           = 0o755
)

// Config holds the resolved application configuration.
type Config struct {
	Paths Paths `json:"paths"`
}

// Paths holds the resolved filesystem paths for configuration, data, and the database.
type Paths struct {
	ConfigFile   string `json:"-"`
	DataDir      string `json:"data_dir"`
	DatabasePath string `json:"database_path"`
	BackupDir    string `json:"backup_dir"`
	ExportDir    string `json:"export_dir"`
}

type fileConfig struct {
	Paths filePaths `json:"paths"`
}

type filePaths struct {
	DataDir      string `json:"data_dir"`
	DatabasePath string `json:"database_path"`
	BackupDir    string `json:"backup_dir"`
	ExportDir    string `json:"export_dir"`
}

// Default returns a Config populated with OS-specific default paths.
func Default() (Config, error) {
	configPath, err := defaultConfigPath()
	if err != nil {
		return Config{}, err
	}

	dataDir, err := defaultDataDir()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Paths: Paths{
			ConfigFile:   configPath,
			DataDir:      dataDir,
			DatabasePath: filepath.Join(dataDir, DefaultDatabase),
			BackupDir:    filepath.Join(dataDir, DefaultBackupDirName),
			ExportDir:    filepath.Join(dataDir, DefaultExportDirName),
		},
	}, nil
}

// Load returns a fully resolved Config by merging defaults, the config file
// (if present), and environment variable overrides. All paths are normalized
// to absolute form.
func Load() (Config, error) {
	cfg, err := Default()
	if err != nil {
		return Config{}, err
	}

	configPath, err := resolveConfigPath()
	if err != nil {
		return Config{}, err
	}

	cfg.Paths.ConfigFile = configPath

	if err := mergeFileConfig(&cfg, configPath); err != nil {
		return Config{}, err
	}

	applyEnvOverrides(&cfg)

	if err := cfg.normalize(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks that all required path fields are set. It returns an error
// describing the first missing field, if any.
func (c *Config) Validate() error {
	switch {
	case c.Paths.ConfigFile == "":
		return errors.New("config file path is required")
	case c.Paths.DataDir == "":
		return errors.New("data directory path is required")
	case c.Paths.DatabasePath == "":
		return errors.New("database path is required")
	case c.Paths.BackupDir == "":
		return errors.New("backup directory path is required")
	case c.Paths.ExportDir == "":
		return errors.New("export directory path is required")
	default:
		return nil
	}
}

// EnsureDirectories creates the config, data, and database parent directories
// if they do not already exist.
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		filepath.Dir(c.Paths.ConfigFile),
		c.Paths.DataDir,
		filepath.Dir(c.Paths.DatabasePath),
		c.Paths.BackupDir,
		c.Paths.ExportDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return fmt.Errorf("create directory %q: %w", dir, err)
		}
	}

	return nil
}

func (c *Config) normalize() error {
	configFile, err := absolutePath(c.Paths.ConfigFile)
	if err != nil {
		return fmt.Errorf("normalize config path: %w", err)
	}
	c.Paths.ConfigFile = configFile

	configDir := filepath.Dir(c.Paths.ConfigFile)
	dataDir, err := absolutePathFromBase(configDir, c.Paths.DataDir)
	if err != nil {
		return fmt.Errorf("normalize data directory: %w", err)
	}
	c.Paths.DataDir = dataDir

	if strings.TrimSpace(c.Paths.DatabasePath) == "" {
		c.Paths.DatabasePath = DefaultDatabase
	}

	databasePath, err := absolutePathFromBase(c.Paths.DataDir, c.Paths.DatabasePath)
	if err != nil {
		return fmt.Errorf("normalize database path: %w", err)
	}
	c.Paths.DatabasePath = databasePath

	if strings.TrimSpace(c.Paths.BackupDir) == "" {
		c.Paths.BackupDir = DefaultBackupDirName
	}

	backupDir, err := absolutePathFromBase(c.Paths.DataDir, c.Paths.BackupDir)
	if err != nil {
		return fmt.Errorf("normalize backup directory: %w", err)
	}
	c.Paths.BackupDir = backupDir

	if strings.TrimSpace(c.Paths.ExportDir) == "" {
		c.Paths.ExportDir = DefaultExportDirName
	}

	exportDir, err := absolutePathFromBase(c.Paths.DataDir, c.Paths.ExportDir)
	if err != nil {
		return fmt.Errorf("normalize export directory: %w", err)
	}
	c.Paths.ExportDir = exportDir

	return c.Validate()
}

func mergeFileConfig(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	var parsed fileConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	if parsed.Paths.DataDir != "" {
		cfg.Paths.DataDir = parsed.Paths.DataDir
		if parsed.Paths.BackupDir == "" {
			cfg.Paths.BackupDir = ""
		}
		if parsed.Paths.ExportDir == "" {
			cfg.Paths.ExportDir = ""
		}
	}

	if parsed.Paths.DatabasePath != "" {
		cfg.Paths.DatabasePath = parsed.Paths.DatabasePath
	}

	if parsed.Paths.BackupDir != "" {
		cfg.Paths.BackupDir = parsed.Paths.BackupDir
	}

	if parsed.Paths.ExportDir != "" {
		cfg.Paths.ExportDir = parsed.Paths.ExportDir
	}

	return nil
}

func applyEnvOverrides(cfg *Config) {
	if value := strings.TrimSpace(os.Getenv(EnvDataDir)); value != "" {
		cfg.Paths.DataDir = value
		if strings.TrimSpace(os.Getenv(EnvBackupDir)) == "" {
			cfg.Paths.BackupDir = ""
		}
		if strings.TrimSpace(os.Getenv(EnvExportDir)) == "" {
			cfg.Paths.ExportDir = ""
		}
	}

	if value := strings.TrimSpace(os.Getenv(EnvDatabasePath)); value != "" {
		cfg.Paths.DatabasePath = value
	}

	if value := strings.TrimSpace(os.Getenv(EnvBackupDir)); value != "" {
		cfg.Paths.BackupDir = value
	}

	if value := strings.TrimSpace(os.Getenv(EnvExportDir)); value != "" {
		cfg.Paths.ExportDir = value
	}
}

func resolveConfigPath() (string, error) {
	if value := strings.TrimSpace(os.Getenv(EnvConfigPath)); value != "" {
		return absolutePath(value)
	}

	return defaultConfigPath()
}

func defaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config directory: %w", err)
	}

	return filepath.Join(dir, AppName, defaultConfigFile), nil
}

func defaultDataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
			return filepath.Join(localAppData, AppName), nil
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", AppName), nil
	default:
		if xdgDataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdgDataHome != "" {
			return filepath.Join(xdgDataHome, AppName), nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}

		return filepath.Join(home, ".local", "share", AppName), nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve fallback data directory: %w", err)
	}

	return filepath.Join(configDir, AppName), nil
}

func absolutePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.Clean(abs), nil
}

func absolutePathFromBase(base string, path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	return absolutePath(filepath.Join(base, path))
}
