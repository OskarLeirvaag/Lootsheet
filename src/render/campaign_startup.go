package render

import (
	"context"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// RunCampaignCreator opens a minimal TUI to prompt for the first campaign name.
// It returns the entered name. Used when no campaigns exist yet.
func RunCampaignCreator(ctx context.Context, factory ScreenFactory) (string, error) {
	theme := resolveTheme(nil)
	terminal, err := OpenTerminal(factory, &theme)
	if err != nil {
		return "", err
	}
	defer terminal.Close()

	var name string
	var errorText string

	draw := func() {
		bounds := terminal.Bounds()
		buffer := NewBuffer(bounds.W, bounds.H, theme.Base)

		lines := []string{
			"Welcome to LootSheet!",
			"",
			"No campaigns found. Enter a name for your first campaign:",
			"",
			"Name: " + name + "_",
		}
		if errorText != "" {
			lines = append(lines, "Error: "+errorText)
		}
		lines = append(lines, "", "Enter confirm  Ctrl+C quit")

		rect := bounds.Inset(1)
		accent := theme.HeaderLabel
		DrawPanel(buffer, modalBounds(rect, lines, 60, 40, 70, 8), &theme, Panel{
			Title:       "New Campaign",
			Lines:       lines,
			BorderStyle: &accent,
			TitleStyle:  &accent,
			Texture:     PanelTextureNone,
		})
		terminal.Present(buffer, true)
	}

	draw()

	for {
		event := terminal.PollEvent()
		switch typed := event.(type) {
		case nil:
			return "", context.Canceled
		case *tcell.EventKey:
			switch typed.Key() {
			case tcell.KeyEnter:
				trimmed := strings.TrimSpace(name)
				if trimmed == "" {
					errorText = "Campaign name is required."
					draw()
					continue
				}
				return trimmed, nil
			case tcell.KeyEsc, tcell.KeyCtrlC:
				return "", context.Canceled
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				runes := []rune(name)
				if len(runes) > 0 {
					name = string(runes[:len(runes)-1])
					errorText = ""
				}
				draw()
			case tcell.KeyCtrlU:
				name = ""
				errorText = ""
				draw()
			case tcell.KeyRune:
				name += string(typed.Rune())
				errorText = ""
				draw()
			default: //nolint:exhaustive // only handle keys relevant to text input
			}
		case *tcell.EventResize:
			draw()
		default:
		}
	}
}

// RunCampaignPicker opens a minimal TUI to select from existing campaigns.
// Used when 2+ campaigns exist at startup.
func RunCampaignPicker(ctx context.Context, factory ScreenFactory, campaigns []model.CampaignOption) (string, error) { //nolint:revive // TUI event handling
	theme := resolveTheme(nil)
	terminal, err := OpenTerminal(factory, &theme)
	if err != nil {
		return "", err
	}
	defer terminal.Close()

	selectedIndex := 0

	draw := func() {
		bounds := terminal.Bounds()
		buffer := NewBuffer(bounds.W, bounds.H, theme.Base)

		lines := make([]string, 0, len(campaigns)+6)
		lines = append(lines, "Select a campaign:", "")
		for i, c := range campaigns {
			prefix := "   "
			if i == selectedIndex {
				prefix = " > "
			}
			lines = append(lines, prefix+c.Name)
		}
		lines = append(lines, "", "Enter select  ↑↓/jk navigate  Ctrl+C quit")

		rect := bounds.Inset(1)
		accent := theme.HeaderLabel
		DrawPanel(buffer, modalBounds(rect, lines, 46, 32, 58, 8), &theme, Panel{
			Title:       "Select Campaign",
			Lines:       lines,
			BorderStyle: &accent,
			TitleStyle:  &accent,
			Texture:     PanelTextureNone,
		})
		terminal.Present(buffer, true)
	}

	draw()

	for {
		event := terminal.PollEvent()
		switch typed := event.(type) {
		case nil:
			return "", context.Canceled
		case *tcell.EventKey:
			switch typed.Key() {
			case tcell.KeyEnter:
				if selectedIndex >= 0 && selectedIndex < len(campaigns) {
					return campaigns[selectedIndex].ID, nil
				}
			case tcell.KeyEsc, tcell.KeyCtrlC:
				return "", context.Canceled
			case tcell.KeyUp:
				if selectedIndex > 0 {
					selectedIndex--
					draw()
				}
			case tcell.KeyDown:
				if selectedIndex < len(campaigns)-1 {
					selectedIndex++
					draw()
				}
			case tcell.KeyRune:
				switch typed.Rune() {
				case 'j':
					if selectedIndex < len(campaigns)-1 {
						selectedIndex++
						draw()
					}
				case 'k':
					if selectedIndex > 0 {
						selectedIndex--
						draw()
					}
				case 'q':
					return "", context.Canceled
				default:
				}
			default: //nolint:exhaustive // only handle keys relevant to list navigation
			}
		case *tcell.EventResize:
			draw()
		default:
		}
	}
}
