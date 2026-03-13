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
	if !strings.Contains(output, "Commands:") {
		t.Fatalf("top-level help missing commands section: %q", output)
	}
	if !strings.Contains(output, "tui        open the full-screen TUI shell") {
		t.Fatalf("top-level help missing tui command: %q", output)
	}
	if !strings.Contains(output, "db         inspect database state and run schema migrations") {
		t.Fatalf("top-level help missing db command: %q", output)
	}
	if !strings.Contains(output, "init       initialize a fresh LootSheet database") {
		t.Fatalf("top-level help missing init command: %q", output)
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
	if !strings.Contains(output, "toggle the selected account active/inactive") {
		t.Fatalf("tui help missing account toggle guidance: %q", output)
	}
	if !strings.Contains(output, "reverse the selected posted journal entry") {
		t.Fatalf("tui help missing journal reverse guidance: %q", output)
	}
	if !strings.Contains(output, "collect the full outstanding balance") {
		t.Fatalf("tui help missing quest collect guidance: %q", output)
	}
	if !strings.Contains(output, "write off the full outstanding balance") {
		t.Fatalf("tui help missing quest write-off guidance: %q", output)
	}
	if !strings.Contains(output, "recognize the selected latest loot appraisal") {
		t.Fatalf("tui help missing loot recognize guidance: %q", output)
	}
	if !strings.Contains(output, "sell the selected recognized loot item") {
		t.Fatalf("tui help missing loot sell guidance: %q", output)
	}
	if !strings.Contains(output, "edit the selected quest or loot item") {
		t.Fatalf("tui help missing edit guidance: %q", output)
	}
	if !strings.Contains(output, "does not set value; appraisal happens later") {
		t.Fatalf("tui help missing loot creation appraisal guidance: %q", output)
	}
	if !strings.Contains(output, "open a glossary modal") {
		t.Fatalf("tui help missing glossary guidance: %q", output)
	}
	if !strings.Contains(output, "Enter                      confirm the open modal") {
		t.Fatalf("tui help missing confirm guidance: %q", output)
	}
	if !strings.Contains(output, "Enter                      submit the guided entry composer when a guided entry form is open") {
		t.Fatalf("tui help missing compose submit guidance: %q", output)
	}
}
