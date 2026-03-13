package report

import (
	"context"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/account"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/testutil"
)

func TestGetTrialBalanceWithPostedEntries(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	// Post an entry: Dr Adventuring Supplies 5000:50, Cr Party Cash 1000:50
	_, err := journal.PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Buy supplies",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 50, Memo: "Rations"},
			{AccountCode: "1000", CreditAmount: 50},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	// Post another entry: Dr Party Cash 1000:200, Cr Quest Income 4000:200
	_, err = journal.PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-02",
		Description: "Quest reward",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 200},
			{AccountCode: "4000", CreditAmount: 200},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	if !report.Balanced {
		t.Fatalf("expected trial balance to be balanced, got debits=%d credits=%d", report.TotalDebits, report.TotalCredits)
	}

	if report.TotalDebits != 250 || report.TotalCredits != 250 {
		t.Fatalf("total debits=%d credits=%d, want 250/250", report.TotalDebits, report.TotalCredits)
	}

	if len(report.Accounts) != 3 {
		t.Fatalf("account count = %d, want 3", len(report.Accounts))
	}

	// Verify per-account totals (ordered by code).
	accountsByCode := map[string]TrialBalanceRow{}
	for _, row := range report.Accounts {
		accountsByCode[row.AccountCode] = row
	}

	// Party Cash (asset): debits=200, credits=50, balance=150
	cash := accountsByCode["1000"]
	if cash.TotalDebits != 200 || cash.TotalCredits != 50 || cash.Balance != 150 {
		t.Fatalf("Party Cash = debits:%d credits:%d balance:%d, want 200/50/150", cash.TotalDebits, cash.TotalCredits, cash.Balance)
	}
	if cash.AccountType != ledger.AccountTypeAsset {
		t.Fatalf("Party Cash type = %q, want asset", cash.AccountType)
	}

	// Quest Income (income): debits=0, credits=200, balance=200
	income := accountsByCode["4000"]
	if income.TotalDebits != 0 || income.TotalCredits != 200 || income.Balance != 200 {
		t.Fatalf("Quest Income = debits:%d credits:%d balance:%d, want 0/200/200", income.TotalDebits, income.TotalCredits, income.Balance)
	}
	if income.AccountType != ledger.AccountTypeIncome {
		t.Fatalf("Quest Income type = %q, want income", income.AccountType)
	}

	// Adventuring Supplies (expense): debits=50, credits=0, balance=50
	supplies := accountsByCode["5000"]
	if supplies.TotalDebits != 50 || supplies.TotalCredits != 0 || supplies.Balance != 50 {
		t.Fatalf("Adventuring Supplies = debits:%d credits:%d balance:%d, want 50/0/50", supplies.TotalDebits, supplies.TotalCredits, supplies.Balance)
	}
	if supplies.AccountType != ledger.AccountTypeExpense {
		t.Fatalf("Adventuring Supplies type = %q, want expense", supplies.AccountType)
	}
}

