package app

import (
	"context"
	"path/filepath"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/account"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/journal"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/loot"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger/quest"
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
}

// sqliteDataLoader implements TUIDataLoader by delegating each method to the
// existing free functions in the domain packages.
type sqliteDataLoader struct {
	databasePath string
	backupDir    string
	assets       config.InitAssets
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

func (s *sqliteDataLoader) GetDatabaseStatus(ctx context.Context) (ledger.DatabaseStatus, error) {
	return ledger.GetDatabaseStatusWithAssets(ctx, s.databasePath, s.assets)
}

func (s *sqliteDataLoader) ListAccounts(ctx context.Context) ([]ledger.AccountRecord, error) {
	return account.ListAccounts(ctx, s.databasePath)
}

func (s *sqliteDataLoader) GetJournalSummary(ctx context.Context) (journal.Summary, error) {
	return journal.GetSummary(ctx, s.databasePath)
}

func (s *sqliteDataLoader) ListBrowseJournalEntries(ctx context.Context) ([]journal.BrowseEntryRecord, error) {
	return journal.ListBrowseEntries(ctx, s.databasePath)
}

func (s *sqliteDataLoader) GetTrialBalance(ctx context.Context) (report.TrialBalanceReport, error) {
	return report.GetTrialBalance(ctx, s.databasePath)
}

func (s *sqliteDataLoader) GetPromisedQuests(ctx context.Context) ([]report.PromisedQuestRow, error) {
	return report.GetPromisedQuests(ctx, s.databasePath)
}

func (s *sqliteDataLoader) GetQuestReceivables(ctx context.Context) ([]report.QuestReceivableRow, error) {
	return report.GetQuestReceivables(ctx, s.databasePath)
}

func (s *sqliteDataLoader) ListQuests(ctx context.Context) ([]quest.QuestRecord, error) {
	return quest.ListQuests(ctx, s.databasePath)
}

func (s *sqliteDataLoader) GetLootSummary(ctx context.Context, itemType string) ([]report.LootSummaryRow, error) {
	return report.GetLootSummary(ctx, s.databasePath, itemType)
}

func (s *sqliteDataLoader) ListBrowseLootItems(ctx context.Context, itemType string) ([]loot.BrowseItemRecord, error) {
	return loot.ListBrowseItems(ctx, s.databasePath, itemType)
}
