package render

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
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

	// Status line message.
	StatusText string
}

// --- Text buffer operations ---

func editorPushUndo(e *editorState) {
	snapshot := undoEntry{
		Lines:  make([]string, len(e.Lines)),
		CurRow: e.CurRow,
		CurCol: e.CurCol,
	}
	copy(snapshot.Lines, e.Lines)
	e.UndoStack = append(e.UndoStack, snapshot)
}

func editorUndo(e *editorState) bool {
	if len(e.UndoStack) == 0 {
		return false
	}
	last := e.UndoStack[len(e.UndoStack)-1]
	e.UndoStack = e.UndoStack[:len(e.UndoStack)-1]
	e.Lines = last.Lines
	e.CurRow = last.CurRow
	e.CurCol = last.CurCol
	e.Dirty = true
	return true
}

func editorInsertRune(e *editorState, r rune) {
	editorPushUndo(e)
	if len(e.Lines) == 0 {
		e.Lines = []string{""}
	}
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))
	newLine := make([]rune, 0, len(line)+1)
	newLine = append(newLine, line[:col]...)
	newLine = append(newLine, r)
	newLine = append(newLine, line[col:]...)
	e.Lines[e.CurRow] = string(newLine)
	e.CurCol = col + 1
	e.Dirty = true
}

func editorBackspace(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	editorPushUndo(e)
	line := []rune(e.Lines[e.CurRow])
	if e.CurCol > 0 {
		col := clampInt(e.CurCol, 0, len(line))
		newLine := make([]rune, 0, len(line)-1)
		newLine = append(newLine, line[:col-1]...)
		newLine = append(newLine, line[col:]...)
		e.Lines[e.CurRow] = string(newLine)
		e.CurCol = col - 1
		e.Dirty = true
	} else if e.CurRow > 0 {
		// Join with previous line.
		prevLine := []rune(e.Lines[e.CurRow-1])
		joinCol := len(prevLine)
		e.Lines[e.CurRow-1] = string(prevLine) + string(line)
		e.Lines = append(e.Lines[:e.CurRow], e.Lines[e.CurRow+1:]...)
		e.CurRow--
		e.CurCol = joinCol
		e.Dirty = true
	}
}

func editorDeleteChar(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	editorPushUndo(e)
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))
	if col < len(line) {
		newLine := make([]rune, 0, len(line)-1)
		newLine = append(newLine, line[:col]...)
		newLine = append(newLine, line[col+1:]...)
		e.Lines[e.CurRow] = string(newLine)
		e.Dirty = true
	}
	editorClampCursorToLine(e)
}

func editorDeleteLine(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	editorPushUndo(e)
	if len(e.Lines) == 1 {
		e.Lines = []string{""}
		e.CurCol = 0
	} else {
		e.Lines = append(e.Lines[:e.CurRow], e.Lines[e.CurRow+1:]...)
		if e.CurRow >= len(e.Lines) {
			e.CurRow = len(e.Lines) - 1
		}
	}
	editorClampCursorToLine(e)
	e.Dirty = true
}

func editorSplitLine(e *editorState) {
	if len(e.Lines) == 0 {
		e.Lines = []string{""}
	}
	editorPushUndo(e)
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))
	before := string(line[:col])
	after := string(line[col:])
	newLines := make([]string, 0, len(e.Lines)+1)
	newLines = append(newLines, e.Lines[:e.CurRow]...)
	newLines = append(newLines, before, after)
	newLines = append(newLines, e.Lines[e.CurRow+1:]...)
	e.Lines = newLines
	e.CurRow++
	e.CurCol = 0
	e.Dirty = true
}

func editorOpenLineBelow(e *editorState) {
	if len(e.Lines) == 0 {
		e.Lines = []string{""}
	}
	editorPushUndo(e)
	newLines := make([]string, 0, len(e.Lines)+1)
	newLines = append(newLines, e.Lines[:e.CurRow+1]...)
	newLines = append(newLines, "")
	newLines = append(newLines, e.Lines[e.CurRow+1:]...)
	e.Lines = newLines
	e.CurRow++
	e.CurCol = 0
	e.Mode = editorModeInsert
	e.Dirty = true
}

