package quest

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
)

// CreateQuest inserts a new quest into the database.
// Status must be "offered" or "accepted". If "accepted", AcceptedOn is required.
func CreateQuest(ctx context.Context, databasePath string, input *CreateQuestInput) (QuestRecord, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return QuestRecord{}, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return QuestRecord{}, fmt.Errorf("quest title is required")
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "offered"
	}

	if status != string(ledger.QuestStatusOffered) && status != string(ledger.QuestStatusAccepted) {
		return QuestRecord{}, fmt.Errorf("quest creation status must be %q or %q, got %q", ledger.QuestStatusOffered, ledger.QuestStatusAccepted, status)
	}

	acceptedOn := strings.TrimSpace(input.AcceptedOn)
	if status == string(ledger.QuestStatusAccepted) && acceptedOn == "" {
		return QuestRecord{}, fmt.Errorf("accepted_on date is required when quest status is %q", ledger.QuestStatusAccepted)
	}

	if input.PromisedBaseReward < 0 {
		return QuestRecord{}, fmt.Errorf("promised_base_reward must be non-negative")
	}

	if input.PartialAdvance < 0 {
		return QuestRecord{}, fmt.Errorf("partial_advance must be non-negative")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return QuestRecord{}, err
	}
	defer db.Close()

	id := uuid.NewString()

	var acceptedOnVal *string
	if acceptedOn != "" {
		acceptedOnVal = &acceptedOn
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO quests (id, title, patron, description, promised_base_reward, partial_advance, bonus_conditions, status, notes, accepted_on)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, '', ?)`,
		id, title, strings.TrimSpace(input.Patron), strings.TrimSpace(input.Description),
		input.PromisedBaseReward, input.PartialAdvance, strings.TrimSpace(input.BonusConditions),
		status, acceptedOnVal,
	); err != nil {
		return QuestRecord{}, fmt.Errorf("insert quest: %w", err)
	}

	return QuestRecord{
		ID:                 id,
		Title:              title,
		Patron:             strings.TrimSpace(input.Patron),
		Description:        strings.TrimSpace(input.Description),
		PromisedBaseReward: input.PromisedBaseReward,
		PartialAdvance:     input.PartialAdvance,
		BonusConditions:    strings.TrimSpace(input.BonusConditions),
		Status:             ledger.QuestStatus(status),
		AcceptedOn:         acceptedOn,
	}, nil
}

// ListQuests returns all quests ordered by status priority then created_at.
func ListQuests(ctx context.Context, databasePath string) ([]QuestRecord, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return nil, err
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT id, title, patron, description, promised_base_reward, partial_advance,
		       bonus_conditions, status, notes,
		       COALESCE(accepted_on, ''), COALESCE(completed_on, ''), COALESCE(closed_on, ''),
		       created_at, updated_at
		FROM quests
		ORDER BY
		  CASE status
		    WHEN 'accepted' THEN 1
		    WHEN 'completed' THEN 2
		    WHEN 'collectible' THEN 3
		    WHEN 'partially_paid' THEN 4
		    WHEN 'offered' THEN 5
		    WHEN 'paid' THEN 6
		    WHEN 'defaulted' THEN 7
		    WHEN 'voided' THEN 8
		  END,
		  created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("query quests: %w", err)
	}
	defer rows.Close()

	quests := []QuestRecord{}
	for rows.Next() {
		var q QuestRecord
		var status string

		if err := rows.Scan(
			&q.ID, &q.Title, &q.Patron, &q.Description,
			&q.PromisedBaseReward, &q.PartialAdvance,
			&q.BonusConditions, &status, &q.Notes,
			&q.AcceptedOn, &q.CompletedOn, &q.ClosedOn,
			&q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan quest row: %w", err)
		}

		q.Status = ledger.QuestStatus(status)
		if !q.Status.Valid() {
			return nil, fmt.Errorf("scan quest row: invalid quest status %q", status)
		}

		quests = append(quests, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quest rows: %w", err)
	}

	return quests, nil
}

// AcceptQuest transitions a quest from 'offered' to 'accepted'.
func AcceptQuest(ctx context.Context, databasePath string, questID string, acceptedDate string) error {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	questID = strings.TrimSpace(questID)
	if questID == "" {
		return fmt.Errorf("quest ID is required")
	}

	acceptedDate = strings.TrimSpace(acceptedDate)
	if acceptedDate == "" {
		return fmt.Errorf("accepted date is required")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	currentStatus, err := getQuestStatus(ctx, db, questID)
	if err != nil {
		return err
	}

	if currentStatus != ledger.QuestStatusOffered {
		return fmt.Errorf("quest %q cannot be accepted: current status is %q, expected %q", questID, currentStatus, ledger.QuestStatusOffered)
	}

	if _, err := db.ExecContext(ctx,
		"UPDATE quests SET status = ?, accepted_on = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(ledger.QuestStatusAccepted), acceptedDate, questID,
	); err != nil {
		return fmt.Errorf("accept quest: %w", err)
	}

	return nil
}

// CompleteQuest transitions a quest from 'accepted' to 'completed'.
func CompleteQuest(ctx context.Context, databasePath string, questID string, completedDate string) error {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return err
	}

	questID = strings.TrimSpace(questID)
	if questID == "" {
		return fmt.Errorf("quest ID is required")
	}

	completedDate = strings.TrimSpace(completedDate)
	if completedDate == "" {
		return fmt.Errorf("completed date is required")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	currentStatus, err := getQuestStatus(ctx, db, questID)
	if err != nil {
		return err
	}

	if currentStatus != ledger.QuestStatusAccepted {
		return fmt.Errorf("quest %q cannot be completed: current status is %q, expected %q", questID, currentStatus, ledger.QuestStatusAccepted)
	}

	if _, err := db.ExecContext(ctx,
		"UPDATE quests SET status = ?, completed_on = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(ledger.QuestStatusCompleted), completedDate, questID,
	); err != nil {
		return fmt.Errorf("complete quest: %w", err)
	}

	return nil
}

// CollectQuestPayment creates a journal entry for quest payment collection.
func CollectQuestPayment(ctx context.Context, databasePath string, input CollectQuestPaymentInput) (ledger.PostedJournalEntry, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	questID := strings.TrimSpace(input.QuestID)
	if questID == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("quest ID is required")
	}

	if input.Amount <= 0 {
		return ledger.PostedJournalEntry{}, fmt.Errorf("payment amount must be positive")
	}

	date := strings.TrimSpace(input.Date)
	if date == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("payment date is required")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}
	defer db.Close()

	var quest QuestRecord
	var statusStr string
	var acceptedOn, completedOn, closedOn sql.NullString

	if err := db.QueryRowContext(ctx,
		"SELECT id, title, promised_base_reward, partial_advance, status, accepted_on, completed_on, closed_on FROM quests WHERE id = ?",
		questID,
	).Scan(&quest.ID, &quest.Title, &quest.PromisedBaseReward, &quest.PartialAdvance, &statusStr, &acceptedOn, &completedOn, &closedOn); err != nil {
		if err == sql.ErrNoRows {
			return ledger.PostedJournalEntry{}, fmt.Errorf("quest %q does not exist", questID)
		}
		return ledger.PostedJournalEntry{}, fmt.Errorf("query quest: %w", err)
	}

	quest.Status = ledger.QuestStatus(statusStr)

	collectibleStatuses := map[ledger.QuestStatus]bool{
		ledger.QuestStatusCompleted:     true,
		ledger.QuestStatusCollectible:   true,
		ledger.QuestStatusPartiallyPaid: true,
	}

	if !collectibleStatuses[quest.Status] {
		return ledger.PostedJournalEntry{}, fmt.Errorf("quest %q cannot be collected: current status is %q, expected one of completed, collectible, partially_paid", questID, quest.Status)
	}

	var totalPaid int64
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(jl.debit_amount), 0) FROM journal_lines jl
		 JOIN journal_entries je ON je.id = jl.journal_entry_id
		 WHERE je.description LIKE ? AND je.status = 'posted'
		 AND jl.account_id = (SELECT id FROM accounts WHERE code = '1000')`,
		fmt.Sprintf("Quest payment: %s%%", quest.Title),
	).Scan(&totalPaid); err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("query total paid: %w", err)
	}

	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = fmt.Sprintf("Quest payment: %s", quest.Title)
	}

	journalInput := ledger.JournalPostInput{
		EntryDate:   date,
		Description: description,
		Lines: []ledger.JournalLineInput{
			{AccountCode: "1000", DebitAmount: input.Amount, Memo: fmt.Sprintf("Quest payment: %s", quest.Title)},
			{AccountCode: "4000", CreditAmount: input.Amount, Memo: fmt.Sprintf("Quest payment: %s", quest.Title)},
		},
	}

	validated, err := ledger.ValidateJournalPostInput(journalInput)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	accountIDsByCode, err := ledger.ResolveActiveAccountIDsByCode(ctx, db, validated.Lines)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	entryNumber, err := ledger.NextJournalEntryNumber(ctx, db)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	entryID := uuid.NewString()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("begin quest payment transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries (id, entry_number, status, entry_date, description, posted_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		entryID, entryNumber, "posted", validated.EntryDate, validated.Description,
	); err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("insert quest payment journal entry: %w", err)
	}

	for index, line := range validated.Lines {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
			uuid.NewString(), entryID, index+1, accountIDsByCode[line.AccountCode], line.Memo, line.DebitAmount, line.CreditAmount,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("insert quest payment journal line %d: %w", index+1, err)
		}
	}

	newTotalPaid := totalPaid + input.Amount
	var newStatus ledger.QuestStatus
	if newTotalPaid >= quest.PromisedBaseReward && quest.PromisedBaseReward > 0 {
		newStatus = ledger.QuestStatusPaid
	} else {
		newStatus = ledger.QuestStatusPartiallyPaid
	}

	if newStatus == ledger.QuestStatusPaid {
		if _, err := tx.ExecContext(ctx,
			"UPDATE quests SET status = ?, closed_on = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			string(newStatus), date, questID,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("update quest status to paid: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx,
			"UPDATE quests SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			string(newStatus), questID,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("update quest status to partially_paid: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("commit quest payment transaction: %w", err)
	}

	return ledger.PostedJournalEntry{
		ID:          entryID,
		EntryNumber: entryNumber,
		EntryDate:   validated.EntryDate,
		Description: validated.Description,
		LineCount:   len(validated.Lines),
		DebitTotal:  validated.Totals.DebitAmount,
		CreditTotal: validated.Totals.CreditAmount,
	}, nil
}

