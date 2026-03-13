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
func CreateQuest(ctx context.Context, databasePath string, campaignID string, input *CreateQuestInput) (QuestRecord, error) {
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
	if status != string(ledger.QuestStatusAccepted) {
		acceptedOn = ""
	}

	if input.PromisedBaseReward < 0 {
		return QuestRecord{}, fmt.Errorf("promised_base_reward must be non-negative")
	}

	if input.PartialAdvance < 0 {
		return QuestRecord{}, fmt.Errorf("partial_advance must be non-negative")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (QuestRecord, error) {
		id := uuid.NewString()

		var acceptedOnVal *string
		if acceptedOn != "" {
			acceptedOnVal = &acceptedOn
		}

		notes := strings.TrimSpace(input.Notes)

		if _, err := db.ExecContext(ctx,
			`INSERT INTO quests (id, campaign_id, title, patron, description, promised_base_reward, partial_advance, bonus_conditions, status, notes, accepted_on)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, campaignID, title, strings.TrimSpace(input.Patron), strings.TrimSpace(input.Description),
			input.PromisedBaseReward, input.PartialAdvance, strings.TrimSpace(input.BonusConditions),
			status, notes, acceptedOnVal,
		); err != nil {
			return QuestRecord{}, fmt.Errorf("insert quest: %w", err)
		}

		if err := rebuildReferences(ctx, db, id, campaignID, title, notes); err != nil {
			return QuestRecord{}, err
		}

		return QuestRecord{
			ID:                 id,
			Title:              title,
			Patron:             strings.TrimSpace(input.Patron),
			Description:        strings.TrimSpace(input.Description),
			PromisedBaseReward: input.PromisedBaseReward,
			PartialAdvance:     input.PartialAdvance,
			BonusConditions:    strings.TrimSpace(input.BonusConditions),
			Notes:              notes,
			Status:             ledger.QuestStatus(status),
			AcceptedOn:         acceptedOn,
		}, nil
	})
}

// UpdateQuest edits operational quest fields without mutating posted journal history.
func UpdateQuest(ctx context.Context, databasePath string, campaignID string, questID string, input *UpdateQuestInput) (QuestRecord, error) {
	questID = strings.TrimSpace(questID)
	if questID == "" {
		return QuestRecord{}, fmt.Errorf("quest ID is required")
	}
	if input == nil {
		return QuestRecord{}, fmt.Errorf("quest input is required")
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return QuestRecord{}, fmt.Errorf("quest title is required")
	}
	if input.PromisedBaseReward < 0 {
		return QuestRecord{}, fmt.Errorf("promised_base_reward must be non-negative")
	}
	if input.PartialAdvance < 0 {
		return QuestRecord{}, fmt.Errorf("partial_advance must be non-negative")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (QuestRecord, error) {
		current, err := getQuestByID(ctx, db, questID)
		if err != nil {
			return QuestRecord{}, err
		}

		acceptedOn := strings.TrimSpace(input.AcceptedOn)
		switch current.Status {
		case ledger.QuestStatusOffered:
			if acceptedOn != "" {
				return QuestRecord{}, fmt.Errorf("accepted_on can only be set after a quest is accepted")
			}
		case ledger.QuestStatusAccepted:
			if acceptedOn == "" {
				return QuestRecord{}, fmt.Errorf("accepted_on date is required when quest status is %q", ledger.QuestStatusAccepted)
			}
		default:
			if input.PromisedBaseReward != current.PromisedBaseReward {
				return QuestRecord{}, fmt.Errorf("promised reward cannot be edited after quest status moves beyond accepted")
			}
			if input.PartialAdvance != current.PartialAdvance {
				return QuestRecord{}, fmt.Errorf("partial advance cannot be edited after quest status moves beyond accepted")
			}
			if acceptedOn != current.AcceptedOn {
				return QuestRecord{}, fmt.Errorf("accepted_on cannot be edited after quest status moves beyond accepted")
			}
			acceptedOn = current.AcceptedOn
		}

		notes := strings.TrimSpace(input.Notes)

		if _, err := db.ExecContext(ctx,
			`UPDATE quests
			 SET title = ?, patron = ?, description = ?, promised_base_reward = ?, partial_advance = ?,
			     bonus_conditions = ?, notes = ?, accepted_on = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			title,
			strings.TrimSpace(input.Patron),
			strings.TrimSpace(input.Description),
			input.PromisedBaseReward,
			input.PartialAdvance,
			strings.TrimSpace(input.BonusConditions),
			notes,
			nullString(acceptedOn),
			questID,
		); err != nil {
			return QuestRecord{}, fmt.Errorf("update quest: %w", err)
		}

		if err := rebuildReferences(ctx, db, questID, campaignID, title, notes); err != nil {
			return QuestRecord{}, err
		}

		return getQuestByID(ctx, db, questID)
	})
}

