package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// TUIService abstracts the domain operations the server needs. The app package
// provides an implementation that wraps buildTUIShellData, handleTUICommand,
// and the search/campaign helpers.
type TUIService interface {
	BuildShellData(ctx context.Context) (model.ShellData, error)
	HandleCommand(ctx context.Context, cmd model.Command) (model.CommandResult, error)
	SetCampaign(ctx context.Context, campaignID string) error
	ListCampaigns(ctx context.Context) ([]model.CampaignOption, error)
	SearchCodexEntries(ctx context.Context, query string) ([]model.ListItemData, error)
	SearchNotes(ctx context.Context, query string) ([]model.ListItemData, error)
	SearchCompendium(ctx context.Context, section model.Section, query string) ([]model.ListItemData, error)
	DatabasePath() string
}

// NewHandler returns a Handler that dispatches protobuf requests to the given
// TUIService.
func NewHandler(svc TUIService) Handler {
	return func(ctx context.Context, req *pb.Request) *pb.Response {
		return dispatch(ctx, req, svc)
	}
}

func dispatch(ctx context.Context, req *pb.Request, svc TUIService) *pb.Response {
	switch req.Method {
	case pb.Method_BUILD_SHELL_DATA:
		return handleBuildShellData(ctx, svc)
	case pb.Method_EXECUTE_COMMAND:
		return handleExecuteCommand(ctx, req.GetExecuteCommand(), svc)
	case pb.Method_SET_CAMPAIGN:
		return handleSetCampaign(ctx, req.GetSetCampaign(), svc)
	case pb.Method_LIST_CAMPAIGNS:
		return handleListCampaigns(ctx, svc)
	case pb.Method_SEARCH_CODEX_ENTRIES:
		return handleSearchCodex(ctx, req.GetSearchCodex(), svc)
	case pb.Method_SEARCH_NOTES:
		return handleSearchNotes(ctx, req.GetSearchNotes(), svc)
	case pb.Method_SEARCH_COMPENDIUM:
		return handleSearchCompendium(ctx, req.GetSearchCompendium(), svc)
	case pb.Method_DOWNLOAD_DATABASE:
		return handleDownloadDatabase(svc)
	case pb.Method_UPLOAD_CAMPAIGN:
		return handleUploadCampaign(ctx, req.GetUploadCampaign(), svc)
	case pb.Method_PING:
		return &pb.Response{Ok: true}
	default:
		return errorResponse(fmt.Sprintf("unknown method: %v", req.Method))
	}
}

func handleBuildShellData(ctx context.Context, svc TUIService) *pb.Response {
	buildStart := time.Now()
	data, err := svc.BuildShellData(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}
	buildElapsed := time.Since(buildStart)

	protoStart := time.Now()
	protoData := pb.ShellDataToProto(&data)
	protoElapsed := time.Since(protoStart)

	slog.InfoContext(ctx, "handleBuildShellData",
		slog.Duration("build", buildElapsed),
		slog.Duration("proto", protoElapsed),
	)
	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_BuildShellData{
			BuildShellData: &pb.BuildShellDataResponse{
				Data: protoData,
			},
		},
	}
}

func handleExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest, svc TUIService) *pb.Response {
	if req == nil || req.Command == nil {
		return errorResponse("execute_command: missing command")
	}

	cmd := pb.CommandFromProto(req.Command)
	cmdStart := time.Now()
	result, err := svc.HandleCommand(ctx, cmd)
	cmdElapsed := time.Since(cmdStart)
	if err != nil {
		// Check for input errors — these should be surfaced to the client
		// as structured error responses so the TUI keeps the modal open.
		var inputErr model.InputError
		if errors.As(err, &inputErr) {
			return &pb.Response{
				Ok:    false,
				Error: "input:" + inputErr.Message,
			}
		}
		return errorResponse(err.Error())
	}

	protoStart := time.Now()
	protoResult := pb.CommandResultToProto(&result)
	protoElapsed := time.Since(protoStart)

	slog.InfoContext(ctx, "handleExecuteCommand",
		slog.String("command_id", cmd.ID),
		slog.Duration("handle", cmdElapsed),
		slog.Duration("proto", protoElapsed),
	)
	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_ExecuteCommand{
			ExecuteCommand: &pb.ExecuteCommandResponse{
				Result: protoResult,
			},
		},
	}
}

func handleSetCampaign(ctx context.Context, req *pb.SetCampaignRequest, svc TUIService) *pb.Response {
	if req == nil {
		return errorResponse("set_campaign: missing request")
	}

	if err := svc.SetCampaign(ctx, req.CampaignId); err != nil {
		return errorResponse(err.Error())
	}

	return &pb.Response{
		Ok:      true,
		Payload: &pb.Response_SetCampaign{SetCampaign: &pb.SetCampaignResponse{}},
	}
}

func handleListCampaigns(ctx context.Context, svc TUIService) *pb.Response {
	campaigns, err := svc.ListCampaigns(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_ListCampaigns{
			ListCampaigns: &pb.ListCampaignsResponse{
				Campaigns: pb.CampaignOptionsToProto(campaigns),
			},
		},
	}
}

func handleSearchCodex(ctx context.Context, req *pb.SearchRequest, svc TUIService) *pb.Response {
	if req == nil {
		return errorResponse("search_codex: missing request")
	}

	items, err := svc.SearchCodexEntries(ctx, req.Query)
	if err != nil {
		return errorResponse(err.Error())
	}

	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_SearchCodex{
			SearchCodex: &pb.SearchResponse{
				Items: pb.ListItemsToProto(items),
			},
		},
	}
}

func handleSearchNotes(ctx context.Context, req *pb.SearchRequest, svc TUIService) *pb.Response {
	if req == nil {
		return errorResponse("search_notes: missing request")
	}

	items, err := svc.SearchNotes(ctx, req.Query)
	if err != nil {
		return errorResponse(err.Error())
	}

	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_SearchNotes{
			SearchNotes: &pb.SearchResponse{
				Items: pb.ListItemsToProto(items),
			},
		},
	}
}

func handleSearchCompendium(ctx context.Context, req *pb.SearchRequest, svc TUIService) *pb.Response {
	if req == nil {
		return errorResponse("search_compendium: missing request")
	}

	items, err := svc.SearchCompendium(ctx, model.Section(req.Section), req.Query)
	if err != nil {
		return errorResponse(err.Error())
	}

	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_SearchCompendium{
			SearchCompendium: &pb.SearchResponse{
				Items: pb.ListItemsToProto(items),
			},
		},
	}
}

func handleDownloadDatabase(svc TUIService) *pb.Response {
	dbPath := svc.DatabasePath()
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return errorResponse(fmt.Sprintf("read database: %v", err))
	}

	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_DownloadDatabase{
			DownloadDatabase: &pb.DownloadDatabaseResponse{
				Data:     data,
				Filename: filepath.Base(dbPath),
			},
		},
	}
}

func errorResponse(msg string) *pb.Response {
	return &pb.Response{Ok: false, Error: msg}
}
