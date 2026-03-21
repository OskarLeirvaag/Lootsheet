package render

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// --- Key dispatch ---

func (s *Shell) handleEditorKeyEvent(event *tcell.EventKey, _ Action) (handleResult, bool) {
	if s.editor == nil || event == nil {
		return handleResult{}, false
	}

	// Ref picker intercepts all keys when open.
	if s.editor.refPicker != nil {
		return s.handleEditorRefPickerKey(event), true
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
		case '^':
			editorMoveToFirstNonBlank(e)
		case '$':
			editorMoveToLineEnd(e)
		case 'w':
			editorMoveWordForward(e)
		case 'e':
			editorMoveWordEnd(e)
		case 'b':
			editorMoveWordBackward(e)
		case 'G':
			editorMoveToBottom(e)
		case 'g':
			e.PendingKey = 'g'
		case 'd':
			e.PendingKey = 'd'
		case 'c':
			e.PendingKey = 'c'
		case 'y':
			e.PendingKey = 'y'
		case 'D':
			editorDeleteToEndOfLine(e)
		case 'C':
			editorChangeToEndOfLine(e)
		case 'J':
			editorJoinLines(e)
		case 'p':
			editorPaste(e)
		case 'i':
			e.Mode = editorModeInsert
		case 'I':
			editorMoveToFirstNonBlank(e)
			e.Mode = editorModeInsert
		case 'a':
			editorMoveRight(e)
			e.Mode = editorModeInsert
		case 'A':
			editorMoveToLineEnd(e)
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
	switch first {
	case 'g':
		if second == 'g' {
			editorMoveToTop(e)
		}
	case 'd':
		switch second {
		case 'd':
			editorDeleteLine(e)
		case 'w':
			editorDeleteWord(e)
		case '$':
			editorDeleteToEndOfLine(e)
		}
	case 'c':
		switch second {
		case 'c':
			editorChangeLine(e)
		case 'w':
			editorChangeWord(e)
		case '$':
			editorChangeToEndOfLine(e)
		}
	case 'y':
		if second == 'y' {
			editorYankLine(e)
			e.StatusText = "1 line yanked"
		}
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
	case tcell.KeyCtrlA:
		s.openEditorRefPicker()
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

func (s *Shell) handleEditorRefPickerKey(event *tcell.EventKey) handleResult {
	p := s.editor.refPicker
	closed, value := handlePickerKey(p, event)
	if closed {
		if value != "" {
			editorInsertString(s.editor, value)
		}
		s.editor.refPicker = nil
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
