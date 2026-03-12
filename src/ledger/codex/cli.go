package codex

import (
	"context"
	"fmt"
	"io"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// RunCreate creates a codex entry and writes the CLI output.
func RunCreate(ctx context.Context, hctx ledger.HandlerContext, input *CreateInput) error {
	if input == nil {
		return fmt.Errorf("codex entry input is required")
	}

	result, err := CreateEntry(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Created codex entry %s\nName: %s\nType: %s\n",
		result.ID,
		result.Name,
		result.TypeName,
	); err != nil {
		return fmt.Errorf("write codex output: %w", err)
	}

	return nil
}

// RunList writes the codex entry listing.
func RunList(ctx context.Context, hctx ledger.HandlerContext) error {
	entries, err := ListEntries(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "TYPE     SECONDARY      NAME"); err != nil {
		return fmt.Errorf("write codex header: %w", err)
	}

	return writeCodexRows(hctx.Stdout, entries)
}

// RunSearch writes matching codex entries for a query.
func RunSearch(ctx context.Context, hctx ledger.HandlerContext, query string) error {
	entries, err := SearchEntries(ctx, hctx.DatabasePath, query)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		if _, err := fmt.Fprintln(hctx.Stdout, "No matching codex entries found."); err != nil {
			return fmt.Errorf("write search output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "TYPE     SECONDARY      NAME"); err != nil {
		return fmt.Errorf("write codex header: %w", err)
	}

	return writeCodexRows(hctx.Stdout, entries)
}

func writeCodexRows(w io.Writer, entries []CodexEntry) error {
	for i := range entries {
		e := &entries[i]
		secondary := e.Disposition
		if e.TypeID == "player" {
			secondary = e.Class
		}
		if _, err := fmt.Fprintf(
			w,
			"%-8s %-14s %s\n",
			e.TypeName,
			secondary,
			e.Name,
		); err != nil {
			return fmt.Errorf("write codex row: %w", err)
		}
	}
	return nil
}
