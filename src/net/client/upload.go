package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/OskarLeirvaag/Lootsheet/src/config"
	"github.com/OskarLeirvaag/Lootsheet/src/ledger"
	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
)

// UploadCampaign extracts a single campaign from localDBPath into a minimal
// SQLite file and sends it to the server via the UPLOAD_CAMPAIGN RPC.
func UploadCampaign(ctx context.Context, c *Client, localDBPath, campaignID string, mode pb.UploadMode) (*pb.UploadCampaignResponse, error) {
	data, err := extractCampaignDB(ctx, localDBPath, campaignID)
	if err != nil {
		return nil, fmt.Errorf("extract campaign: %w", err)
	}

	req := &pb.Request{
		Method: pb.Method_UPLOAD_CAMPAIGN,
		Payload: &pb.Request_UploadCampaign{
			UploadCampaign: &pb.UploadCampaignRequest{
				Data:          data,
				CampaignId:    campaignID,
				SchemaVersion: config.SchemaVersion,
				Mode:          mode,
			},
		},
	}

	resp, err := c.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	ul := resp.GetUploadCampaign()
	if ul == nil {
		return nil, errors.New("upload campaign: empty response")
	}

	return ul, nil
}

// ListRemoteCampaigns fetches the campaign list from the server.
func ListRemoteCampaigns(ctx context.Context, c *Client) ([]*pb.CampaignOptionProto, error) {
	req := &pb.Request{
		Method: pb.Method_LIST_CAMPAIGNS,
		Payload: &pb.Request_ListCampaigns{
			ListCampaigns: &pb.ListCampaignsRequest{},
		},
	}

	resp, err := c.Call(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}

	lc := resp.GetListCampaigns()
	if lc == nil {
		return nil, nil
	}

	return lc.Campaigns, nil
}

// extractCampaignDB creates a minimal SQLite DB file containing just the
// specified campaign's data. Returns the file contents as bytes.
//
//nolint:cyclop,revive // sequential table-copy steps are inherently linear
func extractCampaignDB(ctx context.Context, localDBPath, campaignID string) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "lootsheet-extract-*.db")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer os.Remove(tmpPath)

	// Initialize the temp DB with the full schema.
	assets, err := config.LoadInitAssets()
	if err != nil {
		return nil, fmt.Errorf("load init assets: %w", err)
	}

	if _, err := ledger.EnsureSQLiteInitialized(ctx, tmpPath, assets); err != nil {
		return nil, fmt.Errorf("initialize temp database: %w", err)
	}

	// Open temp DB and attach source.
	db, err := ledger.OpenDB(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("open temp database: %w", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, "ATTACH DATABASE ? AS source", localDBPath); err != nil {
		return nil, fmt.Errorf("attach source database: %w", err)
	}

	// Verify campaign exists in source.
	var exists int
	err = db.QueryRowContext(ctx,
		"SELECT 1 FROM source.campaigns WHERE id = ?", campaignID,
	).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("campaign %q not found in local database", campaignID)
	}
	if err != nil {
		return nil, fmt.Errorf("check source campaign: %w", err)
	}

	// Copy data in FK order within a transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin extract transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete the default campaign and its seed accounts from the temp DB.
	if _, err := tx.ExecContext(ctx, "DELETE FROM accounts"); err != nil {
		return nil, fmt.Errorf("clear seed accounts: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM campaigns"); err != nil {
		return nil, fmt.Errorf("clear default campaign: %w", err)
	}

	// 1. campaigns
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO campaigns SELECT * FROM source.campaigns WHERE id = ?",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy campaigns: %w", err)
	}

	// 2. accounts
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO accounts SELECT * FROM source.accounts WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy accounts: %w", err)
	}

	// 3. quests
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO quests SELECT * FROM source.quests WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy quests: %w", err)
	}

	// 4. loot_items
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO loot_items SELECT * FROM source.loot_items WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy loot_items: %w", err)
	}

	// 5. journal_entries — NULL reverses_entry_id first, then non-NULL
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries SELECT * FROM source.journal_entries WHERE campaign_id = ? AND reverses_entry_id IS NULL",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy journal_entries (base): %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO journal_entries SELECT * FROM source.journal_entries WHERE campaign_id = ? AND reverses_entry_id IS NOT NULL",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy journal_entries (reversals): %w", err)
	}

	// 6. journal_lines
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO journal_lines SELECT jl.* FROM source.journal_lines jl
		 JOIN source.journal_entries je ON je.id = jl.journal_entry_id
		 WHERE je.campaign_id = ?`,
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy journal_lines: %w", err)
	}

	// 7. loot_appraisals
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO loot_appraisals SELECT la.* FROM source.loot_appraisals la
		 JOIN source.loot_items li ON li.id = la.loot_item_id
		 WHERE li.campaign_id = ?`,
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy loot_appraisals: %w", err)
	}

	// 8. asset_template_lines
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO asset_template_lines SELECT atl.* FROM source.asset_template_lines atl
		 JOIN source.loot_items li ON li.id = atl.loot_item_id
		 WHERE li.campaign_id = ?`,
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy asset_template_lines: %w", err)
	}

	// 9. codex_types (INSERT OR IGNORE)
	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO codex_types SELECT ct.* FROM source.codex_types ct
		 WHERE ct.id IN (SELECT type_id FROM source.codex_entries WHERE campaign_id = ?)`,
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy codex_types: %w", err)
	}

	// 10. codex_entries
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO codex_entries SELECT * FROM source.codex_entries WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy codex_entries: %w", err)
	}

	// 11. notes
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO notes SELECT * FROM source.notes WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy notes: %w", err)
	}

	// 12. entity_references
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO entity_references SELECT * FROM source.entity_references WHERE campaign_id = ?",
		campaignID,
	); err != nil {
		return nil, fmt.Errorf("copy entity_references: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit extract: %w", err)
	}

	if _, err := db.ExecContext(ctx, "DETACH DATABASE source"); err != nil {
		return nil, fmt.Errorf("detach source: %w", err)
	}
	db.Close()

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("read temp database: %w", err)
	}

	return data, nil
}
