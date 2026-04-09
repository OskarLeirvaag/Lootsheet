package compendium

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestUpsertAndListMonsters(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	ctx := context.Background()

	monsters := []Monster{
		{DdbID: 16907, Name: "Goblin", CR: "1/4", Type: "Humanoid", Size: "Small", HP: "7", AC: "15", SourceName: "Basic Rules", DetailJSON: `{"id":16907}`},
		{DdbID: 16808, Name: "Dragon", CR: "17", Type: "Dragon", Size: "Huge", HP: "256", AC: "19", SourceName: "PHB", DetailJSON: `{"id":16808}`},
	}
	if err := UpsertMonsters(ctx, dbPath, monsters); err != nil {
		t.Fatalf("upsert monsters: %v", err)
	}

	all, err := ListMonsters(ctx, dbPath, "")
	if err != nil {
		t.Fatalf("list monsters: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 monsters, got %d", len(all))
	}

	// FTS search.
	goblins, err := ListMonsters(ctx, dbPath, "goblin")
	if err != nil {
		t.Fatalf("search monsters: %v", err)
	}
	if len(goblins) != 1 || goblins[0].Name != "Goblin" {
		t.Fatalf("expected 1 goblin, got %d", len(goblins))
	}

	// Upsert is idempotent (updates on conflict).
	monsters[0].CR = "1/2"
	if err := UpsertMonsters(ctx, dbPath, monsters[:1]); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	all, _ = ListMonsters(ctx, dbPath, "")
	if len(all) != 2 {
		t.Fatalf("expected 2 after re-upsert, got %d", len(all))
	}
	for _, m := range all {
		if m.DdbID == 16907 && m.CR != "1/2" {
			t.Fatalf("expected updated CR '1/2', got %q", m.CR)
		}
	}
}

func TestUpsertAndListSpells(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	ctx := context.Background()

	spells := []Spell{
		{DdbID: 2110, Name: "Floating Disk", Level: 1, School: "Conjuration", CastingTime: "1 action", Range: "30 ft", Components: "V,S,M", Duration: "1 hour", Classes: "Wizard", SourceName: "PHB"},
		{DdbID: 2241, Name: "Fireball", Level: 3, School: "Evocation", CastingTime: "1 action", Range: "150 ft", Components: "V,S,M", Duration: "Instantaneous", Classes: "Sorcerer,Wizard", SourceName: "PHB"},
	}
	if err := UpsertSpells(ctx, dbPath, spells); err != nil {
		t.Fatalf("upsert spells: %v", err)
	}

	all, err := ListSpells(ctx, dbPath, "")
	if err != nil {
		t.Fatalf("list spells: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 spells, got %d", len(all))
	}
	// Ordered by level, name: Floating Disk (1) before Fireball (3).
	if all[0].Name != "Floating Disk" {
		t.Fatalf("expected first spell 'Floating Disk', got %q", all[0].Name)
	}

	// FTS search by school.
	evo, err := ListSpells(ctx, dbPath, "evocation")
	if err != nil {
		t.Fatalf("search spells: %v", err)
	}
	if len(evo) != 1 || evo[0].Name != "Fireball" {
		t.Fatalf("expected 1 evocation spell, got %d", len(evo))
	}
}

func TestUpsertAndListItems(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	ctx := context.Background()

	items := []Item{
		{DdbID: 4570, Name: "Amulet of the Planes", Type: "Wondrous item", Rarity: "Very Rare", Attunement: true, SourceName: "PHB"},
		{DdbID: 4572, Name: "Apparatus of the Crab", Type: "Wondrous item", Rarity: "Legendary", Attunement: false, SourceName: "DMG"},
	}
	if err := UpsertItems(ctx, dbPath, items); err != nil {
		t.Fatalf("upsert items: %v", err)
	}

	all, err := ListItems(ctx, dbPath, "")
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 items, got %d", len(all))
	}

	// FTS search by rarity.
	legendary, err := ListItems(ctx, dbPath, "legendary")
	if err != nil {
		t.Fatalf("search items: %v", err)
	}
	if len(legendary) != 1 || legendary[0].Name != "Apparatus of the Crab" {
		t.Fatalf("expected 1 legendary item, got %d", len(legendary))
	}
}

