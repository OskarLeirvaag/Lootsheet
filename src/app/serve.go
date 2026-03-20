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
	var noTLS bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Host the LootSheet database over the network for remote TUI clients",
		Long:  serveHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isLeafHelpArg(args) {
				return a.writeCommandHelp(cmd)
			}
			return a.runServe(cmd.Context(), addr, noTLS)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":7547", "listen address (host:port)")
	cmd.Flags().BoolVar(&noTLS, "no-tls", false, "disable built-in TLS (use behind a TLS-terminating reverse proxy)")

	cmd.AddCommand(a.newServeTokenCommand())

	return cmd
}

func (a *Application) newServeTokenCommand() *cobra.Command {
	return a.newNoArgsLeafCommand("token", "Print the server bearer token", serveCredentialHelpText, func(_ context.Context) error {
		serverDir := filepath.Join(a.config.Paths.DataDir, "server")
		token, err := server.LoadOrGenerateToken(serverDir)
		if err != nil {
			return fmt.Errorf("token: %w", err)
		}
		_, err = fmt.Fprintln(a.stdout, token)
		return err
	})
}

func (a *Application) runServe(ctx context.Context, addr string, noTLS bool) error {
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

	cfg := server.Config{
		Addr:    addr,
		Token:   token,
		Handler: server.NewHandler(&tuiService{loader: loader, databasePath: a.config.Paths.DatabasePath}),
		Logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}

	if noTLS {
		if _, err := fmt.Fprintf(a.stdout, "TLS: disabled (use behind a TLS-terminating reverse proxy)\n"); err != nil {
			return err
		}
	} else {
		tlsCfg, tlsErr := server.LoadOrGenerateTLS(serverDir)
		if tlsErr != nil {
			return fmt.Errorf("tls: %w", tlsErr)
		}
		cfg.TLSConfig = tlsCfg
	}

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

	return server.ListenAndServe(ctx, cfg)
}
