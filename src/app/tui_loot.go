package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func summarizeItemRegister(rows []report.LootSummaryRow, label string) []string {
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
		fmt.Sprintf("Tracked %s: %d", label, len(rows)),
		fmt.Sprintf("Recognized: %d", recognized),
		fmt.Sprintf("Total quantity: %d", totalQuantity),
		"Appraised value: " + currency.FormatAmount(appraisedValue),
	}
}

func buildItemDetailLines(row *loot.BrowseItemRecord, includeSaleState bool) []string {
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
	if includeSaleState && lootSellable(row) {
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
	return detailLines
}

func buildLootItems(rows []loot.BrowseItemRecord, today string) []render.ListItemData {
	items := make([]render.ListItemData, 0, len(rows))
	for index := range rows {
		row := &rows[index]
		name := row.Name
		if strings.TrimSpace(row.Source) != "" {
			name = name + " (" + row.Source + ")"
		}

		detailLines := buildItemDetailLines(row, true)

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
		if row.Status == ledger.LootStatusHeld {
			actions = append(actions, lootAppraiseAction(row, tuiCommandLootAppraise, today))
		}
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
				Placeholder: currency.FormatAmount(row.RecognizedAppraisalValue),
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
			Row:         lootRowLabel(row, name),
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

		detailLines := buildItemDetailLines(row, false)

		// Append template lines to detail panel if present.
		if len(row.TemplateLines) > 0 {
			detailLines = append(detailLines, "", "Entry template:")
			for _, tl := range row.TemplateLines {
				sideLabel := "Dr (to)"
				if tl.Side == "credit" {
					sideLabel = "Cr (from)"
				}
				amtLabel := ""
				if strings.TrimSpace(tl.Amount) != "" {
					amtLabel = " " + tl.Amount
				}
				detailLines = append(detailLines, fmt.Sprintf("  %s %s%s", sideLabel, tl.AccountCode, amtLabel))
			}
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
		if row.Status == ledger.LootStatusHeld {
			actions = append(actions, lootAppraiseAction(row, tuiCommandAssetAppraise, today))
		}
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

		// Edit template — always available on assets.
		actions = append(actions, render.ItemActionData{
			Trigger:      render.ActionEditTemplate,
			ID:           tuiCommandAssetTemplateSave,
			Label:        "f template",
			Mode:         render.ItemActionModeCompose,
			ComposeMode:  "asset_template",
			ComposeTitle: "Edit Entry Template",
			ComposeLines: assetTemplateToCommandLines(row.TemplateLines),
		})

		// Execute template — only if template lines exist.
		if len(row.TemplateLines) > 0 {
			actions = append(actions, render.ItemActionData{
				Trigger:      render.ActionExecuteTemplate,
				ID:           tuiCommandCreateCustom,
				Label:        "x execute",
				Mode:         render.ItemActionModeCompose,
				ComposeMode:  "custom_from_template",
				ComposeTitle: row.Name,
				ComposeLines: assetTemplateToExecuteLines(row.TemplateLines),
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
			Row:         assetRowLabel(row, name),
			DetailTitle: row.Name,
			DetailLines: detailLines,
			Actions:     actions,
		})
	}

	return items
}

func lootRowLabel(row *loot.BrowseItemRecord, name string) string {
	qtyPart := "       "
	if row.Quantity > 1 {
		qtyPart = fmt.Sprintf("qty:%-3d", row.Quantity)
	}
	return fmt.Sprintf("%-12s %s %-11s %s", lootRowAppraisalLabel(row), qtyPart, string(row.Status), name)
}

func assetRowLabel(row *loot.BrowseItemRecord, name string) string {
	tpl := "   "
	if len(row.TemplateLines) > 0 {
		tpl = "[T]"
	}
	holder := "—"
	if h := strings.TrimSpace(row.Holder); h != "" {
		holder = h
	}
	return fmt.Sprintf("%-12s %-11s %-4s %-12s %s", lootRowAppraisalLabel(row), string(row.Status), tpl, holder, name)
}

func assetTemplateToCommandLines(lines []loot.AssetTemplateLineRecord) []render.CommandLine {
	if len(lines) == 0 {
		return nil
	}
	result := make([]render.CommandLine, len(lines))
	for i, line := range lines {
		result[i] = render.CommandLine{
			Side:        line.Side,
			AccountCode: line.AccountCode,
			Amount:      line.Amount,
		}
	}
	return result
}

func assetTemplateToExecuteLines(lines []loot.AssetTemplateLineRecord) []render.CommandLine {
	if len(lines) == 0 {
		return nil
	}
	result := make([]render.CommandLine, len(lines))
	for i, line := range lines {
		result[i] = render.CommandLine{
			Side:        line.Side,
			AccountCode: line.AccountCode,
			Amount:      line.Amount,
		}
	}
	return result
}

func lootAppraiseAction(row *loot.BrowseItemRecord, commandID string, today string) render.ItemActionData {
	placeholder := "10GP"
	helpLines := []string{
		"Appraisal date: " + today,
		"Enter appraised value in GP/SP/CP format.",
	}
	if row.LatestAppraisal != nil && row.LatestAppraisal.AppraisedValue > 0 {
		placeholder = currency.FormatAmount(row.LatestAppraisal.AppraisedValue)
		helpLines = append(helpLines, "Current appraisal: "+placeholder, "This will replace the current appraisal.")
	}
	return render.ItemActionData{
		Trigger:     render.ActionAppraise,
		ID:          commandID,
		Label:       "p appraise",
		Mode:        render.ItemActionModeInput,
		InputTitle:  fmt.Sprintf("Appraise %q", row.Name),
		InputPrompt: "Appraised value",
		InputHelp:   helpLines,
		Placeholder: placeholder,
	}
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

	return currency.FormatAmount(row.LatestAppraisal.AppraisedValue)
}

func lootRecognizedValueText(row *loot.BrowseItemRecord) string {
	if row == nil || !row.HasRecognizedAppraisal {
		return "Unknown / none"
	}

	return currency.FormatAmount(row.RecognizedAppraisalValue)
}

func lootRowAppraisalLabel(row *loot.BrowseItemRecord) string {
	if row == nil || row.LatestAppraisal == nil {
		return "—"
	}

	return currency.FormatAmount(row.LatestAppraisal.AppraisedValue)
}
