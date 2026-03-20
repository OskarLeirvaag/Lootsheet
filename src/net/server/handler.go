package server

import (
	"context"
	"errors"
	"fmt"

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
	default:
		return errorResponse(fmt.Sprintf("unknown method: %v", req.Method))
	}
}

func handleBuildShellData(ctx context.Context, svc TUIService) *pb.Response {
	data, err := svc.BuildShellData(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}
	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_BuildShellData{
			BuildShellData: &pb.BuildShellDataResponse{
				Data: pb.ShellDataToProto(&data),
			},
		},
	}
}

func handleExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest, svc TUIService) *pb.Response {
	if req == nil || req.Command == nil {
		return errorResponse("execute_command: missing command")
	}

	cmd := pb.CommandFromProto(req.Command)
	result, err := svc.HandleCommand(ctx, cmd)
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

	return &pb.Response{
		Ok: true,
		Payload: &pb.Response_ExecuteCommand{
			ExecuteCommand: &pb.ExecuteCommandResponse{
				Result: pb.CommandResultToProto(&result),
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

func errorResponse(msg string) *pb.Response {
	return &pb.Response{Ok: false, Error: msg}
}
