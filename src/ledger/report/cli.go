package report

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/currency"
)

// RunTrialBalance generates and displays the trial balance report.
func RunTrialBalance(ctx context.Context, hctx ledger.HandlerContext) error {
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
			currency.FormatAmount(row.TotalDebits),
			currency.FormatAmount(row.TotalCredits),
			currency.FormatAmount(row.Balance),
		)
	}

	fmt.Fprintf(tw, "\t\t\t\t%s\t%s\n", "---", "---")

	balanceLabel := "BALANCED"
	if !report.Balanced {
		diff := report.TotalDebits - report.TotalCredits
		if diff < 0 {
			diff = -diff
		}
		balanceLabel = fmt.Sprintf("UNBALANCED (diff: %s)", currency.FormatAmount(diff))
	}

	fmt.Fprintf(tw, "\t\tTotals:\t%s\t%s\t%s\n",
		currency.FormatAmount(report.TotalDebits), currency.FormatAmount(report.TotalCredits), balanceLabel,
	)

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write trial balance table: %w", err)
	}

	return nil
}

// RunQuestReceivables generates and displays the quest receivables report.
func RunQuestReceivables(ctx context.Context, hctx ledger.HandlerContext) error {
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
			currency.FormatAmount(r.PromisedReward),
			currency.FormatAmount(r.TotalPaid),
			currency.FormatAmount(r.Outstanding),
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write quest receivables table: %w", err)
	}

	return nil
}

// RunPromisedQuests generates and displays the promised-but-unearned quest report.
func RunPromisedQuests(ctx context.Context, hctx ledger.HandlerContext) error {
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
			currency.FormatAmount(row.PromisedReward),
			currency.FormatAmount(row.PartialAdvance),
			bonusDisplay,
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write promised quests table: %w", err)
	}

	return nil
}

// RunLootSummary generates and displays the loot summary report.
func RunLootSummary(ctx context.Context, hctx ledger.HandlerContext) error {
	rows, err := GetLootSummary(ctx, hctx.DatabasePath, "loot")
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
			appraisedDisplay = currency.FormatAmount(r.LatestAppraisalValue)
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
			currency.FormatAmount(row.PromisedReward),
			currency.FormatAmount(row.TotalPaid),
			currency.FormatAmount(row.Outstanding),
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write write-off candidates table: %w", err)
	}

	return nil
}