func editorOpenLineAbove(e *editorState) {
	if len(e.Lines) == 0 {
		e.Lines = []string{""}
	}
	editorPushUndo(e)
	newLines := make([]string, 0, len(e.Lines)+1)
	newLines = append(newLines, e.Lines[:e.CurRow]...)
	newLines = append(newLines, "")
	newLines = append(newLines, e.Lines[e.CurRow:]...)
	e.Lines = newLines
	e.CurCol = 0
	e.Mode = editorModeInsert
	e.Dirty = true
}

// --- Cursor movement ---

func editorMoveLeft(e *editorState) {
	if e.CurCol > 0 {
		e.CurCol--
	}
}

func editorMoveRight(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	lineLen := len([]rune(e.Lines[e.CurRow]))
	maxCol := lineLen
	if e.Mode == editorModeNormal && lineLen > 0 {
		maxCol = lineLen - 1
	}
	if e.CurCol < maxCol {
		e.CurCol++
	}
}

func editorMoveUp(e *editorState) {
	if e.CurRow > 0 {
		e.CurRow--
		editorClampCursorToLine(e)
	}
}

func editorMoveDown(e *editorState) {
	if e.CurRow < len(e.Lines)-1 {
		e.CurRow++
		editorClampCursorToLine(e)
	}
}

func editorMoveToLineStart(e *editorState) {
	e.CurCol = 0
}

func editorMoveToLineEnd(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	lineLen := len([]rune(e.Lines[e.CurRow]))
	if e.Mode == editorModeNormal && lineLen > 0 {
		e.CurCol = lineLen - 1
	} else {
		e.CurCol = lineLen
	}
}

func editorMoveWordForward(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))

	// Skip current word characters.
	for col < len(line) && !unicode.IsSpace(line[col]) {
		col++
	}
	// Skip whitespace.
	for col < len(line) && unicode.IsSpace(line[col]) {
		col++
	}
	e.CurCol = col
	editorClampCursorToLine(e)
}

func editorMoveWordBackward(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))

	// Skip whitespace backward.
	for col > 0 && unicode.IsSpace(line[col-1]) {
		col--
	}
	// Skip word characters backward.
	for col > 0 && !unicode.IsSpace(line[col-1]) {
		col--
	}
	e.CurCol = col
}

func editorMoveToTop(e *editorState) {
	e.CurRow = 0
	editorClampCursorToLine(e)
}

func editorMoveToBottom(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	e.CurRow = len(e.Lines) - 1
	editorClampCursorToLine(e)
}

func editorClampCursorToLine(e *editorState) {
	if len(e.Lines) == 0 {
		e.CurCol = 0
		return
	}
	e.CurRow = clampInt(e.CurRow, 0, len(e.Lines)-1)
	lineLen := len([]rune(e.Lines[e.CurRow]))
	maxCol := lineLen
	if e.Mode == editorModeNormal && lineLen > 0 {
		maxCol = lineLen - 1
	}
	if maxCol < 0 {
		maxCol = 0
	}
	e.CurCol = clampInt(e.CurCol, 0, maxCol)
}

func editorEnsureCursorVisible(e *editorState, viewHeight int) {
	if viewHeight <= 0 {
		return
	}
	if e.CurRow < e.ScrollRow {
		e.ScrollRow = e.CurRow
	}
	if e.CurRow >= e.ScrollRow+viewHeight {
		e.ScrollRow = e.CurRow - viewHeight + 1
	}
	if e.ScrollRow < 0 {
		e.ScrollRow = 0
	}
}

// editorAdvanceFocus cycles through Session → Title → Body.
func editorAdvanceFocus(e *editorState, delta int) {
	order := []editorFocus{editorFocusSession, editorFocusTitle, editorFocusBody}
	cur := 0
	for i, f := range order {
		if f == e.Focus {
			cur = i
			break
		}
	}
	next := (cur + delta + len(order)) % len(order)
	e.Focus = order[next]
}

// --- Key dispatch ---

func (s *Shell) handleEditorKeyEvent(event *tcell.EventKey, _ Action) (handleResult, bool) {
	if s.editor == nil || event == nil {
		return handleResult{}, false
	}

	e := s.editor
	e.StatusText = ""

	switch e.Mode {
	case editorModeCommand:
		return s.handleEditorCommandKey(event), true
	case editorModeInsert:
		return s.handleEditorInsertKey(event), true
	default:
		return s.handleEditorNormalKey(event), true
	}
}

