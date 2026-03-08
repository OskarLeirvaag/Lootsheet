package service

type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeEquity    AccountType = "equity"
	AccountTypeIncome    AccountType = "income"
	AccountTypeExpense   AccountType = "expense"
)

func (t AccountType) Valid() bool {
	switch t {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeEquity, AccountTypeIncome, AccountTypeExpense:
		return true
	default:
		return false
	}
}

func AccountTypes() []AccountType {
	return []AccountType{
		AccountTypeAsset,
		AccountTypeLiability,
		AccountTypeEquity,
		AccountTypeIncome,
		AccountTypeExpense,
	}
}

type JournalEntryStatus string

const (
	JournalEntryStatusDraft    JournalEntryStatus = "draft"
	JournalEntryStatusPosted   JournalEntryStatus = "posted"
	JournalEntryStatusReversed JournalEntryStatus = "reversed"
)

func (s JournalEntryStatus) Valid() bool {
	switch s {
	case JournalEntryStatusDraft, JournalEntryStatusPosted, JournalEntryStatusReversed:
		return true
	default:
		return false
	}
}

// Immutable returns true if the entry status forbids edits and deletes.
// Posted and reversed entries are immutable; corrections must use reversal or adjustment.
func (s JournalEntryStatus) Immutable() bool {
	switch s {
	case JournalEntryStatusPosted, JournalEntryStatusReversed:
		return true
	default:
		return false
	}
}

func JournalEntryStatuses() []JournalEntryStatus {
	return []JournalEntryStatus{
		JournalEntryStatusDraft,
		JournalEntryStatusPosted,
		JournalEntryStatusReversed,
	}
}

type QuestStatus string

const (
	QuestStatusOffered       QuestStatus = "offered"
	QuestStatusAccepted      QuestStatus = "accepted"
	QuestStatusCompleted     QuestStatus = "completed"
	QuestStatusCollectible   QuestStatus = "collectible"
	QuestStatusPartiallyPaid QuestStatus = "partially_paid"
	QuestStatusPaid          QuestStatus = "paid"
	QuestStatusDefaulted     QuestStatus = "defaulted"
	QuestStatusVoided        QuestStatus = "voided"
)

func (s QuestStatus) Valid() bool {
	switch s {
	case QuestStatusOffered,
		QuestStatusAccepted,
		QuestStatusCompleted,
		QuestStatusCollectible,
		QuestStatusPartiallyPaid,
		QuestStatusPaid,
		QuestStatusDefaulted,
		QuestStatusVoided:
		return true
	default:
		return false
	}
}

func QuestStatuses() []QuestStatus {
	return []QuestStatus{
		QuestStatusOffered,
		QuestStatusAccepted,
		QuestStatusCompleted,
		QuestStatusCollectible,
		QuestStatusPartiallyPaid,
		QuestStatusPaid,
		QuestStatusDefaulted,
		QuestStatusVoided,
	}
}

type LootStatus string

const (
	LootStatusHeld       LootStatus = "held"
	LootStatusRecognized LootStatus = "recognized"
	LootStatusSold       LootStatus = "sold"
	LootStatusAssigned   LootStatus = "assigned"
	LootStatusConsumed   LootStatus = "consumed"
	LootStatusDiscarded  LootStatus = "discarded"
)

func (s LootStatus) Valid() bool {
	switch s {
	case LootStatusHeld,
		LootStatusRecognized,
		LootStatusSold,
		LootStatusAssigned,
		LootStatusConsumed,
		LootStatusDiscarded:
		return true
	default:
		return false
	}
}

func LootStatuses() []LootStatus {
	return []LootStatus{
		LootStatusHeld,
		LootStatusRecognized,
		LootStatusSold,
		LootStatusAssigned,
		LootStatusConsumed,
		LootStatusDiscarded,
	}
}
