package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

func (a *Application) runReport(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing report subcommand\n\n%s", usageText)
	}

	switch args[0] {
	case "trial-balance":
		return a.runTrialBalance(ctx)
	case "quest-receivables":
		return a.runQuestReceivables(ctx)
	case "loot-summary":
		return a.runLootSummary(ctx)
	default:
		return fmt.Errorf("unknown report subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runTrialBalance(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "generating trial balance", slog.String("database_path", a.config.Paths.DatabasePath))

	report, err := repo.GetTrialBalance(ctx, a.config.Paths.DatabasePath)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to generate trial balance", slog.String("error", err.Error()))
		return err
	}

	if _, err := fmt.Fprintln(a.stdout, "Trial Balance"); err != nil {
		return fmt.Errorf("write trial balance header: %w", err)
	}

	if _, err := fmt.Fprintln(a.stdout, ""); err != nil {
		return fmt.Errorf("write trial balance blank line: %w", err)
	}

	if _, err := fmt.Fprintf(a.stdout, "%-6s%-21s%-11s  %-20s  %-20s  %-20s\n",
		"CODE", "ACCOUNT", "TYPE", "DEBITS", "CREDITS", "BALANCE",
	); err != nil {
		return fmt.Errorf("write trial balance column header: %w", err)
	}

	for _, row := range report.Accounts {
		if _, err := fmt.Fprintf(a.stdout, "%-6s%-21s%-11s  %-20s  %-20s  %-20s\n",
			row.AccountCode,
			truncate(row.AccountName, 20),
			string(row.AccountType),
			tools.FormatAmount(row.TotalDebits),
			tools.FormatAmount(row.TotalCredits),
			tools.FormatAmount(row.Balance),
		); err != nil {
			return fmt.Errorf("write trial balance row: %w", err)
		}
	}

	if _, err := fmt.Fprintf(a.stdout, "%-38s  %-20s  %-20s\n", "", "---", "---"); err != nil {
		return fmt.Errorf("write trial balance separator: %w", err)
	}

	balanceLabel := "BALANCED"
	if !report.Balanced {
		diff := report.TotalDebits - report.TotalCredits
		if diff < 0 {
			diff = -diff
		}
		balanceLabel = fmt.Sprintf("UNBALANCED (diff: %s)", tools.FormatAmount(diff))
	}

	if _, err := fmt.Fprintf(a.stdout, "%-27s%-11s  %-20s  %-20s  %s\n",
		"", "Totals:", tools.FormatAmount(report.TotalDebits), tools.FormatAmount(report.TotalCredits), balanceLabel,
	); err != nil {
		return fmt.Errorf("write trial balance totals: %w", err)
	}

	a.log.logger.InfoContext(ctx, "generated trial balance",
		slog.Int("account_count", len(report.Accounts)),
		slog.Bool("balanced", report.Balanced),
	)

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "~"
}
