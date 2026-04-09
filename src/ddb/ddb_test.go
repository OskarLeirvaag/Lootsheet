package ddb

import (
	"encoding/json"
	"testing"
)

func TestFormatCR(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{0, "0"},
		{0.125, "1/8"},
		{0.25, "1/4"},
		{0.5, "1/2"},
		{1, "1"},
		{5, "5"},
		{17, "17"},
		{30, "30"},
	}
	for _, tt := range tests {
		got := formatCR(tt.value)
		if got != tt.want {
			t.Errorf("formatCR(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestConfigLookups(t *testing.T) {
	cfg := &Config{
		ChallengeRatings: []ConfigChallengeRating{
			{ID: 1, Value: 0},
			{ID: 3, Value: 0.25},
			{ID: 27, Value: 17},
		},
		CreatureSizes: []ConfigCreatureSize{
			{ID: 2, Name: "Tiny"},
			{ID: 3, Name: "Small"},
			{ID: 4, Name: "Medium"},
		},
		MonsterTypes: []ConfigMonsterType{
			{ID: 11, Name: "Humanoid"},
			{ID: 13, Name: "Dragon"},
		},
		Sources: []ConfigSource{
			{ID: 1, Description: "Basic Rules"},
			{ID: 2, Description: "Player's Handbook"},
		},
	}

	if got := cfg.ChallengeRatingLabel(3); got != "1/4" {
		t.Errorf("CR label for id=3: got %q, want '1/4'", got)
	}
	if got := cfg.ChallengeRatingLabel(27); got != "17" {
		t.Errorf("CR label for id=27: got %q, want '17'", got)
	}
	if got := cfg.ChallengeRatingLabel(999); got != "?" {
		t.Errorf("CR label for missing: got %q, want '?'", got)
	}
	if got := cfg.CreatureSizeName(3); got != "Small" {
		t.Errorf("size name for id=3: got %q, want 'Small'", got)
	}
	if got := cfg.MonsterTypeName(11); got != "Humanoid" {
		t.Errorf("type name for id=11: got %q, want 'Humanoid'", got)
	}
	if got := cfg.SourceName(2); got != "Player's Handbook" {
		t.Errorf("source name for id=2: got %q, want 'Player's Handbook'", got)
	}
}

func TestParseMonsterResponse(t *testing.T) {
	raw := `{"data":[{"id":16907,"name":"Goblin","armorClass":15,"armorClassDescription":"(leather armor, shield)","averageHitPoints":7,"hitPointDice":{"diceString":"2d6"},"challengeRatingId":3,"sizeId":3,"typeId":11,"tags":["Goblinoid"],"stats":[{"statId":1,"value":8}],"movements":[{"movementId":1,"speed":30}],"sensesHtml":"Darkvision 60 ft.","skillsHtml":"Stealth +6","languageDescription":"Common, Goblin","passivePerception":9,"specialTraitsDescription":"<p>Nimble Escape.</p>","actionsDescription":"<p>Scimitar.</p>","bonusActionsDescription":"","reactionsDescription":"","legendaryActionsDescription":"","characteristicsDescription":"","sources":[{"sourceId":1}],"isHomebrew":false,"isReleased":true,"url":"https://www.dndbeyond.com/monsters/16907-goblin"}]}`

	var resp MonsterResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 monster, got %d", len(resp.Data))
	}

	m := resp.Data[0]
	if m.Name != "Goblin" {
		t.Fatalf("name = %q", m.Name)
	}
	if m.ChallengeRatingID != 3 {
		t.Fatalf("CR id = %d", m.ChallengeRatingID)
	}
	if FormatMonsterHP(&m) != "7 (2d6)" {
		t.Fatalf("HP = %q", FormatMonsterHP(&m))
	}
	if FormatMonsterAC(&m) != "15 (leather armor, shield)" {
		t.Fatalf("AC = %q", FormatMonsterAC(&m))
	}
}

func TestParseSpellResponse(t *testing.T) {
	raw := `{"data":[{"id":2307,"definition":{"id":2110,"name":"Floating Disk","level":1,"school":"Conjuration","components":[1,2,3],"componentsDescription":"a drop of mercury","concentration":false,"ritual":true,"duration":{"durationInterval":1,"durationType":"Time","durationUnit":"Hour"},"range":{"origin":"Ranged","rangeValue":30},"description":"<p>Creates a disk.</p>","sources":[{"sourceId":1}],"isHomebrew":false,"tags":["Utility"]}}]}`

	var resp SpellResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 spell, got %d", len(resp.Data))
	}

	def := &resp.Data[0].Definition
	if def.Name != "Floating Disk" {
		t.Fatalf("name = %q", def.Name)
	}
	if FormatSpellComponents(def) != "V, S, M (a drop of mercury)" {
		t.Fatalf("components = %q", FormatSpellComponents(def))
	}
	if FormatSpellDuration(def) != "1 Hour" {
		t.Fatalf("duration = %q", FormatSpellDuration(def))
	}
	if FormatSpellRange(def) != "30 ft" {
		t.Fatalf("range = %q", FormatSpellRange(def))
	}
}

func TestParseItemResponse(t *testing.T) {
	raw := `{"data":[{"id":4570,"name":"Amulet of the Planes","type":"Wondrous item","filterType":"Wondrous item","rarity":"Very Rare","canAttune":true,"magic":true,"description":"<p>Plane shift.</p>","tags":["Teleportation"],"sources":[{"sourceId":2}],"isHomebrew":false}]}`

	var resp ItemResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Data))
	}

	item := &resp.Data[0]
	if item.Name != "Amulet of the Planes" {
		t.Fatalf("name = %q", item.Name)
	}
	if item.Rarity != "Very Rare" {
		t.Fatalf("rarity = %q", item.Rarity)
	}
	if !item.CanAttune {
		t.Fatal("expected attunement")
	}
	if ItemTypeName(item) != "Wondrous item" {
		t.Fatalf("type = %q", ItemTypeName(item))
	}
}

func TestParseConfigConditions(t *testing.T) {
	raw := `{"sources":[],"conditions":[{"definition":{"id":1,"name":"Blinded","description":"<p>Can't see.</p>","slug":"blinded","type":1}}],"basicActions":[{"id":1,"name":"Attack","description":"<p>Make an attack.</p>"}],"rules":[{"id":34,"name":"Race","description":"Affects your character."}],"weaponProperties":[{"id":1,"name":"Ammunition","description":"<p>Requires ammo.</p>"}],"challengeRatings":[{"id":1,"value":0}],"creatureSizes":[{"id":2,"name":"Tiny"}],"monsterTypes":[{"id":11,"name":"Humanoid"}]}`

	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(cfg.Conditions) != 1 || cfg.Conditions[0].Definition.Name != "Blinded" {
		t.Fatalf("conditions: %+v", cfg.Conditions)
	}
	if len(cfg.BasicActions) != 1 || cfg.BasicActions[0].Name != "Attack" {
		t.Fatalf("actions: %+v", cfg.BasicActions)
	}
	if len(cfg.Rules) != 1 || cfg.Rules[0].Name != "Race" {
		t.Fatalf("rules: %+v", cfg.Rules)
	}
	if len(cfg.WeaponProperties) != 1 || cfg.WeaponProperties[0].Name != "Ammunition" {
		t.Fatalf("weapon props: %+v", cfg.WeaponProperties)
	}
}
