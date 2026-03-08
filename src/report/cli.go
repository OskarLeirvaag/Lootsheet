package report

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"
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

	tw := tabwriter.NewWriter(hctx.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CODE\tACCOUNT\tTYPE\tDEBITS\tCREDITS\tBALANCE")

	for _, row := range report.Accounts {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.AccountCode,
			row.AccountName,
			string(row.AccountType),
			tools.FormatAmount(row.TotalDebits),
			tools.FormatAmount(row.TotalCredits),
			tools.FormatAmount(row.Balance),
		)
	}

	fmt.Fprintf(tw, "\t\t\t\t%s\t%s\n", "---", "---")

	balanceLabel := "BALANCED"
	if !report.Balanced {
		diff := report.TotalDebits - report.TotalCredits
		if diff < 0 {
			diff = -diff
		}
		balanceLabel = fmt.Sprintf("UNBALANCED (diff: %s)", tools.FormatAmount(diff))
	}

	fmt.Fprintf(tw, "\t\tTotals:\t%s\t%s\t%s\n",
		tools.FormatAmount(report.TotalDebits), tools.FormatAmount(report.TotalCredits), balanceLabel,
	)

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write trial balance table: %w", err)
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

	tw := tabwriter.NewWriter(hctx.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "QUEST\tPATRON\tSTATUS\tPROMISED\tPAID\tOUTSTANDING")

	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Title,
			r.Patron,
			string(r.Status),
			tools.FormatAmount(r.PromisedReward),
			tools.FormatAmount(r.TotalPaid),
			tools.FormatAmount(r.Outstanding),
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write quest receivables table: %w", err)
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

	tw := tabwriter.NewWriter(hctx.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "QUEST\tPATRON\tSTATUS\tPROMISED\tADVANCE\tBONUS")

	for _, row := range rows {
		bonusDisplay := "-"
		if row.BonusConditions != "" {
			bonusDisplay = row.BonusConditions
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.Title,
			row.Patron,
			string(row.Status),
			tools.FormatAmount(row.PromisedReward),
			tools.FormatAmount(row.PartialAdvance),
			bonusDisplay,
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write promised quests table: %w", err)
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

	tw := tabwriter.NewWriter(hctx.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSOURCE\tSTATUS\tQTY\tAPPRAISED VALUE")

	for _, r := range rows {
		appraisedDisplay := "-"
		if r.LatestAppraisalValue > 0 {
			appraisedDisplay = tools.FormatAmount(r.LatestAppraisalValue)
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n",
			r.Name,
			r.Source,
			string(r.Status),
			r.Quantity,
			appraisedDisplay,
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write loot summary table: %w", err)
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

	return RunWriteOffCandidates(ctx, hctx, WriteOffCandidateFilter{
		AsOfDate:   asOfDate,
		MinAgeDays: minAgeDays,
	})
}

// RunWriteOffCandidates writes the write-off candidates report.
func RunWriteOffCandidates(ctx context.Context, hctx ledger.HandlerContext, filter WriteOffCandidateFilter) error {
	rows, err := GetWriteOffCandidates(ctx, hctx.DatabasePath, WriteOffCandidateFilter{
		AsOfDate:   filter.AsOfDate,
		MinAgeDays: filter.MinAgeDays,
	})
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		if _, err := fmt.Fprintf(hctx.Stdout, "No write-off candidates as of %s.\n", filter.AsOfDate); err != nil {
			return fmt.Errorf("write write-off candidates output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Write-Off Candidates (as of %s, min age %d days)\n\n", filter.AsOfDate, filter.MinAgeDays); err != nil {
		return fmt.Errorf("write write-off candidates title: %w", err)
	}

	tw := tabwriter.NewWriter(hctx.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "QUEST\tPATRON\tSTATUS\tCOMPLETED\tAGE\tPROMISED\tPAID\tOUTSTANDING")

	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			row.Title,
			row.Patron,
			string(row.Status),
			row.CompletedOn,
			row.AgeDays,
			tools.FormatAmount(row.PromisedReward),
			tools.FormatAmount(row.TotalPaid),
			tools.FormatAmount(row.Outstanding),
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write write-off candidates table: %w", err)
	}

	return nil
}
