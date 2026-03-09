package app

import (
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/currency"
)

func summarizeJournal(summary journal.Summary) []string {
	if summary.TotalEntries == 0 {
		return []string{
			"Entries: 0 total",
			"Posted: 0",
			"Reversal entries: 0",
			"No journal activity yet.",
		}
	}

	return []string{
		fmt.Sprintf("Entries: %d total", summary.TotalEntries),
		fmt.Sprintf("Posted: %d", summary.PostedEntries),
		fmt.Sprintf("Reversal entries: %d", summary.ReversalEntries),
		fmt.Sprintf("Latest: #%d %s", summary.LatestEntryNumber, summary.LatestEntryDate),
	}
}

func buildJournalItems(entries []journal.BrowseEntryRecord) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(entries))
	for index := range entries {
		entry := &entries[index]
		rowStatus := string(entry.Status)
		detailLines := []string{
			"Date: " + entry.EntryDate,
			"Status: " + string(entry.Status),
			"Description: " + entry.Description,
		}
		if entry.ReversesEntryID != "" {
			rowStatus = "reversal"
			detailLines = append(detailLines, fmt.Sprintf("Reverses: entry #%d", entry.ReversesEntryNumber))
		}
		if entry.ReversedByEntryID != "" {
			detailLines = append(detailLines, fmt.Sprintf("Reversed by: entry #%d", entry.ReversedByEntryNumber))
		}
		if entry.Status == ledger.JournalEntryStatusReversed {
			detailLines = append(detailLines, "This entry has been reversed and remains in the audit trail.")
		}
		detailLines = append(detailLines, "", "Lines:")
		if len(entry.Lines) == 0 {
			detailLines = append(detailLines, "No journal lines loaded.")
		}
		for _, line := range entry.Lines {
			detailLines = append(detailLines, formatJournalDetailLine(line))
		}

		var actions []render.ItemActionData
		if entry.Status == ledger.JournalEntryStatusPosted {
			actions = []render.ItemActionData{{
				Trigger:      render.ActionReverse,
				ID:           tuiCommandJournalReverse,
				Label:        "r reverse",
				ConfirmTitle: fmt.Sprintf("Reverse entry #%d?", entry.EntryNumber),
				ConfirmLines: []string{
					entry.Description,
					"Original date: " + entry.EntryDate,
					"Reversal date: " + entry.EntryDate,
					"A new posted reversing entry will be created.",
					fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Reversal of entry #%d", entry.EntryNumber)),
				},
			}}
		}

		items = append(items, render.ListItemData{
			Key:         entry.ID,
			Row:         fmt.Sprintf("#%-4d %-10s %-8s %s", entry.EntryNumber, entry.EntryDate, rowStatus, entry.Description),
			DetailTitle: fmt.Sprintf("Entry #%d", entry.EntryNumber),
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

func formatJournalDetailLine(line journal.BrowseEntryLine) string {
	side := "CR"
	amount := line.CreditAmount
	if line.DebitAmount > 0 {
		side = "DR"
		amount = line.DebitAmount
	}

	text := fmt.Sprintf("%s %s %s %s", line.AccountCode, line.AccountName, side, currency.FormatAmount(amount))
	if strings.TrimSpace(line.Memo) == "" {
		return text
	}

	return text + " (" + line.Memo + ")"
}

func findBrowseEntry(entries []journal.BrowseEntryRecord, entryID string) (journal.BrowseEntryRecord, bool) {
	for index := range entries {
		entry := entries[index]
		if entry.ID == entryID {
			return entry, true
		}
	}

	return journal.BrowseEntryRecord{}, false
}
