package account

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// RunList writes the account listing to the handler context stdout.
func RunList(ctx context.Context, hctx ledger.HandlerContext) error {
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
