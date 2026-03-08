package render

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// Action is the semantic meaning of a keyboard binding.
type Action string

const (
	ActionNone   Action = ""
	ActionQuit   Action = "quit"
	ActionRedraw Action = "redraw"
)

// KeyStroke matches a specific tcell key event.
type KeyStroke struct {
	Key  tcell.Key
	Rune rune
	Mod  tcell.ModMask
}

// Binding maps a keystroke to a user-visible action label.
type Binding struct {
	Action Action
	Stroke KeyStroke
	Label  string
}

// KeyMap stores the available bindings for the current screen.
type KeyMap struct {
	Bindings []Binding
}

// DefaultKeyMap returns the first-slice keyboard controls.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Bindings: []Binding{
			{
				Action: ActionQuit,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'q'},
				Label:  "q quit",
			},
			{
				Action: ActionQuit,
				Stroke: KeyStroke{Key: tcell.KeyEsc},
				Label:  "Esc quit",
			},
			{
				Action: ActionRedraw,
				Stroke: KeyStroke{Key: tcell.KeyCtrlL},
				Label:  "Ctrl+L redraw",
			},
		},
	}
}

// Resolve converts a key event into an action.
func (k KeyMap) Resolve(event *tcell.EventKey) Action {
	if event == nil {
		return ActionNone
	}

	for _, binding := range k.withDefaults().Bindings {
		if binding.Stroke.matches(event) {
			return binding.Action
		}
	}

	return ActionNone
}

// HelpText formats the footer help content.
func (k KeyMap) HelpText() string {
	labels := make([]string, 0, len(k.withDefaults().Bindings))
	for _, binding := range k.withDefaults().Bindings {
		labels = append(labels, binding.Label)
	}

	return strings.Join(labels, "  ")
}

func (k KeyMap) withDefaults() KeyMap {
	if len(k.Bindings) == 0 {
		return DefaultKeyMap()
	}
	return k
}

func (k KeyStroke) matches(event *tcell.EventKey) bool {
	if event.Key() != k.Key {
		return false
	}
	if event.Modifiers() != k.Mod {
		return false
	}
	if k.Key != tcell.KeyRune {
		return true
	}
	return unicode.ToLower(event.Rune()) == unicode.ToLower(k.Rune)
}
