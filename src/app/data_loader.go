package app

import (
	"context"
	"database/sql"
	"path/filepath"
	"time"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/account"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/codex"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/notes"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/quest"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/refs"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/report"
)

// TUIDataLoader abstracts the read-only domain queries the TUI needs to build
// its shell data. Accepting this interface instead of a raw database path lets
// callers substitute test doubles or alternative data sources.
type TUIDataLoader interface {
	// DatabaseName returns a human-readable label for the underlying data source.
	DatabaseName() string
	GetDatabaseStatus(ctx context.Context) (ledger.DatabaseStatus, error)
	ListAccounts(ctx context.Context) ([]ledger.AccountRecord, error)
	GetJournalSummary(ctx context.Context) (journal.Summary, error)
	ListBrowseJournalEntries(ctx context.Context) ([]journal.BrowseEntryRecord, error)
	GetTrialBalance(ctx context.Context) (report.TrialBalanceReport, error)
	GetPromisedQuests(ctx context.Context) ([]report.PromisedQuestRow, error)
	GetQuestReceivables(ctx context.Context) ([]report.QuestReceivableRow, error)
	ListQuests(ctx context.Context) ([]quest.QuestRecord, error)
	GetLootSummary(ctx context.Context, itemType string) ([]report.LootSummaryRow, error)
	ListBrowseLootItems(ctx context.Context, itemType string) ([]loot.BrowseItemRecord, error)
	ListCodexEntries(ctx context.Context) ([]codex.CodexEntry, error)
	ListAllCodexReferences(ctx context.Context) (map[string][]refs.EntityReference, error)
	ListCodexTypes(ctx context.Context) ([]codex.CodexType, error)
	ListNotes(ctx context.Context) ([]notes.NoteRecord, error)
	ListAllNotesReferences(ctx context.Context) (map[string][]refs.EntityReference, error)
	ListAllEntityReferences(ctx context.Context) (map[string][]refs.EntityReference, error)
	GetWriteOffCandidates(ctx context.Context) ([]report.WriteOffCandidateRow, error)
	GetAccountLedger(ctx context.Context, accountCode string) (journal.AccountLedgerReport, error)
	SearchCodexEntries(ctx context.Context, query string) ([]codex.CodexEntry, error)
	SearchNotes(ctx context.Context, query string) ([]notes.NoteRecord, error)
	CampaignID() string
	CampaignName() string
	SetCampaign(id, name string)
	ListCampaigns(ctx context.Context) ([]campaign.Record, error)
	SeedAccounts() []config.SeedAccount
}

// sqliteDataLoader implements TUIDataLoader by delegating each method to the
// existing free functions in the domain packages.
type sqliteDataLoader struct {
	databasePath string
	backupDir    string
	assets       config.InitAssets
	campaignID   string
	campaignName string
}

// EnsureReady auto-migrates the database if it is upgradeable. This lets the
// TUI (and any other caller) work immediately after a binary upgrade without
// requiring a manual `lootsheet db migrate`.
func (s *sqliteDataLoader) EnsureReady(ctx context.Context) error {
	status, err := s.GetDatabaseStatus(ctx)
	if err != nil {
		return err
	}

	if status.State != ledger.DatabaseStateUpgradeable {
		return nil
	}

	_, err = ledger.MigrateSQLiteDatabase(ctx, s.databasePath, s.backupDir, s.assets)
	return err
}

func (s *sqliteDataLoader) DatabaseName() string {
	return filepath.Base(s.databasePath)
}

func (s *sqliteDataLoader) CampaignID() string   { return s.campaignID }
func (s *sqliteDataLoader) CampaignName() string { return s.campaignName }
func (s *sqliteDataLoader) SetCampaign(id, name string) {
	s.campaignID = id
	s.campaignName = name
}

func (s *sqliteDataLoader) SeedAccounts() []config.SeedAccount {
	return s.assets.Accounts
}

