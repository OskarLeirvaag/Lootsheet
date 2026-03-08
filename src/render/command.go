package render

import "context"

// Command is a user action emitted by the interactive shell for the app layer.
type Command struct {
	ID      string
	Section Section
	ItemKey string
	Fields  map[string]string
	Lines   []CommandLine
}

// CommandLine is a structured line payload for multi-line compose actions.
type CommandLine struct {
	Side        string
	AccountCode string
	Amount      string
	Memo        string
}

// CommandResult is the app-facing outcome of a TUI command.
type CommandResult struct {
	Data          ShellData
	Status        StatusMessage
	NavigateTo    Section
	SelectItemKey string
}

// StatusLevel describes the severity of a transient TUI status message.
type StatusLevel string

const (
	StatusInfo    StatusLevel = "info"
	StatusSuccess StatusLevel = "success"
	StatusError   StatusLevel = "error"
)

// StatusMessage is shown above the footer help line.
type StatusMessage struct {
	Level StatusLevel
	Text  string
}

// Empty reports whether the status message contains any visible content.
func (s StatusMessage) Empty() bool {
	return s.Text == ""
}

// InputError keeps an input modal open while surfacing a validation message.
type InputError struct {
	Message string
}

// Error implements error.
func (e InputError) Error() string {
	return e.Message
}

// CommandHandler performs an interactive TUI command and returns refreshed data.
type CommandHandler func(context.Context, Command) (CommandResult, error)
