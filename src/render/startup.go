package render

import (
	"context"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// StartupChoice represents the user's selection from the startup modal.
type StartupChoice struct {
	Mode    string // "local" or "connect"
	Address string // server address (only for "connect")
}

// RunStartupPicker opens a minimal TUI to choose between local mode and
// connecting to a saved (or new) server. Returns the user's choice.
func RunStartupPicker(ctx context.Context, factory ScreenFactory, savedServers []string) (StartupChoice, error) {
	theme := resolveTheme(nil)
	terminal, err := OpenTerminal(factory, &theme)
	if err != nil {
		return StartupChoice{}, err
	}
	defer terminal.Close()

	// Build option list: Local, then saved servers, then Add new.
	type option struct {
		label string
		mode  string // "local", "connect", "new"
		addr  string
	}

	options := make([]option, 0, len(savedServers)+2)
	options = append(options, option{label: "Local database", mode: "local"})
	for _, addr := range savedServers {
		// Strip default port from display label.
		label, _ := strings.CutSuffix(addr, ":7547")
		options = append(options, option{label: label, mode: "connect", addr: addr})
	}
	options = append(options, option{label: "Connect to new server...", mode: "new"})

	selectedIndex := 0

	// New server input state.
	inputMode := false
	inputAddr := ""
	inputError := ""

	draw := func() {
		bounds := terminal.Bounds()
		buffer := NewBuffer(bounds.W, bounds.H, theme.Base)

		var lines []string

		if inputMode {
			lines = []string{
				"Enter server address (host:port):",
				"",
				"Address: " + inputAddr + "_",
			}
			if inputError != "" {
				lines = append(lines, "Error: "+inputError)
			}
			lines = append(lines, "", "Enter connect  Esc back")
		} else {
			lines = make([]string, 0, len(options)+4)
			lines = append(lines, "Select mode:", "")
			for i, opt := range options {
				prefix := "   "
				if i == selectedIndex {
					prefix = " > "
				}
				lines = append(lines, prefix+opt.label)
			}
			lines = append(lines, "", "Enter select  ↑↓/jk navigate  q quit")
		}

		rect := bounds.Inset(1)
		accent := theme.HeaderLabel
		DrawPanel(buffer, modalBounds(rect, lines, 52, 36, 64, 8), &theme, Panel{
			Title:       "LootSheet",
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
			return StartupChoice{}, context.Canceled
		case *tcell.EventKey:
			if inputMode {
				switch typed.Key() {
				case tcell.KeyEnter:
					addr := strings.TrimSpace(inputAddr)
					if addr == "" {
						inputError = "Address is required."
						draw()
						continue
					}
					if !strings.Contains(addr, ":") {
						addr += ":7547"
					}
					return StartupChoice{Mode: "connect", Address: addr}, nil
				case tcell.KeyEsc:
					inputMode = false
					inputAddr = ""
					inputError = ""
					draw()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					runes := []rune(inputAddr)
					if len(runes) > 0 {
						inputAddr = string(runes[:len(runes)-1])
						inputError = ""
					}
					draw()
				case tcell.KeyCtrlU:
					inputAddr = ""
					inputError = ""
					draw()
				case tcell.KeyRune:
					inputAddr += string(typed.Rune())
					inputError = ""
					draw()
				default:
				}
				continue
			}

			switch typed.Key() {
			case tcell.KeyEnter:
				if selectedIndex >= 0 && selectedIndex < len(options) {
					opt := options[selectedIndex]
					switch opt.mode {
					case "local":
						return StartupChoice{Mode: "local"}, nil
					case "connect":
						return StartupChoice{Mode: "connect", Address: opt.addr}, nil
					case "new":
						inputMode = true
						draw()
					}
				}
			case tcell.KeyEsc, tcell.KeyCtrlC:
				return StartupChoice{}, context.Canceled
			case tcell.KeyUp:
				if selectedIndex > 0 {
					selectedIndex--
					draw()
				}
			case tcell.KeyDown:
				if selectedIndex < len(options)-1 {
					selectedIndex++
					draw()
				}
			case tcell.KeyRune:
				switch typed.Rune() {
				case 'j':
					if selectedIndex < len(options)-1 {
						selectedIndex++
						draw()
					}
				case 'k':
					if selectedIndex > 0 {
						selectedIndex--
						draw()
					}
				case 'q':
					return StartupChoice{}, context.Canceled
				}
			default:
			}
		case *tcell.EventResize:
			draw()
		}
	}
}