// WriteOffQuest writes off an outstanding quest receivable as a failed patron loss.
func WriteOffQuest(ctx context.Context, databasePath string, input WriteOffQuestInput) (ledger.PostedJournalEntry, error) {
	if err := ledger.EnsureInitializedDatabase(ctx, databasePath); err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	questID := strings.TrimSpace(input.QuestID)
	if questID == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("quest ID is required")
	}

	date := strings.TrimSpace(input.Date)
	if date == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("write-off date is required")
	}

	db, err := ledger.OpenDB(databasePath)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}
	defer db.Close()

	var quest QuestRecord
	var statusStr string
	var acceptedOn, completedOn, closedOn sql.NullString

	if err := db.QueryRowContext(ctx,
		"SELECT id, title, promised_base_reward, partial_advance, status, accepted_on, completed_on, closed_on FROM quests WHERE id = ?",
		questID,
	).Scan(&quest.ID, &quest.Title, &quest.PromisedBaseReward, &quest.PartialAdvance, &statusStr, &acceptedOn, &completedOn, &closedOn); err != nil {
		if err == sql.ErrNoRows {
			return ledger.PostedJournalEntry{}, fmt.Errorf("quest %q does not exist", questID)
		}
		return ledger.PostedJournalEntry{}, fmt.Errorf("query quest: %w", err)
	}

	quest.Status = ledger.QuestStatus(statusStr)

	writeOffStatuses := map[ledger.QuestStatus]bool{
		ledger.QuestStatusCompleted:     true,
		ledger.QuestStatusCollectible:   true,
		ledger.QuestStatusPartiallyPaid: true,
	}

	if !writeOffStatuses[quest.Status] {
		return ledger.PostedJournalEntry{}, fmt.Errorf("quest %q cannot be written off: current status is %q, expected one of completed, collectible, partially_paid", questID, quest.Status)
	}

	var totalPaid int64
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(jl.debit_amount), 0) FROM journal_lines jl
		 JOIN journal_entries je ON je.id = jl.journal_entry_id
		 WHERE je.description LIKE ? AND je.status = 'posted'
		 AND jl.account_id = (SELECT id FROM accounts WHERE code = '1000')`,
		fmt.Sprintf("Quest payment: %s%%", quest.Title),
	).Scan(&totalPaid); err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("query total paid: %w", err)
	}

	outstanding := quest.PromisedBaseReward - totalPaid
	if outstanding <= 0 {
		return ledger.PostedJournalEntry{}, fmt.Errorf("quest has no outstanding balance to write off")
	}

	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = fmt.Sprintf("Quest write-off: %s", quest.Title)
	}

	journalInput := ledger.JournalPostInput{
		EntryDate:   date,
		Description: description,
		Lines: []ledger.JournalLineInput{
			{AccountCode: "5500", DebitAmount: outstanding, Memo: fmt.Sprintf("Quest write-off: %s", quest.Title)},
			{AccountCode: "1100", CreditAmount: outstanding, Memo: fmt.Sprintf("Quest write-off: %s", quest.Title)},
		},
	}

	validated, err := ledger.ValidateJournalPostInput(journalInput)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	accountIDsByCode, err := ledger.ResolveActiveAccountIDsByCode(ctx, db, validated.Lines)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	entryNumber, err := ledger.NextJournalEntryNumber(ctx, db)
	if err != nil {
		return ledger.PostedJournalEntry{}, err
	}

	entryID := uuid.NewString()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("begin quest write-off transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries (id, entry_number, status, entry_date, description, posted_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		entryID, entryNumber, "posted", validated.EntryDate, validated.Description,
	); err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("insert quest write-off journal entry: %w", err)
	}

	for index, line := range validated.Lines {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
			uuid.NewString(), entryID, index+1, accountIDsByCode[line.AccountCode], line.Memo, line.DebitAmount, line.CreditAmount,
		); err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("insert quest write-off journal line %d: %w", index+1, err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		"UPDATE quests SET status = ?, closed_on = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(ledger.QuestStatusDefaulted), date, questID,
	); err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("update quest status to defaulted: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ledger.PostedJournalEntry{}, fmt.Errorf("commit quest write-off transaction: %w", err)
	}

	return ledger.PostedJournalEntry{
		ID:          entryID,
		EntryNumber: entryNumber,
		EntryDate:   validated.EntryDate,
		Description: validated.Description,
		LineCount:   len(validated.Lines),
		DebitTotal:  validated.Totals.DebitAmount,
		CreditTotal: validated.Totals.CreditAmount,
	}, nil
}

func getQuestStatus(ctx context.Context, db *sql.DB, questID string) (ledger.QuestStatus, error) {
	var status string
	if err := db.QueryRowContext(ctx,
		"SELECT status FROM quests WHERE id = ?", questID,
	).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("quest %q does not exist", questID)
		}
		return "", fmt.Errorf("query quest status: %w", err)
	}

	s := ledger.QuestStatus(status)
	if !s.Valid() {
		return "", fmt.Errorf("quest %s has invalid status %q", questID, status)
	}

	return s, nil
}
