package codex

import (
	"testing"
)

func TestParseReferences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []parsedRef
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
			want: []parsedRef{
				{TargetType: "quest", TargetName: "Clear the Tower"},
			},
		},
		{
			name:  "person ref terminated by period",
			input: "Met @person/Mayor Elra.",
			want: []parsedRef{
				{TargetType: "person", TargetName: "Mayor Elra"},
			},
		},
		{
			name:  "loot ref at end of string",
			input: "Found @loot/Ruby Pendant",
			want: []parsedRef{
				{TargetType: "loot", TargetName: "Ruby Pendant"},
			},
		},
		{
			name:  "asset ref at end of string",
			input: "Stored @asset/Ship",
			want: []parsedRef{
				{TargetType: "asset", TargetName: "Ship"},
			},
		},
		{
			name:  "multiple refs separated by @",
			input: "@quest/Dragon Slaying @person/Garrick",
			want: []parsedRef{
				{TargetType: "quest", TargetName: "Dragon Slaying"},
				{TargetType: "person", TargetName: "Garrick"},
			},
		},
		{
			name:  "trailing period stripped",
			input: "Talked to @person/Garrick.",
			want: []parsedRef{
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
			want: []parsedRef{
				{TargetType: "quest", TargetName: "Rescue"},
			},
		},
		{
			name:  "ref with trailing exclamation",
			input: "Found @loot/Gem!",
			want: []parsedRef{
				{TargetType: "loot", TargetName: "Gem"},
			},
		},
		{
			name:  "ref with trailing semicolon",
			input: "@person/Elra; @person/Garrick.",
			want: []parsedRef{
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