func (s *Shell) handleEditorNormalKey(event *tcell.EventKey) handleResult {
	e := s.editor

	// Handle header fields (session / title) — j/Down/Tab advance, k/Up go back.
	if e.Focus == editorFocusSession || e.Focus == editorFocusTitle {
		switch event.Key() { //nolint:exhaustive // only handle relevant keys
		case tcell.KeyTab, tcell.KeyDown:
			editorAdvanceFocus(e, 1)
			return handleResult{Redraw: true}
		case tcell.KeyBacktab, tcell.KeyUp:
			editorAdvanceFocus(e, -1)
			return handleResult{Redraw: true}
		case tcell.KeyEsc:
			return s.editorTryQuit(false)
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				editorAdvanceFocus(e, 1)
				return handleResult{Redraw: true}
			case 'k':
				editorAdvanceFocus(e, -1)
				return handleResult{Redraw: true}
			case 'i':
				e.Mode = editorModeInsert
				return handleResult{Redraw: true}
			case 'a':
				e.Mode = editorModeInsert
				return handleResult{Redraw: true}
			case ':':
				e.Mode = editorModeCommand
				e.CmdBuffer = ""
				return handleResult{Redraw: true}
			}
		}
		return handleResult{Redraw: true}
	}

	// Body focus — full vim normal mode.
	switch event.Key() { //nolint:exhaustive // only handle relevant keys
	case tcell.KeyEsc:
		return s.editorTryQuit(false)
	case tcell.KeyTab:
		editorAdvanceFocus(e, 1)
		return handleResult{Redraw: true}
	case tcell.KeyBacktab:
		editorAdvanceFocus(e, -1)
		return handleResult{Redraw: true}
	case tcell.KeyLeft:
		editorMoveLeft(e)
		return handleResult{Redraw: true}
	case tcell.KeyRight:
		editorMoveRight(e)
		return handleResult{Redraw: true}
	case tcell.KeyUp:
		editorMoveUp(e)
		return handleResult{Redraw: true}
	case tcell.KeyDown:
		editorMoveDown(e)
		return handleResult{Redraw: true}
	case tcell.KeyRune:
		// Handle two-key sequences.
		if e.PendingKey != 0 {
			pending := e.PendingKey
			e.PendingKey = 0
			return s.handleEditorTwoKey(pending, event.Rune())
		}

		switch event.Rune() {
		case 'h':
			editorMoveLeft(e)
		case 'l':
			editorMoveRight(e)
		case 'j':
			editorMoveDown(e)
		case 'k':
			editorMoveUp(e)
		case '0':
			editorMoveToLineStart(e)
		case '$':
			editorMoveToLineEnd(e)
		case 'w':
			editorMoveWordForward(e)
		case 'b':
			editorMoveWordBackward(e)
		case 'G':
			editorMoveToBottom(e)
		case 'g':
			e.PendingKey = 'g'
		case 'd':
			e.PendingKey = 'd'
		case 'i':
			e.Mode = editorModeInsert
		case 'a':
			editorMoveRight(e)
			e.Mode = editorModeInsert
		case 'o':
			editorOpenLineBelow(e)
		case 'O':
			editorOpenLineAbove(e)
		case 'x':
			editorDeleteChar(e)
		case 'u':
			if !editorUndo(e) {
				e.StatusText = "Already at oldest change"
			}
		case ':':
			e.Mode = editorModeCommand
			e.CmdBuffer = ""
		}
		return handleResult{Redraw: true}
	}

	return handleResult{Redraw: true}
}

func (s *Shell) handleEditorTwoKey(first, second rune) handleResult {
	e := s.editor
	switch {
	case first == 'g' && second == 'g':
		editorMoveToTop(e)
	case first == 'd' && second == 'd':
		editorDeleteLine(e)
	}
	return handleResult{Redraw: true}
}

