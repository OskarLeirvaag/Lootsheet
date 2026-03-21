package export

import (
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
	"github.com/xuri/excelize/v2"
)

// Excel column widths.
const (
	excelColWidthNarrow  = 8  // Code / Entry# columns
	excelColWidthDate    = 12 // Date columns
	excelColWidthName    = 30 // Name / Description columns
	excelColWidthType    = 12 // Type columns (trial balance)
	excelColWidthAmount  = 15 // Debit / Credit / Balance columns
	excelTitleFontSize   = 14 // Title row font size
	excelMaxSheetNameLen = 31 // Excel sheet name character limit
	excelLastCol         = 5  // 0-based index of the last column (F)
)

// WriteTrialBalanceExcel creates a multi-sheet Excel workbook at path containing
// the trial balance and per-account ledger detail sheets.
//
//nolint:revive // cyclomatic complexity is inherent to multi-step Excel builder
func WriteTrialBalanceExcel(path string, tb report.TrialBalanceReport, ledgers map[string]journal.AccountLedgerReport, campaignName string) error {
	f := excelize.NewFile()
	defer f.Close()

	// ── Styles ──────────────────────────────────────────────────────────

	boldStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return fmt.Errorf("create bold style: %w", err)
	}

	italicStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true},
	})
	if err != nil {
		return fmt.Errorf("create italic style: %w", err)
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: excelTitleFontSize},
	})
	if err != nil {
		return fmt.Errorf("create title style: %w", err)
	}

	// ── Sheet 1: Trial Balance ──────────────────────────────────────────

	const sheet = "Trial Balance"
	// Rename the default sheet.
	defaultSheet := f.GetSheetName(0)
	if err := f.SetSheetName(defaultSheet, sheet); err != nil {
		return fmt.Errorf("rename default sheet: %w", err)
	}

	// Column widths: Code, Name, Type, Debits, Credits, Balance.
	colWidths := map[string]float64{
		"A": excelColWidthNarrow, "B": excelColWidthName, "C": excelColWidthType,
		"D": excelColWidthAmount, "E": excelColWidthAmount, "F": excelColWidthAmount,
	}
	for col, w := range colWidths {
		if err := f.SetColWidth(sheet, col, col, w); err != nil {
			return fmt.Errorf("set column width %s: %w", col, err)
		}
	}

	// Row 1: Title (merged across all columns).
	titleText := fmt.Sprintf("Trial Balance \u2014 %s", campaignName)
	if err := f.MergeCell(sheet, "A1", "F1"); err != nil {
		return fmt.Errorf("merge title cells: %w", err)
	}
	if err := f.SetCellValue(sheet, "A1", titleText); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if err := f.SetCellStyle(sheet, "A1", "F1", titleStyle); err != nil {
		return fmt.Errorf("set title style: %w", err)
	}

	// Row 3: Column headers.
	headers := []string{"Code", "Name", "Type", "Debits", "Credits", "Balance"}
	row := 3
	for col, h := range headers {
		cell := cellRef(col, row)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return fmt.Errorf("set header %s: %w", cell, err)
		}
	}
	if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(excelLastCol, row), boldStyle); err != nil {
		return fmt.Errorf("set header style: %w", err)
	}

	row = 4
	groups := groupAccountsByType(tb.Accounts)

	for _, acctType := range accountTypeOrder {
		rows, ok := groups[acctType]
		if !ok {
			continue
		}

		label := accountTypeLabel(acctType)

		// Group header row.
		if err := f.SetCellValue(sheet, cellRef(1, row), label); err != nil {
			return fmt.Errorf("set group header: %w", err)
		}
		if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(excelLastCol, row), boldStyle); err != nil {
			return fmt.Errorf("set group header style: %w", err)
		}
		row++

		var groupDebits, groupCredits, groupBalance int64

		for _, acct := range rows {
			vals := []any{
				acct.AccountCode,
				acct.AccountName,
				string(acct.AccountType),
				currency.FormatAmount(acct.TotalDebits),
				currency.FormatAmount(acct.TotalCredits),
				currency.FormatAmount(acct.Balance),
			}
			for col, v := range vals {
				if err := f.SetCellValue(sheet, cellRef(col, row), v); err != nil {
					return fmt.Errorf("set account cell: %w", err)
				}
			}
			groupDebits += acct.TotalDebits
			groupCredits += acct.TotalCredits
			groupBalance += acct.Balance
			row++
		}

		// Subtotal row (italic).
		subtotals := []any{
			"",
			fmt.Sprintf("Subtotal %s", label),
			"",
			currency.FormatAmount(groupDebits),
			currency.FormatAmount(groupCredits),
			currency.FormatAmount(groupBalance),
		}
		for col, v := range subtotals {
			if err := f.SetCellValue(sheet, cellRef(col, row), v); err != nil {
				return fmt.Errorf("set subtotal cell: %w", err)
			}
		}
		if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(excelLastCol, row), italicStyle); err != nil {
			return fmt.Errorf("set subtotal style: %w", err)
		}
		row++

		// Blank separator row.
		row++
	}

	// Grand totals row (bold).
	grandTotals := []any{
		"",
		"Grand Total",
		"",
		currency.FormatAmount(tb.TotalDebits),
		currency.FormatAmount(tb.TotalCredits),
		"",
	}
	for col, v := range grandTotals {
		if err := f.SetCellValue(sheet, cellRef(col, row), v); err != nil {
			return fmt.Errorf("set grand total cell: %w", err)
		}
	}
	if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(excelLastCol, row), boldStyle); err != nil {
		return fmt.Errorf("set grand total style: %w", err)
	}
	row++

	// Balanced status row.
	if err := f.SetCellValue(sheet, cellRef(1, row), fmt.Sprintf("Balanced: %t", tb.Balanced)); err != nil {
		return fmt.Errorf("set balanced cell: %w", err)
	}

	// ── Per-account detail sheets ───────────────────────────────────────

	for _, acct := range tb.Accounts {
		lr, ok := ledgers[acct.AccountCode]
		if !ok || len(lr.Entries) == 0 {
			continue
		}

		sheetName := truncateSheetName(fmt.Sprintf("%s %s", lr.AccountCode, lr.AccountName))
		if _, err := f.NewSheet(sheetName); err != nil {
			return fmt.Errorf("create sheet %q: %w", sheetName, err)
		}

		// Column widths for detail sheets.
		detailWidths := map[string]float64{
			"A": excelColWidthNarrow, "B": excelColWidthDate, "C": excelColWidthName,
			"D": excelColWidthAmount, "E": excelColWidthAmount, "F": excelColWidthAmount,
		}
		for col, w := range detailWidths {
			if err := f.SetColWidth(sheetName, col, col, w); err != nil {
				return fmt.Errorf("set detail column width %s: %w", col, err)
			}
		}

		// Row 1: Account title.
		accountTitle := fmt.Sprintf("Account %s \u2014 %s", lr.AccountCode, lr.AccountName)
		if err := f.SetCellValue(sheetName, "A1", accountTitle); err != nil {
			return fmt.Errorf("set account title: %w", err)
		}
		if err := f.SetCellStyle(sheetName, "A1", "A1", boldStyle); err != nil {
			return fmt.Errorf("set account title style: %w", err)
		}

		// Row 2: Type and balance.
		info := fmt.Sprintf("Type: %s  Balance: %s", lr.AccountType, currency.FormatAmount(lr.Balance))
		if err := f.SetCellValue(sheetName, "A2", info); err != nil {
			return fmt.Errorf("set account info: %w", err)
		}

		// Row 4: Column headers.
		detailHeaders := []string{"Entry#", "Date", "Description", "Debit", "Credit", "Running Balance"}
		for col, h := range detailHeaders {
			if err := f.SetCellValue(sheetName, cellRef(col, 4), h); err != nil {
				return fmt.Errorf("set detail header: %w", err)
			}
		}
		if err := f.SetCellStyle(sheetName, cellRef(0, 4), cellRef(excelLastCol, 4), boldStyle); err != nil {
			return fmt.Errorf("set detail header style: %w", err)
		}

		// Entry rows.
		dRow := 5
		for _, e := range lr.Entries {
			vals := []any{
				e.EntryNumber,
				e.EntryDate,
				e.Description,
				currency.FormatAmount(e.DebitAmount),
				currency.FormatAmount(e.CreditAmount),
				currency.FormatAmount(e.RunningBalance),
			}
			for col, v := range vals {
				if err := f.SetCellValue(sheetName, cellRef(col, dRow), v); err != nil {
					return fmt.Errorf("set entry cell: %w", err)
				}
			}
			dRow++
		}
	}

	return f.SaveAs(path)
}

// cellRef returns the Excel cell reference (e.g., "A1") for a 0-based column
// index and 1-based row number.
func cellRef(col, row int) string {
	const baseCol = 'A'
	return fmt.Sprintf("%c%d", baseCol+rune(col), row) //nolint:gosec // col is always 0-5
}

// truncateSheetName truncates a sheet name to 31 characters (the Excel maximum)
// and replaces characters that are invalid in Excel sheet names.
func truncateSheetName(name string) string {
	// Replace invalid sheet name characters.
	replacer := strings.NewReplacer(
		":", "-", "\\", "-", "/", "-",
		"?", "", "*", "", "[", "(", "]", ")",
	)
	name = replacer.Replace(name)

	if len(name) > excelMaxSheetNameLen {
		name = name[:excelMaxSheetNameLen]
	}
	return name
}
