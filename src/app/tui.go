package app

import (
	"context"

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
		return render.Run(ctx, &render.Options{})
	})
}
