package report

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

// HandleTrialBalance generates and displays the trial balance report.
func HandleTrialBalance(ctx context.Context, hctx ledger.HandlerContext) error {
	report, err := GetTrialBalance(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "Trial Balance"); err != nil {
		return fmt.Errorf("write trial balance header: %w", err)
	}

	if _, err := fmt.Fprintln(hctx.Stdout, ""); err != nil {
		return fmt.Errorf("write trial balance blank line: %w", err)
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "%-6s%-21s%-11s  %-20s  %-20s  %-20s\n",
		"CODE", "ACCOUNT", "TYPE", "DEBITS", "CREDITS", "BALANCE",
	); err != nil {
		return fmt.Errorf("write trial balance column header: %w", err)
	}

	for _, row := range report.Accounts {
		if _, err := fmt.Fprintf(hctx.Stdout, "%-6s%-21s%-11s  %-20s  %-20s  %-20s\n",
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

	if _, err := fmt.Fprintf(hctx.Stdout, "%-38s  %-20s  %-20s\n", "", "---", "---"); err != nil {
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

	if _, err := fmt.Fprintf(hctx.Stdout, "%-27s%-11s  %-20s  %-20s  %s\n",
		"", "Totals:", tools.FormatAmount(report.TotalDebits), tools.FormatAmount(report.TotalCredits), balanceLabel,
	); err != nil {
		return fmt.Errorf("write trial balance totals: %w", err)
	}

	return nil
}

// HandleQuestReceivables generates and displays the quest receivables report.
func HandleQuestReceivables(ctx context.Context, hctx ledger.HandlerContext) error {
	rows, err := GetQuestReceivables(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		if _, err := fmt.Fprintln(hctx.Stdout, "No outstanding quest receivables."); err != nil {
			return fmt.Errorf("write quest receivables output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-16s %-16s %-16s %s\n",
		"QUEST", "PATRON", "STATUS", "PROMISED", "PAID", "OUTSTANDING",
	); err != nil {
		return fmt.Errorf("write quest receivables header: %w", err)
	}

	for _, r := range rows {
		if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-16s %-16s %-16s %s\n",
			truncate(r.Title, 24),
			truncate(r.Patron, 16),
			string(r.Status),
			tools.FormatAmount(r.PromisedReward),
			tools.FormatAmount(r.TotalPaid),
			tools.FormatAmount(r.Outstanding),
		); err != nil {
			return fmt.Errorf("write quest receivable row: %w", err)
		}
	}

	return nil
}

// HandlePromisedQuests generates and displays the promised-but-unearned quest report.
func HandlePromisedQuests(ctx context.Context, hctx ledger.HandlerContext) error {
	rows, err := GetPromisedQuests(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		if _, err := fmt.Fprintln(hctx.Stdout, "No promised but unearned quests."); err != nil {
			return fmt.Errorf("write promised quests output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintln(hctx.Stdout, "Promised But Unearned Quests"); err != nil {
		return fmt.Errorf("write promised quests title: %w", err)
	}
	if _, err := fmt.Fprintln(hctx.Stdout, ""); err != nil {
		return fmt.Errorf("write promised quests blank line: %w", err)
	}
	if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-10s %-16s %-16s %s\n",
		"QUEST", "PATRON", "STATUS", "PROMISED", "ADVANCE", "BONUS",
	); err != nil {
		return fmt.Errorf("write promised quests header: %w", err)
	}

	for _, row := range rows {
		bonusDisplay := "-"
		if row.BonusConditions != "" {
			bonusDisplay = truncate(row.BonusConditions, 36)
		}

		if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-10s %-16s %-16s %s\n",
			truncate(row.Title, 24),
			truncate(row.Patron, 16),
			string(row.Status),
			tools.FormatAmount(row.PromisedReward),
			tools.FormatAmount(row.PartialAdvance),
			bonusDisplay,
		); err != nil {
			return fmt.Errorf("write promised quest row: %w", err)
		}
	}

	return nil
}

// HandleLootSummary generates and displays the loot summary report.
func HandleLootSummary(ctx context.Context, hctx ledger.HandlerContext) error {
	rows, err := GetLootSummary(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		if _, err := fmt.Fprintln(hctx.Stdout, "No held or recognized loot items."); err != nil {
			return fmt.Errorf("write loot summary output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-12s %5s  %s\n",
		"NAME", "SOURCE", "STATUS", "QTY", "APPRAISED VALUE",
	); err != nil {
		return fmt.Errorf("write loot summary header: %w", err)
	}

	for _, r := range rows {
		appraisedDisplay := "-"
		if r.LatestAppraisalValue > 0 {
			appraisedDisplay = tools.FormatAmount(r.LatestAppraisalValue)
		}

		if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-12s %5d  %s\n",
			truncate(r.Name, 24),
			truncate(r.Source, 16),
			string(r.Status),
			r.Quantity,
			appraisedDisplay,
		); err != nil {
			return fmt.Errorf("write loot summary row: %w", err)
		}
	}

	return nil
}

// HandleWriteOffCandidates generates and displays the write-off candidates report.
func HandleWriteOffCandidates(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	asOfDate := time.Now().Format(reportDateLayout)
	minAgeDays := 30

	flagSet := flag.NewFlagSet("report writeoff-candidates", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&asOfDate, "as-of", asOfDate, "report date in YYYY-MM-DD")
	flagSet.IntVar(&minAgeDays, "min-age-days", minAgeDays, "minimum completed age in days")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	rows, err := GetWriteOffCandidates(ctx, hctx.DatabasePath, WriteOffCandidateFilter{
		AsOfDate:   asOfDate,
		MinAgeDays: minAgeDays,
	})
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		if _, err := fmt.Fprintf(hctx.Stdout, "No write-off candidates as of %s.\n", asOfDate); err != nil {
			return fmt.Errorf("write write-off candidates output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Write-Off Candidates (as of %s, min age %d days)\n\n", asOfDate, minAgeDays); err != nil {
		return fmt.Errorf("write write-off candidates title: %w", err)
	}
	if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-16s %-10s %4s  %-16s %-16s %s\n",
		"QUEST", "PATRON", "STATUS", "COMPLETED", "AGE", "PROMISED", "PAID", "OUTSTANDING",
	); err != nil {
		return fmt.Errorf("write write-off candidates header: %w", err)
	}

	for _, row := range rows {
		if _, err := fmt.Fprintf(hctx.Stdout, "%-24s %-16s %-16s %-10s %4d  %-16s %-16s %s\n",
			truncate(row.Title, 24),
			truncate(row.Patron, 16),
			string(row.Status),
			row.CompletedOn,
			row.AgeDays,
			tools.FormatAmount(row.PromisedReward),
			tools.FormatAmount(row.TotalPaid),
			tools.FormatAmount(row.Outstanding),
		); err != nil {
			return fmt.Errorf("write write-off candidate row: %w", err)
		}
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "~"
}