// ListQuests returns all quests ordered by status priority then created_at.
func ListQuests(ctx context.Context, databasePath string, campaignID string) ([]QuestRecord, error) {
	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) ([]QuestRecord, error) {
		rows, err := db.QueryContext(ctx, `
			SELECT id, title, patron, description, promised_base_reward, partial_advance,
			       bonus_conditions, status, notes,
			       COALESCE(accepted_on, ''), COALESCE(completed_on, ''), COALESCE(closed_on, ''),
			       created_at, updated_at
			FROM quests
			WHERE campaign_id = ?
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
		`, campaignID)
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
	})
}

// CollectQuestPayment creates a journal entry for quest payment collection.
func CollectQuestPayment(ctx context.Context, databasePath string, campaignID string, input CollectQuestPaymentInput) (ledger.PostedJournalEntry, error) {
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

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
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

		totalPaid, err := queryQuestTotalPaid(ctx, db, quest.Title)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
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

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin quest payment transaction: %w", err)
		}
		defer tx.Rollback()

		posted, err := ledger.PostJournalWithinTx(ctx, db, tx, campaignID, journalInput)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
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

		return posted, nil
	})
}

// WriteOffQuest writes off an outstanding quest receivable as a failed patron loss.
func WriteOffQuest(ctx context.Context, databasePath string, campaignID string, input WriteOffQuestInput) (ledger.PostedJournalEntry, error) {
	questID := strings.TrimSpace(input.QuestID)
	if questID == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("quest ID is required")
	}

	date := strings.TrimSpace(input.Date)
	if date == "" {
		return ledger.PostedJournalEntry{}, fmt.Errorf("write-off date is required")
	}

	return ledger.WithDBResult(ctx, databasePath, func(db *sql.DB) (ledger.PostedJournalEntry, error) {
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

		totalPaid, err := queryQuestTotalPaid(ctx, db, quest.Title)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
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

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ledger.PostedJournalEntry{}, fmt.Errorf("begin quest write-off transaction: %w", err)
		}
		defer tx.Rollback()

		posted, err := ledger.PostJournalWithinTx(ctx, db, tx, campaignID, journalInput)
		if err != nil {
			return ledger.PostedJournalEntry{}, err
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

		return posted, nil
	})
}

func getQuestByID(ctx context.Context, db *sql.DB, questID string) (QuestRecord, error) {
	var record QuestRecord
	var status string

	if err := db.QueryRowContext(ctx, `
		SELECT id, title, patron, description, promised_base_reward, partial_advance,
		       bonus_conditions, status, notes,
		       COALESCE(accepted_on, ''), COALESCE(completed_on, ''), COALESCE(closed_on, ''),
		       created_at, updated_at
		FROM quests
		WHERE id = ?
	`, questID).Scan(
		&record.ID, &record.Title, &record.Patron, &record.Description,
		&record.PromisedBaseReward, &record.PartialAdvance,
		&record.BonusConditions, &status, &record.Notes,
		&record.AcceptedOn, &record.CompletedOn, &record.ClosedOn,
		&record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return QuestRecord{}, fmt.Errorf("quest %q does not exist", questID)
		}
		return QuestRecord{}, fmt.Errorf("query quest: %w", err)
	}

	record.Status = ledger.QuestStatus(status)
	if !record.Status.Valid() {
		return QuestRecord{}, fmt.Errorf("quest %s has invalid status %q", questID, status)
	}

	return record, nil
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func queryQuestTotalPaid(ctx context.Context, db *sql.DB, questTitle string) (int64, error) {
	var totalPaid int64
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(jl.debit_amount), 0)
		 FROM journal_lines jl
		 JOIN journal_entries je ON je.id = jl.journal_entry_id
		 JOIN accounts a ON a.id = jl.account_id
		 WHERE je.status = 'posted'
		   AND a.code = '1000'
		   AND jl.memo = ?`,
		fmt.Sprintf("Quest payment: %s", questTitle),
	).Scan(&totalPaid); err != nil {
		return 0, fmt.Errorf("query total paid: %w", err)
	}

	return totalPaid, nil
}
