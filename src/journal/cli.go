package journal

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

// HandlePost parses flags and posts a journal entry.
func HandlePost(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	input, err := parseJournalPostArgs(args)
	if err != nil {
		return err
	}

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

func parseJournalPostArgs(args []string) (ledger.JournalPostInput, error) {
	var (
		entryDate   string
		description string
		debitSpecs  stringListFlag
		creditSpecs stringListFlag
	)

	flagSet := flag.NewFlagSet("journal post", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&entryDate, "date", "", "entry date in YYYY-MM-DD")
	flagSet.StringVar(&description, "description", "", "journal entry description")
	flagSet.Var(&debitSpecs, "debit", "debit line in CODE:AMOUNT[:MEMO] format")
	flagSet.Var(&creditSpecs, "credit", "credit line in CODE:AMOUNT[:MEMO] format")

	if err := flagSet.Parse(args); err != nil {
		return ledger.JournalPostInput{}, fmt.Errorf("%s\n\n%s", err, JournalPostUsageText)
	}

	if flagSet.NArg() > 0 {
		return ledger.JournalPostInput{}, fmt.Errorf("unexpected journal post arguments: %s\n\n%s", strings.Join(flagSet.Args(), " "), JournalPostUsageText)
	}

	lines := make([]ledger.JournalLineInput, 0, len(debitSpecs.values)+len(creditSpecs.values))
	for _, spec := range debitSpecs.values {
		line, err := parseJournalLineSpec(spec, true)
		if err != nil {
			return ledger.JournalPostInput{}, fmt.Errorf("%s\n\n%s", err, JournalPostUsageText)
		}
		lines = append(lines, line)
	}

	for _, spec := range creditSpecs.values {
		line, err := parseJournalLineSpec(spec, false)
		if err != nil {
			return ledger.JournalPostInput{}, fmt.Errorf("%s\n\n%s", err, JournalPostUsageText)
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

type stringListFlag struct {
	values []string
}

func (f *stringListFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *stringListFlag) Set(value string) error {
	f.values = append(f.values, value)
	return nil
}

// HandleReverse parses flags and reverses a journal entry.
func HandleReverse(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var entryID, date, description string

	flagSet := flag.NewFlagSet("journal reverse", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&entryID, "entry-id", "", "UUID of the journal entry to reverse")
	flagSet.StringVar(&date, "date", "", "reversal date in YYYY-MM-DD")
	flagSet.StringVar(&description, "description", "", "optional reversal description")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("%s\n\n%s", err, JournalReverseUsageText)
	}

	if flagSet.NArg() > 0 {
		return fmt.Errorf("unexpected journal reverse arguments: %s\n\n%s", strings.Join(flagSet.Args(), " "), JournalReverseUsageText)
	}

	if entryID == "" {
		return fmt.Errorf("--entry-id is required\n\n%s", JournalReverseUsageText)
	}

	if date == "" {
		return fmt.Errorf("--date is required\n\n%s", JournalReverseUsageText)
	}

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

// HandleAccountLedger parses flags and displays the account ledger.
func HandleAccountLedger(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account ledger", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("--code is required")
	}

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

// JournalPostUsageText is the help text for the journal post command.
const JournalPostUsageText = `LootSheet CLI

Usage:
  lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]

Amounts accept D&D 5e denominations: PP, GP, EP, SP, CP (case insensitive).
  Mixed:   2GP5SP, 1PP 2GP 3SP 5CP
  Decimal: 5.5GP, 0.5SP
  Bare integer (treated as CP): 100

Examples:
  lootsheet journal post --date 2026-03-08 --description "Restock arrows" --debit 5100:2SP5CP:Quiver refill --credit 1000:2SP5CP
  lootsheet journal post --date 2026-03-08 --description "Quest reward earned" --debit 1100:1GP --credit 4000:1GP
`

// JournalReverseUsageText is the help text for the journal reverse command.
const JournalReverseUsageText = `LootSheet CLI

Usage:
  lootsheet journal reverse --entry-id UUID --date YYYY-MM-DD [--description TEXT]

Examples:
  lootsheet journal reverse --entry-id abc-123 --date 2026-03-09
  lootsheet journal reverse --entry-id abc-123 --date 2026-03-09 --description "Correcting duplicate entry"
`
