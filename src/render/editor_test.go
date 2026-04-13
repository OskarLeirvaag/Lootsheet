package render

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

const testHello = "hello"

func newTestEditor() *editorState {
	return &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{"hello world", "second line", "third line"},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
	}
}

func TestEditorInsertRune(t *testing.T) {
	e := newTestEditor()
	e.Mode = editorModeInsert
	e.CurRow = 0
	e.CurCol = 5

	editorInsertRune(e, '!')
	if e.Lines[0] != "hello! world" {
		t.Fatalf("insert rune: got %q, want %q", e.Lines[0], "hello! world")
	}
	if e.CurCol != 6 {
		t.Fatalf("cursor col after insert: got %d, want 6", e.CurCol)
	}
	if !e.Dirty {
		t.Fatal("expected dirty flag after insert")
	}
}

func TestEditorBackspace(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 0
	e.CurCol = 5

	editorBackspace(e)
	if e.Lines[0] != "hell world" {
		t.Fatalf("backspace: got %q, want %q", e.Lines[0], "hell world")
	}
	if e.CurCol != 4 {
		t.Fatalf("cursor col after backspace: got %d, want 4", e.CurCol)
	}
}

func TestEditorBackspaceJoinsLines(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 1
	e.CurCol = 0

	editorBackspace(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines after join, got %d", len(e.Lines))
	}
	if e.Lines[0] != "hello worldsecond line" {
		t.Fatalf("joined line: got %q", e.Lines[0])
	}
	if e.CurRow != 0 || e.CurCol != 11 {
		t.Fatalf("cursor after join: row=%d col=%d, want 0,11", e.CurRow, e.CurCol)
	}
}

func TestEditorDeleteChar(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 0
	e.CurCol = 0

	editorDeleteChar(e)
	if e.Lines[0] != "ello world" {
		t.Fatalf("delete char: got %q, want %q", e.Lines[0], "ello world")
	}
}

func TestEditorDeleteLine(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 1

	editorDeleteLine(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines after delete, got %d", len(e.Lines))
	}
	if e.Lines[1] != "third line" {
		t.Fatalf("remaining line: got %q, want %q", e.Lines[1], "third line")
	}
}

func TestEditorDeleteLastLine(t *testing.T) {
	e := &editorState{
		Lines: []string{"only line"},
	}
	editorDeleteLine(e)
	if len(e.Lines) != 1 || e.Lines[0] != "" {
		t.Fatalf("delete last line: got %v", e.Lines)
	}
}

func TestEditorSplitLine(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 0
	e.CurCol = 5

	editorSplitLine(e)
	if len(e.Lines) != 4 {
		t.Fatalf("expected 4 lines after split, got %d", len(e.Lines))
	}
	if e.Lines[0] != testHello || e.Lines[1] != " world" {
		t.Fatalf("split result: %q, %q", e.Lines[0], e.Lines[1])
	}
	if e.CurRow != 1 || e.CurCol != 0 {
		t.Fatalf("cursor after split: row=%d col=%d", e.CurRow, e.CurCol)
	}
}

func TestEditorOpenLineBelow(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 0

	editorOpenLineBelow(e)
	if len(e.Lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(e.Lines))
	}
	if e.Lines[1] != "" {
		t.Fatalf("new line should be empty, got %q", e.Lines[1])
	}
	if e.CurRow != 1 {
		t.Fatalf("cursor row: got %d, want 1", e.CurRow)
	}
	if e.Mode != editorModeInsert {
		t.Fatal("expected insert mode after o")
	}
}

func TestEditorOpenLineAbove(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 1

	editorOpenLineAbove(e)
	if len(e.Lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(e.Lines))
	}
	if e.Lines[1] != "" {
		t.Fatalf("new line should be empty, got %q", e.Lines[1])
	}
	if e.CurRow != 1 {
		t.Fatalf("cursor row: got %d, want 1", e.CurRow)
	}
}

func TestEditorUndo(t *testing.T) {
	e := newTestEditor()
	original := e.Lines[0]

	editorInsertRune(e, 'X')
	if e.Lines[0] == original {
		t.Fatal("expected line to change after insert")
	}

	if !editorUndo(e) {
		t.Fatal("undo should return true")
	}
	if e.Lines[0] != original {
		t.Fatalf("undo: got %q, want %q", e.Lines[0], original)
	}
}

func TestEditorCursorMovement(t *testing.T) {
	e := newTestEditor()
	e.CurRow = 0
	e.CurCol = 0

	editorMoveRight(e)
	if e.CurCol != 1 {
		t.Fatalf("move right: col=%d, want 1", e.CurCol)
	}

	editorMoveDown(e)
	if e.CurRow != 1 {
		t.Fatalf("move down: row=%d, want 1", e.CurRow)
	}

	editorMoveUp(e)
	if e.CurRow != 0 {
		t.Fatalf("move up: row=%d, want 0", e.CurRow)
	}

	editorMoveLeft(e)
	if e.CurCol != 0 {
		t.Fatalf("move left: col=%d, want 0", e.CurCol)
	}

	editorMoveToLineEnd(e)
	if e.CurCol != 10 { // "hello world" has 11 chars, normal mode max = 10
		t.Fatalf("move to end: col=%d, want 10", e.CurCol)
	}

	editorMoveToLineStart(e)
	if e.CurCol != 0 {
		t.Fatalf("move to start: col=%d, want 0", e.CurCol)
	}

	editorMoveToBottom(e)
	if e.CurRow != 2 {
		t.Fatalf("move to bottom: row=%d, want 2", e.CurRow)
	}

	editorMoveToTop(e)
	if e.CurRow != 0 {
		t.Fatalf("move to top: row=%d, want 0", e.CurRow)
	}
}

