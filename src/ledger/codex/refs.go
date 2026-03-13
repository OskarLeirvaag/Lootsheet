package codex

import (
	"context"
	"database/sql"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
)

// rebuildReferences deletes old references for an entry and inserts new ones
// parsed from the notes text.
func rebuildReferences(ctx context.Context, db *sql.DB, entryID, campaignID, entryName, notes string) error {
	return refs.RebuildReferences(ctx, db, "codex", entryID, campaignID, entryName, notes)
}
