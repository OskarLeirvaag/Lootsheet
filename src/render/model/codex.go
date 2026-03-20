package model

// CodexFormField describes a single input field in a codex form.
type CodexFormField struct {
	ID          string
	Label       string
	Placeholder string
}

// CodexForm defines the fields for a codex entry type.
type CodexForm struct {
	ID     string
	Fields []CodexFormField
}

var codexFormRegistry = map[string]CodexForm{
	"player": {
		ID: "player",
		Fields: []CodexFormField{
			{ID: "name", Label: "Name", Placeholder: "Thorin Ironfist"},
			{ID: "class", Label: "Class", Placeholder: "Fighter"},
			{ID: "race", Label: "Race", Placeholder: "Dwarf"},
			{ID: "background", Label: "Background", Placeholder: "Noble"},
			{ID: "notes", Label: "Notes", Placeholder: "@[quest/...] cross-references"},
		},
	},
	"npc": {
		ID: "npc",
		Fields: []CodexFormField{
			{ID: "name", Label: "Name", Placeholder: "Mayor Elra"},
			{ID: "title", Label: "Title", Placeholder: "Mayor"},
			{ID: "location", Label: "Location", Placeholder: "Millhaven"},
			{ID: "faction", Label: "Faction", Placeholder: "Town Council"},
			{ID: "disposition", Label: "Disposition", Placeholder: "friendly"},
			{ID: "description", Label: "Description", Placeholder: "Tall woman with silver hair"},
			{ID: "notes", Label: "Notes", Placeholder: "Met during @[quest/...]"},
		},
	},
	"settlement": {
		ID: "settlement",
		Fields: []CodexFormField{
			{ID: "name", Label: "Name", Placeholder: "Millhaven"},
			{ID: "title", Label: "Title", Placeholder: "Mining town"},
			{ID: "location", Label: "Location", Placeholder: "Northern foothills"},
			{ID: "faction", Label: "Faction", Placeholder: "Free Cities Alliance"},
			{ID: "description", Label: "Description", Placeholder: "A bustling trade hub at the river fork"},
			{ID: "notes", Label: "Notes", Placeholder: "Visited during @[quest/...]"},
		},
	},
}

// LookupCodexForm returns the form definition for the given form ID.
func LookupCodexForm(formID string) (CodexForm, bool) {
	f, ok := codexFormRegistry[formID]
	return f, ok
}