func TestEditorWordMovement(t *testing.T) {
	e := &editorState{
		Lines: []string{"hello world foo"},
		Mode:  editorModeNormal,
	}

	editorMoveWordForward(e)
	if e.CurCol != 6 {
		t.Fatalf("word forward: col=%d, want 6", e.CurCol)
	}

	editorMoveWordBackward(e)
	if e.CurCol != 0 {
		t.Fatalf("word backward: col=%d, want 0", e.CurCol)
	}
}

func TestEditorKeyDispatchNormalMode(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = newTestEditor()

	// 'j' should move down.
	result, handled := shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), ActionNone)
	if !handled || !result.Redraw {
		t.Fatal("j in normal mode should be handled with redraw")
	}
	if shell.editor.CurRow != 1 {
		t.Fatalf("j: row=%d, want 1", shell.editor.CurRow)
	}

	// 'i' should switch to insert mode.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone), ActionNone)
	if shell.editor.Mode != editorModeInsert {
		t.Fatal("i should enter insert mode")
	}

	// Esc should return to normal mode.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), ActionNone)
	if shell.editor.Mode != editorModeNormal {
		t.Fatal("Esc in insert should return to normal")
	}
}

func TestEditorKeyDispatchInsertMode(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{""},
		Mode:      editorModeInsert,
		Focus:     editorFocusBody,
	}

	// Type "Hi".
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone), ActionNone)

	if shell.editor.Lines[0] != "Hi" {
		t.Fatalf("typed text: got %q, want %q", shell.editor.Lines[0], "Hi")
	}

	// Enter splits line.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)
	if len(shell.editor.Lines) != 2 {
		t.Fatalf("expected 2 lines after Enter, got %d", len(shell.editor.Lines))
	}
}

func TestEditorCommandSave(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Title:     "Test",
		Lines:     []string{"body text"},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
		Dirty:     true,
	}

	// Type :w
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone), ActionNone)
	if shell.editor.Mode != editorModeCommand {
		t.Fatal("expected command mode after ':'")
	}
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone), ActionNone)
	result, _ := shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)

	if result.Command == nil {
		t.Fatal("expected :w to emit a command")
	}
	if result.Command.ID != "notes.create" {
		t.Fatalf("command id = %q, want notes.create", result.Command.ID)
	}
	if result.Command.Fields["title"] != "Test" {
		t.Fatalf("command title = %q, want Test", result.Command.Fields["title"])
	}
	if result.Command.Fields["body"] != "body text" {
		t.Fatalf("command body = %q, want 'body text'", result.Command.Fields["body"])
	}
}

func TestEditorCommandQuitDirtyRejects(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{""},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
		Dirty:     true,
	}

	// :q should reject because dirty.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)

	if shell.editor == nil {
		t.Fatal("expected :q to be rejected when dirty")
	}
	if shell.editor.StatusText == "" {
		t.Fatal("expected status text warning about unsaved changes")
	}
}

func TestEditorCommandForceQuit(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{""},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
		Dirty:     true,
	}

	// :q!
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)

	if shell.editor != nil {
		t.Fatal("expected :q! to force quit")
	}
}

func TestEditorCommandWQ(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Title:     "My Note",
		Lines:     []string{"line 1", "line 2"},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
		Dirty:     true,
	}

	// :wq
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone), ActionNone)
	result, _ := shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)

	if result.Command == nil {
		t.Fatal("expected :wq to emit command")
	}
	if !shell.editorQuitAfterSave {
		t.Fatal("expected editorQuitAfterSave to be true")
	}
	if result.Command.Fields["body"] != "line 1\nline 2" {
		t.Fatalf("body = %q, want 'line 1\\nline 2'", result.Command.Fields["body"])
	}
}

func TestEditorSavePersistsAcrossReload(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Title:     "Test",
		Lines:     []string{"body"},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
		Dirty:     true,
	}

	// :w (save, stay open)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)

	if !shell.editorSaveInFlight {
		t.Fatal("expected editorSaveInFlight after :w")
	}

	// Simulate reload after successful save.
	shell.Reload(&data)

	if shell.editor == nil {
		t.Fatal("expected editor to survive reload after :w")
	}
	if shell.editor.Dirty {
		t.Fatal("expected dirty=false after successful save")
	}
	if shell.editor.StatusText != "Saved." {
		t.Fatalf("status = %q, want 'Saved.'", shell.editor.StatusText)
	}
}

