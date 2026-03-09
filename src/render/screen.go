package render

import "github.com/gdamore/tcell/v2"

// Screen captures the small part of tcell used by the first dashboard slice.
type Screen interface {
	Init() error
	Fini()
	Size() (int, int)
	SetStyle(tcell.Style)
	Clear()
	HideCursor()
	Show()
	Sync()
	EnableMouse(...tcell.MouseFlags)
	PollEvent() tcell.Event
	PostEvent(tcell.Event) error
	SetContent(x int, y int, primary rune, combining []rune, style tcell.Style)
}

// ScreenFactory creates a terminal screen implementation.
type ScreenFactory func() (Screen, error)

// Terminal owns the screen lifecycle and frame presentation.
type Terminal struct {
	screen Screen
}

func defaultScreenFactory() (Screen, error) {
	return tcell.NewScreen()
}

// OpenTerminal initializes the terminal screen and prepares it for drawing.
func OpenTerminal(factory ScreenFactory, theme *Theme) (*Terminal, error) {
	if factory == nil {
		factory = defaultScreenFactory
	}

	screen, err := factory()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.EnableMouse(tcell.MouseButtonEvents)
	screen.SetStyle(theme.Base)
	screen.HideCursor()
	screen.Clear()

	return &Terminal{screen: screen}, nil
}

// Close restores the terminal to its prior state.
func (t *Terminal) Close() {
	if t == nil || t.screen == nil {
		return
	}
	t.screen.Fini()
}

// Bounds returns the full current terminal area.
func (t *Terminal) Bounds() Rect {
	if t == nil || t.screen == nil {
		return Rect{}
	}

	width, height := t.screen.Size()
	return Rect{W: width, H: height}
}

// PollEvent waits for the next terminal event.
func (t *Terminal) PollEvent() tcell.Event {
	if t == nil || t.screen == nil {
		return nil
	}
	return t.screen.PollEvent()
}

// PostEvent injects an event into the terminal event queue.
func (t *Terminal) PostEvent(event tcell.Event) error {
	if t == nil || t.screen == nil {
		return nil
	}
	return t.screen.PostEvent(event)
}

// Present flushes the frame and updates the visible screen.
func (t *Terminal) Present(buffer *Buffer, full bool) {
	if t == nil || t.screen == nil || buffer == nil {
		return
	}

	t.screen.HideCursor()
	buffer.Flush(t.screen)

	if full {
		t.screen.Sync()
		return
	}

	t.screen.Show()
}
