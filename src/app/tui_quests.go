package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

type tuiQuestRow struct {
	Record      quest.QuestRecord
	TotalPaid   int64
	Outstanding int64
	Collectible bool
	OffLedger   bool
}

func summarizeQuests(promised []report.PromisedQuestRow, receivables []report.QuestReceivableRow, writeOffCandidates []report.WriteOffCandidateRow) []string {
	var promisedValue int64
	for _, row := range promised {
		promisedValue += row.PromisedReward
	}

	var receivableValue int64
	for _, row := range receivables {
		receivableValue += row.Outstanding
	}

	lines := []string{
		fmt.Sprintf("Promised quests: %d", len(promised)),
		"Promised value: " + currency.FormatAmount(promisedValue),
		fmt.Sprintf("Receivables: %d", len(receivables)),
		"Outstanding: " + currency.FormatAmount(receivableValue),
	}

	if len(writeOffCandidates) > 0 {
		lines = append(lines, fmt.Sprintf("Stale receivables: %d (%d+ days)", len(writeOffCandidates), writeOffMinAgeDays))
	}

	return lines
}

func buildQuestItems(quests []tuiQuestRow, today string) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(quests))
	for index := range quests {
		row := &quests[index]
		record := &row.Record
		patron := record.Patron
		if strings.TrimSpace(patron) == "" {
			patron = "No patron"
		}

		detailLines := []string{
			"Patron: " + patron,
			"Status: " + string(record.Status),
			"Promised reward: " + currency.FormatAmount(record.PromisedBaseReward),
			"Outstanding: " + currency.FormatAmount(row.Outstanding),
			"Collected so far: " + currency.FormatAmount(row.TotalPaid),
			"Accounting state: " + questAccountingState(row),
		}
		if record.PartialAdvance > 0 {
			detailLines = append(detailLines, "Partial advance: "+currency.FormatAmount(record.PartialAdvance))
		}
		if record.AcceptedOn != "" {
			detailLines = append(detailLines, "Accepted on: "+record.AcceptedOn)
		}
		if record.CompletedOn != "" {
			detailLines = append(detailLines, "Completed on: "+record.CompletedOn)
		}
		if record.ClosedOn != "" {
			detailLines = append(detailLines, "Closed on: "+record.ClosedOn)
		}
		if strings.TrimSpace(record.BonusConditions) != "" {
			detailLines = append(detailLines, "Bonus conditions: "+record.BonusConditions)
		}
		if strings.TrimSpace(record.Notes) != "" {
			detailLines = append(detailLines, "Notes: "+record.Notes)
		}

		var actions []render.ItemActionData
		actions = append(actions, render.ItemActionData{
			Trigger:      render.ActionEdit,
			ID:           tuiCommandQuestUpdate,
			Label:        "u edit",
			Mode:         render.ItemActionModeCompose,
			ComposeMode:  "quest",
			ComposeTitle: "Edit Quest",
			ComposeFields: map[string]string{
				"title":       record.Title,
				"patron":      record.Patron,
				"description": record.Description,
				"reward":      currency.FormatAmount(record.PromisedBaseReward),
				"advance":     currency.FormatAmount(record.PartialAdvance),
				"bonus":       record.BonusConditions,
				"notes":       record.Notes,
				"status":      string(record.Status),
				"accepted_on": record.AcceptedOn,
			},
		})
		if row.Collectible {
			actions = append(actions,
				render.ItemActionData{
					Trigger:      render.ActionCollect,
					ID:           tuiCommandQuestCollectFull,
					Label:        "c collect",
					ConfirmTitle: fmt.Sprintf("Collect full payment for %q?", record.Title),
					ConfirmLines: []string{
						"Outstanding: " + currency.FormatAmount(row.Outstanding),
						"Collection date: " + today,
						"This collects the full remaining receivable.",
						fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Quest payment: %s", record.Title)),
					},
				},
				render.ItemActionData{
					Trigger:      render.ActionWriteOff,
					ID:           tuiCommandQuestWriteOffFull,
					Label:        "w write off",
					ConfirmTitle: fmt.Sprintf("Write off %q?", record.Title),
					ConfirmLines: []string{
						"Outstanding: " + currency.FormatAmount(row.Outstanding),
						"Write-off date: " + today,
						"This records the remaining balance as a failed patron loss.",
						fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Quest write-off: %s", record.Title)),
					},
				},
			)
		}

		items = append(items, render.ListItemData{
			Key:         record.ID,
			Row:         fmt.Sprintf("%-12s %-14s %-12s %s (%s)", currency.FormatAmount(record.PromisedBaseReward), string(record.Status), questOutstandingLabel(row.Outstanding), record.Title, patron),
			DetailTitle: record.Title,
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

func loadTUIQuestRows(ctx context.Context, loader TUIDataLoader) ([]tuiQuestRow, error) {
	quests, err := loader.ListQuests(ctx)
	if err != nil {
		return nil, err
	}

	receivables, err := loader.GetQuestReceivables(ctx)
	if err != nil {
		return nil, err
	}

	receivablesByQuestID := make(map[string]report.QuestReceivableRow, len(receivables))
	for _, row := range receivables {
		receivablesByQuestID[row.QuestID] = row
	}

	rows := make([]tuiQuestRow, 0, len(quests))
	for index := range quests {
		record := quests[index]
		receivable, ok := receivablesByQuestID[record.ID]
		row := tuiQuestRow{
			Record:    record,
			OffLedger: record.Status == ledger.QuestStatusOffered || record.Status == ledger.QuestStatusAccepted,
		}
		if ok {
			row.TotalPaid = receivable.TotalPaid
			row.Outstanding = receivable.Outstanding
		} else if record.Status == ledger.QuestStatusPaid {
			row.TotalPaid = record.PromisedBaseReward
		}

		row.Collectible = row.Outstanding > 0 && questCollectibleStatus(record.Status)
		rows = append(rows, row)
	}

	return rows, nil
}

func findTUIQuestRow(rows []tuiQuestRow, questID string) (tuiQuestRow, bool) {
	for index := range rows {
		if rows[index].Record.ID == questID {
			return rows[index], true
		}
	}

	return tuiQuestRow{}, false
}

func questAccountingState(row *tuiQuestRow) string {
	if row == nil {
		return ""
	}

	switch row.Record.Status {
	case ledger.QuestStatusOffered, ledger.QuestStatusAccepted:
		return "off-ledger promise"
	case ledger.QuestStatusCompleted, ledger.QuestStatusCollectible:
		return "collectible but unpaid"
	case ledger.QuestStatusPartiallyPaid:
		return "partially collected receivable"
	case ledger.QuestStatusPaid:
		return "fully collected"
	case ledger.QuestStatusDefaulted:
		return "written off"
	case ledger.QuestStatusVoided:
		return "closed without collection"
	default:
		return string(row.Record.Status)
	}
}

func questOutstandingLabel(value int64) string {
	if value <= 0 {
		return "-"
	}

	return currency.FormatAmount(value) + " due"
}

func questCollectibleStatus(status ledger.QuestStatus) bool {
	switch status {
	case ledger.QuestStatusCompleted, ledger.QuestStatusCollectible, ledger.QuestStatusPartiallyPaid:
		return true
	default:
		return false
	}
}
