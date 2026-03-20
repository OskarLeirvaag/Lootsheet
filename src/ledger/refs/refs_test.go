package refs

import (
	"testing"
)

func TestParseReferences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []ParsedRef
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "no references",
			input: "Just a plain note with no refs.",
			want:  nil,
		},
		{
			name:  "quest ref with spaces",
			input: "See @[quest/Clear the Tower] for details.",
			want: []ParsedRef{
				{TargetType: "quest", TargetName: "Clear the Tower"},
			},
		},
		{
			name:  "person ref",
			input: "Met @[person/Mayor Elra] at the gate.",
			want: []ParsedRef{
				{TargetType: "person", TargetName: "Mayor Elra"},
			},
		},
		{
			name:  "note referencing another note",
			input: "Continued from @[note/Session 3].",
			want: []ParsedRef{
				{TargetType: "note", TargetName: "Session 3"},
			},
		},
		{
			name:  "loot ref",
			input: "Found @[loot/Ruby Pendant]",
			want: []ParsedRef{
				{TargetType: "loot", TargetName: "Ruby Pendant"},
			},
		},
		{
			name:  "asset ref",
			input: "Stored @[asset/Ship]",
			want: []ParsedRef{
				{TargetType: "asset", TargetName: "Ship"},
			},
		},
		{
			name:  "multiple refs",
			input: "@[quest/Dragon Slaying] and @[person/Garrick]",
			want: []ParsedRef{
				{TargetType: "quest", TargetName: "Dragon Slaying"},
				{TargetType: "person", TargetName: "Garrick"},
			},
		},
		{
			name:  "loot and asset refs",
			input: "@[loot/Ruby Pendant] @[asset/Ship]",
			want: []ParsedRef{
				{TargetType: "loot", TargetName: "Ruby Pendant"},
				{TargetType: "asset", TargetName: "Ship"},
			},
		},
		{
			name:  "unknown type ignored",
			input: "See @[spell/Fireball] for more.",
			want:  nil,
		},
		{
			name:  "single word ref",
			input: "See @[quest/Rescue]",
			want: []ParsedRef{
				{TargetType: "quest", TargetName: "Rescue"},
			},
		},
		{
			name:  "unclosed bracket ignored",
			input: "See @[quest/Rescue for more.",
			want:  nil,
		},
		{
			name:  "old format without brackets ignored",
			input: "See @quest/Rescue",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseReferences(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseReferences(%q) returned %d refs, want %d: %v", tt.input, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i].TargetType != tt.want[i].TargetType {
					t.Errorf("ref[%d].TargetType = %q, want %q", i, got[i].TargetType, tt.want[i].TargetType)
				}
				if got[i].TargetName != tt.want[i].TargetName {
					t.Errorf("ref[%d].TargetName = %q, want %q", i, got[i].TargetName, tt.want[i].TargetName)
				}
			}
		})
	}
}
