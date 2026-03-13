package app

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/notes"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestBuildSearchHandlerCodexReturnsMatchingItems(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()

	_, err := codex.CreateEntry(ctx, databasePath, &codex.CreateInput{
		Name:  "Garrick the Bold",
		Notes: "A fearsome warrior.",
	})
	if err != nil {
		t.Fatalf("create codex entry: %v", err)
	}

	_, err = codex.CreateEntry(ctx, databasePath, &codex.CreateInput{
		Name:  "Elra the Wise",
		Notes: "A quiet scholar.",
	})
	if err != nil {
		t.Fatalf("create codex entry: %v", err)
	}

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	loader := &sqliteDataLoader{databasePath: databasePath, assets: assets}
	handler := buildSearchHandler(ctx, loader)

	items, err := handler(render.SectionCodex, "fearsome")
	if err != nil {
		t.Fatalf("search codex: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 codex result, got %d", len(items))
	}
	if items[0].DetailTitle != "Garrick the Bold" {
		t.Fatalf("result title = %q, want Garrick the Bold", items[0].DetailTitle)
	}
}

func TestBuildSearchHandlerNotesReturnsMatchingItems(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	ctx := context.Background()

	_, err := notes.CreateNote(ctx, databasePath, &notes.CreateNoteInput{
		Title: "Session 1",
		Body:  "The party fought a dragon.",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	_, err = notes.CreateNote(ctx, databasePath, &notes.CreateNoteInput{
		Title: "Shopping List",
		Body:  "Potions and rope.",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	loader := &sqliteDataLoader{databasePath: databasePath, assets: assets}
	handler := buildSearchHandler(ctx, loader)

	items, err := handler(render.SectionNotes, "dragon")
	if err != nil {
		t.Fatalf("search notes: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 note result, got %d", len(items))
	}
	if items[0].DetailTitle != "Session 1" {
		t.Fatalf("result title = %q, want Session 1", items[0].DetailTitle)
	}
}

func TestBuildSearchHandlerReturnsNilForUnsupportedSections(t *testing.T) {
	ctx := context.Background()
	assets, err := config.LoadInitAssets()
	if err != nil {
		t.Fatalf("load init assets: %v", err)
	}

	databasePath := testutil.InitTestDB(t)
	loader := &sqliteDataLoader{databasePath: databasePath, assets: assets}
	handler := buildSearchHandler(ctx, loader)

	items, err := handler(render.SectionJournal, "anything")
	if err != nil {
		t.Fatalf("search journal: %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil for unsupported section, got %d items", len(items))
	}
}
