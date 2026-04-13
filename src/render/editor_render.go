package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// --- Editor rendering ---

func (s *Shell) renderEditor(buffer *Buffer, rect Rect, theme *Theme) { //nolint:revive // TUI editor rendering
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

	// Pre-compute code fence state per line for syntax highlighting.
	inCodeFence := make([]bool, len(e.Lines))
	fenceOpen := false
	for i, ln := range e.Lines {
		trimmed := strings.TrimLeft(ln, " \t")
		if strings.HasPrefix(trimmed, "```") {
			inCodeFence[i] = true
			fenceOpen = !fenceOpen
		} else {
			inCodeFence[i] = fenceOpen
		}
	}

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

		// Line content with syntax highlighting.
		fullLine := []rune(e.Lines[vr.lineIdx])
		lineStyles := editorLineStyles(fullLine, inCodeFence[vr.lineIdx], theme)

		for i, r := range vr.runes {
			bufCol := vr.colStart + i
			style := theme.Text
			if bufCol < len(lineStyles) {
				style = lineStyles[bufCol]
			}
			// Search match highlighting (overlays syntax, under cursor).
			if e.SearchActive {
				if matchIdx := editorMatchAt(e, vr.lineIdx, bufCol); matchIdx >= 0 {
					if matchIdx == e.SearchIndex {
						style = theme.EditorSearchCurrent
					} else {
						style = theme.EditorSearchMatch
					}
				}
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

	switch e.Mode {
	case editorModeCommand:
		cmdText := ":" + e.CmdBuffer + "_"
		buffer.WriteString(content.X, statusY, theme.EditorCommandLine, clipText(cmdText, content.W))
	case editorModeSearch:
		searchText := "/" + e.SearchBuffer + "_"
		buffer.WriteString(content.X, statusY, theme.EditorCommandLine, clipText(searchText, content.W))
	default:
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
		":fmt format doc",
		"i/a/A/I insert",
		"o/O new line",
		"w/e/b word move",
		"0/^/$ line move",
		"gg/G top/bottom",
		"cw change word",
		"cc change line",
		"dw del word",
		"dd del line",
		"D/C del/chg→end",
		"yy yank line",
		"p paste",
		"J join lines",
		"x del char",
		"u undo",
		"/  search",
		"n/N next/prev",
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

// editorMatchAt returns the match index if (row, col) is inside a search match, or -1.
func editorMatchAt(e *editorState, row, col int) int {
	for i, m := range e.SearchMatches {
		if m.Row == row && col >= m.ColStart && col < m.ColEnd {
			return i
		}
	}
	return -1
}

// --- Syntax highlighting ---

// editorLineStyles computes a per-rune style for a line in the editor.
func editorLineStyles(line []rune, inCodeFence bool, theme *Theme) []tcell.Style {
	n := len(line)
	styles := make([]tcell.Style, n)
	for i := range styles {
		styles[i] = theme.Text
	}
	if n == 0 {
		return styles
	}

	// Block-level styles take over the whole line and short-circuit inline parsing.
	if blockStyle, ok := editorBlockLineStyle(line, inCodeFence, theme); ok {
		for i := range styles {
			styles[i] = blockStyle
		}
		return styles
	}

	// List markers: color the bullet/number without overriding the rest.
	editorApplyListMarker(line, styles, theme)

	// Inline formatting (bold, code, references).
	editorApplyInlineStyles(line, styles, theme)

	return styles
}

// editorBlockLineStyle returns the whole-line style for headings, blockquotes,
// code fences, and code fence content. The second return reports whether the
// line has a block-level style that should apply to every rune.
func editorBlockLineStyle(line []rune, inCodeFence bool, theme *Theme) (tcell.Style, bool) {
	if inCodeFence {
		return theme.EditorCode, true
	}
	trimmed := strings.TrimLeft(string(line), " \t")
	if strings.HasPrefix(trimmed, "```") {
		return theme.EditorCode, true
	}
	if strings.HasPrefix(trimmed, "### ") || strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "# ") {
		return theme.EditorHeading, true
	}
	if strings.HasPrefix(trimmed, "> ") {
		return theme.EditorBlockquote, true
	}
	return tcell.Style{}, false
}

// editorApplyListMarker colors the bullet or number prefix on a list line.
func editorApplyListMarker(line []rune, styles []tcell.Style, theme *Theme) {
	n := len(line)
	trimmed := strings.TrimLeft(string(line), " \t")
	offset := n - len([]rune(trimmed))

	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		if offset < n {
			styles[offset] = theme.EditorListMarker
		}
		return
	}

	dotIdx := strings.Index(trimmed, ". ")
	if dotIdx <= 0 || dotIdx > 3 {
		return
	}
	for _, r := range trimmed[:dotIdx] {
		if r < '0' || r > '9' {
			return
		}
	}
	for i := offset; i < offset+dotIdx+1 && i < n; i++ {
		styles[i] = theme.EditorListMarker
	}
}

// editorApplyInlineStyles overlays inline formatting styles onto a per-rune style slice.
func editorApplyInlineStyles(line []rune, styles []tcell.Style, theme *Theme) {
	n := len(line)
	for i := 0; i < n; i++ {
		if skip := editorTryInlineBold(line, styles, i, theme); skip > 0 {
			i += skip
			continue
		}
		if skip := editorTryInlineCode(line, styles, i, theme); skip > 0 {
			i += skip
			continue
		}
		if skip := editorTryInlineRef(line, styles, i, theme); skip > 0 {
			i += skip
			continue
		}
	}
}

// editorTryInlineBold matches **bold** at i and returns the number of runes to skip past.
func editorTryInlineBold(line []rune, styles []tcell.Style, i int, theme *Theme) int {
	n := len(line)
	if i+1 >= n || line[i] != '*' || line[i+1] != '*' {
		return 0
	}
	end := editorFindDoubleChar(line, i+2, '*')
	if end < 0 {
		return 0
	}
	for j := i; j < end+2 && j < n; j++ {
		styles[j] = theme.EditorBold
	}
	return end + 1 - i
}

// editorTryInlineCode matches `code` at i and returns the number of runes to skip past.
func editorTryInlineCode(line []rune, styles []tcell.Style, i int, theme *Theme) int {
	n := len(line)
	if line[i] != '`' {
		return 0
	}
	end := editorFindChar(line, i+1, '`')
	if end < 0 {
		return 0
	}
	for j := i; j <= end && j < n; j++ {
		styles[j] = theme.EditorCode
	}
	return end - i
}

// editorTryInlineRef matches @[type/name] at i and returns the number of runes to skip past.
func editorTryInlineRef(line []rune, styles []tcell.Style, i int, theme *Theme) int {
	n := len(line)
	if line[i] != '@' || i+1 >= n || line[i+1] != '[' {
		return 0
	}
	end := i + 2
	for end < n && line[end] != ']' {
		end++
	}
	if end >= n || line[end] != ']' {
		return 0
	}
	ref := string(line[i : end+1])
	if !strings.Contains(ref, "/") {
		return 0
	}
	for j := i; j <= end; j++ {
		styles[j] = theme.EditorReference
	}
	return end - i
}

// editorFindDoubleChar finds position of two consecutive occurrences of ch.
func editorFindDoubleChar(line []rune, start int, ch rune) int {
	for i := start; i+1 < len(line); i++ {
		if line[i] == ch && line[i+1] == ch {
			return i
		}
	}
	return -1
}

// editorFindChar finds position of a single character.
func editorFindChar(line []rune, start int, ch rune) int {
	for i := start; i < len(line); i++ {
		if line[i] == ch {
			return i
		}
	}
	return -1
}

// editorParseReferences finds @[type/name] patterns in body lines.
func editorParseReferences(e *editorState) []string {
	if e == nil {
		return nil
	}
	seen := make(map[string]bool)
	var refs []string
	for _, line := range e.Lines {
		runes := []rune(line)
		for i := range runes {
			if runes[i] == '@' && i+1 < len(runes) && runes[i+1] == '[' {
				end := i + 2
				for end < len(runes) && runes[end] != ']' {
					end++
				}
				if end < len(runes) && runes[end] == ']' {
					ref := string(runes[i : end+1])
					if strings.Contains(ref, "/") && !seen[ref] {
						seen[ref] = true
						refs = append(refs, ref)
					}
				}
			}
		}
	}
	return refs
}
