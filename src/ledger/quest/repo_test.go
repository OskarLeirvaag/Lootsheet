package quest

import (
	"context"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestCreateQuestOffered(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Clear the Goblin Cave",
		Patron:             "Mayor Thornton",
		Description:        "Goblins infesting the east cave",
		PromisedBaseReward: 500,
		BonusConditions:    "No casualties",
		Status:             "offered",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if quest.ID == "" {
		t.Fatal("quest ID is empty")
	}

	if quest.Title != "Clear the Goblin Cave" {
		t.Fatalf("quest title = %q, want Clear the Goblin Cave", quest.Title)
	}

	if quest.Status != ledger.QuestStatusOffered {
		t.Fatalf("quest status = %q, want offered", quest.Status)
	}

	if quest.PromisedBaseReward != 500 {
		t.Fatalf("quest reward = %d, want 500", quest.PromisedBaseReward)
	}
}

func TestCreateQuestAccepted(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Escort the Merchant",
		PromisedBaseReward: 200,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if quest.Status != ledger.QuestStatusAccepted {
		t.Fatalf("quest status = %q, want accepted", quest.Status)
	}

	if quest.AcceptedOn != "2026-03-01" {
		t.Fatalf("quest accepted_on = %q, want 2026-03-01", quest.AcceptedOn)
	}
}

func TestCreateQuestAcceptedRequiresDate(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	_, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:  "No Date Quest",
		Status: "accepted",
	})
	if err == nil {
		t.Fatal("expected error for accepted quest without date")
	}

	if !strings.Contains(err.Error(), "accepted_on date is required") {
		t.Fatalf("error = %q, want accepted_on date required", err)
	}
}

func TestCreateQuestRejectsEmptyTitle(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	_, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title: "",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}

	if !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("error = %q, want title required", err)
	}
}

func TestCreateQuestRejectsInvalidStatus(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	_, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:  "Bad Status Quest",
		Status: "completed",
	})
	if err == nil {
		t.Fatal("expected error for invalid creation status")
	}

	if !strings.Contains(err.Error(), "must be") {
		t.Fatalf("error = %q, want status validation error", err)
	}
}

func TestListQuests(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	_, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Quest A",
		PromisedBaseReward: 100,
		Status:             "offered",
	})
	if err != nil {
		t.Fatalf("create quest A: %v", err)
	}

	_, err = CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Quest B",
		PromisedBaseReward: 200,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest B: %v", err)
	}

	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if len(quests) != 2 {
		t.Fatalf("quest count = %d, want 2", len(quests))
	}

	// Accepted quests should sort before offered.
	if quests[0].Title != "Quest B" {
		t.Fatalf("first quest = %q, want Quest B (accepted sorts before offered)", quests[0].Title)
	}
}

func TestAcceptQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:  "Accept Me",
		Status: "offered",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := AcceptQuest(context.Background(), databasePath, quest.ID, "2026-03-05"); err != nil {
		t.Fatalf("accept quest: %v", err)
	}

	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusAccepted {
		t.Fatalf("quest status = %q, want accepted", quests[0].Status)
	}

	if quests[0].AcceptedOn != "2026-03-05" {
		t.Fatalf("quest accepted_on = %q, want 2026-03-05", quests[0].AcceptedOn)
	}
}

func TestAcceptQuestRejectsNonOffered(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:      "Already Accepted",
		Status:     "accepted",
		AcceptedOn: "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	err = AcceptQuest(context.Background(), databasePath, quest.ID, "2026-03-05")
	if err == nil {
		t.Fatal("expected error accepting an already accepted quest")
	}

	if !strings.Contains(err.Error(), "cannot be accepted") {
		t.Fatalf("error = %q, want cannot be accepted", err)
	}
}

func TestCompleteQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:      "Complete Me",
		Status:     "accepted",
		AcceptedOn: "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, quest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusCompleted {
		t.Fatalf("quest status = %q, want completed", quests[0].Status)
	}

	if quests[0].CompletedOn != "2026-03-10" {
		t.Fatalf("quest completed_on = %q, want 2026-03-10", quests[0].CompletedOn)
	}
}

