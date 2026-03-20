package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/net/client"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/spf13/cobra"
)

func (a *Application) newConnectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <addr>",
		Short: "Connect to a remote LootSheet server and open the TUI",
		Long:  connectHelpText,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runConnect(cmd.Context(), args[0])
		},
	}

	return cmd
}

func (a *Application) runConnect(ctx context.Context, addr string) error {
	configDir := filepath.Dir(a.config.Paths.ConfigFile)

	// Look up saved token for this address.
	token, found, err := client.LookupToken(configDir, addr)
	if err != nil {
		return fmt.Errorf("lookup token: %w", err)
	}

	if !found {
		// Prompt for token on stderr (stdout is for TUI).
		if _, err := fmt.Fprint(os.Stderr, "Enter server token: "); err != nil {
			return err
		}

		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("no token provided")
		}
		token = strings.TrimSpace(scanner.Text())
		if token == "" {
			return fmt.Errorf("empty token")
		}
	}

	if _, err := fmt.Fprintf(os.Stderr, "Connecting to %s...\n", addr); err != nil {
		return err
	}

	c, authResp, err := client.Dial(ctx, addr, token)
	if err != nil {
		return err
	}
	defer c.Close()

	if _, err := fmt.Fprintf(os.Stderr, "Connected to %s\n", authResp.ServerName); err != nil {
		return err
	}

	// Save token on successful connection.
	if !found {
		if err := client.SaveToken(configDir, addr, token); err != nil {
			// Non-fatal — warn but continue.
			if _, printErr := fmt.Fprintf(os.Stderr, "Warning: could not save token: %v\n", err); printErr != nil {
				return printErr
			}
		}
	}

	return render.Run(ctx, &render.Options{
		ShellLoader:    client.RemoteShellLoader(c),
		CommandHandler: client.RemoteCommandHandler(c),
		SearchHandler:  client.RemoteSearchHandler(c),
	})
}
