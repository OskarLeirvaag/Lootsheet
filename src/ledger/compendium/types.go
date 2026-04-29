// Package compendium provides repository functions for cross-campaign D&D
// Beyond reference data (monsters, spells, items, rules, conditions).
package compendium

// Monster represents a cached monster from D&D Beyond.
type Monster struct {
	ID         int
	DdbID      int
	Name       string
	CR         string
	Type       string
	Size       string
	HP         string
	AC         string
	SourceName string
	DetailJSON string
	SyncedAt   string
}

// Spell represents a cached spell from D&D Beyond.
type Spell struct {
	ID          int
	DdbID       int
	Name        string
	Level       int
	School      string
	CastingTime string
	Range       string
	Components  string
	Duration    string
	Classes     string
	SourceName  string
	DetailJSON  string
	SyncedAt    string
}

// Item represents a cached item from D&D Beyond.
type Item struct {
	ID         int
	DdbID      int
	Name       string
	Type       string
	Rarity     string
	Attunement bool
	SourceName string
	DetailJSON string
	SyncedAt   string
}

// Rule represents a cached rule entry (rules, basic actions, weapon properties).
type Rule struct {
	ID          int
	DdbID       int
	Name        string
	Category    string
	Description string
	SyncedAt    string
}

// Condition represents a cached condition from D&D Beyond.
type Condition struct {
	ID          int
	DdbID       int
	Name        string
	Description string
	SyncedAt    string
}

// Ownership tri-state values stored in compendium_sources.owned.
const (
	OwnershipUnknown = 0
	OwnershipOwned   = 1
	OwnershipLocked  = 2
)

// Source represents a D&D Beyond source book for filtering.
type Source struct {
	ID         int
	Name       string
	Enabled    bool
	Owned      int  // OwnershipUnknown / OwnershipOwned / OwnershipLocked
	HasSpells  bool // false once we observe a sync return zero spells
	HasItems   bool
	IsReleased bool
	CategoryID int
}
