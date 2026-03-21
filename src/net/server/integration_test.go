package server_test

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"testing"
	"time"

	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
	"github.com/OskarLeirvaag/Lootsheet/src/net/server"
	"github.com/OskarLeirvaag/Lootsheet/src/net/wire"
	"github.com/OskarLeirvaag/Lootsheet/src/render/model"
)

const testToken = "test-token"

// testService implements server.TUIService for integration testing.
type testService struct {
	campaignID string
	dbPath     string
}

func (s *testService) DatabasePath() string { return s.dbPath }

func (*testService) BuildShellData(_ context.Context) (model.ShellData, error) {
	return model.ShellData{
		CampaignName: "Test Campaign",
		Dashboard: model.DashboardData{
			HeaderLines: []string{"Test Dashboard"},
		},
	}, nil
}

func (*testService) HandleCommand(_ context.Context, cmd model.Command) (model.CommandResult, error) {
	return model.CommandResult{
		Status: model.StatusMessage{
			Level: model.StatusSuccess,
			Text:  "executed: " + cmd.ID,
		},
	}, nil
}

func (s *testService) SetCampaign(_ context.Context, id string) error {
	s.campaignID = id
	return nil
}

func (*testService) ListCampaigns(_ context.Context) ([]model.CampaignOption, error) {
	return []model.CampaignOption{
		{ID: "c1", Name: "Campaign One"},
		{ID: "c2", Name: "Campaign Two"},
	}, nil
}

func (*testService) SearchCodexEntries(_ context.Context, query string) ([]model.ListItemData, error) {
	return []model.ListItemData{
		{Key: "codex-1", Row: "Result: " + query},
	}, nil
}

func (*testService) SearchNotes(_ context.Context, query string) ([]model.ListItemData, error) {
	return []model.ListItemData{
		{Key: "note-1", Row: "Note: " + query},
	}, nil
}

