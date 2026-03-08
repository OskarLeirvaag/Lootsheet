package loot

import "github.com/OskarLeirvaag/Lootsheet/src/ledger"

// LootItemRecord represents a loot item row from the database.
type LootItemRecord struct {
	ID        string
	Name      string
	Source    string
	Status    ledger.LootStatus
	Quantity  int
	Holder    string
	Notes     string
	CreatedAt string
	UpdatedAt string
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
