package account

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// HandleList writes the account listing to the handler context stdout.
func HandleList(ctx context.Context, hctx ledger.HandlerContext) error {
	accounts, err := ListAccounts(ctx, hctx.DatabasePath)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(hctx.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CODE\tTYPE\tACTIVE\tNAME")

	for _, account := range accounts {
		activeLabel := "no"
		if account.Active {
			activeLabel = "yes"
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			account.Code,
			string(account.Type),
			activeLabel,
			account.Name,
		)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write accounts table: %w", err)
	}

	return nil
}

// HandleCreate parses flags and creates a new account.
func HandleCreate(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var code, name, accountType string

	flagSet := flag.NewFlagSet("account create", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")
	flagSet.StringVar(&name, "name", "", "account name")
	flagSet.StringVar(&accountType, "type", "", "account type (asset, liability, equity, income, expense)")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	return RunCreate(ctx, hctx, code, name, ledger.AccountType(accountType))
}

// RunCreate creates a new account and writes the CLI output.
func RunCreate(ctx context.Context, hctx ledger.HandlerContext, code string, name string, accountType ledger.AccountType) error {
	result, err := CreateAccount(ctx, hctx.DatabasePath, code, name, accountType)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		hctx.Stdout,
		"Created account %s\nCode: %s\nName: %s\nType: %s\n",
		result.ID,
		result.Code,
		result.Name,
		string(result.Type),
	); err != nil {
		return fmt.Errorf("write account output: %w", err)
	}

	return nil
}

// HandleRename parses flags and renames an account.
func HandleRename(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var code, name string

	flagSet := flag.NewFlagSet("account rename", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")
	flagSet.StringVar(&name, "name", "", "new account name")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	return RunRename(ctx, hctx, code, name)
}

// RunRename renames an account and writes the CLI output.
func RunRename(ctx context.Context, hctx ledger.HandlerContext, code string, name string) error {
	if err := RenameAccount(ctx, hctx.DatabasePath, code, name); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Renamed account %s to %q\n", code, name); err != nil {
		return fmt.Errorf("write rename output: %w", err)
	}

	return nil
}

// HandleDeactivate parses flags and deactivates an account.
func HandleDeactivate(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account deactivate", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	return RunDeactivate(ctx, hctx, code)
}

// RunDeactivate deactivates an account and writes the CLI output.
func RunDeactivate(ctx context.Context, hctx ledger.HandlerContext, code string) error {
	if err := DeactivateAccount(ctx, hctx.DatabasePath, code); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Deactivated account %s\n", code); err != nil {
		return fmt.Errorf("write deactivate output: %w", err)
	}

	return nil
}

// HandleActivate parses flags and activates an account.
func HandleActivate(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account activate", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	return RunActivate(ctx, hctx, code)
}

// RunActivate activates an account and writes the CLI output.
func RunActivate(ctx context.Context, hctx ledger.HandlerContext, code string) error {
	if err := ActivateAccount(ctx, hctx.DatabasePath, code); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Activated account %s\n", code); err != nil {
		return fmt.Errorf("write activate output: %w", err)
	}

	return nil
}

// HandleDelete parses flags and deletes an account.
func HandleDelete(ctx context.Context, hctx ledger.HandlerContext, args []string) error {
	var code string

	flagSet := flag.NewFlagSet("account delete", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&code, "code", "", "account code")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	return RunDelete(ctx, hctx, code)
}

// RunDelete deletes an unused account and writes the CLI output.
func RunDelete(ctx context.Context, hctx ledger.HandlerContext, code string) error {
	if err := DeleteAccount(ctx, hctx.DatabasePath, code); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(hctx.Stdout, "Deleted account %s\n", code); err != nil {
		return fmt.Errorf("write delete output: %w", err)
	}

	return nil
}
