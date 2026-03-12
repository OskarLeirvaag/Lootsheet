package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (s *Shell) openCodexPicker() bool {
	if s == nil || len(s.Data.CodexTypes) == 0 {
		return false
	}
	s.codexPicker = &codexPickerState{
		Section: s.Section,
	}
	return true
}

func (s *Shell) handleCodexPickerKeyEvent(event *tcell.EventKey, action Action) (handleResult, bool) {
	if s.codexPicker == nil || event == nil {
		return handleResult{}, false
	}

	switch action {
	case ActionQuit:
		s.codexPicker = nil
		return handleResult{Redraw: true}, true
	case ActionRedraw:
		s.codexPicker = nil
		return handleResult{Reload: true}, true
	}

	picker := s.codexPicker

	switch event.Key() {
	case tcell.KeyEsc:
		s.codexPicker = nil
		return handleResult{Redraw: true}, true
	case tcell.KeyEnter:
		if picker.Focus == 0 {
			// Move from name to type list
			picker.Focus = 1
			return handleResult{Redraw: true}, true
		}
		// Submit
		name := strings.TrimSpace(picker.Name)
		if name == "" {
			picker.ErrorText = "Name is required."
			picker.Focus = 0
			return handleResult{Redraw: true}, true
		}
		if picker.TypeIndex < 0 || picker.TypeIndex >= len(s.Data.CodexTypes) {
			return handleResult{Redraw: true}, true
		}
		t := s.Data.CodexTypes[picker.TypeIndex]
		s.codexPicker = nil
		s.compose = newCodexCompose(t.FormID, t.ID, name, s.Section)
		return handleResult{Redraw: true}, true
	case tcell.KeyTab:
		if picker.Focus == 0 {
			picker.Focus = 1
		} else {
			picker.Focus = 0
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyBacktab:
		if picker.Focus == 1 {
			picker.Focus = 0
		} else {
			picker.Focus = 1
		}
		return handleResult{Redraw: true}, true
	case tcell.KeyUp:
		if picker.Focus == 1 && picker.TypeIndex > 0 {
			picker.TypeIndex--
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyDown:
		if picker.Focus == 1 && picker.TypeIndex < len(s.Data.CodexTypes)-1 {
			picker.TypeIndex++
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if picker.Focus == 0 {
			runes := []rune(picker.Name)
			if len(runes) > 0 {
				picker.Name = string(runes[:len(runes)-1])
				picker.ErrorText = ""
			}
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyCtrlU:
		if picker.Focus == 0 {
			picker.Name = ""
			picker.ErrorText = ""
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true
	case tcell.KeyRune:
		if picker.Focus == 0 {
			picker.Name += string(event.Rune())
			picker.ErrorText = ""
			return handleResult{Redraw: true}, true
		}
		r := event.Rune()
		if picker.Focus == 1 {
			switch r {
			case 'j':
				if picker.TypeIndex < len(s.Data.CodexTypes)-1 {
					picker.TypeIndex++
					return handleResult{Redraw: true}, true
				}
			case 'k':
				if picker.TypeIndex > 0 {
					picker.TypeIndex--
					return handleResult{Redraw: true}, true
				}
			}
		}
		return handleResult{}, true
	default:
		return handleResult{}, true
	}
}

func (s *Shell) renderCodexPickerModal(buffer *Buffer, rect Rect, theme *Theme) {
	if s.codexPicker == nil || rect.Empty() {
		return
	}

	picker := s.codexPicker
	lines := make([]string, 0, len(s.Data.CodexTypes)+8)

	nameDisplay := picker.Name + "_"
	if picker.Focus != 0 {
		nameDisplay = picker.Name
		if nameDisplay == "" {
			nameDisplay = "[enter name]"
		}
	}
	lines = append(lines, "Name: "+nameDisplay)

	if strings.TrimSpace(picker.ErrorText) != "" {
		lines = append(lines, "Error: "+picker.ErrorText)
	}

	lines = append(lines, "", "Type:")
	for i, t := range s.Data.CodexTypes {
		prefix := "   "
		if i == picker.TypeIndex {
			prefix = " > "
		}
		lines = append(lines, prefix+t.Name)
	}

	lines = append(lines, "", fmt.Sprintf("Tab switch  Enter %s  Esc cancel", pickerEnterLabel(picker)))

	accent := s.sectionStyle(theme)
	DrawPanel(buffer, modalBounds(rect, lines, 40, 30, 50, 8), theme, Panel{
		Title:       "Add to Codex",
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
		Texture:     PanelTextureNone,
	})
}

func pickerEnterLabel(picker *codexPickerState) string {
	if picker.Focus == 0 {
		return "next"
	}
	return "create"
}
