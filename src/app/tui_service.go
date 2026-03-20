package app

import (
	"context"
	"fmt"

	"github.com/OskarLeirvaag/Lootsheet/src/ledger/campaign"
	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// tuiService implements server.TUIService by wrapping the existing
// buildTUIShellData, handleTUICommand, and search/campaign helpers.
type tuiService struct {
	loader       TUIDataLoader
	databasePath string
}

func (s *tuiService) BuildShellData(ctx context.Context) (model.ShellData, error) {
	return buildTUIShellData(ctx, s.loader)
}

func (s *tuiService) HandleCommand(ctx context.Context, cmd model.Command) (model.CommandResult, error) {
	return handleTUICommand(ctx, cmd, s.databasePath, s.loader)
}

func (s *tuiService) SetCampaign(ctx context.Context, campaignID string) error {
	campaigns, err := campaign.List(ctx, s.databasePath)
	if err != nil {
		return fmt.Errorf("list campaigns: %w", err)
	}

	for _, c := range campaigns {
		if c.ID == campaignID {
			s.loader.SetCampaign(c.ID, c.Name)
			return nil
		}
	}

	return fmt.Errorf("campaign %q not found", campaignID)
}

func (s *tuiService) ListCampaigns(ctx context.Context) ([]model.CampaignOption, error) {
	campaigns, err := campaign.List(ctx, s.databasePath)
	if err != nil {
		return nil, err
	}

	opts := make([]model.CampaignOption, len(campaigns))
	for i, c := range campaigns {
		opts[i] = model.CampaignOption{ID: c.ID, Name: c.Name}
	}
	return opts, nil
}

func (s *tuiService) SearchCodexEntries(ctx context.Context, query string) ([]model.ListItemData, error) {
	entries, err := s.loader.SearchCodexEntries(ctx, query)
	if err != nil {
		return nil, err
	}
	return buildCodexItems(entries, nil), nil
}

func (s *tuiService) SearchNotes(ctx context.Context, query string) ([]model.ListItemData, error) {
	records, err := s.loader.SearchNotes(ctx, query)
	if err != nil {
		return nil, err
	}
	return buildNotesItems(records, nil), nil
}
