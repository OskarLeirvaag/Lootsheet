package render

import "unicode"

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
