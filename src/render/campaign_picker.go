package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	campaignPickerModeSelect = iota
	campaignPickerModeCreate
	campaignPickerModeRename
)

type campaignPickerState struct {
	selectedIndex int
	mode          int
	inputValue    string
	errorText     string
}

func (s *Shell) openCampaignPicker() bool {
	if s == nil {
		return false
	}
	picker := &campaignPickerState{}
	if len(s.Data.Campaigns) == 0 {
		// No campaigns yet — go straight to create mode.
		picker.mode = campaignPickerModeCreate
	}
	s.campaignPicker = picker
	return true
}

func (s *Shell) handleCampaignPickerKeyEvent(event *tcell.EventKey, action Action) (handleResult, bool) {
	if s.campaignPicker == nil || event == nil {
		return handleResult{}, false
	}

	switch action {
	case ActionQuit:
		s.campaignPicker = nil
		return handleResult{Redraw: true}, true
	case ActionRedraw:
		s.campaignPicker = nil
		return handleResult{Reload: true}, true
	default:
	}

	picker := s.campaignPicker

	switch event.Key() {
	case tcell.KeyEsc:
		if picker.mode != campaignPickerModeSelect {
			picker.mode = campaignPickerModeSelect
			picker.inputValue = ""
			picker.errorText = ""
			return handleResult{Redraw: true}, true
		}
		s.campaignPicker = nil
		return handleResult{Redraw: true}, true

	case tcell.KeyEnter:
		switch picker.mode {
		case campaignPickerModeSelect:
			if picker.selectedIndex < 0 || picker.selectedIndex >= len(s.Data.Campaigns) {
				return handleResult{}, true
			}
			selected := s.Data.Campaigns[picker.selectedIndex]
			s.campaignPicker = nil
			return handleResult{Command: &Command{
				ID:      "campaign.switch",
				ItemKey: selected.ID,
			}}, true

		case campaignPickerModeCreate:
			name := strings.TrimSpace(picker.inputValue)
			if name == "" {
				picker.errorText = "Name is required."
				return handleResult{Redraw: true}, true
			}
			s.campaignPicker = nil
			return handleResult{Command: &Command{
				ID:     "campaign.create",
				Fields: map[string]string{"name": name},
			}}, true

		case campaignPickerModeRename:
			name := strings.TrimSpace(picker.inputValue)
			if name == "" {
				picker.errorText = "Name is required."
				return handleResult{Redraw: true}, true
			}
			s.campaignPicker = nil
			return handleResult{Command: &Command{
				ID:     "campaign.rename",
				Fields: map[string]string{"name": name},
			}}, true
		}
		return handleResult{}, true

	case tcell.KeyUp:
		if picker.mode == campaignPickerModeSelect && picker.selectedIndex > 0 {
			picker.selectedIndex--
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true

	case tcell.KeyDown:
		if picker.mode == campaignPickerModeSelect && picker.selectedIndex < len(s.Data.Campaigns)-1 {
			picker.selectedIndex++
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if picker.mode != campaignPickerModeSelect {
			runes := []rune(picker.inputValue)
			if len(runes) > 0 {
				picker.inputValue = string(runes[:len(runes)-1])
				picker.errorText = ""
			}
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true

	case tcell.KeyCtrlU:
		if picker.mode != campaignPickerModeSelect {
			picker.inputValue = ""
			picker.errorText = ""
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true

	case tcell.KeyRune:
		if picker.mode != campaignPickerModeSelect {
			picker.inputValue += string(event.Rune())
			picker.errorText = ""
			return handleResult{Redraw: true}, true
		}

		switch event.Rune() {
		case 'j':
			if picker.selectedIndex < len(s.Data.Campaigns)-1 {
				picker.selectedIndex++
				return handleResult{Redraw: true}, true
			}
		case 'k':
			if picker.selectedIndex > 0 {
				picker.selectedIndex--
				return handleResult{Redraw: true}, true
			}
		case 'a':
			picker.mode = campaignPickerModeCreate
			picker.inputValue = ""
			picker.errorText = ""
			return handleResult{Redraw: true}, true
		case 'u':
			picker.mode = campaignPickerModeRename
			if picker.selectedIndex >= 0 && picker.selectedIndex < len(s.Data.Campaigns) {
				picker.inputValue = s.Data.Campaigns[picker.selectedIndex].Name
			}
			picker.errorText = ""
			return handleResult{Redraw: true}, true
		}
		return handleResult{}, true

	default:
		return handleResult{}, true
	}
}

func (s *Shell) renderCampaignPickerModal(buffer *Buffer, rect Rect, theme *Theme) {
	if s.campaignPicker == nil || rect.Empty() {
		return
	}

	picker := s.campaignPicker
	lines := make([]string, 0, len(s.Data.Campaigns)+10)

	switch picker.mode {
	case campaignPickerModeCreate:
		display := picker.inputValue + "_"
		lines = append(lines, "New campaign name: "+display)
		if picker.errorText != "" {
			lines = append(lines, "Error: "+picker.errorText)
		}
		lines = append(lines, "", "Enter create  Esc cancel")

	case campaignPickerModeRename:
		display := picker.inputValue + "_"
		lines = append(lines, "Rename campaign: "+display)
		if picker.errorText != "" {
			lines = append(lines, "Error: "+picker.errorText)
		}
		lines = append(lines, "", "Enter rename  Esc cancel")

	default:
		lines = append(lines, "Campaigns:")
		for i, c := range s.Data.Campaigns {
			prefix := "   "
			if i == picker.selectedIndex {
				prefix = " > "
			}
			lines = append(lines, prefix+c.Name)
		}
		if picker.errorText != "" {
			lines = append(lines, "", "Error: "+picker.errorText)
		}
		lines = append(lines, "", fmt.Sprintf("Enter select  a create  u rename  %s  Esc close", "↑↓/jk navigate"))
	}

	accent := s.sectionStyle(theme)
	DrawPanel(buffer, modalBounds(rect, lines, 46, 32, 58, 8), theme, Panel{
		Title:       "Campaigns",
		Lines:       lines,
		BorderStyle: &accent,
		TitleStyle:  &accent,
		Texture:     PanelTextureNone,
	})
}
