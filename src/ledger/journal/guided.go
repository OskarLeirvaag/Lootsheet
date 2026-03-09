package journal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// ExpenseEntryInput is the guided two-line expense posting input.
type ExpenseEntryInput struct {
	Date               string
	Description        string
	ExpenseAccountCode string
	FundingAccountCode string
	Amount             int64
	Memo               string
}

// IncomeEntryInput is the guided two-line income posting input.
type IncomeEntryInput struct {
	Date               string
	Description        string
	IncomeAccountCode  string
	DepositAccountCode string
	Amount             int64
	Memo               string
}

type guidedAccountRecord struct {
	Code   string
	Name   string
	Type   ledger.AccountType
	Active bool
}

// PostExpenseEntry validates a guided expense input and posts the resulting
// two-line journal entry.
func PostExpenseEntry(ctx context.Context, databasePath string, input *ExpenseEntryInput) (ledger.PostedJournalEntry, error) {
	if input == nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("expense entry input is required")
	}
	sanitizeExpenseEntryInput(input)
	if err := validateGuidedAmount(input.Amount); err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
		accounts, err := loadGuidedAccountsByCode(ctx, db, input.ExpenseAccountCode, input.FundingAccountCode)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		expense := accounts[input.ExpenseAccountCode]
		if expense.Type != ledger.AccountTypeExpense {
			return ledger.PostedJournalEntry{}, fmt.Errorf("account code %q must be an active expense account", input.ExpenseAccountCode)
		}

		funding := accounts[input.FundingAccountCode]
		if funding.Type != ledger.AccountTypeAsset && funding.Type != ledger.AccountTypeLiability {
			return ledger.PostedJournalEntry{}, fmt.Errorf("account code %q must be an active asset or liability account", input.FundingAccountCode)
		}

		return PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
			EntryDate:   input.Date,
			Description: input.Description,
			Lines: []ledger.JournalLineInput{
				{AccountCode: input.ExpenseAccountCode, DebitAmount: input.Amount, Memo: input.Memo},
				{AccountCode: input.FundingAccountCode, CreditAmount: input.Amount},
			},
		})
	})
}

// PostIncomeEntry validates a guided income input and posts the resulting
// two-line journal entry.
func PostIncomeEntry(ctx context.Context, databasePath string, input *IncomeEntryInput) (ledger.PostedJournalEntry, error) {
	if input == nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("income entry input is required")
	}
	sanitizeIncomeEntryInput(input)
	if err := validateGuidedAmount(input.Amount); err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
		accounts, err := loadGuidedAccountsByCode(ctx, db, input.IncomeAccountCode, input.DepositAccountCode)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
		}

		income := accounts[input.IncomeAccountCode]
		if income.Type != ledger.AccountTypeIncome {
			return ledger.PostedJournalEntry{}, fmt.Errorf("account code %q must be an active income account", input.IncomeAccountCode)
		}

		deposit := accounts[input.DepositAccountCode]
		if deposit.Type != ledger.AccountTypeAsset {
			return ledger.PostedJournalEntry{}, fmt.Errorf("account code %q must be an active asset account", input.DepositAccountCode)
		}

		return PostJournalEntry(ctx, databasePath, ledger.JournalPostInput{
			EntryDate:   input.Date,
			Description: input.Description,
			Lines: []ledger.JournalLineInput{
				{AccountCode: input.DepositAccountCode, DebitAmount: input.Amount},
				{AccountCode: input.IncomeAccountCode, CreditAmount: input.Amount, Memo: input.Memo},
			},
		})
	})
}

func sanitizeExpenseEntryInput(input *ExpenseEntryInput) {
	input.Date = strings.TrimSpace(input.Date)
	input.Description = strings.TrimSpace(input.Description)
	input.ExpenseAccountCode = strings.TrimSpace(input.ExpenseAccountCode)
	input.FundingAccountCode = strings.TrimSpace(input.FundingAccountCode)
	input.Memo = strings.TrimSpace(input.Memo)
}

func sanitizeIncomeEntryInput(input *IncomeEntryInput) {
	input.Date = strings.TrimSpace(input.Date)
	input.Description = strings.TrimSpace(input.Description)
	input.IncomeAccountCode = strings.TrimSpace(input.IncomeAccountCode)
	input.DepositAccountCode = strings.TrimSpace(input.DepositAccountCode)
	input.Memo = strings.TrimSpace(input.Memo)
}

func validateGuidedAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("journal entry amount must be positive")
	}
	return nil
}

func loadGuidedAccountsByCode(ctx context.Context, db *sql.DB, codes ...string) (map[string]guidedAccountRecord, error) {
	trimmed := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			return nil, fmt.Errorf("account code is required")
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		trimmed = append(trimmed, code)
	}

	placeholders := make([]string, len(trimmed))
	args := make([]any, len(trimmed))
	for index := range trimmed {
		placeholders[index] = "?"
		args[index] = trimmed[index]
	}

	query := "SELECT code, name, type, active FROM accounts WHERE code IN (" + strings.Join(placeholders, ", ") + ")" //nolint:gosec // placeholders are fixed "?" literals
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query guided entry accounts: %w", err)
	}
	defer rows.Close()

	records := make(map[string]guidedAccountRecord, len(trimmed))
	for rows.Next() {
		var record guidedAccountRecord
		var accountType string
		var active int
		if err := rows.Scan(&record.Code, &record.Name, &accountType, &active); err != nil {
			return nil, fmt.Errorf("scan guided entry account: %w", err)
		}

		record.Type = ledger.AccountType(accountType)
		if !record.Type.Valid() {
			return nil, fmt.Errorf("account code %q has invalid type %q", record.Code, accountType)
		}
		record.Active = active == 1
		records[record.Code] = record
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate guided entry accounts: %w", err)
	}

	for _, code := range trimmed {
		record, ok := records[code]
		if !ok {
			return nil, fmt.Errorf("account code %q does not exist", code)
		}
		if !record.Active {
			return nil, fmt.Errorf("account code %q is inactive", code)
		}
	}

	return records, nil
}
