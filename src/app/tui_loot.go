package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/report"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

func summarizeLoot(rows []report.LootSummaryRow) []string {
	totalQuantity := 0
	recognized := 0
	var appraisedValue int64

	for _, row := range rows {
		totalQuantity += row.Quantity
		appraisedValue += row.LatestAppraisalValue
		if row.Status == ledger.LootStatusRecognized {
			recognized++
		}
	}

	return []string{
		fmt.Sprintf("Tracked items: %d", len(rows)),
		fmt.Sprintf("Recognized: %d", recognized),
		fmt.Sprintf("Total quantity: %d", totalQuantity),
		"Appraised value: " + tools.FormatAmount(appraisedValue),
	}
}

func summarizeAssets(rows []report.LootSummaryRow) []string {
	totalQuantity := 0
	recognized := 0
	var appraisedValue int64

	for _, row := range rows {
		totalQuantity += row.Quantity
		appraisedValue += row.LatestAppraisalValue
		if row.Status == ledger.LootStatusRecognized {
			recognized++
		}
	}

	return []string{
		fmt.Sprintf("Tracked assets: %d", len(rows)),
		fmt.Sprintf("Recognized: %d", recognized),
		fmt.Sprintf("Total quantity: %d", totalQuantity),
		"Appraised value: " + tools.FormatAmount(appraisedValue),
	}
}

