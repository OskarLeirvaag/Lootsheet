package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	manDirPerm  = 0o755
	manFilePerm = 0o600
)

var manPageDate = time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)

// GenerateManPages writes a reproducible set of section-1 man pages for the
// current Cobra command tree.
func GenerateManPages(outputDir string) error {
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return errors.New("man page output directory is required")
	}

	if err := os.MkdirAll(outputDir, manDirPerm); err != nil {
		return fmt.Errorf("create man page directory: %w", err)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("read man page directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "lootsheet") || filepath.Ext(entry.Name()) != ".1" {
			continue
		}
		if err := os.Remove(filepath.Join(outputDir, entry.Name())); err != nil {
			return fmt.Errorf("remove stale man page %q: %w", entry.Name(), err)
		}
	}

	docApp := &Application{stdout: io.Discard}
	root := docApp.newRootCommand()

	if err := generateManPageTree(root, outputDir); err != nil {
		return fmt.Errorf("generate man pages: %w", err)
	}

	return nil
}

func generateManPageTree(cmd *cobra.Command, outputDir string) error {
	if err := writeManPageFile(cmd, outputDir); err != nil {
		return err
	}

	children := cmd.Commands()
	slices.SortFunc(children, func(a, b *cobra.Command) int {
		if a.Name() < b.Name() {
			return -1
		}
		if a.Name() > b.Name() {
			return 1
		}
		return 0
	})

	for _, child := range children {
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := generateManPageTree(child, outputDir); err != nil {
			return err
		}
	}

	return nil
}

func writeManPageFile(cmd *cobra.Command, outputDir string) error {
	filename := strings.ReplaceAll(cmd.CommandPath(), " ", "-") + ".1"
	path := filepath.Join(outputDir, filename)

	content := renderManPage(cmd)
	if err := os.WriteFile(path, []byte(content), manFilePerm); err != nil {
		return fmt.Errorf("write man page %q: %w", filename, err)
	}

	return nil
}

func renderManPage(cmd *cobra.Command) string {
	cmd.InitDefaultHelpFlag()

	var buf bytes.Buffer

	title := strings.ToUpper(strings.ReplaceAll(cmd.CommandPath(), " ", "-"))
	_, _ = fmt.Fprintf(&buf, ".TH %s 1 %q %q %q\n", title, manPageDate.Format("Jan 2006"), "LootSheet", "LootSheet Manual")
	writeSection(&buf, "NAME", fmt.Sprintf("%s \\- %s", strings.ReplaceAll(cmd.CommandPath(), " ", "-"), cmd.Short))
	writeSection(&buf, "SYNOPSIS", cmd.UseLine())

	description := strings.TrimSpace(cmd.Long)
	if description == "" {
		description = strings.TrimSpace(cmd.Short)
	}
	writeSection(&buf, "DESCRIPTION", description)

	writeFlagsSection(&buf, "OPTIONS", cmd.NonInheritedFlags())
	writeFlagsSection(&buf, "OPTIONS INHERITED FROM PARENT COMMANDS", cmd.InheritedFlags())
	writeSeeAlsoSection(&buf, cmd)

	return buf.String()
}

func writeSection(buf *bytes.Buffer, title string, body string) {
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}

	_, _ = buf.WriteString(".SH ")
	_, _ = buf.WriteString(title)
	_ = buf.WriteByte('\n')
	for line := range strings.SplitSeq(body, "\n") {
		_, _ = buf.WriteString(roffLine(line))
		_ = buf.WriteByte('\n')
	}
}

func writeFlagsSection(buf *bytes.Buffer, title string, flags *pflag.FlagSet) {
	if flags == nil || !flags.HasAvailableFlags() {
		return
	}

	_, _ = buf.WriteString(".SH ")
	_, _ = buf.WriteString(title)
	_ = buf.WriteByte('\n')

	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden || flag.Deprecated != "" {
			return
		}

		_, _ = buf.WriteString(".TP\n")
		_, _ = buf.WriteString(roffLine(renderFlagUsage(flag)))
		_ = buf.WriteByte('\n')
		_, _ = buf.WriteString(roffLine(renderFlagDescription(flag)))
		_ = buf.WriteByte('\n')
	})
}

func writeSeeAlsoSection(buf *bytes.Buffer, cmd *cobra.Command) {
	var refs []string

	if parent := cmd.Parent(); parent != nil {
		refs = append(refs, strings.ReplaceAll(parent.CommandPath(), " ", "-")+"(1)")
	}

	children := cmd.Commands()
	slices.SortFunc(children, func(a, b *cobra.Command) int {
		if a.Name() < b.Name() {
			return -1
		}
		if a.Name() > b.Name() {
			return 1
		}
		return 0
	})
	for _, child := range children {
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}
		refs = append(refs, strings.ReplaceAll(child.CommandPath(), " ", "-")+"(1)")
	}

	if len(refs) == 0 {
		return
	}

	writeSection(buf, "SEE ALSO", strings.Join(refs, "\n"))
}

func renderFlagUsage(flag *pflag.Flag) string {
	var parts []string
	if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
		parts = append(parts, "-"+flag.Shorthand)
	}

	long := "--" + flag.Name
	if flag.Value.Type() != "bool" {
		long += "=" + flag.Value.Type()
	}
	parts = append(parts, long)

	return strings.Join(parts, ", ")
}

func renderFlagDescription(flag *pflag.Flag) string {
	description := strings.TrimSpace(flag.Usage)
	if flag.Value.Type() == "bool" {
		return description
	}

	return fmt.Sprintf("%s (default: %s)", description, flag.DefValue)
}

func roffLine(line string) string {
	line = strings.ReplaceAll(line, `\`, `\\`)
	line = strings.TrimRight(line, " \t")
	if line == "" {
		return ""
	}
	if strings.HasPrefix(line, ".") || strings.HasPrefix(line, "'") {
		line = `\&` + line
	}
	return line
}
