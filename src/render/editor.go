package render

import (
	"fmt"
	"strings"
)

type editorVimMode int

const (
	editorModeNormal  editorVimMode = iota
	editorModeInsert                // typing text
	editorModeCommand               // after typing ':'
)

type editorFocus int

const (
	editorFocusBody    editorFocus = iota
	editorFocusTitle               // title field selected
	editorFocusSession             // session number field
)

type undoEntry struct {
	Lines  []string
	CurRow int
	CurCol int
}

type editorState struct {
	// Command identity (same pattern as composeState).
	CommandID string
	ItemKey   string
	Section   Section

	// Content.
	SessionNum int // auto-incrementing session label; 0 = none
	Title      string
	Lines      []string
	Dirty      bool

	// Cursor & viewport.
	CurRow    int
	CurCol    int
	ScrollRow int

	// Mode.
	Mode  editorVimMode
	Focus editorFocus

	// Command-line mode buffer (after ':').
	CmdBuffer string

	// Two-key sequences (dd, gg).
	PendingKey rune

	// Undo.
	UndoStack []undoEntry

	// Clipboard for yank/paste.
	Clipboard []string // nil = empty, single-element = inline, multi = line-wise

	// Status line message.
	StatusText string

	// Reference picker (Ctrl+A in insert mode).
	refPicker *pickerState
}

// --- Open editor ---

func (s *Shell) openEditor() {
	nextSession := editorNextSessionNum(s.Data.Notes.Items)
	s.editor = &editorState{
		CommandID:  "notes.create",
		Section:    SectionNotes,
		SessionNum: nextSession,
		Lines:      []string{""},
		Mode:       editorModeInsert,
		Focus:      editorFocusTitle,
	}
}

func (s *Shell) openEditorFromAction(itemKey string, action *ItemActionData) {
	title := ""
	body := ""
	if action != nil && action.ComposeFields != nil {
		title = action.ComposeFields["title"]
		body = action.ComposeFields["body"]
	}

	lines := strings.Split(body, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	sessionNum, parsedTitle := editorParseTitle(title)

	s.editor = &editorState{
		CommandID:  "notes.update",
		ItemKey:    itemKey,
		Section:    SectionNotes,
		SessionNum: sessionNum,
		Title:      parsedTitle,
		Lines:      lines,
		Mode:       editorModeNormal,
		Focus:      editorFocusBody,
	}
}

func (s *Shell) openEditorRefPicker() bool {
	if s.editor == nil {
		return false
	}
	optsCap := len(s.Data.Quests.Items) + len(s.Data.Loot.Items) + len(s.Data.Assets.Items) +
		len(s.Data.Codex.Items) + len(s.Data.Notes.Items)
	opts := make([]pickerOption, 0, optsCap)
	for _, item := range s.Data.Quests.Items {
		opts = append(opts, pickerOption{Value: "@quest/" + item.DetailTitle, Label: item.DetailTitle, Kind: "quest"})
	}
	for _, item := range s.Data.Loot.Items {
		opts = append(opts, pickerOption{Value: "@loot/" + item.DetailTitle, Label: item.DetailTitle, Kind: "loot"})
	}
	for _, item := range s.Data.Assets.Items {
		opts = append(opts, pickerOption{Value: "@asset/" + item.DetailTitle, Label: item.DetailTitle, Kind: "asset"})
	}
	for _, item := range s.Data.Codex.Items {
		opts = append(opts, pickerOption{Value: "@person/" + item.DetailTitle, Label: item.DetailTitle, Kind: "person"})
	}
	for _, item := range s.Data.Notes.Items {
		opts = append(opts, pickerOption{Value: "@note/" + item.DetailTitle, Label: item.DetailTitle, Kind: "note"})
	}
	s.editor.refPicker = newPicker("Insert @reference", opts)
	return true
}

// editorNextSessionNum finds the highest "Session N" number in existing notes
// and returns N+1.
func editorNextSessionNum(items []ListItemData) int {
	maxNum := 0
	for _, item := range items {
		n, _ := editorParseTitle(item.DetailTitle)
		if n == 0 {
			// Also try from the Row which contains the title.
			n, _ = editorParseTitle(item.Row)
		}
		if n > maxNum {
			maxNum = n
		}
	}
	return maxNum + 1
}

// editorParseTitle splits "Session N: rest" into (N, rest).
// Returns (0, original) if no session prefix is found.
func editorParseTitle(title string) (int, string) {
	trimmed := strings.TrimSpace(title)
	if !strings.HasPrefix(trimmed, "Session ") {
		return 0, trimmed
	}
	rest := trimmed[len("Session "):]

	// Extract the number.
	numEnd := 0
	for numEnd < len(rest) && rest[numEnd] >= '0' && rest[numEnd] <= '9' {
		numEnd++
	}
	if numEnd == 0 {
		return 0, trimmed
	}

	num := 0
	for _, c := range rest[:numEnd] {
		num = num*10 + int(c-'0')
	}

	after := rest[numEnd:]
	if strings.HasPrefix(after, ": ") {
		return num, strings.TrimSpace(after[2:])
	}
	if after == "" || after == ":" {
		return num, ""
	}

	return 0, trimmed
}

// editorComposeTitle builds the stored title from session number + title.
func editorComposeTitle(e *editorState) string {
	title := strings.TrimSpace(e.Title)
	if e.SessionNum > 0 {
		prefix := fmt.Sprintf("Session %d", e.SessionNum)
		if title != "" {
			return prefix + ": " + title
		}
		return prefix
	}
	return title
}
