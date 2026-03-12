package render

import "github.com/OskarLeirvaag/Lootsheet/src/render/model"

// Type aliases re-export model command types.
type Command = model.Command
type CommandLine = model.CommandLine
type CommandResult = model.CommandResult
type StatusLevel = model.StatusLevel
type StatusMessage = model.StatusMessage
type InputError = model.InputError
type CommandHandler = model.CommandHandler

const (
	StatusInfo    = model.StatusInfo
	StatusSuccess = model.StatusSuccess
	StatusError   = model.StatusError
)
