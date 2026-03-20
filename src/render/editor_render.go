package render

import (
	"fmt"
	"strings"
)

// --- Editor rendering ---

func (s *Shell) renderEditor(buffer *Buffer, rect Rect, theme *Theme) {
	if s.editor == nil || rect.Empty() {
		return
	}
	e := s.editor
	ss := sectionStyleFor(SectionNotes, theme)

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

	// Body area with soft word-wrapping.
	bodyHeight := max(1, content.Y+content.H-y-1) // reserve 1 for status

	gutterW := 4
	textW := max(1, content.W-gutterW)
	textX := content.X + gutterW

	// Build visual rows from buffer lines.
	type vrow struct {
		lineIdx  int
		colStart int    // first rune index in the buffer line
		runes    []rune // slice of runes for this visual row
		isFirst  bool   // first visual row for this buffer line
	}
	var vrows []vrow
	cursorVRow := 0

	for lineIdx := range e.Lines {
		line := []rune(e.Lines[lineIdx])
		if len(line) == 0 {
			vr := vrow{lineIdx: lineIdx, colStart: 0, runes: nil, isFirst: true}
			if lineIdx == e.CurRow {
				cursorVRow = len(vrows)
			}
			vrows = append(vrows, vr)
			continue
		}
		first := true
		for off := 0; off < len(line); off += textW {
			end := min(off+textW, len(line))
			vr := vrow{lineIdx: lineIdx, colStart: off, runes: line[off:end], isFirst: first}
			if lineIdx == e.CurRow && e.CurCol >= off && (e.CurCol < end || end == len(line)) {
				cursorVRow = len(vrows)
			}
			vrows = append(vrows, vr)
			first = false
		}
	}

	// Scroll to keep cursor visible.
	if e.ScrollRow > cursorVRow {
		e.ScrollRow = cursorVRow
	}
	if cursorVRow >= e.ScrollRow+bodyHeight {
		e.ScrollRow = cursorVRow - bodyHeight + 1
	}
	if e.ScrollRow < 0 {
		e.ScrollRow = 0
	}

	// Render visible visual rows.
	for screenRow := range bodyHeight {
		vrowIdx := e.ScrollRow + screenRow
		lineY := y + screenRow

		if vrowIdx >= len(vrows) {
			buffer.WriteString(content.X, lineY, theme.EditorLineNumber, "  ~ ")
			continue
		}

		vr := vrows[vrowIdx]

		// Line number (only on first visual row of a buffer line).
		if vr.isFirst {
			numStr := fmt.Sprintf("%3d ", vr.lineIdx+1)
			buffer.WriteString(content.X, lineY, theme.EditorLineNumber, numStr)
		} else {
			buffer.WriteString(content.X, lineY, theme.EditorLineNumber, "    ")
		}

		// Line content with @ref highlighting.
		fullLine := []rune(e.Lines[vr.lineIdx])
		refSpans := editorRefSpans(fullLine)

		for i, r := range vr.runes {
			bufCol := vr.colStart + i
			style := theme.Text
			if editorColInRefSpan(bufCol, refSpans) {
				style = theme.EditorReference
			}
			if e.Focus == editorFocusBody && vr.lineIdx == e.CurRow && bufCol == e.CurCol {
				style = theme.EditorCursor
			}
			buffer.Set(textX+i, lineY, r, style)
		}

		// Draw cursor on empty line or past end of visual row.
		if e.Focus == editorFocusBody && vr.lineIdx == e.CurRow {
			visualCol := e.CurCol - vr.colStart
			if visualCol >= 0 && visualCol >= len(vr.runes) && visualCol < textW {
				buffer.Set(textX+visualCol, lineY, ' ', theme.EditorCursor)
			}
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

	// Ref picker overlay.
	if e.refPicker != nil {
		notesStyle := sectionStyleFor(SectionNotes, theme)
		renderPicker(e.refPicker, buffer, editorRect, theme, &notesStyle)
	}
}

func (s *Shell) renderEditorSidebar(buffer *Buffer, rect Rect, theme *Theme) {
	e := s.editor
	ss := sectionStyleFor(SectionNotes, theme)

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
		"C-a insert @ref",
	}
	for _, line := range helpLines {
		if y >= content.Y+content.H {
			break
		}
		buffer.WriteString(content.X, y, theme.Text, clipText("  "+line, content.W))
		y++
	}
}

// editorRefSpan marks start..end columns of a @ref in a single line.
type editorRefSpan struct{ start, end int }

// editorRefSpans returns the column ranges of @type/name references in line.
func editorRefSpans(line []rune) []editorRefSpan {
	var spans []editorRefSpan
	for i := range line {
		if line[i] == '@' && i+1 < len(line) {
			end := i + 1
			for end < len(line) && !isRefTerminatorRune(line[end]) {
				end++
			}
			ref := string(line[i:end])
			if strings.Contains(ref, "/") && len(ref) > 2 {
				spans = append(spans, editorRefSpan{i, end})
			}
		}
	}
	return spans
}

func editorColInRefSpan(col int, spans []editorRefSpan) bool {
	for _, sp := range spans {
		if col >= sp.start && col < sp.end {
			return true
		}
	}
	return false
}

// editorParseReferences finds @type/name patterns in body lines.
func editorParseReferences(e *editorState) []string {
	if e == nil {
		return nil
	}
	seen := make(map[string]bool)
	var refs []string
	for _, line := range e.Lines {
		runes := []rune(line)
		for i := range runes {
			if runes[i] == '@' && i+1 < len(runes) {
				end := i + 1
				for end < len(runes) && !isRefTerminatorRune(runes[end]) {
					end++
				}
				ref := string(runes[i:end])
				if strings.Contains(ref, "/") && len(ref) > 2 && !seen[ref] {
					seen[ref] = true
					refs = append(refs, ref)
				}
			}
		}
	}
	return refs
}
