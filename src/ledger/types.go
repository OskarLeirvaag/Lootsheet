// Package ledger provides the core domain types, enumerations, database helpers,
// and validation logic for the LootSheet double-entry bookkeeping system.
package ledger

import (
	"io"
	"log/slog"
)

// HandlerContext provides shared dependencies for CLI handler functions
// in domain packages. This avoids circular imports between domain packages
// and the app package.
type HandlerContext struct {
	DatabasePath string
	Stdout       io.Writer
	Logger       *slog.Logger
}

// AccountType represents the classification of a ledger account.
type AccountType string

const (
	// AccountTypeAsset represents asset accounts (debit-normal balance).
	AccountTypeAsset AccountType = "asset"
	// AccountTypeLiability represents liability accounts (credit-normal balance).
	AccountTypeLiability AccountType = "liability"
	// AccountTypeEquity represents equity accounts (credit-normal balance).
	AccountTypeEquity AccountType = "equity"
	// AccountTypeIncome represents income accounts (credit-normal balance).
	AccountTypeIncome AccountType = "income"
	// AccountTypeExpense represents expense accounts (debit-normal balance).
	AccountTypeExpense AccountType = "expense"
)

// Valid returns true if the AccountType is one of the recognized account classifications.
func (t AccountType) Valid() bool {
	switch t {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeEquity, AccountTypeIncome, AccountTypeExpense:
		return true
	default:
		return false
	}
}

// AccountTypes returns all valid account type values.
func AccountTypes() []AccountType {
	return []AccountType{
		AccountTypeAsset,
		AccountTypeLiability,
		AccountTypeEquity,
		AccountTypeIncome,
		AccountTypeExpense,
	}
}

// JournalEntryStatus represents the lifecycle state of a journal entry.
type JournalEntryStatus string

const (
	// JournalEntryStatusDraft indicates the entry has not yet been posted and may be edited.
	JournalEntryStatusDraft JournalEntryStatus = "draft"
	// JournalEntryStatusPosted indicates the entry is final and immutable.
	JournalEntryStatusPosted JournalEntryStatus = "posted"
	// JournalEntryStatusReversed indicates the entry has been reversed by a correcting entry.
	JournalEntryStatusReversed JournalEntryStatus = "reversed"
)

// Valid returns true if the JournalEntryStatus is one of the recognized lifecycle states.
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

// JournalEntryStatuses returns all valid journal entry status values.
func JournalEntryStatuses() []JournalEntryStatus {
	return []JournalEntryStatus{
		JournalEntryStatusDraft,
		JournalEntryStatusPosted,
		JournalEntryStatusReversed,
	}
}

// QuestStatus represents the lifecycle state of a quest.
type QuestStatus string

const (
	// QuestStatusOffered indicates the quest has been offered but not yet accepted.
	QuestStatusOffered QuestStatus = "offered"
	// QuestStatusAccepted indicates the party has accepted the quest.
	QuestStatusAccepted QuestStatus = "accepted"
	// QuestStatusCompleted indicates the quest objectives have been fulfilled.
	QuestStatusCompleted QuestStatus = "completed"
	// QuestStatusCollectible indicates the quest reward is ready to be collected.
	QuestStatusCollectible QuestStatus = "collectible"
	// QuestStatusPartiallyPaid indicates some but not all of the reward has been collected.
	QuestStatusPartiallyPaid QuestStatus = "partially_paid"
	// QuestStatusPaid indicates the full quest reward has been collected.
	QuestStatusPaid QuestStatus = "paid"
	// QuestStatusDefaulted indicates the quest receivable has been written off.
	QuestStatusDefaulted QuestStatus = "defaulted"
	// QuestStatusVoided indicates the quest has been cancelled.
	QuestStatusVoided QuestStatus = "voided"
)

// Valid returns true if the QuestStatus is one of the recognized lifecycle states.
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

// QuestStatuses returns all valid quest status values.
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

// LootStatus represents the lifecycle state of a loot item.
type LootStatus string

const (
	// LootStatusHeld indicates the item is in the party's possession and off-ledger.
	LootStatusHeld LootStatus = "held"
	// LootStatusRecognized indicates the item has been appraised and recorded on-ledger.
	LootStatusRecognized LootStatus = "recognized"
	// LootStatusSold indicates the item has been sold.
	LootStatusSold LootStatus = "sold"
	// LootStatusAssigned indicates the item has been assigned to a party member.
	LootStatusAssigned LootStatus = "assigned"
	// LootStatusConsumed indicates the item has been used or consumed.
	LootStatusConsumed LootStatus = "consumed"
	// LootStatusDiscarded indicates the item has been discarded.
	LootStatusDiscarded LootStatus = "discarded"
)

// Valid returns true if the LootStatus is one of the recognized lifecycle states.
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

// LootStatuses returns all valid loot status values.
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

// InitResult describes the outcome of a database initialization attempt.
type InitResult struct {
	Initialized  bool
	SeededCounts SeededCounts
}

// SeededCounts tracks how many rows were inserted during database seeding.
type SeededCounts struct {
	Accounts int
}

// AccountRecord represents a row from the accounts table.
type AccountRecord struct {
	ID     string
	Code   string
	Name   string
	Type   AccountType
	Active bool
}

// PostedJournalEntry represents a successfully posted journal entry with
// summary information about its lines and totals.
type PostedJournalEntry struct {
	ID          string
	EntryNumber int
	EntryDate   string
	Description string
	LineCount   int
	DebitTotal  int64
	CreditTotal int64
}

// DatabaseStatus describes the current state of a LootSheet database,
// including schema version, migration history, and whether upgrades are available.
type DatabaseStatus struct {
	Exists              bool
	Initialized         bool
	State               DatabaseLifecycleState
	Detail              string
	UserTableCount      int
	SchemaVersion       string
	TargetSchemaVersion string
	AppliedMigrations   []AppliedMigration
	PendingMigrations   []PendingMigration
}

// AppliedMigration records a migration that has been applied to the database.
type AppliedMigration struct {
	Version   string
	Name      string
	AppliedAt string
}

// PendingMigration represents a migration that has not yet been applied.
type PendingMigration struct {
	Version string
	Name    string
}

// MigrationResult describes the outcome of a database migration operation.
type MigrationResult struct {
	Migrated          bool
	MetadataRepaired  bool
	FromSchemaVersion string
	ToSchemaVersion   string
	BackupPath        string
	AppliedMigrations []PendingMigration
}

// DatabaseLifecycleState categorizes the overall state of a LootSheet database
// relative to the application's expected schema.
type DatabaseLifecycleState string

const (
	// DatabaseStateUninitialized indicates no LootSheet schema has been applied.
	DatabaseStateUninitialized DatabaseLifecycleState = "uninitialized"
	// DatabaseStateCurrent indicates the schema matches the application's latest version.
	DatabaseStateCurrent DatabaseLifecycleState = "current"
	// DatabaseStateUpgradeable indicates pending migrations are available.
	DatabaseStateUpgradeable DatabaseLifecycleState = "upgradeable"
	// DatabaseStateForeign indicates the file is a valid database, but not a
	// LootSheet database this build can safely manage.
	DatabaseStateForeign DatabaseLifecycleState = "foreign"
	// DatabaseStateDamaged indicates the file exists but SQLite could not read it
	// safely as a healthy database.
	DatabaseStateDamaged DatabaseLifecycleState = "damaged"
)
