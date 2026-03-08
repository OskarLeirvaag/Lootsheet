package render

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

// Options configures the first dashboard shell.
type Options struct {
	ScreenFactory ScreenFactory
	Theme         Theme
	KeyMap        KeyMap
}

type cancelInterrupt struct{}

// Run opens the first boxed-dashboard TUI slice and blocks until exit.
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

	dashboard := Dashboard{}
	drawFrame(terminal, dashboard, &theme, keymap, false)

	for {
		event := terminal.PollEvent()
		switch typed := event.(type) {
		case nil:
			return nil
		case *tcell.EventKey:
			switch keymap.Resolve(typed) {
			case ActionQuit:
				return nil
			case ActionRedraw:
				drawFrame(terminal, dashboard, &theme, keymap, true)
			case ActionNone:
			}
		case *tcell.EventResize:
			drawFrame(terminal, dashboard, &theme, keymap, true)
		case *tcell.EventInterrupt:
			if _, ok := typed.Data().(cancelInterrupt); ok {
				return nil
			}
			drawFrame(terminal, dashboard, &theme, keymap, true)
		}
	}
}

func drawFrame(terminal *Terminal, dashboard Dashboard, theme *Theme, keymap KeyMap, full bool) {
	bounds := terminal.Bounds()
	buffer := NewBuffer(bounds.W, bounds.H, theme.Base)
	dashboard.Render(buffer, theme, keymap)
	terminal.Present(buffer, full)
}