func TestGetTrialBalanceEmptyLedger(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	report, err := GetTrialBalance(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	if len(report.Accounts) != 0 {
		t.Fatalf("account count = %d, want 0 for empty ledger", len(report.Accounts))
	}

	if report.TotalDebits != 0 || report.TotalCredits != 0 {
		t.Fatalf("totals = %d/%d, want 0/0 for empty ledger", report.TotalDebits, report.TotalCredits)
	}

	if !report.Balanced {
		t.Fatal("expected empty trial balance to be balanced")
	}
}

func TestGetTrialBalanceExcludesAccountsWithNoPostedTransactions(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	// Post a single entry touching only two accounts.
	_, err := journal.PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Simple entry",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 100},
			{AccountCode: "1000", CreditAmount: 100},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	// Only the 2 accounts with transactions should appear, not all 16 seed accounts.
	if len(report.Accounts) != 2 {
		t.Fatalf("account count = %d, want 2", len(report.Accounts))
	}
}

func TestGetTrialBalanceReversedEntriesDoNotDoubleCount(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	// Post an entry: Dr 5000:75, Cr 1000:75
	posted, err := journal.PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Buy supplies",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5000", DebitAmount: 75},
			{AccountCode: "1000", CreditAmount: 75},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	// Also post a second entry that stays posted: Dr 1000:200, Cr 4000:200
	_, err = journal.PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Quest reward",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 200},
			{AccountCode: "4000", CreditAmount: 200},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	// Reverse the first entry.
	_, err = journal.ReverseJournalEntry(ctx, databasePath, campaignID, posted.ID, "2026-03-02", "")
	if err != nil {
		t.Fatalf("reverse journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	if !report.Balanced {
		t.Fatalf("expected balanced trial balance, got debits=%d credits=%d", report.TotalDebits, report.TotalCredits)
	}

	if report.TotalDebits != 275 || report.TotalCredits != 275 {
		t.Fatalf("totals = debits:%d credits:%d, want 275/275", report.TotalDebits, report.TotalCredits)
	}

	accountsByCode := map[string]TrialBalanceRow{}
	for _, row := range report.Accounts {
		accountsByCode[row.AccountCode] = row
	}

	// Party Cash 1000 (asset): debits=200+75=275, credits=0
	cash := accountsByCode["1000"]
	if cash.TotalDebits != 275 || cash.TotalCredits != 0 {
		t.Fatalf("Party Cash = debits:%d credits:%d, want 275/0", cash.TotalDebits, cash.TotalCredits)
	}

	// Quest Income 4000 (income): debits=0, credits=200
	income := accountsByCode["4000"]
	if income.TotalDebits != 0 || income.TotalCredits != 200 {
		t.Fatalf("Quest Income = debits:%d credits:%d, want 0/200", income.TotalDebits, income.TotalCredits)
	}

	// Adventuring Supplies 5000 (expense): debits=0, credits=75 (from reversal only)
	supplies := accountsByCode["5000"]
	if supplies.TotalDebits != 0 || supplies.TotalCredits != 75 {
		t.Fatalf("Adventuring Supplies = debits:%d credits:%d, want 0/75", supplies.TotalDebits, supplies.TotalCredits)
	}
}

func TestGetTrialBalanceNormalBalanceDirections(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	// Create a liability account for testing.
	_, err := account.CreateAccount(ctx, databasePath, campaignID, "2100", "Test Liability", ledger.AccountTypeLiability)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	// Create an equity account for testing.
	_, err = account.CreateAccount(ctx, databasePath, campaignID, "3100", "Test Equity", ledger.AccountTypeEquity)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	// Dr Party Cash 1000:500, Cr Test Liability 2100:300, Cr Test Equity 3100:200
	_, err = journal.PostJournalEntry(ctx, databasePath, campaignID, ledger.JournalPostInput{
		EntryDate:   "2026-03-01",
		Description: "Test normal balances",
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: 500},
			{AccountCode: "2100", CreditAmount: 300},
			{AccountCode: "3100", CreditAmount: 200},
		},
	})
	if err != nil {
		t.Fatalf("post journal entry: %v", err)
	}

	report, err := GetTrialBalance(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}

	accountsByCode := map[string]TrialBalanceRow{}
	for _, row := range report.Accounts {
		accountsByCode[row.AccountCode] = row
	}

	// Asset: normal debit balance = debits - credits
	cash := accountsByCode["1000"]
	if cash.Balance != 500 {
		t.Fatalf("Party Cash balance = %d, want 500", cash.Balance)
	}

	// Liability: normal credit balance = credits - debits
	liability := accountsByCode["2100"]
	if liability.Balance != 300 {
		t.Fatalf("Test Liability balance = %d, want 300", liability.Balance)
	}

	// Equity: normal credit balance = credits - debits
	equity := accountsByCode["3100"]
	if equity.Balance != 200 {
		t.Fatalf("Test Equity balance = %d, want 200", equity.Balance)
	}
}

func TestGetQuestReceivablesCountsCustomDescriptionCollections(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	createdQuest, err := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
		Title:              "Custom Description Payment",
		Patron:             "Guildmaster Rena",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	testutil.CompleteQuest(t, databasePath, createdQuest.ID, "2026-03-05")

	if _, err := quest.CollectQuestPayment(ctx, databasePath, campaignID, quest.CollectQuestPaymentInput{
		QuestID:     createdQuest.ID,
		Amount:      200,
		Date:        "2026-03-06",
		Description: "Guild paid first installment",
	}); err != nil {
		t.Fatalf("collect quest payment: %v", err)
	}

	rows, err := GetQuestReceivables(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get quest receivables: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("receivable count = %d, want 1", len(rows))
	}

	if rows[0].TotalPaid != 200 {
		t.Fatalf("total paid = %d, want 200", rows[0].TotalPaid)
	}

	if rows[0].Outstanding != 300 {
		t.Fatalf("outstanding = %d, want 300", rows[0].Outstanding)
	}
}

