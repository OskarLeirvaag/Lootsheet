package ledger

import "testing"

func TestAccountTypesAreValid(t *testing.T) {
	for _, value := range AccountTypes() {
		if !value.Valid() {
			t.Fatalf("account type %q should be valid", value)
		}
	}
}

func TestJournalEntryStatusesAreValid(t *testing.T) {
	for _, value := range JournalEntryStatuses() {
		if !value.Valid() {
			t.Fatalf("journal entry status %q should be valid", value)
		}
	}
}

func TestQuestStatusesAreValid(t *testing.T) {
	for _, value := range QuestStatuses() {
		if !value.Valid() {
			t.Fatalf("quest status %q should be valid", value)
		}
	}
}

func TestLootStatusesAreValid(t *testing.T) {
	for _, value := range LootStatuses() {
		if !value.Valid() {
			t.Fatalf("loot status %q should be valid", value)
		}
	}
}
