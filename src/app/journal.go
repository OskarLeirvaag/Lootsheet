package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
	"github.com/OskarLeirvaag/Lootsheet/src/service"
)

func (a *Application) runJournal(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing journal subcommand\n\n%s", usageText)
	}

	switch args[0] {
	case "post":
		return a.runJournalPost(ctx, args[1:])
	default:
		return fmt.Errorf("unknown journal subcommand %q\n\n%s", args[0], usageText)
	}
}

func (a *Application) runJournalPost(ctx context.Context, args []string) error {
	input, err := parseJournalPostArgs(args)
	if err != nil {
		return err
	}

	a.log.logger.InfoContext(
		ctx,
		"posting journal entry",
		slog.String("database_path", a.config.Paths.DatabasePath),
		slog.String("entry_date", input.EntryDate),
		slog.String("description", input.Description),
		slog.Int("line_count", len(input.Lines)),
	)

	result, err := repo.PostJournalEntry(ctx, a.config.Paths.DatabasePath, input)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to post journal entry", slog.String("error", err.Error()))
		return err
	}

	a.log.logger.InfoContext(
		ctx,
		"posted journal entry",
		slog.Int("entry_number", result.EntryNumber),
		slog.Int("line_count", result.LineCount),
		slog.Int64("debit_total", result.DebitTotal),
		slog.Int64("credit_total", result.CreditTotal),
	)

	if _, err := fmt.Fprintf(
		a.stdout,
		"Posted journal entry #%d\nDate: %s\nDescription: %s\nLines: %d\nDebits: %d\nCredits: %d\n",
		result.EntryNumber,
		result.EntryDate,
		result.Description,
		result.LineCount,
		result.DebitTotal,
		result.CreditTotal,
	); err != nil {
		return fmt.Errorf("write journal output: %w", err)
	}

	return nil
}

func parseJournalPostArgs(args []string) (service.JournalPostInput, error) {
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
		return service.JournalPostInput{}, fmt.Errorf("%s\n\n%s", err, journalPostUsageText)
	}

	if flagSet.NArg() > 0 {
		return service.JournalPostInput{}, fmt.Errorf("unexpected journal post arguments: %s\n\n%s", strings.Join(flagSet.Args(), " "), journalPostUsageText)
	}

	lines := make([]service.JournalLineInput, 0, len(debitSpecs.values)+len(creditSpecs.values))
	for _, spec := range debitSpecs.values {
		line, err := parseJournalLineSpec(spec, true)
		if err != nil {
			return service.JournalPostInput{}, fmt.Errorf("%s\n\n%s", err, journalPostUsageText)
		}
		lines = append(lines, line)
	}

	for _, spec := range creditSpecs.values {
		line, err := parseJournalLineSpec(spec, false)
		if err != nil {
			return service.JournalPostInput{}, fmt.Errorf("%s\n\n%s", err, journalPostUsageText)
		}
		lines = append(lines, line)
	}

	return service.JournalPostInput{
		EntryDate:   entryDate,
		Description: description,
		Lines:       lines,
	}, nil
}

func parseJournalLineSpec(value string, isDebit bool) (service.JournalLineInput, error) {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) < 2 {
		return service.JournalLineInput{}, fmt.Errorf("journal line %q must use CODE:AMOUNT[:MEMO] format", value)
	}

	accountCode := strings.TrimSpace(parts[0])
	if accountCode == "" {
		return service.JournalLineInput{}, fmt.Errorf("journal line %q is missing an account code", value)
	}

	amount, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return service.JournalLineInput{}, fmt.Errorf("journal line %q has an invalid amount", value)
	}

	memo := ""
	if len(parts) == 3 {
		memo = strings.TrimSpace(parts[2])
	}

	line := service.JournalLineInput{
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

const journalPostUsageText = `LootSheet CLI

Usage:
  lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]

Examples:
  lootsheet journal post --date 2026-03-08 --description "Restock arrows" --debit 5100:25:Quiver refill --credit 1000:25
  lootsheet journal post --date 2026-03-08 --description "Quest reward earned" --debit 1100:100 --credit 4000:100
`
