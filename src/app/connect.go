package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/net/client"
	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
	"github.com/OskarLeirvaag/Lootsheet/src/render"
	"github.com/spf13/cobra"
)

const downloadFilePerm = 0o600

func (a *Application) newConnectCommand() *cobra.Command {
	var skipVerify bool
	var download string
	var upload bool

	cmd := &cobra.Command{
		Use:   "connect <addr>",
		Short: "Connect to a remote LootSheet server and open the TUI",
		Long:  connectHelpText,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if download != "" {
				return a.runDownload(cmd.Context(), args[0], skipVerify, download)
			}
			if upload {
				return a.runUpload(cmd.Context(), args[0], skipVerify)
			}
			return a.runConnect(cmd.Context(), args[0], skipVerify)
		},
	}

	cmd.Flags().BoolVar(&skipVerify, "tls-skip-verify", false, "skip TLS certificate verification (for self-signed certs)")
	cmd.Flags().StringVar(&download, "download", "", "download the server database to this path and exit")
	cmd.Flags().BoolVar(&upload, "upload", false, "upload a local campaign to the server and exit")

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

func (a *Application) runUpload(ctx context.Context, addr string, skipVerify bool) error {
	if !strings.Contains(addr, ":") {
		addr += ":7547"
	}
	configDir := filepath.Dir(a.config.Paths.ConfigFile)
	localDBPath := a.config.Paths.DatabasePath

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

	// List local campaigns.
	campaigns, err := campaign.List(ctx, localDBPath)
	if err != nil {
		return fmt.Errorf("list local campaigns: %w", err)
	}
	if len(campaigns) == 0 {
		return fmt.Errorf("no campaigns found in local database")
	}

	// Pick a campaign.
	var selected campaign.Record
	if len(campaigns) == 1 {
		selected = campaigns[0]
	} else {
		if _, err := fmt.Fprintln(os.Stderr, "Select a campaign to upload:"); err != nil {
			return err
		}
		for i, c := range campaigns {
			if _, err := fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, c.Name); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(os.Stderr, "Choice: "); err != nil {
			return err
		}

		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("no choice provided")
		}
		var choice int
		if _, err := fmt.Sscanf(strings.TrimSpace(scanner.Text()), "%d", &choice); err != nil || choice < 1 || choice > len(campaigns) {
			return fmt.Errorf("invalid choice")
		}
		selected = campaigns[choice-1]
	}

	// Check if campaign already exists on server.
	mode := pb.UploadMode_UPLOAD_NEW
	remoteCampaigns, err := client.ListRemoteCampaigns(ctx, c)
	if err != nil {
		return fmt.Errorf("list server campaigns: %w", err)
	}

	for _, rc := range remoteCampaigns {
		if rc.Id == selected.ID {
			if _, err := fmt.Fprintf(os.Stderr, "Campaign %q already exists on server. Overwrite? [y/N] ", selected.Name); err != nil {
				return err
			}
			scanner := bufio.NewScanner(os.Stdin)
			if !scanner.Scan() {
				return fmt.Errorf("no confirmation provided")
			}
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "y" && answer != "yes" {
				return fmt.Errorf("upload cancelled")
			}
			mode = pb.UploadMode_UPLOAD_OVERWRITE
			break
		}
	}

	if _, err := fmt.Fprintf(os.Stderr, "Uploading campaign %q to %s...\n", selected.Name, addr); err != nil {
		return err
	}

	resp, err := client.UploadCampaign(ctx, c, localDBPath, selected.ID, mode)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(a.stdout, "Uploaded campaign %q (%s)\n", resp.CampaignName, resp.CampaignId); err != nil {
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
		ShellLoader:       client.RemoteShellLoader(c),
		CommandHandler:    client.RemoteCommandHandler(c),
		SearchHandler:     client.RemoteSearchHandler(c),
		ConnectionChecker: c.Ping,
	})
}