func (s *sqliteDataLoader) ListCampaigns(ctx context.Context) ([]campaign.Record, error) {
	return campaign.List(ctx, s.databasePath)
}

func (s *sqliteDataLoader) GetDatabaseStatus(ctx context.Context) (ledger.DatabaseStatus, error) {
	return ledger.GetDatabaseStatusWithAssets(ctx, s.databasePath, s.assets)
}

func (s *sqliteDataLoader) ListAccounts(ctx context.Context) ([]ledger.AccountRecord, error) {
	return account.ListAccounts(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) GetJournalSummary(ctx context.Context) (journal.Summary, error) {
	return journal.GetSummary(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) ListBrowseJournalEntries(ctx context.Context) ([]journal.BrowseEntryRecord, error) {
	return journal.ListBrowseEntries(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) GetTrialBalance(ctx context.Context) (report.TrialBalanceReport, error) {
	return report.GetTrialBalance(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) GetPromisedQuests(ctx context.Context) ([]report.PromisedQuestRow, error) {
	return report.GetPromisedQuests(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) GetQuestReceivables(ctx context.Context) ([]report.QuestReceivableRow, error) {
	return report.GetQuestReceivables(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) ListQuests(ctx context.Context) ([]quest.QuestRecord, error) {
	return quest.ListQuests(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) GetLootSummary(ctx context.Context, itemType string) ([]report.LootSummaryRow, error) {
	return report.GetLootSummary(ctx, s.databasePath, s.campaignID, itemType)
}

func (s *sqliteDataLoader) ListBrowseLootItems(ctx context.Context, itemType string) ([]loot.BrowseItemRecord, error) {
	return loot.ListBrowseItems(ctx, s.databasePath, s.campaignID, itemType)
}

func (s *sqliteDataLoader) ListCodexEntries(ctx context.Context) ([]codex.CodexEntry, error) {
	return codex.ListEntries(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) ListAllCodexReferences(ctx context.Context) (map[string][]refs.EntityReference, error) {
	return codex.ListAllReferences(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) ListCodexTypes(ctx context.Context) ([]codex.CodexType, error) {
	return codex.ListTypes(ctx, s.databasePath)
}

func (s *sqliteDataLoader) ListNotes(ctx context.Context) ([]notes.NoteRecord, error) {
	return notes.ListNotes(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) ListAllNotesReferences(ctx context.Context) (map[string][]refs.EntityReference, error) {
	return notes.ListAllReferences(ctx, s.databasePath, s.campaignID)
}

func (s *sqliteDataLoader) ListAllEntityReferences(ctx context.Context) (map[string][]refs.EntityReference, error) {
	return ledger.WithDBResult(ctx, s.databasePath, func(db *sql.DB) (map[string][]refs.EntityReference, error) {
		return refs.ListAllByTarget(ctx, db, s.campaignID)
	})
}

const writeOffMinAgeDays = 30

func (s *sqliteDataLoader) GetWriteOffCandidates(ctx context.Context) ([]report.WriteOffCandidateRow, error) {
	return report.GetWriteOffCandidates(ctx, s.databasePath, s.campaignID, report.WriteOffCandidateFilter{
		AsOfDate:   time.Now().Format("2006-01-02"),
		MinAgeDays: writeOffMinAgeDays,
	})
}

func (s *sqliteDataLoader) GetAccountLedger(ctx context.Context, accountCode string) (journal.AccountLedgerReport, error) {
	return journal.GetAccountLedger(ctx, s.databasePath, s.campaignID, accountCode)
}

func (s *sqliteDataLoader) SearchCodexEntries(ctx context.Context, query string) ([]codex.CodexEntry, error) {
	return codex.SearchEntries(ctx, s.databasePath, s.campaignID, query)
}

func (s *sqliteDataLoader) SearchNotes(ctx context.Context, query string) ([]notes.NoteRecord, error) {
	return notes.SearchNotes(ctx, s.databasePath, s.campaignID, query)
}
