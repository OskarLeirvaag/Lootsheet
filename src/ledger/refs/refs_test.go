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
			name:  "quest ref terminated by period",
			input: "See @quest/Clear the Tower.",
			want: []ParsedRef{
				{TargetType: "quest", TargetName: "Clear the Tower"},
			},
		},
		{
			name:  "person ref terminated by period",
			input: "Met @person/Mayor Elra.",
			want: []ParsedRef{
				{TargetType: "person", TargetName: "Mayor Elra"},
			},
		},
		{
			name:  "note referencing another note",
			input: "Continued from @note/Session 3.",
			want: []ParsedRef{
				{TargetType: "note", TargetName: "Session 3"},
			},
		},
		{
			name:  "loot ref at end of string",
			input: "Found @loot/Ruby Pendant",
			want: []ParsedRef{
				{TargetType: "loot", TargetName: "Ruby Pendant"},
			},
		},
		{
			name:  "asset ref at end of string",
			input: "Stored @asset/Ship",
			want: []ParsedRef{
				{TargetType: "asset", TargetName: "Ship"},
			},
		},
		{
			name:  "multiple refs separated by @",
			input: "@quest/Dragon Slaying @person/Garrick",
			want: []ParsedRef{
				{TargetType: "quest", TargetName: "Dragon Slaying"},
				{TargetType: "person", TargetName: "Garrick"},
			},
		},
		{
			name:  "loot and asset refs separated by @",
			input: "@loot/Ruby Pendant @asset/Ship",
			want: []ParsedRef{
				{TargetType: "loot", TargetName: "Ruby Pendant"},
				{TargetType: "asset", TargetName: "Ship"},
			},
		},
		{
			name:  "trailing period stripped",
			input: "Talked to @person/Garrick.",
			want: []ParsedRef{
				{TargetType: "person", TargetName: "Garrick"},
			},
		},
		{
			name:  "unknown type ignored",
			input: "See @spell/Fireball for more.",
			want:  nil,
		},
		{
			name:  "single word ref at end",
			input: "See @quest/Rescue",
			want: []ParsedRef{
				{TargetType: "quest", TargetName: "Rescue"},
			},
		},
		{
			name:  "ref with trailing exclamation",
			input: "Found @loot/Gem!",
			want: []ParsedRef{
				{TargetType: "loot", TargetName: "Gem"},
			},
		},
		{
			name:  "ref with trailing semicolon",
			input: "@person/Elra; @person/Garrick.",
			want: []ParsedRef{
				{TargetType: "person", TargetName: "Elra"},
				{TargetType: "person", TargetName: "Garrick"},
			},
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
