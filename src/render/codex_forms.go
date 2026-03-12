package render

import "github.com/OskarLeirvaag/Lootsheet/src/render/model"

// Type aliases re-export model codex types.
type CodexFormField = model.CodexFormField
type CodexForm = model.CodexForm
type CodexTypeOption = model.CodexTypeOption

// LookupCodexForm returns the form definition for the given form ID.
var LookupCodexForm = model.LookupCodexForm
