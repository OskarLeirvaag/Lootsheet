package quest

import "github.com/OskarLeirvaag/Lootsheet/src/ledger"

// QuestRecord represents a quest row from the database.
type QuestRecord struct {
	ID                 string
	Title              string
	Patron             string
	Description        string
	PromisedBaseReward int64
	PartialAdvance     int64
	BonusConditions    string
	Status             ledger.QuestStatus
	Notes              string
	AcceptedOn         string
	CompletedOn        string
	ClosedOn           string
	CreatedAt          string
	UpdatedAt          string
}

// CreateQuestInput holds the parameters for creating a new quest.
type CreateQuestInput struct {
	Title              string
	Patron             string
	Description        string
	PromisedBaseReward int64
	PartialAdvance     int64
	BonusConditions    string
	Status             string // "offered" or "accepted"
	AcceptedOn         string // required if status is "accepted"
}

// CollectQuestPaymentInput holds the parameters for collecting a quest payment.
type CollectQuestPaymentInput struct {
	QuestID     string
	Amount      int64
	Date        string
	Description string
}

// WriteOffQuestInput holds the parameters for writing off a quest receivable.
type WriteOffQuestInput struct {
	QuestID     string
	Date        string
	Description string
}
