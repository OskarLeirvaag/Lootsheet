package render

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

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
	if e.Lines[0] != "hello" || e.Lines[1] != " world" {
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