func TestEditorTwoKeySequences(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = newTestEditor()
	shell.editor.CurRow = 2

	// gg should go to top.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone), ActionNone)
	if shell.editor.CurRow != 0 {
		t.Fatalf("gg: row=%d, want 0", shell.editor.CurRow)
	}

	// dd should delete line.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone), ActionNone)
	if len(shell.editor.Lines) != 2 {
		t.Fatalf("dd: expected 2 lines, got %d", len(shell.editor.Lines))
	}
}

func TestEditorTabCyclesFocus(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = newTestEditor()

	// Body → Tab → Session.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), ActionNone)
	if shell.editor.Focus != editorFocusSession {
		t.Fatalf("Tab from body: got focus %d, want session", shell.editor.Focus)
	}

	// Session → Tab → Title.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), ActionNone)
	if shell.editor.Focus != editorFocusTitle {
		t.Fatalf("Tab from session: got focus %d, want title", shell.editor.Focus)
	}

	// Title → Tab → Body.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), ActionNone)
	if shell.editor.Focus != editorFocusBody {
		t.Fatalf("Tab from title: got focus %d, want body", shell.editor.Focus)
	}
}

func TestEditorJDownNavigatesFromTitleToBody(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = newTestEditor()
	shell.editor.Focus = editorFocusTitle

	// j from title should go to body.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), ActionNone)
	if shell.editor.Focus != editorFocusBody {
		t.Fatalf("j from title: got focus %d, want body", shell.editor.Focus)
	}
}

func TestEditorDownArrowNavigatesFromTitleToBody(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = newTestEditor()
	shell.editor.Focus = editorFocusTitle

	// Down arrow from title should go to body.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), ActionNone)
	if shell.editor.Focus != editorFocusBody {
		t.Fatalf("Down from title: got focus %d, want body", shell.editor.Focus)
	}
}

func TestEditorTitleInsert(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{""},
		Mode:      editorModeInsert,
		Focus:     editorFocusTitle,
	}

	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone), ActionNone)

	if shell.editor.Title != "AB" {
		t.Fatalf("title = %q, want AB", shell.editor.Title)
	}

	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone), ActionNone)
	if shell.editor.Title != "A" {
		t.Fatalf("title after backspace = %q, want A", shell.editor.Title)
	}
}

func TestEditorRender(t *testing.T) {
	theme := DefaultTheme()
	keymap := DefaultKeyMap()
	buffer := NewBuffer(100, 30, theme.Base)

	data := DefaultShellData()
	shell := NewShell(&data)
	shell.Section = SectionNotes
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Title:     "Session 5",
		Lines:     []string{"# Heading", "Some body text", ""},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
	}

	shell.Render(buffer, &theme, keymap)
	output := buffer.PlainText()

	for _, token := range []string{
		"New Note",
		"Session: 0",
		"Title:",
		"Session 5",
		"# Heading",
		"Some body text",
		"NORMAL",
		"Info",
		"Help:",
		":w save",
	} {
		if !strings.Contains(output, token) {
			t.Fatalf("editor output missing %q:\n%s", token, output)
		}
	}
}

func TestEditorOpenFromNewCustomAction(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.Section = SectionNotes

	result := shell.HandleAction(ActionNewCustom)
	if !result.Redraw {
		t.Fatal("expected redraw when opening editor")
	}
	if shell.editor == nil {
		t.Fatal("expected editor to open for notes section")
	}
	if shell.editor.CommandID != "notes.create" {
		t.Fatalf("command id = %q, want notes.create", shell.editor.CommandID)
	}
}

func TestEditorOpenFromEditAction(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Notes: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "note-1",
					Row:         "2026-03-12 Test Note",
					DetailTitle: "Test Note",
					DetailLines: []string{"Updated: 2026-03-12"},
					DetailBody:  "Some body",
					Actions: []ItemActionData{{
						Trigger:      ActionEdit,
						ID:           "notes.update",
						Label:        "u edit",
						Mode:         ItemActionModeCompose,
						ComposeMode:  "notes",
						ComposeTitle: "Edit Note",
						ComposeFields: map[string]string{
							"title": "Test Note",
							"body":  "line 1\nline 2",
						},
					}},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowNotes)

	result := shell.HandleAction(ActionEdit)
	if !result.Redraw {
		t.Fatal("expected redraw when opening editor from edit")
	}
	if shell.editor == nil {
		t.Fatal("expected editor to open from edit action")
	}
	if shell.editor.CommandID != "notes.update" {
		t.Fatalf("command id = %q, want notes.update", shell.editor.CommandID)
	}
	if shell.editor.Title != "Test Note" {
		t.Fatalf("title = %q, want 'Test Note'", shell.editor.Title)
	}
	if len(shell.editor.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(shell.editor.Lines))
	}
	if shell.editor.Lines[0] != "line 1" || shell.editor.Lines[1] != "line 2" {
		t.Fatalf("lines = %v", shell.editor.Lines)
	}
}

func TestEditorParseReferences(t *testing.T) {
	e := &editorState{
		Lines: []string{
			"Met @[person/Mayor Elra] at the gate.",
			"Related to @[quest/Clear the Tower].",
			"@[person/Mayor Elra] mentioned again.",
		},
	}

	refs := editorParseReferences(e)
	if len(refs) != 2 {
		t.Fatalf("expected 2 unique refs, got %d: %v", len(refs), refs)
	}
	if refs[0] != "@[person/Mayor Elra]" {
		t.Fatalf("ref[0] = %q, want @[person/Mayor Elra]", refs[0])
	}
	if refs[1] != "@[quest/Clear the Tower]" {
		t.Fatalf("ref[1] = %q, want @[quest/Clear the Tower]", refs[1])
	}
}

