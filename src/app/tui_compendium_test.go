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