func startTestServer(t *testing.T, token string) (string, context.CancelFunc) {
	t.Helper()

	dir := t.TempDir()
	tlsCfg, err := server.LoadOrGenerateTLS(dir)
	if err != nil {
		t.Fatalf("TLS: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec // cancel returned to caller

	// Use a dynamic port via :0 — server will bind and we read back the addr.
	// We have to pass a known addr though, so pick one and let the OS assign.
	addr := "127.0.0.1:0"

	// Start server in background. We use a channel to detect when it's ready.
	ready := make(chan string, 1)

	go func() {
		// Pre-listen to get a free port, close, then let the server bind to it.
		ln, listenErr := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
		if listenErr != nil {
			ready <- "127.0.0.1:17547"
			_ = server.ListenAndServe(ctx, server.Config{
				Addr:      "127.0.0.1:17547",
				TLSConfig: tlsCfg,
				Token:     token,
				Handler:   server.NewHandler(&testService{}),
			})
			return
		}
		boundAddr := ln.Addr().String()
		_ = ln.Close()
		ready <- boundAddr
		_ = server.ListenAndServe(ctx, server.Config{
			Addr:      boundAddr,
			TLSConfig: tlsCfg,
			Token:     token,
			Handler:   server.NewHandler(&testService{}),
		})
	}()

	boundAddr := <-ready
	// Give the server a moment to start listening after we released the port.
	time.Sleep(50 * time.Millisecond)

	return boundAddr, cancel
}

func dialTLS(t *testing.T, addr string) *tls.Conn {
	t.Helper()

	dialer := &tls.Dialer{Config: &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // self-signed test cert
		MinVersion:         tls.VersionTLS13,
	}}

	conn, err := dialer.DialContext(context.Background(), "tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		t.Fatal("expected *tls.Conn from dialer")
	}
	return tlsConn
}

func authenticate(t *testing.T, conn *tls.Conn, token string) {
	t.Helper()

	req := &pb.Request{
		Method: pb.Method_AUTH,
		Payload: &pb.Request_Auth{
			Auth: &pb.AuthRequest{
				Token:           token,
				ProtocolVersion: pb.ProtocolVersion,
				AppVersion:      pb.AppVersion,
			},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write auth: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read auth response: %v", err)
	}

	if !resp.Ok {
		t.Fatalf("auth failed: %s", resp.Error)
	}
}

func TestIntegrationAuthAndBuildShellData(t *testing.T) {
	token := "test-token-123"
	addr, cancel := startTestServer(t, token)
	defer cancel()

	conn := dialTLS(t, addr)
	defer conn.Close()

	authenticate(t, conn, token)

	// BuildShellData
	req := &pb.Request{
		Method: pb.Method_BUILD_SHELL_DATA,
		Payload: &pb.Request_BuildShellData{
			BuildShellData: &pb.BuildShellDataRequest{},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read: %v", err)
	}

	if !resp.Ok {
		t.Fatalf("build shell data failed: %s", resp.Error)
	}

	bsd := resp.GetBuildShellData()
	if bsd == nil || bsd.Data == nil {
		t.Fatal("empty build shell data response")
	}

	if bsd.Data.CampaignName != "Test Campaign" {
		t.Errorf("campaign = %q, want %q", bsd.Data.CampaignName, "Test Campaign")
	}
}

func TestIntegrationInvalidToken(t *testing.T) {
	token := "correct-token"
	addr, cancel := startTestServer(t, token)
	defer cancel()

	conn := dialTLS(t, addr)
	defer conn.Close()

	req := &pb.Request{
		Method: pb.Method_AUTH,
		Payload: &pb.Request_Auth{
			Auth: &pb.AuthRequest{
				Token:           "wrong-token",
				ProtocolVersion: pb.ProtocolVersion,
				AppVersion:      pb.AppVersion,
			},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read: %v", err)
	}

	if resp.Ok {
		t.Fatal("expected auth rejection")
	}
	if resp.Error != "invalid token" {
		t.Errorf("error = %q, want %q", resp.Error, "invalid token")
	}
}

func TestIntegrationExecuteCommand(t *testing.T) {
	token := testToken
	addr, cancel := startTestServer(t, token)
	defer cancel()

	conn := dialTLS(t, addr)
	defer conn.Close()

	authenticate(t, conn, token)

	req := &pb.Request{
		Method: pb.Method_EXECUTE_COMMAND,
		Payload: &pb.Request_ExecuteCommand{
			ExecuteCommand: &pb.ExecuteCommandRequest{
				Command: &pb.CommandProto{
					Id:      "test.command",
					Section: int32(model.SectionDashboard),
				},
			},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read: %v", err)
	}

	if !resp.Ok {
		t.Fatalf("execute command failed: %s", resp.Error)
	}

	ec := resp.GetExecuteCommand()
	if ec == nil || ec.Result == nil {
		t.Fatal("empty execute command response")
	}

	if ec.Result.Status.Text != "executed: test.command" {
		t.Errorf("status = %q, want %q", ec.Result.Status.Text, "executed: test.command")
	}
}

func TestIntegrationListCampaigns(t *testing.T) {
	token := testToken
	addr, cancel := startTestServer(t, token)
	defer cancel()

	conn := dialTLS(t, addr)
	defer conn.Close()

	authenticate(t, conn, token)

	req := &pb.Request{
		Method: pb.Method_LIST_CAMPAIGNS,
		Payload: &pb.Request_ListCampaigns{
			ListCampaigns: &pb.ListCampaignsRequest{},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read: %v", err)
	}

	if !resp.Ok {
		t.Fatalf("list campaigns failed: %s", resp.Error)
	}

	lc := resp.GetListCampaigns()
	if lc == nil {
		t.Fatal("empty list campaigns response")
	}

	if len(lc.Campaigns) != 2 {
		t.Fatalf("campaigns count = %d, want 2", len(lc.Campaigns))
	}

	if lc.Campaigns[0].Name != "Campaign One" {
		t.Errorf("campaign[0] = %q, want %q", lc.Campaigns[0].Name, "Campaign One")
	}
}

func TestIntegrationSearchCodex(t *testing.T) {
	token := testToken
	addr, cancel := startTestServer(t, token)
	defer cancel()

	conn := dialTLS(t, addr)
	defer conn.Close()

	authenticate(t, conn, token)

	req := &pb.Request{
		Method: pb.Method_SEARCH_CODEX_ENTRIES,
		Payload: &pb.Request_SearchCodex{
			SearchCodex: &pb.SearchRequest{Query: "dragon"},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read: %v", err)
	}

	if !resp.Ok {
		t.Fatalf("search failed: %s", resp.Error)
	}

	sr := resp.GetSearchCodex()
	if sr == nil || len(sr.Items) == 0 {
		t.Fatal("empty search response")
	}

	if sr.Items[0].Row != "Result: dragon" {
		t.Errorf("item row = %q, want %q", sr.Items[0].Row, "Result: dragon")
	}
}

func TestIntegrationVersionMismatch(t *testing.T) {
	token := testToken
	addr, cancel := startTestServer(t, token)
	defer cancel()

	conn := dialTLS(t, addr)
	defer conn.Close()

	// Send AUTH with a wrong (future) protocol version.
	req := &pb.Request{
		Method: pb.Method_AUTH,
		Payload: &pb.Request_Auth{
			Auth: &pb.AuthRequest{
				Token:           token,
				ProtocolVersion: pb.ProtocolVersion + 99,
			},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read: %v", err)
	}

	if resp.Ok {
		t.Fatal("expected version mismatch rejection")
	}

	if !strings.Contains(resp.Error, "protocol version mismatch") {
		t.Errorf("error = %q, want substring %q", resp.Error, "protocol version mismatch")
	}

	if !strings.Contains(resp.Error, "upgrade the server") {
		t.Errorf("error = %q, want suggestion to upgrade server", resp.Error)
	}
}

func TestIntegrationOldClientVersion(t *testing.T) {
	token := testToken
	addr, cancel := startTestServer(t, token)
	defer cancel()

	conn := dialTLS(t, addr)
	defer conn.Close()

	// Send AUTH with an old (zero) protocol version.
	req := &pb.Request{
		Method: pb.Method_AUTH,
		Payload: &pb.Request_Auth{
			Auth: &pb.AuthRequest{
				Token:           token,
				ProtocolVersion: 0,
			},
		},
	}

	if err := wire.WriteMessage(conn, req); err != nil {
		t.Fatalf("write: %v", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		t.Fatalf("read: %v", err)
	}

	if resp.Ok {
		t.Fatal("expected version mismatch rejection")
	}

	if !strings.Contains(resp.Error, "upgrade the client") {
		t.Errorf("error = %q, want suggestion to upgrade client", resp.Error)
	}
}
