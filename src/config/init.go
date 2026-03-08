package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

//go:embed setup/migrations/*.sql setup/seed_accounts.json
var initFS embed.FS

type InitAssets struct {
	Migrations    []InitMigration
	SchemaVersion string
	Accounts      []SeedAccount
}

type InitMigration struct {
	Version string
	Name    string
	SQL     string
}

type SeedAccount struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Active bool   `json:"active"`
}

func LoadInitAssets() (InitAssets, error) {
	migrationEntries, err := initFS.ReadDir("setup/migrations")
	if err != nil {
		return InitAssets{}, fmt.Errorf("read migration assets: %w", err)
	}

	migrationNames := make([]string, 0, len(migrationEntries))
	for _, entry := range migrationEntries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		migrationNames = append(migrationNames, entry.Name())
	}

	if len(migrationNames) == 0 {
		return InitAssets{}, fmt.Errorf("no init migrations found")
	}

	slices.Sort(migrationNames)

	migrations := make([]InitMigration, 0, len(migrationNames))
	for _, name := range migrationNames {
		version, err := parseMigrationVersion(name)
		if err != nil {
			return InitAssets{}, err
		}

		sqlBytes, err := initFS.ReadFile(filepath.Join("setup/migrations", name))
		if err != nil {
			return InitAssets{}, fmt.Errorf("read migration asset %q: %w", name, err)
		}

		migrations = append(migrations, InitMigration{
			Version: version,
			Name:    name,
			SQL:     strings.TrimSpace(string(sqlBytes)),
		})
	}

	seedAccountsJSON, err := initFS.ReadFile("setup/seed_accounts.json")
	if err != nil {
		return InitAssets{}, fmt.Errorf("read seed accounts asset: %w", err)
	}

	var accounts []SeedAccount
	if err := json.Unmarshal(seedAccountsJSON, &accounts); err != nil {
		return InitAssets{}, fmt.Errorf("parse seed accounts asset: %w", err)
	}

	return InitAssets{
		Migrations:    migrations,
		SchemaVersion: migrations[len(migrations)-1].Version,
		Accounts:      accounts,
	}, nil
}

func parseMigrationVersion(name string) (string, error) {
	prefix, _, found := strings.Cut(name, "_")
	if !found {
		return "", fmt.Errorf("migration file %q must use NNN_name.sql format", name)
	}

	version, err := strconv.Atoi(prefix)
	if err != nil {
		return "", fmt.Errorf("migration file %q has invalid numeric prefix", name)
	}

	if version <= 0 {
		return "", fmt.Errorf("migration file %q must use a positive version", name)
	}

	return strconv.Itoa(version), nil
}
