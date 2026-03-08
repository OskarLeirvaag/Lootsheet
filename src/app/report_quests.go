package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

func (a *Application) runQuestReceivables(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "generating quest receivables report",
		slog.String("database_path", a.config.Paths.DatabasePath),
	)

	rows, err := repo.GetQuestReceivables(ctx, a.config.Paths.DatabasePath)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to generate quest receivables report",
			slog.String("error", err.Error()),
		)
		return err
	}

	if len(rows) == 0 {
		if _, err := fmt.Fprintln(a.stdout, "No outstanding quest receivables."); err != nil {
			return fmt.Errorf("write quest receivables output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(a.stdout, "%-24s %-16s %-16s %-16s %-16s %s\n",
		"QUEST", "PATRON", "STATUS", "PROMISED", "PAID", "OUTSTANDING",
	); err != nil {
		return fmt.Errorf("write quest receivables header: %w", err)
	}

	for _, r := range rows {
		if _, err := fmt.Fprintf(a.stdout, "%-24s %-16s %-16s %-16s %-16s %s\n",
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

	a.log.logger.InfoContext(ctx, "generated quest receivables report",
		slog.Int("count", len(rows)),
	)

	return nil
}
