package journal

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/OskarLeirvaag/Lootsheet/src/currency"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// RunPost posts a journal entry and writes the CLI output.
func RunPost(ctx context.Context, hctx ledger.HandlerContext, input ledger.JournalPostInput) error {
	result, err := PostJournalEntry(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	return writePostedEntryOutput(hctx, "Posted journal entry", &result, "")
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

	amount, err := currency.ParseAmount(parts[1])
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
		currency.FormatAmount(result.DebitTotal),
		currency.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write reversal output: %w", err)
	}

	return nil
}

// RunExpense posts a guided expense entry and writes the CLI output.
func RunExpense(ctx context.Context, hctx ledger.HandlerContext, input *ExpenseEntryInput) error {
	result, err := PostExpenseEntry(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	return writePostedEntryOutput(hctx, "Recorded expense as journal entry", &result, "Amount: "+currency.FormatAmount(input.Amount))
}

// RunIncome posts a guided income entry and writes the CLI output.
func RunIncome(ctx context.Context, hctx ledger.HandlerContext, input *IncomeEntryInput) error {
	result, err := PostIncomeEntry(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	return writePostedEntryOutput(hctx, "Recorded income as journal entry", &result, "Amount: "+currency.FormatAmount(input.Amount))
}

// RunCustom posts a guided custom entry and writes the CLI output.
func RunCustom(ctx context.Context, hctx ledger.HandlerContext, input ledger.JournalPostInput) error {
	result, err := PostJournalEntry(ctx, hctx.DatabasePath, input)
	if err != nil {
		return err
	}

	return writePostedEntryOutput(hctx, "Recorded custom entry as journal entry", &result, "")
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

	tw := tabwriter.NewWriter(hctx.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "DATE\t#\tDESCRIPTION\tMEMO\tDEBIT\tCREDIT\tBALANCE")

	for _, entry := range report.Entries {
		debitStr := ""
		if entry.DebitAmount != 0 {
			debitStr = currency.FormatAmount(entry.DebitAmount)
		}
		creditStr := ""
		if entry.CreditAmount != 0 {
			creditStr = currency.FormatAmount(entry.CreditAmount)
		}
		balanceStr := currency.FormatAmount(entry.RunningBalance)

		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			entry.EntryDate, entry.EntryNumber, entry.Description, entry.Memo, debitStr, creditStr, balanceStr,
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write ledger table: %w", err)
	}

	// Print final balance line.
	if _, err := fmt.Fprintf(hctx.Stdout, "Balance: %s\n", currency.FormatAmount(report.Balance)); err != nil {
		return fmt.Errorf("write ledger balance: %w", err)
	}

	return nil
}

func writePostedEntryOutput(hctx ledger.HandlerContext, headline string, result *ledger.PostedJournalEntry, extraLine string) error {
	if result == nil {
		return fmt.Errorf("posted journal result is required")
	}
	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"%s #%d\nDate: %s\nDescription: %s\nLines: %d\nDebits: %s\nCredits: %s\n",
		headline,
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		result.LineCount,
		currency.FormatAmount(result.DebitTotal),
		currency.FormatAmount(result.CreditTotal),
	); err != nil {
		return fmt.Errorf("write journal output: %w", err)
	}

	if strings.TrimSpace(extraLine) == "" {
		return nil
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "%s\n", extraLine); err != nil {
		return fmt.Errorf("write journal extra output: %w", err)
	}

	return nil
}
