package quest

import (
	"context"
	"database/sql"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
)

// rebuildReferences deletes old references for a quest and inserts new ones
// parsed from the notes text.
func rebuildReferences(ctx context.Context, db *sql.DB, questID, questTitle, notes string) error {
	return refs.RebuildReferences(ctx, db, "quest", questID, questTitle, notes)
}
