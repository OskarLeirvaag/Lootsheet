package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/net/server"
	"github.com/spf13/cobra"
)

func (a *Application) newServeCommand() *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Host the LootSheet database over the network for remote TUI clients",
		Long:  serveHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isLeafHelpArg(args) {
				return a.writeCommandHelp(cmd)
			}
			return a.runServe(cmd.Context(), addr)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":7547", "listen address (host:port)")

	return cmd
}

func (a *Application) runServe(ctx context.Context, addr string) error {
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

	// Auto-select a campaign without interactive prompts.
	// Try the previously active campaign, then fall back to the first one.
	// If no campaigns exist yet, clients can create one via the TUI.
	if active, err := campaign.GetActive(ctx, a.config.Paths.DatabasePath); err == nil {
		loader.SetCampaign(active.ID, active.Name)
	} else if campaigns, listErr := campaign.List(ctx, a.config.Paths.DatabasePath); listErr == nil && len(campaigns) > 0 {
		loader.SetCampaign(campaigns[0].ID, campaigns[0].Name)
	}

	serverDir := filepath.Join(a.config.Paths.DataDir, "server")

	token, err := server.LoadOrGenerateToken(serverDir)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}

	tlsCfg, err := server.LoadOrGenerateTLS(serverDir)
	if err != nil {
		return fmt.Errorf("tls: %w", err)
	}

	svc := &tuiService{
		loader:       loader,
		databasePath: a.config.Paths.DatabasePath,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if _, err := fmt.Fprintf(a.stdout, "Server token: %s\n", token); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(a.stdout, "Listening on %s\n", addr); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(a.stdout, "Database: %s\n", a.config.Paths.DatabasePath); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return server.ListenAndServe(ctx, server.Config{
		Addr:      addr,
		TLSConfig: tlsCfg,
		Token:     token,
		Handler:   server.NewHandler(svc),
		Logger:    logger,
	})
}
