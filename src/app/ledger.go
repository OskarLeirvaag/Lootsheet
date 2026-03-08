package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

func (a *Application) runAccountLedger(ctx context.Context, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account ledger", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, usageText)
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("--code is required\n\n%s", usageText)
	}

	a.log.logger.InfoContext(ctx, "fetching account ledger",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("code", code),
	)

	report, err := repo.GetAccountLedger(ctx, a.config.Paths.DatabasePath, code)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to fetch account ledger", slog.String("error", err.Error()))
		return err
	}

	if _, err := fmt.Fprintf(a.stdout, "Account: %s %s (%s)\n\n",
		report.AccountCode, report.AccountName, string(report.AccountType),
	); err != nil {
		return fmt.Errorf("write ledger header: %w", err)
	}

	if len(report.Entries) == 0 {
		if _, err := fmt.Fprintln(a.stdout, "No transactions."); err != nil {
			return fmt.Errorf("write empty ledger: %w", err)
		}
		a.log.logger.InfoContext(ctx, "account ledger is empty", slog.String("code", code))
		return nil
	}

	// Calculate column widths for description and memo based on data.
	descWidth := 20
	memoWidth := 16
	for _, entry := range report.Entries {
		if len(entry.Description) > descWidth {
			descWidth = len(entry.Description)
		}
		if len(entry.Memo) > memoWidth {
			memoWidth = len(entry.Memo)
		}
	}

	// Cap widths at reasonable maximums.
	if descWidth > 40 {
		descWidth = 40
	}
	if memoWidth > 30 {
		memoWidth = 30
	}

	// Print header.
	headerFmt := fmt.Sprintf("%%-%ds  %%-5s  %%-%ds  %%-%ds  %%20s  %%20s  %%20s\n", 10, descWidth, memoWidth)
	if _, err := fmt.Fprintf(a.stdout, headerFmt,
		"DATE", "#", "DESCRIPTION", "MEMO", "DEBIT", "CREDIT", "BALANCE",
	); err != nil {
		return fmt.Errorf("write ledger column headers: %w", err)
	}

	// Print entries.
	rowFmt := fmt.Sprintf("%%-%ds  %%-5d  %%-%ds  %%-%ds  %%20s  %%20s  %%20s\n", 10, descWidth, memoWidth)
	for _, entry := range report.Entries {
		desc := entry.Description
		if len(desc) > descWidth {
			desc = desc[:descWidth-3] + "..."
		}
		memo := entry.Memo
		if len(memo) > memoWidth {
			memo = memo[:memoWidth-3] + "..."
		}

		debitStr := ""
		if entry.DebitAmount != 0 {
			debitStr = tools.FormatAmount(entry.DebitAmount)
		}
		creditStr := ""
		if entry.CreditAmount != 0 {
			creditStr = tools.FormatAmount(entry.CreditAmount)
		}
		balanceStr := tools.FormatAmount(entry.RunningBalance)

		if _, err := fmt.Fprintf(a.stdout, rowFmt,
			entry.EntryDate, entry.EntryNumber, desc, memo, debitStr, creditStr, balanceStr,
		); err != nil {
			return fmt.Errorf("write ledger row: %w", err)
		}
	}

	// Print final balance line.
	balanceFormatted := tools.FormatAmount(report.Balance)
	balanceLabelWidth := 10 + 2 + 5 + 2 + descWidth + 2 + memoWidth + 2 + 20 + 2 + 20 + 2
	if _, err := fmt.Fprintf(a.stdout, "%*s%s\n", balanceLabelWidth-len(balanceFormatted), "Balance: ", balanceFormatted); err != nil {
		return fmt.Errorf("write ledger balance: %w", err)
	}

	a.log.logger.InfoContext(ctx, "fetched account ledger",
		slog.String("code", code),
		slog.Int("entries", len(report.Entries)),
		slog.Int64("balance", report.Balance),
	)

	return nil
}
