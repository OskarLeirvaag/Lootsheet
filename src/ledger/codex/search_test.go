package codex

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestSearchEntries_PrefixMatch(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, dbPath)
	ctx := context.Background()

	if _, err := CreateType(ctx, dbPath, "custom-player", "CustomPlayer", "player"); err != nil {
		t.Fatalf("create type: %v", err)
	}
	if _, err := CreateEntry(ctx, dbPath, campaignID, &CreateInput{
		TypeID: "custom-player", Name: "Thalion Moonwhisper", PlayerName: "Bryn Sander",
	}); err != nil {
		t.Fatalf("create entry: %v", err)
	}

	for _, tc := range []struct {
		query string
		want  int
	}{
		{"Bryn", 1},
		{"bryn", 1},
		{"Moon", 1},       // prefix match — previously returned 0
		{"Tha", 1},        // prefix match
		{"whisper", 1},    // substring match via LIKE fallback
		{"xyzfoo", 0},
		{"@[person", 0},   // special chars — previously crashed FTS
		{"person/Bryn", 0}, // special chars — previously crashed FTS
	} {
		results, err := SearchEntries(ctx, dbPath, campaignID, tc.query)
		if err != nil {
			t.Errorf("search %q error: %v", tc.query, err)
			continue
		}
		if len(results) != tc.want {
			t.Errorf("search %q: got %d, want %d", tc.query, len(results), tc.want)
		}
	}
}

func TestSearchEntries_ByReference(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, dbPath)
	ctx := context.Background()

	if _, err := CreateType(ctx, dbPath, "custom-player", "CustomPlayer", "player"); err != nil {
		t.Fatalf("create type: %v", err)
	}

	// Alice is a codex entry with a reference to "Bryn Sander" in her notes.
	// Her own name doesn't contain "Bryn".
	_, err := CreateEntry(ctx, dbPath, campaignID, &CreateInput{
		TypeID: "custom-player",
		Name:   "Alice",
		Notes:  "Met @[person/Bryn Sander] at the tavern.",
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	// Searching "Bryn" should find Alice because her refs point at Bryn Sander.
	results, err := SearchEntries(ctx, dbPath, campaignID, "Bryn")
	if err != nil {
		t.Fatalf("search Bryn: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result via reference, got %d", len(results))
	}
	if results[0].Name != "Alice" {
		t.Errorf("expected Alice, got %q", results[0].Name)
	}
}
