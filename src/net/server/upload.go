package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
)

func handleUploadCampaign(ctx context.Context, req *pb.UploadCampaignRequest, svc TUIService) *pb.Response {
	if req == nil {
		return errorResponse("upload_campaign: missing request")
	}
	if len(req.Data) == 0 {
		return errorResponse("upload_campaign: data is empty")
	}
	if req.CampaignId == "" {
		return errorResponse("upload_campaign: campaign_id is required")
	}
	if req.SchemaVersion != config.SchemaVersion {
		return errorResponse(fmt.Sprintf(
			"schema version mismatch: upload=%s, server=%s",
			req.SchemaVersion, config.SchemaVersion,
		))
	}

	// Write upload data to temp file.
	tmpFile, err := os.CreateTemp("", "lootsheet-upload-*.db")
	if err != nil {
		return errorResponse(fmt.Sprintf("create temp file: %v", err))
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(req.Data); err != nil {
		tmpFile.Close()
		return errorResponse(fmt.Sprintf("write temp file: %v", err))
	}
	if err := tmpFile.Close(); err != nil {
		return errorResponse(fmt.Sprintf("close temp file: %v", err))
	}

	// Create backup before overwrite.
	dbPath := svc.DatabasePath()
	if req.Mode == pb.UploadMode_UPLOAD_OVERWRITE {
		backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
		if _, err := ledger.CreateDatabaseBackup(dbPath, backupDir); err != nil {
			return errorResponse(fmt.Sprintf("create backup: %v", err))
		}
	}

	campaignName, err := importCampaign(ctx, dbPath, tmpPath, req.CampaignId, req.Mode)
	if err != nil {
		return errorResponse(err.Error())
	}

	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_UploadCampaign{
			UploadCampaign: &pb.UploadCampaignResponse{
				CampaignId:   req.CampaignId,
				CampaignName: campaignName,
			},
		},
	}
}

func importCampaign(ctx context.Context, dbPath, uploadPath, campaignID string, mode pb.UploadMode) (string, error) {
	db, err := ledger.OpenDB(dbPath)
	if err != nil {
		return "", fmt.Errorf("open server database: %w", err)
	}
	defer db.Close()

	// Disable FK checks for the import (same pattern as MigrateSQLiteDatabase).
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return "", fmt.Errorf("disable foreign keys: %w", err)
	}

	if _, err := db.ExecContext(ctx, "ATTACH DATABASE ? AS upload", uploadPath); err != nil {
		return "", fmt.Errorf("attach upload database: %w", err)
	}
	defer db.ExecContext(ctx, "DETACH DATABASE upload") //nolint:errcheck

	// Verify campaign exists in upload DB.
	var campaignName string
	err = db.QueryRowContext(ctx,
		"SELECT name FROM upload.campaigns WHERE id = ?", campaignID,
	).Scan(&campaignName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("campaign %q not found in uploaded database", campaignID)
	}
	if err != nil {
		return "", fmt.Errorf("query upload campaign: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin import transaction: %w", err)
	}
	defer tx.Rollback()

	switch mode {
	case pb.UploadMode_UPLOAD_OVERWRITE:
		if err := deleteCampaignData(ctx, tx, campaignID); err != nil {
			return "", fmt.Errorf("delete existing campaign data: %w", err)
		}
	case pb.UploadMode_UPLOAD_NEW:
		var exists int
		err := tx.QueryRowContext(ctx,
			"SELECT 1 FROM campaigns WHERE id = ?", campaignID,
		).Scan(&exists)
		if err == nil {
			return "", fmt.Errorf("campaign %q already exists on server", campaignID)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("check campaign exists: %w", err)
		}
	}

	if err := insertCampaignData(ctx, tx, campaignID); err != nil {
		return "", fmt.Errorf("insert campaign data: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit import: %w", err)
	}

	// Verify FK integrity and re-enable foreign keys.
	if err := verifyAndReenableForeignKeys(ctx, db); err != nil {
		return "", err
	}

	return campaignName, nil
}

