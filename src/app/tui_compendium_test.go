package app

import (
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ddb"
)

func TestFilterDDBSpellsBySource(t *testing.T) {
	spells := []ddb.RawSpellEntry{
		{Definition: ddb.RawSpellDef{ID: 1, Sources: []ddb.SourceRef{{SourceID: 1}}}},
		{Definition: ddb.RawSpellDef{ID: 2, Sources: []ddb.SourceRef{{SourceID: 2}}}},
		{Definition: ddb.RawSpellDef{ID: 3}},
	}

	got := filterDDBSpellsBySource(spells, []int{2})
	if len(got) != 1 || got[0].Definition.ID != 2 {
		t.Fatalf("filtered spells = %#v, want only spell 2", got)
	}
}

func TestFilterDDBItemsBySource(t *testing.T) {
	items := []ddb.RawItem{
		{ID: 1, Sources: []ddb.SourceRef{{SourceID: 1}}},
		{ID: 2, Sources: []ddb.SourceRef{{SourceID: 2}}},
		{ID: 3},
	}

	got := filterDDBItemsBySource(items, []int{1})
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("filtered items = %#v, want only item 1", got)
	}
}

func TestFilterDDBMonstersBySourceDropsNonEnabledSources(t *testing.T) {
	monsters := []ddb.RawMonster{
		{ID: 1, Sources: []ddb.SourceRef{{SourceID: 5}}},   // MM 2014 only → keep
		{ID: 2, Sources: []ddb.SourceRef{{SourceID: 198}}},  // MM 2024 only → drop
		{ID: 3, Sources: []ddb.SourceRef{{SourceID: 5}, {SourceID: 198}}}, // both → keep (is in 2014)
		{ID: 4},                                              // no sources → drop
	}

	got := filterDDBMonstersBySource(monsters, []int{5})
	if len(got) != 2 {
		t.Fatalf("filterDDBMonstersBySource = %d results, want 2: %+v", len(got), got)
	}
	if got[0].ID != 1 || got[1].ID != 3 {
		t.Fatalf("unexpected IDs: %d %d", got[0].ID, got[1].ID)
	}
}

func TestConvertDDBSourcesDropsBadAndUnreleasedIDs(t *testing.T) {
	in := []ddb.ConfigSource{
		{ID: 1, Description: "PHB", IsReleased: true, SourceCategoryID: 10},
		{ID: 4, Description: "EE players", IsReleased: true},      // BAD
		{ID: 31, Description: "CR data", IsReleased: true},        // BAD
		{ID: 99, Description: "Unreleased", IsReleased: false},    // dropped
		{ID: 2, Description: "MM", IsReleased: true, SourceCategoryID: 10},
	}

	got := convertDDBSources(in)
	if len(got) != 2 {
		t.Fatalf("convertDDBSources kept %d rows, want 2: %#v", len(got), got)
	}
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("convertDDBSources order/IDs unexpected: %#v", got)
	}
	if !got[0].IsReleased || got[0].CategoryID != 10 {
		t.Fatalf("convertDDBSources lost flags: %#v", got[0])
	}
}

func TestParseSyncCobaltExtractsForceFlag(t *testing.T) {
	tests := []struct {
		in        string
		wantToken string
		wantForce bool
	}{
		{"abc123", "abc123", false},
		{"abc123:force", "abc123", true},
		{"  abc123:FORCE  ", "abc123", true},
		{"abc123:force:force", "abc123:force", true}, // strip only one trailing
		{":force", "", true},
		{"", "", false},
	}
	for _, tc := range tests {
		gotTok, gotForce := parseSyncCobalt(tc.in)
		if gotTok != tc.wantToken || gotForce != tc.wantForce {
			t.Errorf("parseSyncCobalt(%q) = (%q, %v), want (%q, %v)",
				tc.in, gotTok, gotForce, tc.wantToken, tc.wantForce)
		}
	}
}

func TestObservedSourceIDsCollectsAllReferences(t *testing.T) {
	spells := []ddb.RawSpellEntry{
		{Definition: ddb.RawSpellDef{ID: 1, Sources: []ddb.SourceRef{{SourceID: 1}, {SourceID: 5}}}},
		{Definition: ddb.RawSpellDef{ID: 2, Sources: []ddb.SourceRef{{SourceID: 1}}}},
		{Definition: ddb.RawSpellDef{ID: 3}},
	}
	got := observedSourceIDs(spells, spellEntrySourceIDs)
	if _, ok := got[1]; !ok {
		t.Errorf("observed missing source 1: %v", got)
	}
	if _, ok := got[5]; !ok {
		t.Errorf("observed missing source 5: %v", got)
	}
	if len(got) != 2 {
		t.Errorf("observed has %d entries, want 2: %v", len(got), got)
	}
}
