package notes

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
			input: "Just a plain note body.",
			want:  nil,
		},
		{
			name:  "quest ref terminated by period",
			input: "Relates to @quest/Dragon Slaying.",
			want: []parsedRef{
				{TargetType: "quest", TargetName: "Dragon Slaying"},
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
			name:  "note referencing another note",
			input: "Continued from @note/Session 3.",
			want: []parsedRef{
				{TargetType: "note", TargetName: "Session 3"},
			},
		},
		{
			name:  "loot and asset refs separated by @",
			input: "@loot/Ruby Pendant @asset/Ship",
			want: []parsedRef{
				{TargetType: "loot", TargetName: "Ruby Pendant"},
				{TargetType: "asset", TargetName: "Ship"},
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
			input: "Cast @spell/Fireball.",
			want:  nil,
		},
		{
			name:  "ref at end of string",
			input: "See @quest/Rescue",
			want: []parsedRef{
				{TargetType: "quest", TargetName: "Rescue"},
			},
		},
		{
			name:  "refs separated by semicolons",
			input: "@person/Elra; @person/Garrick.",
			want: []parsedRef{
				{TargetType: "person", TargetName: "Elra"},
				{TargetType: "person", TargetName: "Garrick"},
			},
		},
		{
			name:  "ref with trailing exclamation",
			input: "Found @loot/Gem!",
			want: []parsedRef{
				{TargetType: "loot", TargetName: "Gem"},
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
