package app

import (
	"context"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/spf13/cobra"
)

func (a *Application) newTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Open the full-screen terminal dashboard shell",
		Long:  tuiHelpText,
	}

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		assets, err := config.LoadInitAssets()
		if err != nil {
			return err
		}

		return render.Run(ctx, &render.Options{
			DashboardLoader: func(ctx context.Context) (render.DashboardData, error) {
				return buildTUIDashboardData(ctx, a.config.Paths.DatabasePath, assets)
			},
		})
	})
}