func TestEditorFooterHelp(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = newTestEditor()

	help := shell.footerHelpText(DefaultKeyMap())
	if !strings.Contains(help, ":w save") {
		t.Fatalf("editor footer = %q, missing ':w save'", help)
	}
	if !strings.Contains(help, ":q quit") {
		t.Fatalf("editor footer = %q, missing ':q quit'", help)
	}
}

func TestEditorSessionAutoIncrement(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Notes: ListScreenData{
			Items: []ListItemData{
				{Key: "1", DetailTitle: "Session 3: Planning"},
				{Key: "2", DetailTitle: "Session 5: Battle"},
				{Key: "3", DetailTitle: "Random Note"},
			},
		},
	}
	shell := NewShell(&data)
	shell.Section = SectionNotes
	shell.HandleAction(ActionNewCustom)

	if shell.editor == nil {
		t.Fatal("expected editor to open")
	}
	if shell.editor.SessionNum != 6 {
		t.Fatalf("session num = %d, want 6 (max 5 + 1)", shell.editor.SessionNum)
	}
}

func TestEditorParseTitleWithSession(t *testing.T) {
	tests := []struct {
		input     string
		wantNum   int
		wantTitle string
	}{
		{"Session 3: Planning", 3, "Planning"},
		{"Session 10: Long Name", 10, "Long Name"},
		{"Session 1", 1, ""},
		{"Random Note", 0, "Random Note"},
		{"Session ABC", 0, "Session ABC"},
		{"", 0, ""},
	}

	for _, tt := range tests {
		num, title := editorParseTitle(tt.input)
		if num != tt.wantNum || title != tt.wantTitle {
			t.Errorf("editorParseTitle(%q) = (%d, %q), want (%d, %q)", tt.input, num, title, tt.wantNum, tt.wantTitle)
		}
	}
}

func TestEditorComposeTitleWithSession(t *testing.T) {
	e := &editorState{SessionNum: 5, Title: "Battle"}
	got := editorComposeTitle(e)
	if got != "Session 5: Battle" {
		t.Fatalf("compose title = %q, want 'Session 5: Battle'", got)
	}

	e2 := &editorState{SessionNum: 3, Title: ""}
	got2 := editorComposeTitle(e2)
	if got2 != "Session 3" {
		t.Fatalf("compose title = %q, want 'Session 3'", got2)
	}

	e3 := &editorState{SessionNum: 0, Title: "Plain"}
	got3 := editorComposeTitle(e3)
	if got3 != "Plain" {
		t.Fatalf("compose title = %q, want 'Plain'", got3)
	}
}

func TestEditorSaveComposesSessionTitle(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID:  "notes.create",
		Section:    SectionNotes,
		SessionNum: 7,
		Title:      "Dungeon Crawl",
		Lines:      []string{"body"},
		Mode:       editorModeNormal,
		Focus:      editorFocusBody,
		Dirty:      true,
	}

	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone), ActionNone)
	result, _ := shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)

	if result.Command == nil {
		t.Fatal("expected command")
	}
	if result.Command.Fields["title"] != "Session 7: Dungeon Crawl" {
		t.Fatalf("saved title = %q, want 'Session 7: Dungeon Crawl'", result.Command.Fields["title"])
	}
}

func TestEditorEditParsesSessionFromTitle(t *testing.T) {
	data := ShellData{
		Dashboard: DefaultDashboardData(),
		Notes: ListScreenData{
			Items: []ListItemData{
				{
					Key:         "note-1",
					DetailTitle: "Session 4: Tavern Meeting",
					Actions: []ItemActionData{{
						Trigger:     ActionEdit,
						ID:          "notes.update",
						Mode:        ItemActionModeCompose,
						ComposeMode: "notes",
						ComposeFields: map[string]string{
							"title": "Session 4: Tavern Meeting",
							"body":  "content",
						},
					}},
				},
			},
		},
	}
	shell := NewShell(&data)
	shell.HandleAction(ActionShowNotes)
	shell.HandleAction(ActionEdit)

	if shell.editor == nil {
		t.Fatal("expected editor")
	}
	if shell.editor.SessionNum != 4 {
		t.Fatalf("session num = %d, want 4", shell.editor.SessionNum)
	}
	if shell.editor.Title != "Tavern Meeting" {
		t.Fatalf("title = %q, want 'Tavern Meeting'", shell.editor.Title)
	}
}

func TestEditorSessionInsertMode(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID:  "notes.create",
		Section:    SectionNotes,
		SessionNum: 1,
		Lines:      []string{""},
		Mode:       editorModeInsert,
		Focus:      editorFocusSession,
	}

	// Type "2" → session becomes 12.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone), ActionNone)
	if shell.editor.SessionNum != 12 {
		t.Fatalf("session after typing 2: got %d, want 12", shell.editor.SessionNum)
	}

	// Backspace → session becomes 1.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone), ActionNone)
	if shell.editor.SessionNum != 1 {
		t.Fatalf("session after backspace: got %d, want 1", shell.editor.SessionNum)
	}
}

