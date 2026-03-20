package client

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

// RemoteShellLoader returns a ShellLoader that fetches the full shell data
// snapshot from the remote server.
func RemoteShellLoader(c *Client) func(context.Context) (model.ShellData, error) {
	return func(ctx context.Context) (model.ShellData, error) {
		req := &pb.Request{
			Method: pb.Method_BUILD_SHELL_DATA,
			Payload: &pb.Request_BuildShellData{
				BuildShellData: &pb.BuildShellDataRequest{},
			},
		}

		resp, err := c.Call(ctx, req)
		if err != nil {
			return model.ShellData{}, fmt.Errorf("build shell data: %w", err)
		}

		bsd := resp.GetBuildShellData()
		if bsd == nil || bsd.Data == nil {
			return model.ShellData{}, fmt.Errorf("build shell data: empty response")
		}

		return pb.ShellDataFromProto(bsd.Data), nil
	}
}

// RemoteCommandHandler returns a CommandHandler that sends commands to the
// remote server for execution.
func RemoteCommandHandler(c *Client) func(context.Context, model.Command) (model.CommandResult, error) {
	return func(ctx context.Context, cmd model.Command) (model.CommandResult, error) {
		req := &pb.Request{
			Method: pb.Method_EXECUTE_COMMAND,
			Payload: &pb.Request_ExecuteCommand{
				ExecuteCommand: &pb.ExecuteCommandRequest{
					Command: pb.CommandToProto(cmd),
				},
			},
		}

		resp, err := c.Call(ctx, req)
		if err != nil {
			return model.CommandResult{}, fmt.Errorf("execute command: %w", err)
		}

		// Check for input error (modal should stay open).
		if !resp.Ok && strings.HasPrefix(resp.Error, "input:") {
			return model.CommandResult{}, model.InputError{
				Message: strings.TrimPrefix(resp.Error, "input:"),
			}
		}

		ec := resp.GetExecuteCommand()
		if ec == nil || ec.Result == nil {
			return model.CommandResult{}, fmt.Errorf("execute command: empty response")
		}

		return pb.CommandResultFromProto(ec.Result), nil
	}
}

// RemoteSearchHandler returns a SearchHandler that delegates codex and notes
// searches to the remote server. Other sections return nil to fall back to
// client-side filtering.
func RemoteSearchHandler(c *Client) func(model.Section, string) ([]model.ListItemData, error) {
	return func(section model.Section, query string) ([]model.ListItemData, error) {
		switch section {
		case model.SectionCodex:
			return remoteSearch(c, &pb.Request{
				Method:  pb.Method_SEARCH_CODEX_ENTRIES,
				Payload: &pb.Request_SearchCodex{SearchCodex: &pb.SearchRequest{Query: query}},
			}, func(r *pb.Response) *pb.SearchResponse { return r.GetSearchCodex() })
		case model.SectionNotes:
			return remoteSearch(c, &pb.Request{
				Method:  pb.Method_SEARCH_NOTES,
				Payload: &pb.Request_SearchNotes{SearchNotes: &pb.SearchRequest{Query: query}},
			}, func(r *pb.Response) *pb.SearchResponse { return r.GetSearchNotes() })
		default:
			return nil, nil
		}
	}
}

func remoteSearch(c *Client, req *pb.Request, extract func(*pb.Response) *pb.SearchResponse) ([]model.ListItemData, error) {
	resp, err := c.Call(context.Background(), req)
	if err != nil {
		return nil, err
	}

	sr := extract(resp)
	if sr == nil {
		return nil, nil
	}

	return pb.ListItemsFromProto(sr.Items), nil
}
