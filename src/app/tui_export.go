package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/export"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
)

func handleExportCommand(ctx context.Context, commandID string, loader TUIDataLoader) (render.CommandResult, error) {
	tb, err := loader.GetTrialBalance(ctx)
	if err != nil {
		return render.CommandResult{}, fmt.Errorf("load trial balance: %w", err)
	}

	accounts, err := loader.ListAccounts(ctx)
	if err != nil {
		return render.CommandResult{}, fmt.Errorf("load accounts: %w", err)
	}

	ledgers := make(map[string]journal.AccountLedgerReport, len(accounts))
	for _, acct := range accounts {
		if rpt, ledgerErr := loader.GetAccountLedger(ctx, acct.Code); ledgerErr == nil {
			ledgers[acct.Code] = rpt
		}
	}

	campaignName := loader.CampaignName()
	timestamp := time.Now().Format("2006-01-02_150405")

	switch commandID {
	case tuiCommandExportCSV:
		filename := fmt.Sprintf("trial_balance_%s.csv", timestamp)
		f, err := os.Create(filepath.Clean(filename))
		if err != nil {
			return render.CommandResult{}, fmt.Errorf("create CSV file: %w", err)
		}
		defer f.Close() //nolint:errcheck // best-effort close on export file

		if err := export.WriteTrialBalanceCSV(f, tb, ledgers); err != nil {
			return render.CommandResult{}, fmt.Errorf("write CSV: %w", err)
		}

		abs, _ := filepath.Abs(filename)
		return render.CommandResult{
			Status: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Exported CSV: %s", abs)},
		}, nil

	case tuiCommandExportExcel:
		filename := fmt.Sprintf("trial_balance_%s.xlsx", timestamp)
		if err := export.WriteTrialBalanceExcel(filename, tb, ledgers, campaignName); err != nil {
			return render.CommandResult{}, fmt.Errorf("write Excel: %w", err)
		}

		abs, _ := filepath.Abs(filename)
		return render.CommandResult{
			Status: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Exported Excel: %s", abs)},
		}, nil

	case tuiCommandExportPDF:
		filename := fmt.Sprintf("trial_balance_%s.pdf", timestamp)
		if err := export.WriteTrialBalancePDF(filename, tb, campaignName); err != nil {
			return render.CommandResult{}, fmt.Errorf("write PDF: %w", err)
		}

		abs, _ := filepath.Abs(filename)
		return render.CommandResult{
			Status: render.StatusMessage{Level: render.StatusSuccess, Text: fmt.Sprintf("Exported PDF: %s", abs)},
		}, nil

	default:
		return render.CommandResult{}, fmt.Errorf("unsupported export command %q", commandID)
	}
}
