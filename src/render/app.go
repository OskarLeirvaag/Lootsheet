package render

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type animationTick struct{}

// Options configures the interactive TUI shell.
type Options struct {
	ScreenFactory   ScreenFactory
	DashboardLoader DashboardLoader
	ShellLoader     ShellLoader
	CommandHandler  CommandHandler
	SearchHandler   SearchHandler
	Theme           Theme
	KeyMap          KeyMap
}

type cancelInterrupt struct{}

// Run opens the interactive boxed TUI shell and blocks until exit.
func Run(ctx context.Context, options *Options) error {
	theme := resolveTheme(nil)
	keymap := DefaultKeyMap()
	if options != nil {
		theme = resolveTheme(&options.Theme)
		keymap = options.KeyMap.WithDefaults()
	}

	var factory ScreenFactory
	if options != nil {
		factory = options.ScreenFactory
	}

	terminal, err := OpenTerminal(factory, &theme)
	if err != nil {
		return err
	}
	defer terminal.Close()

	cancelDone := make(chan struct{})
	defer close(cancelDone)

	go func() {
		select {
		case <-ctx.Done():
			_ = terminal.PostEvent(tcell.NewEventInterrupt(cancelInterrupt{}))
		case <-cancelDone:
		}
	}()

	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				_ = terminal.PostEvent(tcell.NewEventInterrupt(animationTick{}))
			case <-cancelDone:
				return
			}
		}
	}()

	data, _ := loadShellData(ctx, options)
	shell := NewShell(&data)
	if options != nil {
		shell.SetSearchHandler(options.SearchHandler)
	}
	drawFrame(terminal, shell, &theme, keymap, false)

	for {
		event := terminal.PollEvent()
		switch typed := event.(type) {
		case nil:
			return nil
		case *tcell.EventKey:
			result := shell.HandleKeyEvent(typed, keymap)
			if result.Quit {
				return nil
			}
			if result.Reload {
				data, err := loadShellData(ctx, options)
				if err != nil && isDisconnectError(err) {
					shell.SetDisconnected()
					drawFrame(terminal, shell, &theme, keymap, true)
					continue
				}
				shell.Reload(&data)
				drawFrame(terminal, shell, &theme, keymap, true)
				continue
			}
			if result.Command != nil {
				if options == nil || options.CommandHandler == nil {
					shell.SetStatus(StatusMessage{
						Level: StatusError,
						Text:  "TUI action handler unavailable.",
					})
					drawFrame(terminal, shell, &theme, keymap, true)
					continue
				}

				commandResult, err := options.CommandHandler(ctx, *result.Command)
				if err != nil {
					if isDisconnectError(err) {
						shell.SetDisconnected()
						drawFrame(terminal, shell, &theme, keymap, true)
						continue
					}
					var inputErr InputError
					if errors.As(err, &inputErr) {
						shell.ApplyInputError(inputErr.Error())
					} else {
						shell.CloseModal()
						shell.SetStatus(StatusMessage{
							Level: StatusError,
							Text:  err.Error(),
						})
					}
				} else {
					shell.Reload(&commandResult.Data)
					if commandResult.NavigateTo != SectionDashboard || commandResult.SelectItemKey != "" {
						shell.Navigate(commandResult.NavigateTo, commandResult.SelectItemKey)
					}
					shell.SetStatus(commandResult.Status)
				}
				drawFrame(terminal, shell, &theme, keymap, true)
				continue
			}
			if result.Redraw {
				drawFrame(terminal, shell, &theme, keymap, true)
			}
		case *tcell.EventMouse:
			var mouseResult handleResult
			switch typed.Buttons() { //nolint:exhaustive // only wheel events are relevant
			case tcell.WheelUp:
				mouseResult = shell.HandleAction(ActionMoveUp)
			case tcell.WheelDown:
				mouseResult = shell.HandleAction(ActionMoveDown)
			}
			if mouseResult.Redraw {
				drawFrame(terminal, shell, &theme, keymap, false)
			}
		case *tcell.EventResize:
			drawFrame(terminal, shell, &theme, keymap, true)
		case *tcell.EventInterrupt:
			switch typed.Data().(type) {
			case cancelInterrupt:
				return nil
			case animationTick:
				shell.TickRain()
				drawFrame(terminal, shell, &theme, keymap, true)
			default:
				drawFrame(terminal, shell, &theme, keymap, true)
			}
		}
	}
}

// DashboardLoader produces the dashboard snapshot shown in the TUI.
type DashboardLoader func(context.Context) (DashboardData, error)

// ShellLoader produces the full TUI snapshot shown in the TUI.
type ShellLoader func(context.Context) (ShellData, error)

func loadDashboardData(ctx context.Context, options *Options) DashboardData {
	if options == nil || options.DashboardLoader == nil {
		return DefaultDashboardData()
	}

	data, err := options.DashboardLoader(ctx)
	if err != nil {
		return ErrorDashboardData("Dashboard data unavailable.", err.Error())
	}

	return resolveDashboardData(&data)
}

func loadShellData(ctx context.Context, options *Options) (ShellData, error) {
	if options != nil && options.ShellLoader != nil {
		data, err := options.ShellLoader(ctx)
		if err != nil {
			return ErrorShellData("TUI data unavailable.", err.Error()), err
		}

		return resolveShellData(&data), nil
	}

	return ShellData{
		Dashboard: loadDashboardData(ctx, options),
	}, nil
}

// isDisconnectError returns true when the error indicates the server
// connection is gone (graceful shutdown, broken pipe, or EOF).
func isDisconnectError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "server shutting down") ||
		strings.HasPrefix(msg, "read response:") ||
		strings.HasPrefix(msg, "write request:")
}

func drawFrame(terminal *Terminal, shell ShellUI, theme *Theme, keymap KeyMap, full bool) {
	if shell == nil {
		return
	}

	bounds := terminal.Bounds()
	buffer := NewBuffer(bounds.W, bounds.H, theme.Base)
	shell.Render(buffer, theme, keymap)
	terminal.Present(buffer, full)
}