// --- List continuation tests ---

func TestEditorListPrefix(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"- item", "- "},
		{"* item", "* "},
		{"1. item", "1. "},
		{"12. item", "12. "},
		{"123. item", "123. "},
		{"> quote", "> "},
		{"  - indented", "  - "},
		{"  1. indented", "  1. "},
		{"plain text", ""},
		{"", ""},
		{"- ", "- "},
		{"1234. too many digits", ""},
		{"hello. world", ""},
	}
	for _, tt := range tests {
		got := editorListPrefix(tt.line)
		if got != tt.want {
			t.Errorf("editorListPrefix(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestEditorNextListPrefix(t *testing.T) {
	tests := []struct {
		prefix string
		want   string
	}{
		{"- ", "- "},
		{"* ", "* "},
		{"> ", "> "},
		{"1. ", "2. "},
		{"9. ", "10. "},
		{"  1. ", "  2. "},
		{"99. ", "100. "},
	}
	for _, tt := range tests {
		got := editorNextListPrefix(tt.prefix)
		if got != tt.want {
			t.Errorf("editorNextListPrefix(%q) = %q, want %q", tt.prefix, got, tt.want)
		}
	}
}

func TestEditorSplitLineContinuesBullet(t *testing.T) {
	e := &editorState{
		Lines:  []string{"- hello"},
		Mode:   editorModeInsert,
		CurRow: 0,
		CurCol: 7,
	}
	editorSplitLine(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(e.Lines))
	}
	if e.Lines[0] != "- hello" {
		t.Fatalf("line[0] = %q, want '- hello'", e.Lines[0])
	}
	if e.Lines[1] != "- " {
		t.Fatalf("line[1] = %q, want '- '", e.Lines[1])
	}
	if e.CurCol != 2 {
		t.Fatalf("cursor col = %d, want 2", e.CurCol)
	}
}

func TestEditorSplitLineContinuesNumber(t *testing.T) {
	e := &editorState{
		Lines:  []string{"1. first"},
		Mode:   editorModeInsert,
		CurRow: 0,
		CurCol: 8,
	}
	editorSplitLine(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(e.Lines))
	}
	if e.Lines[1] != "2. " {
		t.Fatalf("line[1] = %q, want '2. '", e.Lines[1])
	}
	if e.CurCol != 3 {
		t.Fatalf("cursor col = %d, want 3", e.CurCol)
	}
}

func TestEditorSplitLineExitsEmptyBullet(t *testing.T) {
	e := &editorState{
		Lines:  []string{"- previous", "- "},
		Mode:   editorModeInsert,
		CurRow: 1,
		CurCol: 2,
	}
	editorSplitLine(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(e.Lines), e.Lines)
	}
	if e.Lines[1] != "" {
		t.Fatalf("line[1] = %q, want empty", e.Lines[1])
	}
	if e.CurCol != 0 {
		t.Fatalf("cursor col = %d, want 0", e.CurCol)
	}
}

func TestEditorSplitLineContinuesBlockquote(t *testing.T) {
	e := &editorState{
		Lines:  []string{"> some text"},
		Mode:   editorModeInsert,
		CurRow: 0,
		CurCol: 11,
	}
	editorSplitLine(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(e.Lines))
	}
	if e.Lines[1] != "> " {
		t.Fatalf("line[1] = %q, want '> '", e.Lines[1])
	}
}

func TestEditorSplitLineNoPrefixUnchanged(t *testing.T) {
	e := &editorState{
		Lines:  []string{"hello world"},
		Mode:   editorModeInsert,
		CurRow: 0,
		CurCol: 5,
	}
	editorSplitLine(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(e.Lines))
	}
	if e.Lines[0] != testHello || e.Lines[1] != " world" {
		t.Fatalf("split = %q, %q; want 'hello', ' world'", e.Lines[0], e.Lines[1])
	}
	if e.CurCol != 0 {
		t.Fatalf("cursor col = %d, want 0", e.CurCol)
	}
}

func TestEditorOpenBelowContinuesList(t *testing.T) {
	e := &editorState{
		Lines:  []string{"- item one"},
		Mode:   editorModeNormal,
		CurRow: 0,
	}
	editorOpenLineBelow(e)
	if len(e.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(e.Lines))
	}
	if e.Lines[1] != "- " {
		t.Fatalf("line[1] = %q, want '- '", e.Lines[1])
	}
	if e.CurCol != 2 {
		t.Fatalf("cursor col = %d, want 2", e.CurCol)
	}
	if e.Mode != editorModeInsert {
		t.Fatal("expected insert mode")
	}
}

func TestEditorOpenBelowNumberedList(t *testing.T) {
	e := &editorState{
		Lines:  []string{"3. third item"},
		Mode:   editorModeNormal,
		CurRow: 0,
	}
	editorOpenLineBelow(e)
	if e.Lines[1] != "4. " {
		t.Fatalf("line[1] = %q, want '4. '", e.Lines[1])
	}
	if e.CurCol != 3 {
		t.Fatalf("cursor col = %d, want 3", e.CurCol)
	}
}

