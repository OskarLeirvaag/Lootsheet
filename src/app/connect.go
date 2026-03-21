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

const downloadFilePerm = 0o600

func (a *Application) newConnectCommand() *cobra.Command {
	var skipVerify bool
	var download string

	cmd := &cobra.Command{
		Use:   "connect <addr>",
		Short: "Connect to a remote LootSheet server and open the TUI",
		Long:  connectHelpText,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if download != "" {
				return a.runDownload(cmd.Context(), args[0], skipVerify, download)
			}
			return a.runConnect(cmd.Context(), args[0], skipVerify)
		},
	}

	cmd.Flags().BoolVar(&skipVerify, "tls-skip-verify", false, "skip TLS certificate verification (for self-signed certs)")
	cmd.Flags().StringVar(&download, "download", "", "download the server database to this path and exit")

	return cmd
}

func (a *Application) runDownload(ctx context.Context, addr string, skipVerify bool, outPath string) error {
	if !strings.Contains(addr, ":") {
		addr += ":7547"
	}
	configDir := filepath.Dir(a.config.Paths.ConfigFile)

	token, found, err := client.LookupToken(configDir, addr)
	if err != nil {
		return fmt.Errorf("lookup token: %w", err)
	}

	if !found {
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

	c, _, err := client.Dial(ctx, addr, token, &client.DialOptions{SkipTLSVerify: skipVerify})
	if err != nil {
		return err
	}
	defer c.Close()

	if !found {
		_ = client.SaveToken(configDir, addr, token)
	}

	if _, err := fmt.Fprintf(os.Stderr, "Downloading database from %s...\n", addr); err != nil {
		return err
	}

	data, filename, err := client.DownloadDatabase(ctx, c)
	if err != nil {
		return err
	}

	// If outPath is a directory, use the server's filename inside it.
	if info, statErr := os.Stat(outPath); statErr == nil && info.IsDir() {
		outPath = filepath.Join(outPath, filename)
	}

	if err := os.WriteFile(outPath, data, downloadFilePerm); err != nil {
		return fmt.Errorf("write database: %w", err)
	}

	if _, err := fmt.Fprintf(a.stdout, "Downloaded %s (%d bytes)\n", outPath, len(data)); err != nil {
		return err
	}

	return nil
}

func (a *Application) runConnect(ctx context.Context, addr string, skipVerify bool) error {
	if !strings.Contains(addr, ":") {
		addr += ":7547"
	}

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

	c, authResp, err := client.Dial(ctx, addr, token, &client.DialOptions{
		SkipTLSVerify: skipVerify,
	})
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
