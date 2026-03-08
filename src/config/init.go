package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed setup/schema.sql setup/seed_accounts.json
var initFS embed.FS

type InitAssets struct {
	SchemaSQL string
	Accounts  []SeedAccount
}

type SeedAccount struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Active bool   `json:"active"`
}

func LoadInitAssets() (InitAssets, error) {
	schemaSQL, err := initFS.ReadFile("setup/schema.sql")
	if err != nil {
		return InitAssets{}, fmt.Errorf("read schema asset: %w", err)
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
		SchemaSQL: strings.TrimSpace(string(schemaSQL)),
		Accounts:  accounts,
	}, nil
}
