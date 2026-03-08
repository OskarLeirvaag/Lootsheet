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
		Short: "Open the full-screen terminal TUI shell",
		Long:  tuiHelpText,
	}

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		assets, err := config.LoadInitAssets()
		if err != nil {
			return err
		}

		return render.Run(ctx, &render.Options{
			ShellLoader: func(ctx context.Context) (render.ShellData, error) {
				return buildTUIShellData(ctx, a.config.Paths.DatabasePath, assets)
			},
			CommandHandler: func(ctx context.Context, command render.Command) (render.ShellData, render.StatusMessage, error) {
				return handleTUICommand(ctx, command, a.config.Paths.DatabasePath, assets)
			},
		})
	})
}
