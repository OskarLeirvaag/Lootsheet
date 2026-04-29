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

// Ownership state values stored in compendium_sources.owned.
//
//   - Unknown: never probed (no cobalt cookie supplied to Phase A yet).
//   - Owned: user purchased this book — `isOwned: true` in available-user-content.
//   - Locked: user did NOT purchase this book and we have not observed any
//     content for it via the active DDB campaign.
//
// "Shared" is encoded separately on the `shared` column rather than as a
// fourth ownership value — a book can be shared into the active campaign
// regardless of ownership, and we don't want to lose the underlying ownership
// signal when sharing flips.
const (
	OwnershipUnknown = 0
	OwnershipOwned   = 1
	OwnershipLocked  = 2
)

// Source represents a D&D Beyond source book combined with the active
// campaign's per-source selection state.
type Source struct {
	ID         int
	Name       string
	// global / account-level
	Owned      int  // OwnershipUnknown / OwnershipOwned / OwnershipLocked
	Shared     bool // confirmed accessible via the active DDB campaign
	IsReleased bool
	CategoryID int
	// per-campaign (from compendium_campaign_sources)
	Enabled   bool
	HasSpells bool
	HasItems  bool
}
