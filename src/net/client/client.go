// Package client implements the LootSheet TCP client that connects to a
// remote server over TLS and exchanges protobuf messages.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
	"github.com/OskarLeirvaag/Lootsheet/src/net/wire"
)

// Client wraps a TLS connection to a LootSheet server and provides typed
// RPC-style calls. It serialises concurrent Call invocations.
type Client struct {
	conn *tls.Conn
	mu   sync.Mutex
}

// Dial connects to the server at addr using the given bearer token. It
// performs the TLS handshake and AUTH exchange, returning the authenticated
// client and server greeting on success.
func Dial(ctx context.Context, addr, token string) (*Client, *pb.AuthResponse, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // self-signed server cert
		MinVersion:         tls.VersionTLS13,
	}

	dialer := tls.Dialer{Config: tlsCfg}
	rawConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	conn := rawConn.(*tls.Conn)
	c := &Client{conn: conn}

	// Send AUTH request with protocol version.
	authReq := &pb.Request{
		Method: pb.Method_AUTH,
		Payload: &pb.Request_Auth{
			Auth: &pb.AuthRequest{
				Token:           token,
				ProtocolVersion: pb.ProtocolVersion,
			},
		},
	}

	if err := wire.WriteMessage(conn, authReq); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("send auth: %w", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(conn, resp); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("read auth response: %w", err)
	}

	if !resp.Ok {
		conn.Close()
		return nil, nil, fmt.Errorf("auth rejected: %s", resp.Error)
	}

	authResp := resp.GetAuth()
	if authResp == nil {
		conn.Close()
		return nil, nil, fmt.Errorf("auth response missing payload")
	}

	// Verify the server's protocol version. An old server that predates the
	// version field will report 0, which will not match ProtocolVersion (>= 1).
	if sv := authResp.ProtocolVersion; sv != pb.ProtocolVersion {
		conn.Close()
		target := "server"
		if sv > pb.ProtocolVersion {
			target = "client"
		}
		return nil, nil, fmt.Errorf(
			"protocol version mismatch: server=%d, client=%d — upgrade the %s binary",
			sv, pb.ProtocolVersion, target,
		)
	}

	return c, authResp, nil
}

// Call sends a request and returns the response. It is safe for concurrent use
// but serialises requests (the TUI is single-threaded anyway).
func (c *Client) Call(_ context.Context, req *pb.Request) (*pb.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := wire.WriteMessage(c.conn, req); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	resp := new(pb.Response)
	if err := wire.ReadMessage(c.conn, resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if !resp.Ok {
		// Check for input error prefix.
		if strings.HasPrefix(resp.Error, "input:") {
			return resp, nil // let caller inspect
		}
		return nil, fmt.Errorf("server error: %s", resp.Error)
	}

	return resp, nil
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
