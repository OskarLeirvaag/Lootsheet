package model

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
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
			{Action: ActionNextSection, Stroke: KeyStroke{Key: tcell.KeyRight}, Label: "←→ section"},
			{Action: ActionPrevSection, Stroke: KeyStroke{Key: tcell.KeyLeft}},
			{Action: ActionNextSection, Stroke: KeyStroke{Key: tcell.KeyTab}},
			{Action: ActionPrevSection, Stroke: KeyStroke{Key: tcell.KeyBacktab}},
			{Action: ActionNextSection, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'l'}},
			{Action: ActionPrevSection, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'h'}},
			{Action: ActionMoveDown, Stroke: KeyStroke{Key: tcell.KeyDown}, Label: "↑↓ select"},
			{Action: ActionMoveUp, Stroke: KeyStroke{Key: tcell.KeyUp}},
			{Action: ActionMoveDown, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'j'}},
			{Action: ActionMoveUp, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'k'}},
			{Action: ActionPageDown, Stroke: KeyStroke{Key: tcell.KeyPgDn}},
			{Action: ActionPageUp, Stroke: KeyStroke{Key: tcell.KeyPgUp}},
			{Action: ActionMoveTop, Stroke: KeyStroke{Key: tcell.KeyHome}},
			{Action: ActionMoveBottom, Stroke: KeyStroke{Key: tcell.KeyEnd}},
			{Action: ActionShowDashboard, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '1'}, Label: "1-7 jump"},
			{Action: ActionShowJournal, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '2'}},
			{Action: ActionShowQuests, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '3'}},
			{Action: ActionShowLoot, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '4'}},
			{Action: ActionShowAssets, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '5'}},
			{Action: ActionShowCodex, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '6'}},
			{Action: ActionShowNotes, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '7'}},
			{Action: ActionShowSettings, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '@'}, Label: "@ settings"},
			{Action: ActionNewExpense, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'e'}},
			{Action: ActionNewIncome, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'i'}},
			{Action: ActionNewCustom, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'a'}, Label: "e/i/a entry"},
			{Action: ActionEdit, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'u'}},
			{Action: ActionDelete, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'd'}},
			{Action: ActionToggle, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 't'}},
			{Action: ActionReverse, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'r'}},
			{Action: ActionCollect, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'c'}},
			{Action: ActionWriteOff, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'w'}},
			{Action: ActionAppraise, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'p'}},
			{Action: ActionRecognize, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'n'}},
			{Action: ActionSell, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 's'}},
			{Action: ActionEditTemplate, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'f'}},
			{Action: ActionExecuteTemplate, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'x'}},
			{Action: ActionSubmitCompose, Stroke: KeyStroke{Key: tcell.KeyCtrlS}},
			{Action: ActionConfirm, Stroke: KeyStroke{Key: tcell.KeyEnter}},
			{Action: ActionSearch, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '/'}, Label: "/ search"},
			{Action: ActionHelp, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: '?'}, Label: "? terms"},
			{Action: ActionQuit, Stroke: KeyStroke{Key: tcell.KeyRune, Rune: 'q'}, Label: "q quit"},
			{Action: ActionQuit, Stroke: KeyStroke{Key: tcell.KeyEsc}, Label: "Esc quit"},
			{Action: ActionRedraw, Stroke: KeyStroke{Key: tcell.KeyCtrlL}, Label: "Ctrl+L refresh"},
		},
	}
}

// Resolve converts a key event into an action.
func (k KeyMap) Resolve(event *tcell.EventKey) Action {
	if event == nil {
		return ActionNone
	}

	for _, binding := range k.WithDefaults().Bindings {
		if binding.Stroke.Matches(event) {
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

	labels := make([]string, 0, len(k.WithDefaults().Bindings))
	seen := make(map[string]struct{}, len(k.WithDefaults().Bindings))
	for _, binding := range k.WithDefaults().Bindings {
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

// WithDefaults returns the keymap, falling back to DefaultKeyMap if empty.
func (k KeyMap) WithDefaults() KeyMap {
	if len(k.Bindings) == 0 {
		return DefaultKeyMap()
	}
	return k
}

// Matches reports whether the keystroke matches the given event.
func (k KeyStroke) Matches(event *tcell.EventKey) bool {
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
