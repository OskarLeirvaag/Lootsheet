package notes

import (
	"context"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// RunCreate creates a note and writes the CLI output.
func RunCreate(ctx context.Context, hctx ledger.HandlerContext, input *CreateNoteInput) error {
	if input == nil {
		return fmt.Errorf("note input is required")
	}

	result, err := CreateNote(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Created note %s\nTitle: %s\n",
		result.ID,
		result.Title,
	); err != nil {
		return fmt.Errorf("write note output: %w", err)
	}

	return nil
}

// RunList writes the note listing.
func RunList(ctx context.Context, hctx ledger.HandlerContext) error {
	noteList, err := ListNotes(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "UPDATED     TITLE"); err != nil {
		return fmt.Errorf("write notes header: %w", err)
	}

	for i := range noteList {
		updated := noteList[i].UpdatedAt
		if len(updated) > 10 {
			updated = updated[:10]
		}
		if _, err := fmt.Fprintf(
			hctx.Stdout,
			"%-11s %s\n",
			updated,
			noteList[i].Title,
		); err != nil {
			return fmt.Errorf("write note row: %w", err)
		}
	}

	return nil
}

// RunSearch writes matching notes for a query.
func RunSearch(ctx context.Context, hctx ledger.HandlerContext, query string) error {
	noteList, err := SearchNotes(ctx, hctx.DatabasePath, query)
	if err != nil {
		return err
	}

	if len(noteList) == 0 {
		if _, err := fmt.Fprintln(hctx.Stdout, "No matching notes found."); err != nil {
			return fmt.Errorf("write search output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "UPDATED     TITLE"); err != nil {
		return fmt.Errorf("write notes header: %w", err)
	}

	for i := range noteList {
		updated := noteList[i].UpdatedAt
		if len(updated) > 10 {
			updated = updated[:10]
		}
		if _, err := fmt.Fprintf(
			hctx.Stdout,
			"%-11s %s\n",
			updated,
			noteList[i].Title,
		); err != nil {
			return fmt.Errorf("write note row: %w", err)
		}
	}

	return nil
}
