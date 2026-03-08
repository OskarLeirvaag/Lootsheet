package journal

import (
	"context"
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

// RunPost posts a journal entry and writes the CLI output.
func RunPost(ctx context.Context, hctx ledger.HandlerContext, input ledger.JournalPostInput) error {
	result, err := PostJournalEntry(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Posted journal entry #%d\nDate: %s\nDescription: %s\nLines: %d\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		result.LineCount,
		tools.FormatAmount(result.DebitTotal),
		tools.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write journal output: %w", err)
	}

	return nil
}

// BuildJournalPostInput converts parsed journal post flag values into a post input.
func BuildJournalPostInput(entryDate string, description string, debitSpecs []string, creditSpecs []string) (ledger.JournalPostInput, error) {
	lines := make([]ledger.JournalLineInput, 0, len(debitSpecs)+len(creditSpecs))
	for _, spec := range debitSpecs {
		line, err := parseJournalLineSpec(spec, true)
		if err != nil {
			return ledger.JournalPostInput{}, err
		}
		lines = append(lines, line)
	}

	for _, spec := range creditSpecs {
		line, err := parseJournalLineSpec(spec, false)
		if err != nil {
			return ledger.JournalPostInput{}, err
		}
		lines = append(lines, line)
	}

	return ledger.JournalPostInput{
		EntryDate:   entryDate,
		Description: description,
		Lines:       lines,
	}, nil
}

func parseJournalLineSpec(value string, isDebit bool) (ledger.JournalLineInput, error) {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) < 2 {
		return ledger.JournalLineInput{}, fmt.Errorf("journal line %q must use CODE:AMOUNT[:MEMO] format", value)
	}

	accountCode := strings.TrimSpace(parts[0])
	if accountCode == "" {
		return ledger.JournalLineInput{}, fmt.Errorf("journal line %q is missing an account code", value)
	}

	amount, err := tools.ParseAmount(parts[1])
	if err != nil {
		return ledger.JournalLineInput{}, fmt.Errorf("journal line %q has an invalid amount: %w", value, err)
	}

	memo := ""
	if len(parts) == 3 {
		memo = strings.TrimSpace(parts[2])
	}

	line := ledger.JournalLineInput{
		AccountCode: accountCode,
		Memo:        memo,
	}
	if isDebit {
		line.DebitAmount = amount
	} else {
		line.CreditAmount = amount
	}

	return line, nil
}

// RunReverse reverses a journal entry and writes the CLI output.
func RunReverse(ctx context.Context, hctx ledger.HandlerContext, entryID string, date string, description string) error {
	result, err := ReverseJournalEntry(ctx, hctx.DatabasePath, entryID, date, description)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Reversed journal entry as #%d\nDate: %s\nDescription: %s\nLines: %d\nDebits: %s\nCredits: %s\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		result.LineCount,
		tools.FormatAmount(result.DebitTotal),
		tools.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write reversal output: %w", err)
	}

	return nil
}

// RunAccountLedger writes the ledger report for a single account.
func RunAccountLedger(ctx context.Context, hctx ledger.HandlerContext, code string) error {
	report, err := GetAccountLedger(ctx, hctx.DatabasePath, code)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Account: %s %s (%s)\n\n",
		report.AccountCode, report.AccountName, string(report.AccountType),
	); err != nil {
		return fmt.Errorf("write ledger header: %w", err)
	}

	if len(report.Entries) == 0 {
		if _, err := fmt.Fprintln(hctx.Stdout, "No transactions."); err != nil {
			return fmt.Errorf("write empty ledger: %w", err)
		}
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
	if _, err := fmt.Fprintf(hctx.Stdout, headerFmt,
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

		if _, err := fmt.Fprintf(hctx.Stdout, rowFmt,
			entry.EntryDate, entry.EntryNumber, desc, memo, debitStr, creditStr, balanceStr,
		); err != nil {
			return fmt.Errorf("write ledger row: %w", err)
		}
	}

	// Print final balance line.
	balanceFormatted := tools.FormatAmount(report.Balance)
	balanceLabelWidth := 10 + 2 + 5 + 2 + descWidth + 2 + memoWidth + 2 + 20 + 2 + 20 + 2
	if _, err := fmt.Fprintf(hctx.Stdout, "%*s%s\n", balanceLabelWidth-len(balanceFormatted), "Balance: ", balanceFormatted); err != nil {
		return fmt.Errorf("write ledger balance: %w", err)
	}

	return nil
}