func (s *Shell) handleEditorInsertKey(event *tcell.EventKey) handleResult {
	e := s.editor

	if e.Focus == editorFocusSession {
		return s.handleEditorInsertSession(event)
	}
	if e.Focus == editorFocusTitle {
		return s.handleEditorInsertTitle(event)
	}

	switch event.Key() { //nolint:exhaustive // only handle relevant keys
	case tcell.KeyEsc:
		e.Mode = editorModeNormal
		editorClampCursorToLine(e)
		return handleResult{Redraw: true}
	case tcell.KeyLeft:
		editorMoveLeft(e)
		return handleResult{Redraw: true}
	case tcell.KeyRight:
		editorMoveRight(e)
		return handleResult{Redraw: true}
	case tcell.KeyUp:
		editorMoveUp(e)
		return handleResult{Redraw: true}
	case tcell.KeyDown:
		editorMoveDown(e)
		return handleResult{Redraw: true}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		editorBackspace(e)
		return handleResult{Redraw: true}
	case tcell.KeyEnter:
		editorSplitLine(e)
		return handleResult{Redraw: true}
	case tcell.KeyRune:
		editorInsertRune(e, event.Rune())
		return handleResult{Redraw: true}
	}

	return handleResult{Redraw: true}
}

func (s *Shell) handleEditorInsertTitle(event *tcell.EventKey) handleResult {
	e := s.editor
	switch event.Key() { //nolint:exhaustive // only handle relevant keys
	case tcell.KeyEsc:
		e.Mode = editorModeNormal
		return handleResult{Redraw: true}
	case tcell.KeyTab, tcell.KeyDown:
		editorAdvanceFocus(e, 1)
		return handleResult{Redraw: true}
	case tcell.KeyBacktab, tcell.KeyUp:
		editorAdvanceFocus(e, -1)
		return handleResult{Redraw: true}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		runes := []rune(e.Title)
		if len(runes) > 0 {
			e.Title = string(runes[:len(runes)-1])
			e.Dirty = true
		}
		return handleResult{Redraw: true}
	case tcell.KeyRune:
		e.Title += string(event.Rune())
		e.Dirty = true
		return handleResult{Redraw: true}
	}
	return handleResult{Redraw: true}
}

func (s *Shell) handleEditorInsertSession(event *tcell.EventKey) handleResult {
	e := s.editor
	switch event.Key() { //nolint:exhaustive // only handle relevant keys
	case tcell.KeyEsc:
		e.Mode = editorModeNormal
		return handleResult{Redraw: true}
	case tcell.KeyTab, tcell.KeyDown:
		editorAdvanceFocus(e, 1)
		return handleResult{Redraw: true}
	case tcell.KeyBacktab, tcell.KeyUp:
		editorAdvanceFocus(e, -1)
		return handleResult{Redraw: true}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.SessionNum > 0 {
			e.SessionNum /= 10
			e.Dirty = true
		}
		return handleResult{Redraw: true}
	case tcell.KeyRune:
		if unicode.IsDigit(event.Rune()) {
			e.SessionNum = e.SessionNum*10 + int(event.Rune()-'0')
			e.Dirty = true
		}
		return handleResult{Redraw: true}
	}
	return handleResult{Redraw: true}
}

func (s *Shell) handleEditorCommandKey(event *tcell.EventKey) handleResult {
	e := s.editor
	switch event.Key() { //nolint:exhaustive // only handle relevant keys
	case tcell.KeyEsc:
		e.Mode = editorModeNormal
		e.CmdBuffer = ""
		return handleResult{Redraw: true}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.CmdBuffer) > 0 {
			runes := []rune(e.CmdBuffer)
			e.CmdBuffer = string(runes[:len(runes)-1])
		} else {
			e.Mode = editorModeNormal
		}
		return handleResult{Redraw: true}
	case tcell.KeyEnter:
		return s.executeEditorCommand()
	case tcell.KeyRune:
		e.CmdBuffer += string(event.Rune())
		return handleResult{Redraw: true}
	}
	return handleResult{Redraw: true}
}

func (s *Shell) executeEditorCommand() handleResult {
	e := s.editor
	cmd := strings.TrimSpace(e.CmdBuffer)
	e.Mode = editorModeNormal
	e.CmdBuffer = ""

	switch cmd {
	case "w":
		return s.editorSave(false)
	case "q":
		return s.editorTryQuit(false)
	case "q!":
		return s.editorForceQuit()
	case "wq", "x":
		return s.editorSave(true)
	default:
		e.StatusText = "Unknown command: :" + cmd
		return handleResult{Redraw: true}
	}
}

