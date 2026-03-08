package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/OskarLeirvaag/Lootsheet/src/repo"
	"github.com/OskarLeirvaag/Lootsheet/src/tools"
)

func (a *Application) runLootSummary(ctx context.Context) error {
	a.log.logger.InfoContext(ctx, "generating loot summary report",
		slog.String("database_path", a.config.Paths.DatabasePath),
	)

	rows, err := repo.GetLootSummary(ctx, a.config.Paths.DatabasePath)
	if err != nil {
		a.log.logger.ErrorContext(ctx, "failed to generate loot summary report",
			slog.String("error", err.Error()),
		)
		return err
	}

	if len(rows) == 0 {
		if _, err := fmt.Fprintln(a.stdout, "No held or recognized loot items."); err != nil {
			return fmt.Errorf("write loot summary output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(a.stdout, "%-24s %-16s %-12s %5s  %s\n",
		"NAME", "SOURCE", "STATUS", "QTY", "APPRAISED VALUE",
	); err != nil {
		return fmt.Errorf("write loot summary header: %w", err)
	}

	for _, r := range rows {
		appraisedDisplay := "-"
		if r.LatestAppraisalValue > 0 {
			appraisedDisplay = tools.FormatAmount(r.LatestAppraisalValue)
		}

		if _, err := fmt.Fprintf(a.stdout, "%-24s %-16s %-12s %5d  %s\n",
			truncate(r.Name, 24),
			truncate(r.Source, 16),
			string(r.Status),
			r.Quantity,
			appraisedDisplay,
		); err != nil {
			return fmt.Errorf("write loot summary row: %w", err)
		}
	}

	a.log.logger.InfoContext(ctx, "generated loot summary report",
		slog.Int("count", len(rows)),
	)

	return nil
}
