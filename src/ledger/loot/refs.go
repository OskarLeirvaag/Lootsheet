package loot

import (
	"context"
	"database/sql"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
)

// rebuildReferences deletes old references for a loot/asset item and inserts
// new ones parsed from the notes text.
func rebuildReferences(ctx context.Context, db *sql.DB, itemID, itemName, notes string) error {
	return refs.RebuildReferences(ctx, db, "loot", itemID, itemName, notes)
}
