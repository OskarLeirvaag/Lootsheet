// Package loot provides repository and CLI handler functions for managing loot
// items and their lifecycle: creation, appraisal (off-ledger), recognition
// (on-ledger), and sale.
package loot

import "github.com/OskarLeirvaag/Lootsheet/src/ledger"

// LootItemRecord represents a loot item row from the database.
type LootItemRecord struct {
	ID        string
	Name      string
	Source    string
	Status    ledger.LootStatus
	ItemType  string
	Quantity  int
	Holder    string
	Notes     string
	CreatedAt string
	UpdatedAt string
}

// UpdateLootItemInput holds the editable fields for a loot register row.
type UpdateLootItemInput struct {
	Name     string
	Source   string
	Quantity int
	Holder   string
	Notes    string
}

// LootAppraisalRecord represents a loot appraisal row from the database.
type LootAppraisalRecord struct {
	ID                string
	LootItemID        string
	AppraisedValue    int64
	Appraiser         string
	Notes             string
	AppraisedAt       string
	RecognizedEntryID string
	CreatedAt         string
}
