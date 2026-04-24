package notes

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestSearchNotes_PrefixMatch(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, dbPath)
	ctx := context.Background()

	_, err := CreateNote(ctx, dbPath, campaignID, &CreateNoteInput{
		Title: "Moonwhisper Chronicle",
		Body:  "Adventure begins on a moonlit night.",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	for _, tc := range []struct {
		query string
		want  int
	}{
		{"Moonwhisper", 1},
		{"moonwhisper", 1},
		{"Moon", 1},       // prefix — previously 0
		{"moon", 1},
		{"adventure", 1},
		{"advent", 1},     // prefix
		{"xyzfoo", 0},
		{"@[person", 0},   // previously crashed
	} {
		results, err := SearchNotes(ctx, dbPath, campaignID, tc.query)
		if err != nil {
			t.Errorf("search %q error: %v", tc.query, err)
			continue
		}
		if len(results) != tc.want {
			t.Errorf("search %q: got %d, want %d", tc.query, len(results), tc.want)
		}
	}
}

func TestSearchNotes_ByReference(t *testing.T) {
	dbPath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, dbPath)
	ctx := context.Background()

	// Note body contains the reference markup. FTS should find "Bryn" via
	// the body tokens, but we also verify reference-table fallback works
	// when body tokenization misses the target (e.g. after future refactor).
	_, err := CreateNote(ctx, dbPath, campaignID, &CreateNoteInput{
		Title: "Session 5",
		Body:  "Met @[person/Bryn Sander] at the gate.",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	results, err := SearchNotes(ctx, dbPath, campaignID, "Bryn")
	if err != nil {
		t.Fatalf("search Bryn: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