func TestCompleteQuestRejectsOffered(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:  "Still Offered",
		Status: "offered",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	err = CompleteQuest(context.Background(), databasePath, quest.ID, "2026-03-10")
	if err == nil {
		t.Fatal("expected error completing an offered quest")
	}

	if !strings.Contains(err.Error(), "cannot be completed") {
		t.Fatalf("error = %q, want cannot be completed", err)
	}
}

func TestCollectQuestFullPayment(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Paid Quest",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, quest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	entry, err := CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: quest.ID,
		Amount:  500,
		Date:    "2026-03-12",
	})
	if err != nil {
		t.Fatalf("collect quest payment: %v", err)
	}

	if entry.EntryNumber < 1 {
		t.Fatalf("entry number = %d, want >= 1", entry.EntryNumber)
	}

	if entry.DebitTotal != 500 || entry.CreditTotal != 500 {
		t.Fatalf("entry totals = %d/%d, want 500/500", entry.DebitTotal, entry.CreditTotal)
	}

	// Verify quest moved to 'paid'.
	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusPaid {
		t.Fatalf("quest status = %q, want paid", quests[0].Status)
	}

	if quests[0].ClosedOn != "2026-03-12" {
		t.Fatalf("quest closed_on = %q, want 2026-03-12", quests[0].ClosedOn)
	}

	// Verify journal entry exists.
	lineCount := strings.TrimSpace(testutil.RunSQLiteQueryForTest(t, databasePath, "SELECT COUNT(*) FROM journal_lines;"))
	if lineCount != "2" {
		t.Fatalf("journal line count = %q, want 2", lineCount)
	}
}

func TestCollectQuestPartialPayment(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Partial Quest",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, quest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	_, err = CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: quest.ID,
		Amount:  200,
		Date:    "2026-03-12",
	})
	if err != nil {
		t.Fatalf("collect partial payment: %v", err)
	}

	// Verify quest moved to 'partially_paid'.
	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusPartiallyPaid {
		t.Fatalf("quest status = %q, want partially_paid", quests[0].Status)
	}

	// Collect the remainder.
	_, err = CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: quest.ID,
		Amount:  300,
		Date:    "2026-03-15",
	})
	if err != nil {
		t.Fatalf("collect remaining payment: %v", err)
	}

	quests, err = ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests after full payment: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusPaid {
		t.Fatalf("quest status = %q, want paid after full collection", quests[0].Status)
	}
}

func TestCollectQuestPaymentRejectsOfferedQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Not Ready",
		PromisedBaseReward: 100,
		Status:             "offered",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	_, err = CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: quest.ID,
		Amount:  100,
		Date:    "2026-03-12",
	})
	if err == nil {
		t.Fatal("expected error collecting from an offered quest")
	}

	if !strings.Contains(err.Error(), "cannot be collected") {
		t.Fatalf("error = %q, want cannot be collected", err)
	}
}

func TestCollectQuestPaymentRejectsAcceptedQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Still In Progress",
		PromisedBaseReward: 100,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	_, err = CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: quest.ID,
		Amount:  100,
		Date:    "2026-03-12",
	})
	if err == nil {
		t.Fatal("expected error collecting from an accepted quest")
	}

	if !strings.Contains(err.Error(), "cannot be collected") {
		t.Fatalf("error = %q, want cannot be collected", err)
	}
}

func TestCollectQuestPaymentRejectsNonexistentQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	_, err := CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: "nonexistent-id",
		Amount:  100,
		Date:    "2026-03-12",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent quest")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does not exist", err)
	}
}

