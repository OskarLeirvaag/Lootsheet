package config

import "testing"

func TestLoadInitAssets(t *testing.T) {
	assets, err := LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	if assets.SchemaSQL == "" {
		t.Fatal("schema SQL must not be empty")
	}

	if len(assets.Accounts) != 16 {
		t.Fatalf("seed account count = %d, want 16", len(assets.Accounts))
	}

	seenIDs := make(map[string]struct{})
	seenCodes := make(map[string]struct{})

	for _, account := range assets.Accounts {
		if account.ID == "" {
			t.Fatal("seed account ID must not be empty")
		}

		if account.Code == "" {
			t.Fatalf("seed account %q must have a code", account.ID)
		}

		if account.Name == "" {
			t.Fatalf("seed account %q must have a name", account.ID)
		}

		if account.Type == "" {
			t.Fatalf("seed account %q must have a type", account.ID)
		}

		if _, exists := seenIDs[account.ID]; exists {
			t.Fatalf("seed account ID %q must be unique", account.ID)
		}
		seenIDs[account.ID] = struct{}{}

		if _, exists := seenCodes[account.Code]; exists {
			t.Fatalf("seed account code %q must be unique", account.Code)
		}
		seenCodes[account.Code] = struct{}{}
	}
}
