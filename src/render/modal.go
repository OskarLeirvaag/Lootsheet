package render

import "github.com/gdamore/tcell/v2"

// Modal is the interface implemented by overlay components (search, confirm,
// input, compose, editor, glossary, codex picker). Each modal handles its
// own key events and rendering while the Shell manages the lifecycle.
type Modal interface {
	// HandleKeyEvent processes a raw key event. It returns a result and
	// whether the event was consumed. Consumed events are not forwarded
	// to the default key map.
	HandleKeyEvent(event *tcell.EventKey, action Action) (HandleResult, bool)

	// Render draws the modal onto the provided buffer region.
	Render(buffer *Buffer, rect Rect, theme *Theme)

	// FooterHelp returns the help text shown in the footer while this
	// modal is active.
	FooterHelp() string
}

// ShellUI is the interface used by the event loop (app.go) to drive the
// interactive TUI. It decouples the run loop from the concrete Shell type,
// making it easier to test and extend.
type ShellUI interface {
	HandleKeyEvent(event *tcell.EventKey, keymap KeyMap) HandleResult
	HandleAction(action Action) HandleResult
	Render(buffer *Buffer, theme *Theme, keymap KeyMap)
	Reload(data *ShellData)
	SetStatus(status StatusMessage)
	ApplyInputError(message string)
	CloseModal()
	Navigate(section Section, selectedKey string)
	TickRain()
}

// Verify Shell satisfies ShellUI at compile time.
var _ ShellUI = (*Shell)(nil)
