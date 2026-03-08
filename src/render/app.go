package render

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// Options configures the interactive TUI shell.
type Options struct {
	ScreenFactory   ScreenFactory
	DashboardLoader DashboardLoader
	ShellLoader     ShellLoader
	CommandHandler  CommandHandler
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
		keymap = options.KeyMap.withDefaults()
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

	data := loadShellData(ctx, options)
	shell := NewShell(&data)
	drawFrame(terminal, shell, &theme, keymap, false)

	for {
		event := terminal.PollEvent()
		switch typed := event.(type) {
		case nil:
			return nil
		case *tcell.EventKey:
			action := keymap.Resolve(typed)
			result := shell.HandleAction(action)
			if result.Quit {
				return nil
			}
			if result.Reload {
				data := loadShellData(ctx, options)
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

				data, status, err := options.CommandHandler(ctx, *result.Command)
				if err != nil {
					shell.SetStatus(StatusMessage{
						Level: StatusError,
						Text:  err.Error(),
					})
				} else {
					shell.Reload(&data)
					shell.SetStatus(status)
				}
				drawFrame(terminal, shell, &theme, keymap, true)
				continue
			}
			if result.Redraw {
				drawFrame(terminal, shell, &theme, keymap, true)
			}
		case *tcell.EventResize:
			drawFrame(terminal, shell, &theme, keymap, true)
		case *tcell.EventInterrupt:
			if _, ok := typed.Data().(cancelInterrupt); ok {
				return nil
			}
			drawFrame(terminal, shell, &theme, keymap, true)
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

func loadShellData(ctx context.Context, options *Options) ShellData {
	if options != nil && options.ShellLoader != nil {
		data, err := options.ShellLoader(ctx)
		if err != nil {
			return ErrorShellData("TUI data unavailable.", err.Error())
		}

		return resolveShellData(&data)
	}

	return ShellData{
		Dashboard: loadDashboardData(ctx, options),
	}
}

func drawFrame(terminal *Terminal, shell *Shell, theme *Theme, keymap KeyMap, full bool) {
	if shell == nil {
		return
	}

	bounds := terminal.Bounds()
	buffer := NewBuffer(bounds.W, bounds.H, theme.Base)
	shell.Render(buffer, theme, keymap)
	terminal.Present(buffer, full)
}