func (s *Shell) editorSave(quitAfter bool) handleResult {
	e := s.editor
	s.editorSaveInFlight = true
	s.editorQuitAfterSave = quitAfter

	command := &Command{
		ID:      e.CommandID,
		Section: e.Section,
		ItemKey: e.ItemKey,
		Fields: map[string]string{
			"title": editorComposeTitle(e),
			"body":  strings.Join(e.Lines, "\n"),
		},
	}
	return handleResult{Command: command}
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

func (s *Shell) editorTryQuit(force bool) handleResult {
	e := s.editor
	if !force && e.Dirty {
		e.StatusText = "Unsaved changes! Use :q! to force quit, or :w to save."
		e.Mode = editorModeNormal
		return handleResult{Redraw: true}
	}
	s.editor = nil
	return handleResult{Redraw: true}
}

func (s *Shell) editorForceQuit() handleResult {
	s.editor = nil
	return handleResult{Redraw: true}
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

// --- Editor rendering ---

func (s *Shell) renderEditor(buffer *Buffer, rect Rect, theme *Theme) {
	if s.editor == nil || rect.Empty() {
		return
	}
	e := s.editor
	ss := SectionNotes.Style(theme)

	// Layout: split into left (editor) and right (sidebar).
	sidebarW := clampInt(rect.W/5, 18, 26)
	if rect.W < 60 {
		sidebarW = 0
	}

	var editorRect, sidebarRect Rect
	if sidebarW > 0 {
		editorRect, sidebarRect = rect.SplitVertical(rect.W-sidebarW-1, 1)
	} else {
		editorRect = rect
	}

	// Draw editor panel.
	editorTitle := "Edit Note"
	if e.CommandID == "notes.create" {
		editorTitle = "New Note"
	}
	DrawPanel(buffer, editorRect, theme, Panel{
		Title:       editorTitle,
		BorderStyle: &ss.Accent,
		TitleStyle:  &ss.Accent,
		Texture:     PanelTextureNone,
	})

	content := panelContentRect(editorRect, buffer.Bounds())
	if content.Empty() {
		return
	}

	y := content.Y

	// Session line.
	sessionLabel := "Session: "
	sessionStyle := theme.Muted
	sessionValueStyle := theme.Text
	if e.Focus == editorFocusSession {
		sessionValueStyle = theme.EditorCursor
	}
	buffer.WriteString(content.X, y, sessionStyle, sessionLabel)
	sessionX := content.X + len([]rune(sessionLabel))
	sessionW := content.W - len([]rune(sessionLabel))
	sessionDisplay := fmt.Sprintf("%d", e.SessionNum)
	if e.Focus == editorFocusSession && e.Mode == editorModeInsert {
		sessionDisplay += "_"
	}
	buffer.WriteString(sessionX, y, sessionValueStyle, clipText(sessionDisplay, sessionW))
	y++

	// Title line.
	titleLabel := "Title:   "
	titleStyle := theme.Muted
	titleValueStyle := theme.Text
	if e.Focus == editorFocusTitle {
		titleValueStyle = theme.EditorCursor
	}
	buffer.WriteString(content.X, y, titleStyle, titleLabel)
	titleX := content.X + len([]rune(titleLabel))
	titleW := content.W - len([]rune(titleLabel))
	titleDisplay := e.Title
	if e.Focus == editorFocusTitle && e.Mode == editorModeInsert {
		titleDisplay += "_"
	}
	buffer.WriteString(titleX, y, titleValueStyle, clipText(titleDisplay, titleW))
	y++

	// Separator.
	for x := content.X; x < content.X+content.W; x++ {
		buffer.Set(x, y, '─', theme.Muted)
	}
	y++

	// Body area.
	bodyHeight := max(1, content.Y+content.H-y-1) // reserve 1 for status

	gutterW := 4
	editorEnsureCursorVisible(e, bodyHeight)

	for row := range bodyHeight {
		lineIdx := e.ScrollRow + row
		lineY := y + row

		if lineIdx < len(e.Lines) {
			// Line number.
			numStr := fmt.Sprintf("%3d ", lineIdx+1)
			buffer.WriteString(content.X, lineY, theme.EditorLineNumber, numStr)

			// Line content.
			line := []rune(e.Lines[lineIdx])
			textX := content.X + gutterW
			textW := content.W - gutterW
			for col := 0; col < textW && col < len(line); col++ {
				style := theme.Text
				if e.Focus == editorFocusBody && lineIdx == e.CurRow && col == e.CurCol {
					style = theme.EditorCursor
				}
				buffer.Set(textX+col, lineY, line[col], style)
			}

			// Draw cursor on empty line or past end of text.
			if e.Focus == editorFocusBody && lineIdx == e.CurRow {
				cursorCol := clampInt(e.CurCol, 0, len(line))
				if cursorCol >= len(line) && cursorCol < textW {
					buffer.Set(textX+cursorCol, lineY, ' ', theme.EditorCursor)
				}
			}
		} else {
			// Past EOF: show tilde.
			buffer.WriteString(content.X, lineY, theme.EditorLineNumber, "  ~ ")
		}
	}

	// Status bar.
	statusY := content.Y + content.H - 1
	buffer.FillRect(Rect{X: content.X, Y: statusY, W: content.W, H: 1}, ' ', theme.EditorStatusBar)

	if e.Mode == editorModeCommand {
		cmdText := ":" + e.CmdBuffer + "_"
		buffer.WriteString(content.X, statusY, theme.EditorCommandLine, clipText(cmdText, content.W))
	} else {
		modeStr := "NORMAL"
		if e.Mode == editorModeInsert {
			modeStr = "INSERT"
		}
		focusStr := ""
		switch e.Focus {
		case editorFocusSession:
			focusStr = " [session]"
		case editorFocusTitle:
			focusStr = " [title]"
		default:
		}

		statusLeft := fmt.Sprintf("-- %s --%s", modeStr, focusStr)
		statusRight := fmt.Sprintf("ln:%d col:%d", e.CurRow+1, e.CurCol+1)

		if e.StatusText != "" {
			statusLeft = e.StatusText
		}

		buffer.WriteString(content.X, statusY, theme.EditorStatusBar, clipText(statusLeft, content.W))
		rightX := content.X + content.W - len([]rune(statusRight))
		if rightX > content.X+len([]rune(statusLeft))+2 {
			buffer.WriteString(rightX, statusY, theme.EditorStatusBar, statusRight)
		}
	}

	// Draw sidebar.
	if sidebarW > 0 && !sidebarRect.Empty() {
		s.renderEditorSidebar(buffer, sidebarRect, theme)
	}
}

func (s *Shell) renderEditorSidebar(buffer *Buffer, rect Rect, theme *Theme) {
	e := s.editor
	ss := SectionNotes.Style(theme)

	DrawPanel(buffer, rect, theme, Panel{
		Title:       "Info",
		BorderStyle: &ss.Accent,
		TitleStyle:  &ss.Accent,
		Texture:     PanelTextureNone,
	})

	content := panelContentRect(rect, buffer.Bounds())
	if content.Empty() {
		return
	}

	y := content.Y

	// References parsed from body.
	refs := editorParseReferences(e)
	if len(refs) > 0 {
		buffer.WriteString(content.X, y, theme.Muted, "References:")
		y++
		for _, ref := range refs {
			if y >= content.Y+content.H {
				break
			}
			buffer.WriteString(content.X, y, theme.SectionNotes, clipText("  "+ref, content.W))
			y++
		}
		y++
	}

	// Help.
	if y < content.Y+content.H {
		buffer.WriteString(content.X, y, theme.Muted, "Help:")
		y++
	}
	helpLines := []string{
		":w save",
		":q quit",
		":wq save+quit",
		"i insert",
		"o new line",
		"dd del line",
		"u undo",
		"Tab title/body",
	}
	for _, line := range helpLines {
		if y >= content.Y+content.H {
			break
		}
		buffer.WriteString(content.X, y, theme.Text, clipText("  "+line, content.W))
		y++
	}
}

// editorParseReferences finds @type/name patterns in body lines.
func editorParseReferences(e *editorState) []string {
	if e == nil {
		return nil
	}
	seen := make(map[string]bool)
	var refs []string
	for _, line := range e.Lines {
		for i := range len(line) {
			if line[i] == '@' && i+1 < len(line) {
				end := i + 1
				for end < len(line) && !isRefTerminator(line[end]) {
					end++
				}
				ref := line[i:end]
				if strings.Contains(ref, "/") && len(ref) > 2 && !seen[ref] {
					seen[ref] = true
					refs = append(refs, ref)
				}
			}
		}
	}
	return refs
}

func isRefTerminator(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == ',' || b == '.' || b == ')' || b == ']'
}