func TestEditorOpenBelowEmptyBulletNoPrefix(t *testing.T) {
	e := &editorState{
		Lines:  []string{"- "},
		Mode:   editorModeNormal,
		CurRow: 0,
	}
	editorOpenLineBelow(e)
	if e.Lines[1] != "" {
		t.Fatalf("line[1] = %q, want empty (no continuation for empty bullet)", e.Lines[1])
	}
}

// --- Syntax highlighting tests ---

func TestEditorLineStylesHeading(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("# My Heading")
	styles := editorLineStyles(line, false, &theme)
	if len(styles) != len(line) {
		t.Fatalf("styles len = %d, want %d", len(styles), len(line))
	}
	for i, s := range styles {
		if s != theme.EditorHeading {
			t.Fatalf("style[%d] should be EditorHeading", i)
		}
	}
}

func TestEditorLineStylesBulletMarker(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("- item text")
	styles := editorLineStyles(line, false, &theme)
	if styles[0] != theme.EditorListMarker {
		t.Fatal("bullet '-' should use EditorListMarker style")
	}
	// Text after bullet should be default text.
	if styles[2] != theme.Text {
		t.Fatal("item text should use Text style")
	}
}

func TestEditorLineStylesNumberedMarker(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("12. item")
	styles := editorLineStyles(line, false, &theme)
	// "12." (3 chars) should be list marker.
	for i := range 3 {
		if styles[i] != theme.EditorListMarker {
			t.Fatalf("style[%d] should be EditorListMarker for '12.'", i)
		}
	}
	// Space and text should be default.
	if styles[3] != theme.Text {
		t.Fatal("style[3] should be Text, got something else")
	}
}

func TestEditorLineStylesCodeFence(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("some code here")
	styles := editorLineStyles(line, true, &theme)
	for i, s := range styles {
		if s != theme.EditorCode {
			t.Fatalf("style[%d] should be EditorCode inside code fence", i)
		}
	}
}

func TestEditorLineStylesBlockquote(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("> quoted text")
	styles := editorLineStyles(line, false, &theme)
	for i, s := range styles {
		if s != theme.EditorBlockquote {
			t.Fatalf("style[%d] should be EditorBlockquote", i)
		}
	}
}

func TestEditorLineStylesBold(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("normal **bold** normal")
	styles := editorLineStyles(line, false, &theme)
	// "normal " = indices 0-6 (Text)
	if styles[0] != theme.Text {
		t.Fatal("leading text should be Text style")
	}
	// "**bold**" = indices 7-14 (Bold)
	for i := 7; i <= 14; i++ {
		if styles[i] != theme.EditorBold {
			t.Fatalf("style[%d] should be EditorBold", i)
		}
	}
	// " normal" = indices 15-21 (Text)
	if styles[15] != theme.Text {
		t.Fatal("trailing text should be Text style")
	}
}

func TestEditorLineStylesInlineCode(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("use `code` here")
	styles := editorLineStyles(line, false, &theme)
	// "`code`" = indices 4-9
	for i := 4; i <= 9; i++ {
		if styles[i] != theme.EditorCode {
			t.Fatalf("style[%d] should be EditorCode", i)
		}
	}
}

func TestEditorLineStylesReference(t *testing.T) {
	theme := DefaultTheme()
	line := []rune("see @[quest/Goblin Cave] here")
	styles := editorLineStyles(line, false, &theme)
	// "@[quest/Goblin Cave]" = indices 4-23
	for i := 4; i <= 23; i++ {
		if styles[i] != theme.EditorReference {
			t.Fatalf("style[%d] should be EditorReference", i)
		}
	}
}

// --- :fmt command tests ---

