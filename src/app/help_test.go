package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunHelpShowsTopLevelTopics(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"help"}, &stdout); err != nil {
		t.Fatalf("run help: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Command groups:") {
		t.Fatalf("top-level help missing command groups: %q", output)
	}
	if !strings.Contains(output, "tui        open the full-screen TUI shell") {
		t.Fatalf("top-level help missing tui command: %q", output)
	}
	if !strings.Contains(output, "lootsheet account help") {
		t.Fatalf("top-level help missing nested help example: %q", output)
	}
}

func TestRunGroupHelpShowsAccountCommands(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"account", "help"}, &stdout); err != nil {
		t.Fatalf("run account help: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "lootsheet account create --code CODE --name NAME --type TYPE") {
		t.Fatalf("account help missing create usage: %q", output)
	}
	if !strings.Contains(output, "ledger      print the posting history") {
		t.Fatalf("account help missing ledger summary: %q", output)
	}
}

func TestRunLeafHelpSupportsHelpSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"quest", "collect", "help"}, &stdout); err != nil {
		t.Fatalf("run quest collect help: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "lootsheet quest collect --id ID --amount AMOUNT --date YYYY-MM-DD") {
		t.Fatalf("quest collect help missing usage: %q", output)
	}
	if !strings.Contains(output, "Second pouch from the mayor") {
		t.Fatalf("quest collect help missing example: %q", output)
	}
}

func TestRunLeafHelpSupportsShortHelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"account", "list", "-h"}, &stdout); err != nil {
		t.Fatalf("run account list -h: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "lootsheet account list") {
		t.Fatalf("account list help missing usage: %q", output)
	}
}

func TestRunLeafHelpSupportsHelpFlagAfterOtherArgs(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"journal", "post", "--date", "2026-03-08", "--help"}, &stdout); err != nil {
		t.Fatalf("run journal post --help: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Amounts accept D&D 5e denominations") {
		t.Fatalf("journal post help missing amount guidance: %q", output)
	}
	if !strings.Contains(output, "lootsheet journal post --date 2026-03-08") {
		t.Fatalf("journal post help missing example: %q", output)
	}
}

func TestRunHelpSupportsNestedHelpCommandPath(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"help", "report", "writeoff-candidates"}, &stdout); err != nil {
		t.Fatalf("run nested help: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "lootsheet report writeoff-candidates") {
		t.Fatalf("writeoff-candidates help missing usage: %q", output)
	}
	if !strings.Contains(output, "--min-age-days  30") {
		t.Fatalf("writeoff-candidates help missing defaults: %q", output)
	}
}

func TestRunTUIHelpShowsKeyboardControls(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(context.Background(), []string{"tui", "help"}, &stdout); err != nil {
		t.Fatalf("run tui help: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "lootsheet tui") {
		t.Fatalf("tui help missing usage: %q", output)
	}
	if !strings.Contains(output, "Ctrl+L") {
		t.Fatalf("tui help missing redraw key: %q", output)
	}
}
