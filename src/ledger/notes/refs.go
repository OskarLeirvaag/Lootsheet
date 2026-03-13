package notes

import (
	"context"
	"database/sql"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
)

// rebuildReferences deletes old references for a note and inserts new ones
// parsed from the body text.
func rebuildReferences(ctx context.Context, db *sql.DB, noteID, campaignID, noteTitle, body string) error {
	return refs.RebuildReferences(ctx, db, "note", noteID, campaignID, noteTitle, body)
}