func buildLootItems(rows []loot.BrowseItemRecord, today string) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(rows))
	for index := range rows {
		row := &rows[index]
		name := row.Name
		if strings.TrimSpace(row.Source) != "" {
			name = name + " (" + row.Source + ")"
		}

		detailLines := []string{
			"Status: " + string(row.Status),
			fmt.Sprintf("Quantity: %d", row.Quantity),
			"Accounting state: " + lootAccountingState(row),
			"Latest appraisal: " + lootLatestAppraisalText(row),
			fmt.Sprintf("Appraisals tracked: %d", row.AppraisalCount),
		}
		if row.HasRecognizedAppraisal {
			detailLines = append(detailLines, "Recognized value: "+lootRecognizedValueText(row))
		}
		if lootSellable(row) {
			detailLines = append(detailLines, "Sale state: sellable from recognized basis")
		}
		if row.LatestAppraisal != nil && row.LatestAppraisal.AppraisedAt != "" {
			detailLines = append(detailLines, "Appraised on: "+row.LatestAppraisal.AppraisedAt)
		}
		if row.LatestAppraisal != nil && strings.TrimSpace(row.LatestAppraisal.Appraiser) != "" {
			detailLines = append(detailLines, "Appraiser: "+row.LatestAppraisal.Appraiser)
		}
		if strings.TrimSpace(row.Source) != "" {
			detailLines = append(detailLines, "Source: "+row.Source)
		}
		if strings.TrimSpace(row.Holder) != "" {
			detailLines = append(detailLines, "Holder: "+row.Holder)
		}
		if row.LatestAppraisal != nil && strings.TrimSpace(row.LatestAppraisal.Notes) != "" {
			detailLines = append(detailLines, "Appraisal notes: "+row.LatestAppraisal.Notes)
		}
		if strings.TrimSpace(row.Notes) != "" {
			detailLines = append(detailLines, "Item notes: "+row.Notes)
		}

		var actions []render.ItemActionData
		actions = append(actions, render.ItemActionData{
			Trigger:      render.ActionEdit,
			ID:           tuiCommandLootUpdate,
			Label:        "u edit",
			Mode:         render.ItemActionModeCompose,
			ComposeMode:  "loot",
			ComposeTitle: "Edit Loot",
			ComposeFields: map[string]string{
				"name":     row.Name,
				"source":   row.Source,
				"quantity": strconv.Itoa(row.Quantity),
				"holder":   row.Holder,
				"notes":    row.Notes,
			},
		})
		if lootRecognizable(row) {
			appraisalDetail := "This uses the latest appraisal."
			if row.AppraisalCount > 1 {
				appraisalDetail = fmt.Sprintf("This uses the latest of %d appraisals.", row.AppraisalCount)
			}

			actions = append(actions, render.ItemActionData{
				Trigger:      render.ActionRecognize,
				ID:           tuiCommandLootRecognize,
				Label:        "n recognize",
				Mode:         render.ItemActionModeConfirm,
				ConfirmTitle: fmt.Sprintf("Recognize %q?", row.Name),
				ConfirmLines: []string{
					"Latest appraisal: " + lootLatestAppraisalText(row),
					"Appraisal date: " + row.LatestAppraisal.AppraisedAt,
					"Recognition date: " + today,
					appraisalDetail,
					"A new posted journal entry will be created.",
					fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Recognize loot appraisal: %s", row.LatestAppraisal.ID)),
				},
			})
		} else if lootSellable(row) {
			actions = append(actions, render.ItemActionData{
				Trigger:     render.ActionSell,
				ID:          tuiCommandLootSell,
				Label:       "s sell",
				Mode:        render.ItemActionModeInput,
				InputTitle:  fmt.Sprintf("Sell %q?", row.Name),
				InputPrompt: "Sale amount",
				InputHelp: []string{
					"Sale date: " + today,
					"Recognized value: " + lootRecognizedValueText(row),
					"Enter sale proceeds in GP/SP/CP format.",
					fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Sale of loot item: %s", row.ID)),
				},
				Placeholder: tools.FormatAmount(row.RecognizedAppraisalValue),
			})
		}
		actions = append(actions, render.ItemActionData{
			Trigger:      render.ActionToggle,
			ID:           tuiCommandLootTransferToAsset,
			Label:        "t to assets",
			Mode:         render.ItemActionModeConfirm,
			ConfirmTitle: fmt.Sprintf("Transfer %q to assets?", row.Name),
			ConfirmLines: []string{
				"Move this item to the asset register for keeping.",
				"The item will appear in the Assets tab.",
			},
		})

		items = append(items, render.ListItemData{
			Key:         row.ID,
			Row:         fmt.Sprintf("%-12s qty:%-3d %-11s %s", lootRowAppraisalLabel(row), row.Quantity, string(row.Status), name),
			DetailTitle: row.Name,
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

func buildAssetItems(rows []loot.BrowseItemRecord, today string) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(rows))
	for index := range rows {
		row := &rows[index]
		name := row.Name
		if strings.TrimSpace(row.Source) != "" {
			name = name + " (" + row.Source + ")"
		}

		detailLines := []string{
			"Status: " + string(row.Status),
			fmt.Sprintf("Quantity: %d", row.Quantity),
			"Accounting state: " + lootAccountingState(row),
			"Latest appraisal: " + lootLatestAppraisalText(row),
			fmt.Sprintf("Appraisals tracked: %d", row.AppraisalCount),
		}
		if row.HasRecognizedAppraisal {
			detailLines = append(detailLines, "Recognized value: "+lootRecognizedValueText(row))
		}
		if row.LatestAppraisal != nil && row.LatestAppraisal.AppraisedAt != "" {
			detailLines = append(detailLines, "Appraised on: "+row.LatestAppraisal.AppraisedAt)
		}
		if row.LatestAppraisal != nil && strings.TrimSpace(row.LatestAppraisal.Appraiser) != "" {
			detailLines = append(detailLines, "Appraiser: "+row.LatestAppraisal.Appraiser)
		}
		if strings.TrimSpace(row.Source) != "" {
			detailLines = append(detailLines, "Source: "+row.Source)
		}
		if strings.TrimSpace(row.Holder) != "" {
			detailLines = append(detailLines, "Holder: "+row.Holder)
		}
		if row.LatestAppraisal != nil && strings.TrimSpace(row.LatestAppraisal.Notes) != "" {
			detailLines = append(detailLines, "Appraisal notes: "+row.LatestAppraisal.Notes)
		}
		if strings.TrimSpace(row.Notes) != "" {
			detailLines = append(detailLines, "Item notes: "+row.Notes)
		}

		var actions []render.ItemActionData
		actions = append(actions, render.ItemActionData{
			Trigger:      render.ActionEdit,
			ID:           tuiCommandAssetUpdate,
			Label:        "u edit",
			Mode:         render.ItemActionModeCompose,
			ComposeMode:  "asset",
			ComposeTitle: "Edit Asset",
			ComposeFields: map[string]string{
				"name":     row.Name,
				"source":   row.Source,
				"quantity": strconv.Itoa(row.Quantity),
				"holder":   row.Holder,
				"notes":    row.Notes,
			},
		})
		if lootRecognizable(row) {
			appraisalDetail := "This uses the latest appraisal."
			if row.AppraisalCount > 1 {
				appraisalDetail = fmt.Sprintf("This uses the latest of %d appraisals.", row.AppraisalCount)
			}

			actions = append(actions, render.ItemActionData{
				Trigger:      render.ActionRecognize,
				ID:           tuiCommandAssetRecognize,
				Label:        "n recognize",
				Mode:         render.ItemActionModeConfirm,
				ConfirmTitle: fmt.Sprintf("Recognize %q?", row.Name),
				ConfirmLines: []string{
					"Latest appraisal: " + lootLatestAppraisalText(row),
					"Appraisal date: " + row.LatestAppraisal.AppraisedAt,
					"Recognition date: " + today,
					appraisalDetail,
					"A new posted journal entry will be created.",
					fmt.Sprintf("Description defaults to %q.", fmt.Sprintf("Recognize loot appraisal: %s", row.LatestAppraisal.ID)),
				},
			})
		}
		actions = append(actions, render.ItemActionData{
			Trigger:      render.ActionToggle,
			ID:           tuiCommandAssetTransferToLoot,
			Label:        "t to loot",
			Mode:         render.ItemActionModeConfirm,
			ConfirmTitle: fmt.Sprintf("Transfer %q to loot?", row.Name),
			ConfirmLines: []string{
				"Move this item to the loot register for sale.",
				"The item will appear in the Loot tab.",
			},
		})

		items = append(items, render.ListItemData{
			Key:         row.ID,
			Row:         fmt.Sprintf("%-12s qty:%-3d %-11s %s", lootRowAppraisalLabel(row), row.Quantity, string(row.Status), name),
			DetailTitle: row.Name,
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

func findBrowseLootItem(rows []loot.BrowseItemRecord, itemID string) (loot.BrowseItemRecord, bool) {
	for index := range rows {
		if rows[index].ID == itemID {
			return rows[index], true
		}
	}

	return loot.BrowseItemRecord{}, false
}

func lootAccountingState(row *loot.BrowseItemRecord) string {
	if row == nil {
		return ""
	}

	switch row.Status {
	case ledger.LootStatusRecognized:
		return "on-ledger recognized inventory"
	case ledger.LootStatusHeld:
		if row.LatestAppraisal != nil && row.LatestAppraisal.AppraisedValue >= 1 {
			return "appraised but off-ledger"
		}
		return "held off-ledger"
	default:
		return string(row.Status)
	}
}

func lootRecognizable(row *loot.BrowseItemRecord) bool {
	if row == nil || row.Status != ledger.LootStatusHeld || row.LatestAppraisal == nil {
		return false
	}

	if row.LatestAppraisal.AppraisedValue < 1 {
		return false
	}

	return strings.TrimSpace(row.LatestAppraisal.RecognizedEntryID) == ""
}

func lootSellable(row *loot.BrowseItemRecord) bool {
	if row == nil || row.Status != ledger.LootStatusRecognized {
		return false
	}

	return row.HasRecognizedAppraisal
}

func lootLatestAppraisalText(row *loot.BrowseItemRecord) string {
	if row == nil || row.LatestAppraisal == nil {
		return "Unknown / none"
	}

	return tools.FormatAmount(row.LatestAppraisal.AppraisedValue)
}

func lootRecognizedValueText(row *loot.BrowseItemRecord) string {
	if row == nil || !row.HasRecognizedAppraisal {
		return "Unknown / none"
	}

	return tools.FormatAmount(row.RecognizedAppraisalValue)
}

func lootRowAppraisalLabel(row *loot.BrowseItemRecord) string {
	if row == nil || row.LatestAppraisal == nil {
		return "unknown"
	}

	return tools.FormatAmount(row.LatestAppraisal.AppraisedValue)
}
