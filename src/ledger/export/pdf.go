package export

import (
	"fmt"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
	"github.com/go-pdf/fpdf"
)

// PDF layout constants (all measurements in mm, landscape A4).
const (
	// Column widths for the trial balance table.
	pdfColCode    = 20
	pdfColName    = 70
	pdfColType    = 30
	pdfColDebits  = 45
	pdfColCredits = 45
	pdfColBalance = 45

	// Row / cell heights.
	pdfRowHeight       = 7
	pdfCellHeightSmall = 5
	pdfCellHeightLarge = 8
	pdfTitleHeight     = 10

	// Font sizes.
	pdfFontSizeTitle    = 16
	pdfFontSizeSubtitle = 11
	pdfFontSizeBody     = 9
	pdfFontSizeSmall    = 8
	pdfFontSizeGroup    = 10

	// Spacing and margins.
	pdfAutoPageBreakMargin = 15
	pdfSeparatorSpacing    = 2
	pdfSectionSpacing      = 4
)

// WriteTrialBalancePDF generates a formatted landscape A4 PDF of the trial
// balance report and writes it to path.
func WriteTrialBalancePDF(path string, tb report.TrialBalanceReport, campaignName string) error {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, pdfAutoPageBreakMargin)

	// Header/footer with page numbers.
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "I", pdfFontSizeSmall)
		pdf.CellFormat(0, pdfCellHeightSmall, fmt.Sprintf("Trial Balance \u2014 %s", campaignName), "", 0, "L", false, 0, "")
		pdf.Ln(pdfCellHeightLarge)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-pdfTitleHeight)
		pdf.SetFont("Arial", "I", pdfFontSizeSmall)
		pdf.CellFormat(0, pdfCellHeightSmall, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	// Title.
	pdf.SetFont("Arial", "B", pdfFontSizeTitle)
	pdf.CellFormat(0, pdfTitleHeight, "Trial Balance", "", 1, "C", false, 0, "")

	// Subtitle: campaign name + date.
	pdf.SetFont("Arial", "", pdfFontSizeSubtitle)
	subtitle := fmt.Sprintf("%s \u2014 %s", campaignName, time.Now().Format("2006-01-02"))
	pdf.CellFormat(0, pdfRowHeight, subtitle, "", 1, "C", false, 0, "")
	pdf.Ln(pdfSectionSpacing)

	// Table column headers.
	writeTableHeader(pdf)

	groups := groupAccountsByType(tb.Accounts)

	for _, acctType := range accountTypeOrder {
		rows, ok := groups[acctType]
		if !ok {
			continue
		}

		label := accountTypeLabel(acctType)

		// Group header row (bold).
		pdf.SetFont("Arial", "B", pdfFontSizeGroup)
		pdf.CellFormat(pdfColCode+pdfColName, pdfRowHeight, label, "1", 0, "L", false, 0, "")
		pdf.CellFormat(pdfColType, pdfRowHeight, "", "1", 0, "", false, 0, "")
		pdf.CellFormat(pdfColDebits, pdfRowHeight, "", "1", 0, "", false, 0, "")
		pdf.CellFormat(pdfColCredits, pdfRowHeight, "", "1", 0, "", false, 0, "")
		pdf.CellFormat(pdfColBalance, pdfRowHeight, "", "1", 1, "", false, 0, "")

		var groupDebits, groupCredits, groupBalance int64

		pdf.SetFont("Arial", "", pdfFontSizeBody)
		for _, acct := range rows {
			pdf.CellFormat(pdfColCode, pdfRowHeight, acct.AccountCode, "1", 0, "L", false, 0, "")
			pdf.CellFormat(pdfColName, pdfRowHeight, acct.AccountName, "1", 0, "L", false, 0, "")
			pdf.CellFormat(pdfColType, pdfRowHeight, string(acct.AccountType), "1", 0, "L", false, 0, "")
			pdf.CellFormat(pdfColDebits, pdfRowHeight, currency.FormatAmount(acct.TotalDebits), "1", 0, "R", false, 0, "")
			pdf.CellFormat(pdfColCredits, pdfRowHeight, currency.FormatAmount(acct.TotalCredits), "1", 0, "R", false, 0, "")
			pdf.CellFormat(pdfColBalance, pdfRowHeight, currency.FormatAmount(acct.Balance), "1", 1, "R", false, 0, "")

			groupDebits += acct.TotalDebits
			groupCredits += acct.TotalCredits
			groupBalance += acct.Balance
		}

		// Subtotal row.
		pdf.SetFont("Arial", "I", pdfFontSizeBody)
		pdf.CellFormat(pdfColCode, pdfRowHeight, "", "1", 0, "", false, 0, "")
		pdf.CellFormat(pdfColName, pdfRowHeight, fmt.Sprintf("Subtotal %s", label), "1", 0, "L", false, 0, "")
		pdf.CellFormat(pdfColType, pdfRowHeight, "", "1", 0, "", false, 0, "")
		pdf.CellFormat(pdfColDebits, pdfRowHeight, currency.FormatAmount(groupDebits), "1", 0, "R", false, 0, "")
		pdf.CellFormat(pdfColCredits, pdfRowHeight, currency.FormatAmount(groupCredits), "1", 0, "R", false, 0, "")
		pdf.CellFormat(pdfColBalance, pdfRowHeight, currency.FormatAmount(groupBalance), "1", 1, "R", false, 0, "")

		// Blank separator.
		pdf.Ln(pdfSeparatorSpacing)
	}

	// Grand totals row (bold).
	pdf.SetFont("Arial", "B", pdfFontSizeGroup)
	pdf.CellFormat(pdfColCode, pdfRowHeight, "", "1", 0, "", false, 0, "")
	pdf.CellFormat(pdfColName, pdfRowHeight, "Grand Total", "1", 0, "L", false, 0, "")
	pdf.CellFormat(pdfColType, pdfRowHeight, "", "1", 0, "", false, 0, "")
	pdf.CellFormat(pdfColDebits, pdfRowHeight, currency.FormatAmount(tb.TotalDebits), "1", 0, "R", false, 0, "")
	pdf.CellFormat(pdfColCredits, pdfRowHeight, currency.FormatAmount(tb.TotalCredits), "1", 0, "R", false, 0, "")
	pdf.CellFormat(pdfColBalance, pdfRowHeight, "", "1", 1, "R", false, 0, "")

	// Balanced status line.
	pdf.Ln(pdfSectionSpacing)
	pdf.SetFont("Arial", "B", pdfFontSizeSubtitle)
	pdf.CellFormat(0, pdfCellHeightLarge, fmt.Sprintf("Balanced: %t", tb.Balanced), "", 1, "L", false, 0, "")

	return pdf.OutputFileAndClose(path)
}

// writeTableHeader writes the column header row for the PDF table.
func writeTableHeader(pdf *fpdf.Fpdf) {
	pdf.SetFont("Arial", "B", pdfFontSizeBody)
	pdf.CellFormat(pdfColCode, pdfRowHeight, "Code", "1", 0, "C", false, 0, "")
	pdf.CellFormat(pdfColName, pdfRowHeight, "Name", "1", 0, "C", false, 0, "")
	pdf.CellFormat(pdfColType, pdfRowHeight, "Type", "1", 0, "C", false, 0, "")
	pdf.CellFormat(pdfColDebits, pdfRowHeight, "Debits", "1", 0, "C", false, 0, "")
	pdf.CellFormat(pdfColCredits, pdfRowHeight, "Credits", "1", 0, "C", false, 0, "")
	pdf.CellFormat(pdfColBalance, pdfRowHeight, "Balance", "1", 1, "C", false, 0, "")
}