func TestUpsertAndListRules(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	ctx := context.Background()

	rules := []Rule{
		{DdbID: 1, Name: "Attack", Category: "Action", Description: "Make a weapon attack."},
		{DdbID: 2, Name: "Dash", Category: "Action", Description: "Double your movement."},
		{DdbID: 100, Name: "Ammunition", Category: "Weapon Property", Description: "Requires ammo."},
	}
	if err := UpsertRules(ctx, dbPath, rules); err != nil {
		t.Fatalf("upsert rules: %v", err)
	}

	all, err := ListRules(ctx, dbPath, "")
	if err != nil {
		t.Fatalf("list rules: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(all))
	}

	// FTS search.
	actions, err := ListRules(ctx, dbPath, "action")
	if err != nil {
		t.Fatalf("search rules: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 action rules, got %d", len(actions))
	}
}

func TestUpsertAndListConditions(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	ctx := context.Background()

	conditions := []Condition{
		{DdbID: 1, Name: "Blinded", Description: "Can't see."},
		{DdbID: 2, Name: "Charmed", Description: "Can't attack the charmer."},
	}
	if err := UpsertConditions(ctx, dbPath, conditions); err != nil {
		t.Fatalf("upsert conditions: %v", err)
	}

	all, err := ListConditions(ctx, dbPath, "")
	if err != nil {
		t.Fatalf("list conditions: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(all))
	}

	// FTS search.
	blind, err := ListConditions(ctx, dbPath, "blind")
	if err != nil {
		t.Fatalf("search conditions: %v", err)
	}
	if len(blind) != 1 || blind[0].Name != "Blinded" {
		t.Fatalf("expected 1 blinded condition, got %d", len(blind))
	}
}

func TestUpsertAndListSources(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	ctx := context.Background()

	sources := []Source{
		{ID: 1, Name: "Basic Rules", Enabled: true},
		{ID: 2, Name: "Player's Handbook", Enabled: false},
	}
	if err := UpsertSources(ctx, dbPath, sources); err != nil {
		t.Fatalf("upsert sources: %v", err)
	}

	all, err := ListSources(ctx, dbPath)
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(all))
	}

	enabled, err := EnabledSourceIDs(ctx, dbPath)
	if err != nil {
		t.Fatalf("enabled sources: %v", err)
	}
	if len(enabled) != 1 || enabled[0] != 1 {
		t.Fatalf("expected [1] enabled, got %v", enabled)
	}

	// Toggle PHB on.
	if err := ToggleSource(ctx, dbPath, 2); err != nil {
		t.Fatalf("toggle source: %v", err)
	}
	enabled, _ = EnabledSourceIDs(ctx, dbPath)
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled after toggle, got %d", len(enabled))
	}

	// Toggle PHB off.
	if err := ToggleSource(ctx, dbPath, 2); err != nil {
		t.Fatalf("toggle source off: %v", err)
	}
	enabled, _ = EnabledSourceIDs(ctx, dbPath)
	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled after toggle off, got %d", len(enabled))
	}

	// Upsert preserves enabled state (ON CONFLICT updates name only).
	sources[1].Name = "PHB (2014)"
	if err := UpsertSources(ctx, dbPath, sources); err != nil {
		t.Fatalf("re-upsert sources: %v", err)
	}
	all, _ = ListSources(ctx, dbPath)
	for _, s := range all {
		if s.ID == 2 && s.Name != "PHB (2014)" {
			t.Fatalf("expected updated name, got %q", s.Name)
		}
		if s.ID == 2 && s.Enabled {
			t.Fatal("expected PHB to stay disabled after re-upsert")
		}
	}
}
