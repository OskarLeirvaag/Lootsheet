package proto

import (
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

func TestShellDataConversionIncludesCompendiumSections(t *testing.T) {
	data := model.ShellData{
		SettingsCompendium: model.ListScreenData{
			SummaryLines: []string{"Sources: 1 enabled / 2 total"},
			Items: []model.ListItemData{{
				Key: "source-1",
				Row: "[x] Player's Handbook",
			}},
		},
		CompendiumMonsters: model.ListScreenData{
			SummaryLines: []string{"Total monsters: 1"},
			Items: []model.ListItemData{{
				Key:         "monster-1",
				Row:         "1/4   beast        Wolf",
				DetailTitle: "Wolf",
			}},
		},
		CompendiumSpells: model.ListScreenData{
			SummaryLines: []string{"Total spells: 1"},
			Items:        []model.ListItemData{{Key: "spell-1", Row: "Cntrp Evocation Fire Bolt"}},
		},
		CompendiumItems: model.ListScreenData{
			SummaryLines: []string{"Total items: 1"},
			Items:        []model.ListItemData{{Key: "item-1", Row: "Common Potion Healing Potion"}},
		},
		CompendiumRules: model.ListScreenData{
			SummaryLines: []string{"Total rules: 1"},
			Items:        []model.ListItemData{{Key: "rule-1", Row: "Combat Attack"}},
		},
		CompendiumConditions: model.ListScreenData{
			SummaryLines: []string{"Total conditions: 1"},
			Items:        []model.ListItemData{{Key: "condition-1", Row: "Blinded"}},
		},
	}

	roundTrip := ShellDataFromProto(ShellDataToProto(&data))

	if got := roundTrip.SettingsCompendium.Items[0].Row; got != "[x] Player's Handbook" {
		t.Fatalf("settings compendium row = %q", got)
	}
	if got := roundTrip.CompendiumMonsters.Items[0].DetailTitle; got != "Wolf" {
		t.Fatalf("monster detail title = %q", got)
	}
	if got := roundTrip.CompendiumSpells.Items[0].Key; got != "spell-1" {
		t.Fatalf("spell key = %q", got)
	}
	if got := roundTrip.CompendiumItems.Items[0].Key; got != "item-1" {
		t.Fatalf("item key = %q", got)
	}
	if got := roundTrip.CompendiumRules.Items[0].Key; got != "rule-1" {
		t.Fatalf("rule key = %q", got)
	}
	if got := roundTrip.CompendiumConditions.Items[0].Key; got != "condition-1" {
		t.Fatalf("condition key = %q", got)
	}
}