// deleteCampaignData removes all data for a campaign in reverse FK order.
func deleteCampaignData(ctx context.Context, tx *sql.Tx, campaignID string) error {
	// Reverse FK order: children first, parents last.
	tables := []string{
		"entity_references",
		"notes",
		"codex_entries",
		"asset_template_lines",
		"loot_appraisals",
		"journal_lines",
		"journal_entries",
		"loot_items",
		"quests",
		"accounts",
		"campaigns",
	}

	for _, table := range tables {
		var query string
		switch table {
		case "journal_lines":
			query = "DELETE FROM journal_lines WHERE journal_entry_id IN (SELECT id FROM journal_entries WHERE campaign_id = ?)"
		case "loot_appraisals":
			query = "DELETE FROM loot_appraisals WHERE loot_item_id IN (SELECT id FROM loot_items WHERE campaign_id = ?)"
		case "asset_template_lines":
			query = "DELETE FROM asset_template_lines WHERE loot_item_id IN (SELECT id FROM loot_items WHERE campaign_id = ?)"
		case "campaigns":
			query = "DELETE FROM campaigns WHERE id = ?"
		default:
			query = fmt.Sprintf("DELETE FROM %s WHERE campaign_id = ?", table)
		}
		if _, err := tx.ExecContext(ctx, query, campaignID); err != nil {
			return fmt.Errorf("delete %s: %w", table, err)
		}
	}

	return nil
}

// insertCampaignData copies data from the attached upload DB in FK order.
func insertCampaignData(ctx context.Context, tx *sql.Tx, campaignID string) error {
	// 1. campaigns
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO campaigns SELECT * FROM upload.campaigns WHERE id = ?",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert campaigns: %w", err)
	}

	// 2. accounts
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO accounts SELECT * FROM upload.accounts WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert accounts: %w", err)
	}

	// 3. quests
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO quests SELECT * FROM upload.quests WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert quests: %w", err)
	}

	// 4. loot_items
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO loot_items SELECT * FROM upload.loot_items WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert loot_items: %w", err)
	}

	// 5. journal_entries — NULL reverses_entry_id first, then non-NULL
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries SELECT * FROM upload.journal_entries WHERE campaign_id = ? AND reverses_entry_id IS NULL",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert journal_entries (base): %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries SELECT * FROM upload.journal_entries WHERE campaign_id = ? AND reverses_entry_id IS NOT NULL",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert journal_entries (reversals): %w", err)
	}

	// 6. journal_lines (join to parent for campaign filter)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO journal_lines SELECT jl.* FROM upload.journal_lines jl
		 JOIN upload.journal_entries je ON je.id = jl.journal_entry_id
		 WHERE je.campaign_id = ?`,
		campaignID,
	); err != nil {
		return fmt.Errorf("insert journal_lines: %w", err)
	}

	// 7. loot_appraisals (join to parent for campaign filter)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO loot_appraisals SELECT la.* FROM upload.loot_appraisals la
		 JOIN upload.loot_items li ON li.id = la.loot_item_id
		 WHERE li.campaign_id = ?`,
		campaignID,
	); err != nil {
		return fmt.Errorf("insert loot_appraisals: %w", err)
	}

	// 8. asset_template_lines (join to parent for campaign filter)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO asset_template_lines SELECT atl.* FROM upload.asset_template_lines atl
		 JOIN upload.loot_items li ON li.id = atl.loot_item_id
		 WHERE li.campaign_id = ?`,
		campaignID,
	); err != nil {
		return fmt.Errorf("insert asset_template_lines: %w", err)
	}

	// 9. codex_types (INSERT OR IGNORE — shared across campaigns)
	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO codex_types SELECT ct.* FROM upload.codex_types ct
		 WHERE ct.id IN (SELECT type_id FROM upload.codex_entries WHERE campaign_id = ?)`,
		campaignID,
	); err != nil {
		return fmt.Errorf("insert codex_types: %w", err)
	}

	// 10. codex_entries
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO codex_entries SELECT * FROM upload.codex_entries WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert codex_entries: %w", err)
	}

	// 11. notes
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO notes SELECT * FROM upload.notes WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert notes: %w", err)
	}

	// 12. entity_references
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO entity_references SELECT * FROM upload.entity_references WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return fmt.Errorf("insert entity_references: %w", err)
	}

	return nil
}

func verifyAndReenableForeignKeys(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("foreign key check: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var table, rowid, parent, fkid string
		if scanErr := rows.Scan(&table, &rowid, &parent, &fkid); scanErr == nil {
			return fmt.Errorf("foreign key violation after import: table=%s rowid=%s parent=%s", table, rowid, parent)
		}
		return fmt.Errorf("foreign key violation detected after import")
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("foreign key check iteration: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("re-enable foreign keys: %w", err)
	}

	return nil
}