func TestCollectQuestPaymentCountsCustomDescriptionPayments(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	createdQuest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Custom Description Balance",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, createdQuest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	if _, err := CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID:     createdQuest.ID,
		Amount:      200,
		Date:        "2026-03-12",
		Description: "Patron paid a partial amount in marked trade bars",
	}); err != nil {
		t.Fatalf("collect partial payment: %v", err)
	}

	if _, err := CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: createdQuest.ID,
		Amount:  300,
		Date:    "2026-03-15",
	}); err != nil {
		t.Fatalf("collect remaining payment: %v", err)
	}

	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusPaid {
		t.Fatalf("quest status = %q, want paid after both payments", quests[0].Status)
	}
}

func TestAcceptQuestRejectsNonexistentQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	err := AcceptQuest(context.Background(), databasePath, "nonexistent-id", "2026-03-05")
	if err == nil {
		t.Fatal("expected error for nonexistent quest")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does not exist", err)
	}
}

func TestCompleteQuestRejectsNonexistentQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	err := CompleteQuest(context.Background(), databasePath, "nonexistent-id", "2026-03-10")
	if err == nil {
		t.Fatal("expected error for nonexistent quest")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %q, want does not exist", err)
	}
}

func TestWriteOffCompletedQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Slay the Dragon",
		Patron:             "King Aldric",
		PromisedBaseReward: 1000,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, quest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	entry, err := WriteOffQuest(context.Background(), databasePath, WriteOffQuestInput{
		QuestID: quest.ID,
		Date:    "2026-03-20",
	})
	if err != nil {
		t.Fatalf("write off quest: %v", err)
	}

	if entry.EntryNumber < 1 {
		t.Fatalf("entry number = %d, want >= 1", entry.EntryNumber)
	}

	if entry.DebitTotal != 1000 || entry.CreditTotal != 1000 {
		t.Fatalf("entry totals = %d/%d, want 1000/1000", entry.DebitTotal, entry.CreditTotal)
	}

	if entry.Description != "Quest write-off: Slay the Dragon" {
		t.Fatalf("entry description = %q, want default write-off description", entry.Description)
	}

	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusDefaulted {
		t.Fatalf("quest status = %q, want defaulted", quests[0].Status)
	}

	if quests[0].ClosedOn != "2026-03-20" {
		t.Fatalf("quest closed_on = %q, want 2026-03-20", quests[0].ClosedOn)
	}
}

func TestWriteOffPartiallyPaidQuest(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Escort the Caravan",
		Patron:             "Merchant Guild",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, quest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	_, err = CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: quest.ID,
		Amount:  200,
		Date:    "2026-03-12",
	})
	if err != nil {
		t.Fatalf("collect partial payment: %v", err)
	}

	entry, err := WriteOffQuest(context.Background(), databasePath, WriteOffQuestInput{
		QuestID:     quest.ID,
		Date:        "2026-03-25",
		Description: "Merchant Guild defaulted on remainder",
	})
	if err != nil {
		t.Fatalf("write off quest: %v", err)
	}

	if entry.DebitTotal != 300 || entry.CreditTotal != 300 {
		t.Fatalf("entry totals = %d/%d, want 300/300", entry.DebitTotal, entry.CreditTotal)
	}

	if entry.Description != "Merchant Guild defaulted on remainder" {
		t.Fatalf("entry description = %q, want custom description", entry.Description)
	}

	quests, err := ListQuests(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("list quests: %v", err)
	}

	if quests[0].Status != ledger.QuestStatusDefaulted {
		t.Fatalf("quest status = %q, want defaulted", quests[0].Status)
	}
}

func TestWriteOffQuestCountsCustomDescriptionPayments(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	createdQuest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Partial Payment Before Default",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, createdQuest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	if _, err := CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID:     createdQuest.ID,
		Amount:      200,
		Date:        "2026-03-12",
		Description: "Patron sent part of the payment with a courier",
	}); err != nil {
		t.Fatalf("collect partial payment: %v", err)
	}

	entry, err := WriteOffQuest(context.Background(), databasePath, WriteOffQuestInput{
		QuestID: createdQuest.ID,
		Date:    "2026-03-20",
	})
	if err != nil {
		t.Fatalf("write off quest: %v", err)
	}

	if entry.DebitTotal != 300 || entry.CreditTotal != 300 {
		t.Fatalf("entry totals = %d/%d, want 300/300", entry.DebitTotal, entry.CreditTotal)
	}
}

