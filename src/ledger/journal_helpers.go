package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
)

// querier is satisfied by both *sql.DB and *sql.Tx, allowing callers to
// choose whether a query runs inside or outside a transaction.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type accountLookupRecord struct {
	ID     string
	Active bool
}

// NextJournalEntryNumber returns the next available journal entry number.
// Pass a *sql.Tx to ensure the read is part of the same transaction that
// will insert the new entry, preventing numbering races.
func NextJournalEntryNumber(ctx context.Context, q querier, campaignID string) (int, error) {
	var entryNumber int

	if err := q.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(entry_number), 0) + 1 FROM journal_entries WHERE campaign_id = ?",
		campaignID,
	).Scan(&entryNumber); err != nil {
		return 0, fmt.Errorf("query next journal entry number: %w", err)
	}

	return entryNumber, nil
}

// ResolveActiveAccountIDsByCode resolves account codes to account IDs,
// verifying that each account exists and is active.
func ResolveActiveAccountIDsByCode(ctx context.Context, db *sql.DB, campaignID string, lines []JournalLineInput) (map[string]string, error) {
	accountCodes := make([]string, 0, len(lines))
	seenCodes := make(map[string]struct{}, len(lines))

	for _, line := range lines {
		if _, seen := seenCodes[line.AccountCode]; seen {
			continue
		}

		seenCodes[line.AccountCode] = struct{}{}
		accountCodes = append(accountCodes, line.AccountCode)
	}

	slices.Sort(accountCodes)

	placeholders := make([]string, len(accountCodes))
	args := make([]any, 0, len(accountCodes)+1)
	args = append(args, campaignID)
	for i, code := range accountCodes {
		placeholders[i] = "?"
		args = append(args, code)
	}

	query := "SELECT code, id, active FROM accounts WHERE campaign_id = ? AND code IN (" + strings.Join(placeholders, ", ") + ")" //nolint:gosec // placeholders are "?" literals, not user input
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query account codes: %w", err)
	}
	defer rows.Close()

	records := map[string]accountLookupRecord{}
	for rows.Next() {
		var code string
		var r accountLookupRecord
		var active int

		if err := rows.Scan(&code, &r.ID, &active); err != nil {
			return nil, fmt.Errorf("scan account lookup row: %w", err)
		}

		r.Active = active == 1
		records[code] = r
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate account lookup rows: %w", err)
	}

	resolved := make(map[string]string, len(accountCodes))
	for _, code := range accountCodes {
		record, ok := records[code]
		if !ok {
			return nil, fmt.Errorf("account code %q does not exist", code)
		}

		if !record.Active {
			return nil, fmt.Errorf("account code %q is inactive", code)
		}

		resolved[code] = record.ID
	}

	return resolved, nil
}
