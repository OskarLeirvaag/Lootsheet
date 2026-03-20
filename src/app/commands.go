package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/net/client"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
	"github.com/spf13/cobra"
)

func (a *Application) executeRootCommand(ctx context.Context, args []string) error {
	root := a.newRootCommand()
	root.SetArgs(args)
	root.SetOut(a.stdout)
	root.SetErr(io.Discard)

	return root.ExecuteContext(ctx)
}

func (a *Application) newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:               "lootsheet",
		Short:             "Local-first D&D 5e double-entry bookkeeping CLI/TUI",
		Long:              rootHelpText,
		SilenceErrors:     true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_ = a.writeCommandHelp(cmd)
	})

	root.AddCommand(
		a.newDatabaseCommand(),
		a.newInitCommand(),
		a.newTUICommand(),
		a.newServeCommand(),
		a.newConnectCommand(),
	)

	return root
}

func (a *Application) newDatabaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Inspect database state and run schema migrations",
		Long:  dbHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.writeCommandHelp(cmd)
		},
	}

	cmd.AddCommand(
		a.newNoArgsLeafCommand("status", "Show configured database lifecycle state", dbStatusHelpText, func(ctx context.Context) error {
			return a.runDatabase(ctx, []string{"status"})
		}),
		a.newNoArgsLeafCommand("migrate", "Apply pending embedded schema migrations", dbMigrateHelpText, func(ctx context.Context) error {
			return a.runDatabase(ctx, []string{"migrate"})
		}),
	)

	return cmd
}

func (a *Application) newInitCommand() *cobra.Command {
	return a.newNoArgsLeafCommand("init", "Initialize a fresh LootSheet database", initHelpText, func(ctx context.Context) error {
		return a.runInit(ctx)
	})
}

func (a *Application) newTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Open the full-screen terminal TUI shell",
		Long:  tuiHelpText,
	}

	return a.newLeafCommand(cmd, func(ctx context.Context) error {
		configDir := filepath.Dir(a.config.Paths.ConfigFile)
		savedServers, _ := client.ListSavedServers(configDir)

		choice, err := render.RunStartupPicker(ctx, nil, savedServers)
		if err != nil {
			return err
		}

		if choice.Mode == "connect" {
			return a.runConnect(ctx, choice.Address, false)
		}

		return a.runLocalTUI(ctx)
	})
}

func (a *Application) runLocalTUI(ctx context.Context) error {
	assets, err := config.LoadInitAssets()
	if err != nil {
		return err
	}

	loader := &sqliteDataLoader{
		databasePath: a.config.Paths.DatabasePath,
		backupDir:    a.config.Paths.BackupDir,
		assets:       assets,
	}

	if err := loader.EnsureReady(ctx); err != nil {
		return err
	}

	if err := resolveCampaign(ctx, loader, assets); err != nil {
		return err
	}

	return render.Run(ctx, &render.Options{
		ShellLoader: func(ctx context.Context) (render.ShellData, error) {
			return buildTUIShellData(ctx, loader)
		},
		CommandHandler: func(ctx context.Context, command render.Command) (render.CommandResult, error) {
			return handleTUICommand(ctx, command, a.config.Paths.DatabasePath, loader)
		},
		SearchHandler: buildSearchHandler(ctx, loader),
	})
}

// resolveCampaign determines which campaign to use for the TUI session.
// 0 campaigns: prompt for a name, create it.
// 1 campaign: auto-select.
// 2+ campaigns: show a picker.
func resolveCampaign(ctx context.Context, loader *sqliteDataLoader, assets config.InitAssets) error {
	campaigns, err := campaign.List(ctx, loader.databasePath)
	if err != nil {
		return fmt.Errorf("list campaigns: %w", err)
	}

	switch len(campaigns) {
	case 0:
		name, err := render.RunCampaignCreator(ctx, nil)
		if err != nil {
			return err
		}
		created, err := campaign.Create(ctx, loader.databasePath, name, assets.Accounts)
		if err != nil {
			return fmt.Errorf("create campaign: %w", err)
		}
		if err := campaign.SetActive(ctx, loader.databasePath, created.ID); err != nil {
			return fmt.Errorf("set active campaign: %w", err)
		}
		loader.SetCampaign(created.ID, created.Name)

	case 1:
		loader.SetCampaign(campaigns[0].ID, campaigns[0].Name)

	default:
		options := make([]model.CampaignOption, len(campaigns))
		for i, c := range campaigns {
			options[i] = model.CampaignOption{ID: c.ID, Name: c.Name}
		}
		selectedID, err := render.RunCampaignPicker(ctx, nil, options)
		if err != nil {
			return err
		}
		for _, c := range campaigns {
			if c.ID == selectedID {
				loader.SetCampaign(c.ID, c.Name)
				break
			}
		}
	}

	return nil
}

func (a *Application) writeCommandHelp(cmd *cobra.Command) error {
	helpText := strings.TrimSpace(cmd.Long)
	if helpText == "" {
		helpText = strings.TrimSpace(cmd.Short)
	}
	if helpText == "" {
		return nil
	}

	_, err := io.WriteString(a.stdout, helpText+"\n")
	return err
}

func (a *Application) newNoArgsLeafCommand(use string, short string, helpText string, run func(context.Context) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  helpText,
	}

	return a.newLeafCommand(cmd, run)
}

func (a *Application) newLeafCommand(cmd *cobra.Command, run func(context.Context) error) *cobra.Command {
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if isLeafHelpArg(args) {
			return a.writeCommandHelp(cmd)
		}

		if len(args) > 0 {
			return unexpectedLeafArgsError(cmd, args)
		}

		return run(cmd.Context())
	}

	return cmd
}

func isLeafHelpArg(args []string) bool {
	return len(args) == 1 && args[0] == "help"
}

func unexpectedLeafArgsError(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("unexpected arguments for %s: %s\n\n%s", cmd.CommandPath(), strings.Join(args, " "), cmd.Long)
}