func TestWriteOffQuestNotCompleted(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Offered Only Quest",
		PromisedBaseReward: 100,
		Status:             "offered",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	_, err = WriteOffQuest(context.Background(), databasePath, WriteOffQuestInput{
		QuestID: quest.ID,
		Date:    "2026-03-20",
	})
	if err == nil {
		t.Fatal("expected error writing off an offered quest")
	}

	if !strings.Contains(err.Error(), "cannot be written off") {
		t.Fatalf("error = %q, want cannot be written off", err)
	}
}

func TestWriteOffQuestFullyPaid(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	quest, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Fully Paid Quest",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if err := CompleteQuest(context.Background(), databasePath, quest.ID, "2026-03-10"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	_, err = CollectQuestPayment(context.Background(), databasePath, CollectQuestPaymentInput{
		QuestID: quest.ID,
		Amount:  500,
		Date:    "2026-03-12",
	})
	if err != nil {
		t.Fatalf("collect full payment: %v", err)
	}

	_, err = WriteOffQuest(context.Background(), databasePath, WriteOffQuestInput{
		QuestID: quest.ID,
		Date:    "2026-03-25",
	})
	if err == nil {
		t.Fatal("expected error writing off a fully paid quest")
	}

	if !strings.Contains(err.Error(), "cannot be written off") {
		t.Fatalf("error = %q, want cannot be written off", err)
	}
}

func TestUpdateQuestEditsAcceptedQuestFields(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	record, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Goblin Bounty",
		Patron:             "Mayor Rowan",
		Description:        "Clear the cave",
		PromisedBaseReward: 2500,
		PartialAdvance:     0,
		BonusConditions:    "No civilian harm",
		Notes:              "Bring proof",
		Status:             "accepted",
		AcceptedOn:         "2026-03-08",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	updated, err := UpdateQuest(context.Background(), databasePath, record.ID, &UpdateQuestInput{
		Title:              "Goblin Cave Cleanup",
		Patron:             "Mayor Rowan",
		Description:        "Sweep the tunnels",
		PromisedBaseReward: 3000,
		PartialAdvance:     500,
		BonusConditions:    "No civilian losses",
		Notes:              "Bring witnesses",
		AcceptedOn:         "2026-03-08",
	})
	if err != nil {
		t.Fatalf("update quest: %v", err)
	}

	if updated.Title != "Goblin Cave Cleanup" || updated.PromisedBaseReward != 3000 || updated.PartialAdvance != 500 {
		t.Fatalf("updated quest = %#v", updated)
	}
	if updated.Notes != "Bring witnesses" {
		t.Fatalf("updated notes = %q, want Bring witnesses", updated.Notes)
	}
}

func TestUpdateQuestRejectsRewardChangeAfterCompletion(t *testing.T) {
	databasePath := testutil.InitTestDB(t)

	record, err := CreateQuest(context.Background(), databasePath, &CreateQuestInput{
		Title:              "Late Reward Change",
		PromisedBaseReward: 2500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-08",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}
	if err := CompleteQuest(context.Background(), databasePath, record.ID, "2026-03-09"); err != nil {
		t.Fatalf("complete quest: %v", err)
	}

	_, err = UpdateQuest(context.Background(), databasePath, record.ID, &UpdateQuestInput{
		Title:              "Late Reward Change",
		PromisedBaseReward: 3000,
		PartialAdvance:     0,
		AcceptedOn:         "2026-03-08",
	})
	if err == nil {
		t.Fatal("expected reward edit to fail after completion")
	}
	if !strings.Contains(err.Error(), "promised reward cannot be edited") {
		t.Fatalf("error = %q, want promised reward edit restriction", err)
	}
}