func TestGetPromisedQuestsIncludesOfferedAndAcceptedOnly(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	offeredQuest, err := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
		Title:              "Scout the Ruins",
		Patron:             "Archivist Pell",
		PromisedBaseReward: 350,
		PartialAdvance:     25,
		BonusConditions:    "Bonus if the maps survive intact",
		Status:             "offered",
	})
	if err != nil {
		t.Fatalf("create offered quest: %v", err)
	}

	acceptedQuest, err := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
		Title:              "Guard the Caravan",
		Patron:             "Merchant Hall",
		PromisedBaseReward: 275,
		Status:             "accepted",
		AcceptedOn:         "2026-03-02",
	})
	if err != nil {
		t.Fatalf("create accepted quest: %v", err)
	}

	completedQuest, err := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
		Title:              "Already Earned",
		PromisedBaseReward: 999,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create completed quest: %v", err)
	}

	testutil.CompleteQuest(t, databasePath, completedQuest.ID, "2026-03-04")

	rows, err := GetPromisedQuests(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get promised quests: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("promised quest count = %d, want 2", len(rows))
	}

	if rows[0].QuestID != acceptedQuest.ID {
		t.Fatalf("first quest id = %q, want accepted quest %q", rows[0].QuestID, acceptedQuest.ID)
	}

	rowsByID := map[string]PromisedQuestRow{}
	for _, row := range rows {
		rowsByID[row.QuestID] = row
	}

	offeredRow := rowsByID[offeredQuest.ID]
	if offeredRow.Status != ledger.QuestStatusOffered {
		t.Fatalf("offered row status = %q, want offered", offeredRow.Status)
	}
	if offeredRow.PromisedReward != 350 || offeredRow.PartialAdvance != 25 {
		t.Fatalf("offered row reward/advance = %d/%d, want 350/25", offeredRow.PromisedReward, offeredRow.PartialAdvance)
	}
	if offeredRow.BonusConditions != "Bonus if the maps survive intact" {
		t.Fatalf("offered row bonus = %q, want original bonus conditions", offeredRow.BonusConditions)
	}
}

func TestGetWriteOffCandidatesFiltersByAgeAndOutstanding(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	campaignID := testutil.DefaultCampaignID(t, databasePath)
	ctx := context.Background()

	oldPartialQuest, err := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
		Title:              "Old Partial Balance",
		Patron:             "Baron Voss",
		PromisedBaseReward: 500,
		Status:             "accepted",
		AcceptedOn:         "2026-01-01",
	})
	if err != nil {
		t.Fatalf("create old partial quest: %v", err)
	}

	testutil.CompleteQuest(t, databasePath, oldPartialQuest.ID, "2026-01-02")

	if _, err := quest.CollectQuestPayment(ctx, databasePath, campaignID, quest.CollectQuestPaymentInput{
		QuestID:     oldPartialQuest.ID,
		Amount:      200,
		Date:        "2026-01-05",
		Description: "Baron sent a runner with partial payment",
	}); err != nil {
		t.Fatalf("collect old partial quest payment: %v", err)
	}

	recentQuest, err := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
		Title:              "Recent Balance",
		Patron:             "Captain Ilya",
		PromisedBaseReward: 400,
		Status:             "accepted",
		AcceptedOn:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("create recent quest: %v", err)
	}

	testutil.CompleteQuest(t, databasePath, recentQuest.ID, "2026-03-10")

	fullyPaidQuest, err := quest.CreateQuest(ctx, databasePath, campaignID, &quest.CreateQuestInput{
		Title:              "Settled Balance",
		Patron:             "Temple of Dawn",
		PromisedBaseReward: 250,
		Status:             "accepted",
		AcceptedOn:         "2026-01-03",
	})
	if err != nil {
		t.Fatalf("create fully paid quest: %v", err)
	}

	testutil.CompleteQuest(t, databasePath, fullyPaidQuest.ID, "2026-01-04")

	if _, err := quest.CollectQuestPayment(ctx, databasePath, campaignID, quest.CollectQuestPaymentInput{
		QuestID: fullyPaidQuest.ID,
		Amount:  250,
		Date:    "2026-01-10",
	}); err != nil {
		t.Fatalf("collect fully paid quest payment: %v", err)
	}

	rows, err := GetWriteOffCandidates(ctx, databasePath, campaignID, WriteOffCandidateFilter{
		AsOfDate:   "2026-03-15",
		MinAgeDays: 30,
	})
	if err != nil {
		t.Fatalf("get write-off candidates: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("candidate count = %d, want 1", len(rows))
	}

	candidate := rows[0]
	if candidate.QuestID != oldPartialQuest.ID {
		t.Fatalf("candidate id = %q, want %q", candidate.QuestID, oldPartialQuest.ID)
	}
	if candidate.Status != ledger.QuestStatusPartiallyPaid {
		t.Fatalf("candidate status = %q, want partially_paid", candidate.Status)
	}
	if candidate.TotalPaid != 200 {
		t.Fatalf("candidate total paid = %d, want 200", candidate.TotalPaid)
	}
	if candidate.Outstanding != 300 {
		t.Fatalf("candidate outstanding = %d, want 300", candidate.Outstanding)
	}
	if candidate.AgeDays != 72 {
		t.Fatalf("candidate age days = %d, want 72", candidate.AgeDays)
	}
}

