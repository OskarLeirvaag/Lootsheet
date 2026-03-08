package render

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// Action is the semantic meaning of a keyboard binding.
type Action string

const (
	ActionNone          Action = ""
	ActionQuit          Action = "quit"
	ActionRedraw        Action = "redraw"
	ActionNextSection   Action = "next_section"
	ActionPrevSection   Action = "prev_section"
	ActionShowDashboard Action = "show_dashboard"
	ActionShowAccounts  Action = "show_accounts"
	ActionShowJournal   Action = "show_journal"
	ActionShowQuests    Action = "show_quests"
	ActionShowLoot      Action = "show_loot"
	ActionMoveUp        Action = "move_up"
	ActionMoveDown      Action = "move_down"
	ActionPageUp        Action = "page_up"
	ActionPageDown      Action = "page_down"
	ActionMoveTop       Action = "move_top"
	ActionMoveBottom    Action = "move_bottom"
	ActionToggle        Action = "toggle"
	ActionReverse       Action = "reverse"
	ActionCollect       Action = "collect"
	ActionWriteOff      Action = "write_off"
	ActionConfirm       Action = "confirm"
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

// DefaultKeyMap returns the interactive keyboard controls.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Bindings: []Binding{
			{
				Action: ActionNextSection,
				Stroke: KeyStroke{Key: tcell.KeyRight},
				Label:  "←→ section",
			},
			{
				Action: ActionPrevSection,
				Stroke: KeyStroke{Key: tcell.KeyLeft},
			},
			{
				Action: ActionNextSection,
				Stroke: KeyStroke{Key: tcell.KeyTab},
			},
			{
				Action: ActionPrevSection,
				Stroke: KeyStroke{Key: tcell.KeyBacktab},
			},
			{
				Action: ActionNextSection,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'l'},
			},
			{
				Action: ActionPrevSection,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'h'},
			},
			{
				Action: ActionMoveDown,
				Stroke: KeyStroke{Key: tcell.KeyDown},
				Label:  "↑↓ select",
			},
			{
				Action: ActionMoveUp,
				Stroke: KeyStroke{Key: tcell.KeyUp},
			},
			{
				Action: ActionMoveDown,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'j'},
			},
			{
				Action: ActionMoveUp,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'k'},
			},
			{
				Action: ActionPageDown,
				Stroke: KeyStroke{Key: tcell.KeyPgDn},
			},
			{
				Action: ActionPageUp,
				Stroke: KeyStroke{Key: tcell.KeyPgUp},
			},
			{
				Action: ActionMoveTop,
				Stroke: KeyStroke{Key: tcell.KeyHome},
			},
			{
				Action: ActionMoveBottom,
				Stroke: KeyStroke{Key: tcell.KeyEnd},
			},
			{
				Action: ActionShowDashboard,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '1'},
				Label:  "1-5 jump",
			},
			{
				Action: ActionShowAccounts,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '2'},
			},
			{
				Action: ActionShowJournal,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '3'},
			},
			{
				Action: ActionShowQuests,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '4'},
			},
			{
				Action: ActionShowLoot,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '5'},
			},
			{
				Action: ActionToggle,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 't'},
			},
			{
				Action: ActionReverse,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'r'},
			},
			{
				Action: ActionCollect,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'c'},
			},
			{
				Action: ActionWriteOff,
				Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'w'},
			},
			{
				Action: ActionConfirm,
				Stroke: KeyStroke{Key: tcell.KeyEnter},
			},
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
				Label:  "Ctrl+L refresh",
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
	return k.HelpTextFor()
}

// HelpTextFor formats footer help for the provided action subset.
func (k KeyMap) HelpTextFor(actions ...Action) string {
	allowed := make(map[Action]struct{}, len(actions))
	for _, action := range actions {
		allowed[action] = struct{}{}
	}

	labels := make([]string, 0, len(k.withDefaults().Bindings))
	seen := make(map[string]struct{}, len(k.withDefaults().Bindings))
	for _, binding := range k.withDefaults().Bindings {
		if binding.Label == "" {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[binding.Action]; !ok {
				continue
			}
		}
		if _, ok := seen[binding.Label]; ok {
			continue
		}
		labels = append(labels, binding.Label)
		seen[binding.Label] = struct{}{}
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