func TestEditorFixHeadingSpace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"# heading", "# heading"},       // already correct
		{"#heading", "# heading"},         // missing space
		{"##heading", "## heading"},       // level 2
		{"###heading", "### heading"},     // level 3
		{"####heading", "####heading"},    // level 4+ ignored
		{"# ", "# "},                      // just prefix, no change
		{"#", "#"},                         // bare hash, no change
		{"normal text", "normal text"},    // no heading
		{"  #indented", "  # indented"},   // indented heading
	}
	for _, tt := range tests {
		got := editorFixHeadingSpace(tt.input)
		if got != tt.want {
			t.Errorf("editorFixHeadingSpace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEditorRenumberLists(t *testing.T) {
	e := &editorState{
		Lines: []string{
			"1. first",
			"5. second",
			"3. third",
		},
	}
	editorRenumberLists(e)
	want := []string{"1. first", "2. second", "3. third"}
	for i, line := range e.Lines {
		if line != want[i] {
			t.Errorf("line[%d] = %q, want %q", i, line, want[i])
		}
	}
}

func TestEditorRenumberListsSeparatedByBlank(t *testing.T) {
	e := &editorState{
		Lines: []string{
			"1. alpha",
			"1. beta",
			"",
			"5. gamma",
			"5. delta",
		},
	}
	editorRenumberLists(e)
	want := []string{"1. alpha", "2. beta", "", "1. gamma", "2. delta"}
	for i, line := range e.Lines {
		if line != want[i] {
			t.Errorf("line[%d] = %q, want %q", i, line, want[i])
		}
	}
}

func TestEditorRenumberNestedLists(t *testing.T) {
	e := &editorState{
		Lines: []string{
			"1. first",
			"  1. nested a",
			"  5. nested b",
			"3. second",
		},
	}
	editorRenumberLists(e)
	want := []string{"1. first", "  1. nested a", "  2. nested b", "2. second"}
	for i, line := range e.Lines {
		if line != want[i] {
			t.Errorf("line[%d] = %q, want %q", i, line, want[i])
		}
	}
}

func TestEditorFormatDocument(t *testing.T) {
	e := &editorState{
		Lines: []string{
			"#heading without space",
			"\t- tab indented item",
			"trailing spaces   ",
			"1. first",
			"1. second",
			"1. third",
			"```",
			"\tcode with tabs preserved",
			"```",
			"##also needs space",
		},
	}
	editorFormatDocument(e)

	want := []string{
		"# heading without space",
		"  - tab indented item",
		"trailing spaces",
		"1. first",
		"2. second",
		"3. third",
		"```",
		"\tcode with tabs preserved",
		"```",
		"## also needs space",
	}
	for i, line := range e.Lines {
		if line != want[i] {
			t.Errorf("line[%d] = %q, want %q", i, line, want[i])
		}
	}
	if !e.Dirty {
		t.Fatal("expected dirty flag after format")
	}
}

func TestEditorFormatCommandDispatch(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{"#hello", "1. a", "1. b"},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
	}

	// :fmt
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)

	if shell.editor.Lines[0] != "# hello" {
		t.Fatalf("heading = %q, want '# hello'", shell.editor.Lines[0])
	}
	if shell.editor.Lines[2] != "2. b" {
		t.Fatalf("list item = %q, want '2. b'", shell.editor.Lines[2])
	}
	if shell.editor.StatusText != "Formatted." {
		t.Fatalf("status = %q, want 'Formatted.'", shell.editor.StatusText)
	}
}

func TestEditorTabInsertsTwoSpaces(t *testing.T) {
	e := &editorState{
		Lines:  []string{testHello},
		Mode:   editorModeInsert,
		Focus:  editorFocusBody,
		CurRow: 0,
		CurCol: 0,
	}
	editorInsertTab(e)
	if e.Lines[0] != "  hello" {
		t.Fatalf("after tab: got %q, want '  hello'", e.Lines[0])
	}
	if e.CurCol != 2 {
		t.Fatalf("cursor col = %d, want 2", e.CurCol)
	}
	// Should be undoable as single action.
	if !editorUndo(e) {
		t.Fatal("undo should succeed")
	}
	if e.Lines[0] != testHello {
		t.Fatalf("after undo: got %q, want 'hello'", e.Lines[0])
	}
}

func TestEditorTabKeyDispatch(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{"text"},
		Mode:      editorModeInsert,
		Focus:     editorFocusBody,
		CurRow:    0,
		CurCol:    0,
	}

	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), ActionNone)
	if shell.editor.Lines[0] != "  text" {
		t.Fatalf("tab in insert body: got %q, want '  text'", shell.editor.Lines[0])
	}
}

func TestEditorFormatUndoable(t *testing.T) {
	e := &editorState{
		Lines: []string{"#heading", "trailing  "},
		Focus: editorFocusBody,
	}
	editorFormatDocument(e)
	if e.Lines[0] != "# heading" {
		t.Fatalf("after fmt: %q", e.Lines[0])
	}

	if !editorUndo(e) {
		t.Fatal("undo should succeed after format")
	}
	if e.Lines[0] != "#heading" {
		t.Fatalf("after undo: %q, want '#heading'", e.Lines[0])
	}
}

// --- Search tests ---

func TestEditorExecuteSearch(t *testing.T) {
	e := &editorState{
		Lines: []string{"hello world", "hello again", "goodbye"},
		Focus: editorFocusBody,
	}
	e.SearchQuery = testHello
	editorExecuteSearch(e)

	if len(e.SearchMatches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(e.SearchMatches))
	}
	if !e.SearchActive {
		t.Fatal("expected SearchActive")
	}
	if e.SearchMatches[0].Row != 0 || e.SearchMatches[0].ColStart != 0 {
		t.Fatalf("match[0] = (%d,%d), want (0,0)", e.SearchMatches[0].Row, e.SearchMatches[0].ColStart)
	}
	if e.SearchMatches[1].Row != 1 || e.SearchMatches[1].ColStart != 0 {
		t.Fatalf("match[1] = (%d,%d), want (1,0)", e.SearchMatches[1].Row, e.SearchMatches[1].ColStart)
	}
	// Cursor should jump to first match.
	if e.CurRow != 0 || e.CurCol != 0 {
		t.Fatalf("cursor = (%d,%d), want (0,0)", e.CurRow, e.CurCol)
	}
}

func TestEditorSearchCaseInsensitive(t *testing.T) {
	e := &editorState{
		Lines: []string{"Hello HELLO hello"},
		Focus: editorFocusBody,
	}
	e.SearchQuery = testHello
	editorExecuteSearch(e)

	if len(e.SearchMatches) != 3 {
		t.Fatalf("expected 3 case-insensitive matches, got %d", len(e.SearchMatches))
	}
}