func TestSampleCampaignFixtureCoversCoreReports(t *testing.T) {
	databasePath := testutil.InitTestDB(t)
	testutil.ApplyFixtureForTest(t, databasePath, "sample_campaign.sql")
	campaignID := "sample-campaign"
	ctx := context.Background()

	trialBalance, err := GetTrialBalance(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get trial balance: %v", err)
	}
	if !trialBalance.Balanced {
		t.Fatalf("fixture trial balance should be balanced, got debits=%d credits=%d", trialBalance.TotalDebits, trialBalance.TotalCredits)
	}

	trialBalanceByCode := map[string]TrialBalanceRow{}
	for _, row := range trialBalance.Accounts {
		trialBalanceByCode[row.AccountCode] = row
	}

	if row := trialBalanceByCode["5125"]; row.AccountName != "Tavern Reparations" || row.TotalDebits != 350 {
		t.Fatalf("custom account row = %+v, want Tavern Reparations with 350 debit", row)
	}
	if row := trialBalanceByCode["5400"]; row.Balance != 200 {
		t.Fatalf("loss on sale row balance = %d, want 200", row.Balance)
	}

	promised, err := GetPromisedQuests(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get promised quests: %v", err)
	}
	if len(promised) != 2 {
		t.Fatalf("promised quest count = %d, want 2", len(promised))
	}
	if promised[0].Title != "Escort the Archivist" || promised[0].Status != ledger.QuestStatusAccepted {
		t.Fatalf("first promised quest = %+v, want accepted Escort the Archivist", promised[0])
	}
	if promised[1].Title != "Clear the Old Watchtower" || promised[1].Status != ledger.QuestStatusOffered {
		t.Fatalf("second promised quest = %+v, want offered Clear the Old Watchtower", promised[1])
	}

	receivables, err := GetQuestReceivables(ctx, databasePath, campaignID)
	if err != nil {
		t.Fatalf("get quest receivables: %v", err)
	}
	if len(receivables) != 1 {
		t.Fatalf("quest receivable count = %d, want 1", len(receivables))
	}
	if receivables[0].Title != "Moonlit Escort" || receivables[0].Outstanding != 500 || receivables[0].TotalPaid != 700 {
		t.Fatalf("quest receivable row = %+v, want Moonlit Escort with 700 paid and 500 outstanding", receivables[0])
	}

	writeoffCandidates, err := GetWriteOffCandidates(ctx, databasePath, campaignID, WriteOffCandidateFilter{
		AsOfDate:   "2026-03-20",
		MinAgeDays: 30,
	})
	if err != nil {
		t.Fatalf("get write-off candidates: %v", err)
	}
	if len(writeoffCandidates) != 1 {
		t.Fatalf("write-off candidate count = %d, want 1", len(writeoffCandidates))
	}
	if writeoffCandidates[0].Title != "Moonlit Escort" || writeoffCandidates[0].AgeDays != 38 {
		t.Fatalf("write-off candidate row = %+v, want Moonlit Escort aged 38 days", writeoffCandidates[0])
	}

	lootSummary, err := GetLootSummary(ctx, databasePath, campaignID, "loot")
	if err != nil {
		t.Fatalf("get loot summary: %v", err)
	}
	if len(lootSummary) != 2 {
		t.Fatalf("loot summary count = %d, want 2", len(lootSummary))
	}

	lootByName := map[string]LootSummaryRow{}
	for _, row := range lootSummary {
		lootByName[row.Name] = row
	}

	if row := lootByName["Wyvern Tooth Necklace"]; row.Status != ledger.LootStatusHeld || row.LatestAppraisalValue != 650 {
		t.Fatalf("held loot row = %+v, want held Wyvern Tooth Necklace appraised at 650", row)
	}
	if row := lootByName["Emerald Idol"]; row.Status != ledger.LootStatusRecognized || row.LatestAppraisalValue != 800 {
		t.Fatalf("recognized loot row = %+v, want recognized Emerald Idol appraised at 800", row)
	}
	if _, exists := lootByName["Cracked Ruby Crown"]; exists {
		t.Fatal("sold loot item should not appear in loot summary")
	}
}
