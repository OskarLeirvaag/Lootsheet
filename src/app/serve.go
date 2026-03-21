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
	proto "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
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

	fmt.Fprintf(a.stdout, "LootSheet server (protocol v%d, schema v%s)\n", proto.ProtocolVersion, config.SchemaVersion)

	// Database readiness.
	dbStatus, err := loader.GetDatabaseStatus(ctx)
	if err != nil {
		fmt.Fprintf(a.stdout, "Database status: error (%v)\n", err)
		return err
	}
	fmt.Fprintf(a.stdout, "Database: %s\n", a.config.Paths.DatabasePath)
	fmt.Fprintf(a.stdout, "Database state: %s\n", dbStatus.State)
	fmt.Fprintf(a.stdout, "Schema version: %s\n", dbStatus.SchemaVersion)

	if err := loader.EnsureReady(ctx); err != nil {
		fmt.Fprintf(a.stdout, "EnsureReady failed: %v\n", err)
		return err
	}

	// Re-check after migration.
	dbStatus, _ = loader.GetDatabaseStatus(ctx)
	fmt.Fprintf(a.stdout, "Database state after migrate: %s\n", dbStatus.State)

	// Campaign selection.
	campaigns, listErr := campaign.List(ctx, a.config.Paths.DatabasePath)
	if listErr != nil {
		fmt.Fprintf(a.stdout, "Campaign list error: %v\n", listErr)
	} else {
		fmt.Fprintf(a.stdout, "Campaigns found: %d\n", len(campaigns))
		for _, c := range campaigns {
			fmt.Fprintf(a.stdout, "  - %s (%s)\n", c.Name, c.ID)
		}
	}

	if active, getErr := campaign.GetActive(ctx, a.config.Paths.DatabasePath); getErr == nil {
		loader.SetCampaign(active.ID, active.Name)
		fmt.Fprintf(a.stdout, "Active campaign: %s\n", active.Name)
	} else if len(campaigns) > 0 {
		loader.SetCampaign(campaigns[0].ID, campaigns[0].Name)
		fmt.Fprintf(a.stdout, "Active campaign (fallback): %s\n", campaigns[0].Name)
	} else {
		fmt.Fprintf(a.stdout, "No campaigns — clients can create one via the TUI (#)\n")
	}

	// Server setup.
	serverDir := filepath.Join(a.config.Paths.DataDir, "server")

	token, err := server.LoadOrGenerateToken(serverDir)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := server.Config{
		Addr:    addr,
		Token:   token,
		Handler: server.NewHandler(&tuiService{loader: loader, databasePath: a.config.Paths.DatabasePath}),
		Logger:  logger,
	}

	if noTLS {
		fmt.Fprintf(a.stdout, "TLS: disabled (behind reverse proxy)\n")
	} else {
		tlsCfg, tlsErr := server.LoadOrGenerateTLS(serverDir)
		if tlsErr != nil {
			return fmt.Errorf("tls: %w", tlsErr)
		}
		cfg.TLSConfig = tlsCfg
		fmt.Fprintf(a.stdout, "TLS: self-signed\n")
	}

	fmt.Fprintf(a.stdout, "Server token: %s\n", token)
	fmt.Fprintf(a.stdout, "Listening on %s\n", addr)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return server.ListenAndServe(ctx, cfg)
}