func TestEditorSearchNoMatches(t *testing.T) {
	e := &editorState{
		Lines: []string{"hello world"},
		Focus: editorFocusBody,
	}
	e.SearchQuery = "xyz"
	editorExecuteSearch(e)

	if len(e.SearchMatches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(e.SearchMatches))
	}
	if e.SearchActive {
		t.Fatal("SearchActive should be false when no matches")
	}
}

func TestEditorSearchNext(t *testing.T) {
	e := &editorState{
		Lines: []string{"aa bb aa"},
		Focus: editorFocusBody,
	}
	e.SearchQuery = "aa"
	editorExecuteSearch(e)

	if len(e.SearchMatches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(e.SearchMatches))
	}
	if e.SearchIndex != 0 {
		t.Fatalf("initial index = %d, want 0", e.SearchIndex)
	}

	// n → next match.
	editorSearchNext(e, 1)
	if e.SearchIndex != 1 {
		t.Fatalf("after n: index = %d, want 1", e.SearchIndex)
	}
	if e.CurCol != 6 {
		t.Fatalf("after n: col = %d, want 6", e.CurCol)
	}

	// n → wrap to first.
	editorSearchNext(e, 1)
	if e.SearchIndex != 0 {
		t.Fatalf("after wrap: index = %d, want 0", e.SearchIndex)
	}

	// N → wrap to last.
	editorSearchNext(e, -1)
	if e.SearchIndex != 1 {
		t.Fatalf("after N wrap: index = %d, want 1", e.SearchIndex)
	}
}

func TestEditorSearchJumpsToNearestMatch(t *testing.T) {
	e := &editorState{
		Lines:  []string{"first match", "second match", "third match"},
		Focus:  editorFocusBody,
		CurRow: 1,
		CurCol: 5,
	}
	e.SearchQuery = "match"
	editorExecuteSearch(e)

	// Should jump to "match" on line 1 (col 7), since cursor is at (1,5).
	if e.SearchIndex != 1 {
		t.Fatalf("index = %d, want 1 (nearest at or after cursor)", e.SearchIndex)
	}
	if e.CurRow != 1 || e.CurCol != 7 {
		t.Fatalf("cursor = (%d,%d), want (1,7)", e.CurRow, e.CurCol)
	}
}

func TestEditorSearchKeyDispatch(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{"hello world", "hello again"},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
	}

	// / enters search mode.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone), ActionNone)
	if shell.editor.Mode != editorModeSearch {
		t.Fatal("expected search mode after /")
	}

	// Type "hello" and Enter.
	for _, r := range testHello {
		shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone), ActionNone)
	}
	if shell.editor.SearchBuffer != testHello {
		t.Fatalf("search buffer = %q, want 'hello'", shell.editor.SearchBuffer)
	}

	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), ActionNone)
	if shell.editor.Mode != editorModeNormal {
		t.Fatal("expected normal mode after search Enter")
	}
	if len(shell.editor.SearchMatches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(shell.editor.SearchMatches))
	}
	if !shell.editor.SearchActive {
		t.Fatal("expected SearchActive after search")
	}

	// n goes to next match.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone), ActionNone)
	if shell.editor.SearchIndex != 1 {
		t.Fatalf("after n: index = %d, want 1", shell.editor.SearchIndex)
	}
	if shell.editor.CurRow != 1 {
		t.Fatalf("after n: row = %d, want 1", shell.editor.CurRow)
	}

	// N goes back.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone), ActionNone)
	if shell.editor.SearchIndex != 0 {
		t.Fatalf("after N: index = %d, want 0", shell.editor.SearchIndex)
	}
}

func TestEditorSearchEscCancels(t *testing.T) {
	data := DefaultShellData()
	shell := NewShell(&data)
	shell.editor = &editorState{
		CommandID: "notes.create",
		Section:   SectionNotes,
		Lines:     []string{testHello},
		Mode:      editorModeNormal,
		Focus:     editorFocusBody,
	}

	// / then Esc should cancel.
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone), ActionNone)
	shell.handleEditorKeyEvent(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), ActionNone)

	if shell.editor.Mode != editorModeNormal {
		t.Fatal("expected normal mode after Esc")
	}
	if shell.editor.SearchActive {
		t.Fatal("SearchActive should be false after Esc")
	}
}

func TestEditorMatchAt(t *testing.T) {
	e := &editorState{
		SearchActive: true,
		SearchMatches: []editorMatch{
			{Row: 0, ColStart: 0, ColEnd: 5},
			{Row: 1, ColStart: 3, ColEnd: 8},
		},
		SearchIndex: 1,
	}

	if idx := editorMatchAt(e, 0, 2); idx != 0 {
		t.Fatalf("matchAt(0,2) = %d, want 0", idx)
	}
	if idx := editorMatchAt(e, 1, 5); idx != 1 {
		t.Fatalf("matchAt(1,5) = %d, want 1", idx)
	}
	if idx := editorMatchAt(e, 2, 0); idx != -1 {
		t.Fatalf("matchAt(2,0) = %d, want -1", idx)
	}
}
