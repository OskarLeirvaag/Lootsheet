package render

import (
	"strings"
	"unicode"
)

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

func editorInsertString(e *editorState, s string) {
	for _, r := range s {
		editorInsertRune(e, r)
	}
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
		e.Clipboard = []string{string(line[col : col+1])}
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
	e.Clipboard = []string{e.Lines[e.CurRow]}
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

// editorDeleteToEndOfLine deletes from cursor to end of line, storing deleted text.
func editorDeleteToEndOfLine(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	editorPushUndo(e)
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))
	deleted := string(line[col:])
	e.Lines[e.CurRow] = string(line[:col])
	e.Clipboard = []string{deleted}
	editorClampCursorToLine(e)
	e.Dirty = true
}

// editorDeleteWord deletes from cursor to the start of the next word.
func editorDeleteWord(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	editorPushUndo(e)
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))
	end := col
	// Skip current word characters.
	for end < len(line) && !unicode.IsSpace(line[end]) {
		end++
	}
	// Skip trailing whitespace.
	for end < len(line) && unicode.IsSpace(line[end]) {
		end++
	}
	deleted := string(line[col:end])
	newLine := make([]rune, 0, len(line)-(end-col))
	newLine = append(newLine, line[:col]...)
	newLine = append(newLine, line[end:]...)
	e.Lines[e.CurRow] = string(newLine)
	e.Clipboard = []string{deleted}
	editorClampCursorToLine(e)
	e.Dirty = true
}

// editorChangeLine clears the current line content and enters insert mode.
func editorChangeLine(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	editorPushUndo(e)
	e.Clipboard = []string{e.Lines[e.CurRow]}
	e.Lines[e.CurRow] = ""
	e.CurCol = 0
	e.Mode = editorModeInsert
	e.Dirty = true
}

// editorChangeToEndOfLine deletes from cursor to end and enters insert mode.
func editorChangeToEndOfLine(e *editorState) {
	editorDeleteToEndOfLine(e)
	e.Mode = editorModeInsert
}

// editorChangeWord deletes the word at cursor and enters insert mode.
func editorChangeWord(e *editorState) {
	editorDeleteWord(e)
	e.Mode = editorModeInsert
}

// editorYankLine copies the current line to the clipboard.
func editorYankLine(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	e.Clipboard = []string{e.Lines[e.CurRow]}
}

// editorPaste inserts the clipboard contents after the cursor/line.
func editorPaste(e *editorState) {
	if len(e.Clipboard) == 0 {
		return
	}
	editorPushUndo(e)
	if len(e.Lines) == 0 {
		e.Lines = []string{""}
	}

	// Single-element clipboard from inline delete: insert after cursor.
	if len(e.Clipboard) == 1 {
		line := []rune(e.Lines[e.CurRow])
		col := clampInt(e.CurCol+1, 0, len(line)+1)
		text := []rune(e.Clipboard[0])
		newLine := make([]rune, 0, len(line)+len(text))
		newLine = append(newLine, line[:col]...)
		newLine = append(newLine, text...)
		newLine = append(newLine, line[col:]...)
		e.Lines[e.CurRow] = string(newLine)
		e.CurCol = max(0, col+len(text)-1)
	}

	editorClampCursorToLine(e)
	e.Dirty = true
}

// editorJoinLines joins the current line with the next one.
func editorJoinLines(e *editorState) {
	if len(e.Lines) == 0 || e.CurRow >= len(e.Lines)-1 {
		return
	}
	editorPushUndo(e)
	curLine := e.Lines[e.CurRow]
	nextLine := strings.TrimLeft(e.Lines[e.CurRow+1], " \t")
	joinCol := len([]rune(curLine))
	if curLine != "" && nextLine != "" {
		e.Lines[e.CurRow] = curLine + " " + nextLine
	} else {
		e.Lines[e.CurRow] = curLine + nextLine
	}
	e.Lines = append(e.Lines[:e.CurRow+1], e.Lines[e.CurRow+2:]...)
	e.CurCol = joinCol
	e.Dirty = true
}

// editorMoveWordEnd moves cursor to the end of the current/next word.
func editorMoveWordEnd(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	line := []rune(e.Lines[e.CurRow])
	col := clampInt(e.CurCol, 0, len(line))
	if col < len(line) {
		col++ // move past current char
	}
	// Skip whitespace.
	for col < len(line) && unicode.IsSpace(line[col]) {
		col++
	}
	// Move to end of word.
	for col < len(line)-1 && !unicode.IsSpace(line[col+1]) {
		col++
	}
	e.CurCol = clampInt(col, 0, max(0, len(line)-1))
}

// editorMoveToFirstNonBlank moves cursor to the first non-whitespace character.
func editorMoveToFirstNonBlank(e *editorState) {
	if len(e.Lines) == 0 {
		return
	}
	line := []rune(e.Lines[e.CurRow])
	for i, r := range line {
		if !unicode.IsSpace(r) {
			e.CurCol = i
			return
		}
	}
	e.CurCol = 0
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

// editorAdvanceFocus cycles through Session -> Title -> Body.
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
