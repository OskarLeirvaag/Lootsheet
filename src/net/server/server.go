// Package server implements the LootSheet TCP server that serves protobuf
// messages over TLS connections authenticated with a shared bearer token.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	pb "github.com/OskarLeirvaag/Lootsheet/src/net/proto"
	"github.com/OskarLeirvaag/Lootsheet/src/net/wire"
)

// Handler dispatches a single request and returns a response. The handler
// is called from per-connection goroutines and must be safe for concurrent use.
type Handler func(ctx context.Context, req *pb.Request) *pb.Response

// Server accepts TLS connections and runs the protobuf request-response loop.
type Server struct {
	listener net.Listener
	token    string
	handler  Handler
	log      *slog.Logger
	wg       sync.WaitGroup
}

// Config holds the parameters for creating a new server.
type Config struct {
	Addr      string
	TLSConfig *tls.Config
	Token     string
	Handler   Handler
	Logger    *slog.Logger
}

// ListenAndServe creates a listener and serves connections until ctx is
// cancelled. It blocks until all active connections have drained. If
// TLSConfig is nil, it listens on plain TCP (for use behind a TLS-terminating
// reverse proxy).
func ListenAndServe(ctx context.Context, cfg Config) error {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	var listener net.Listener
	var err error
	if cfg.TLSConfig != nil {
		listener, err = tls.Listen("tcp", cfg.Addr, cfg.TLSConfig)
	} else {
		listener, err = (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Addr)
	}
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s := &Server{
		listener: listener,
		token:    cfg.Token,
		handler:  cfg.Handler,
		log:      cfg.Logger,
	}

	// Close listener when context is cancelled to unblock Accept.
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	s.log.InfoContext(ctx, "server listening", slog.String("addr", listener.Addr().String()))

	var acceptErr error
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if we're shutting down.
			if ctx.Err() != nil {
				break
			}
			acceptErr = err
			s.log.ErrorContext(ctx, "accept error", slog.String("error", err.Error()))
			break
		}

		s.wg.Go(func() {
			s.handleConnection(ctx, conn)
		})
	}

	s.log.InfoContext(ctx, "server shutting down, waiting for connections to drain")
	s.wg.Wait()
	s.log.InfoContext(ctx, "server stopped")
	return acceptErr
}

//nolint:revive // cognitive-complexity marginally over threshold (26 vs 25); splitting would harm readability
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	s.log.InfoContext(ctx, "connection accepted", slog.String("remote", remote))

	// First message must be AUTH.
	authReq := new(pb.Request)
	if err := wire.ReadMessage(conn, authReq); err != nil {
		s.log.WarnContext(ctx, "read auth failed", slog.String("remote", remote), slog.String("error", err.Error()))
		return
	}

	if authReq.Method != pb.Method_AUTH || authReq.GetAuth() == nil {
		s.writeError(conn, "first message must be AUTH")
		s.log.WarnContext(ctx, "bad first message", slog.String("remote", remote))
		return
	}

	if !ValidateToken(authReq.GetAuth().Token, s.token) {
		s.writeError(conn, "invalid token")
		s.log.WarnContext(ctx, "invalid token", slog.String("remote", remote))
		return
	}

	clientVersion := authReq.GetAuth().ProtocolVersion
	serverVersion := pb.ProtocolVersion
	if clientVersion != serverVersion {
		msg := fmt.Sprintf(
			"protocol version mismatch: server=%d, client=%d — upgrade the %s binary",
			serverVersion, clientVersion, versionUpgradeTarget(serverVersion, clientVersion),
		)
		s.writeError(conn, msg)
		s.log.WarnContext(ctx, "version mismatch",
			slog.String("remote", remote),
			slog.Uint64("server_version", uint64(serverVersion)),
			slog.Uint64("client_version", uint64(clientVersion)),
		)
		return
	}

	// Send auth success.
	authResp := &pb.Response{
		Ok: true,
		Payload: &pb.Response_Auth{
			Auth: &pb.AuthResponse{
				ServerName:      "LootSheet Server",
				ProtocolVersion: serverVersion,
			},
		},
	}
	if err := wire.WriteMessage(conn, authResp); err != nil {
		s.log.WarnContext(ctx, "write auth response failed", slog.String("remote", remote), slog.String("error", err.Error()))
		return
	}

	s.log.InfoContext(ctx, "client authenticated", slog.String("remote", remote))

	// Read messages in a separate goroutine so the main loop can select on
	// ctx.Done() for clean shutdown.
	type readResult struct {
		req *pb.Request
		err error
	}
	reads := make(chan readResult, 1)

	go func() {
		for {
			req := new(pb.Request)
			err := wire.ReadMessage(conn, req)
			reads <- readResult{req, err}
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.log.InfoContext(ctx, "shutting down connection", slog.String("remote", remote))
			s.writeError(conn, "server shutting down")
			return

		case rr := <-reads:
			if rr.err != nil {
				if errors.Is(rr.err, io.EOF) || errors.Is(rr.err, io.ErrUnexpectedEOF) || errors.Is(rr.err, net.ErrClosed) {
					s.log.InfoContext(ctx, "client disconnected", slog.String("remote", remote))
				} else if ctx.Err() == nil {
					s.log.WarnContext(ctx, "read request failed", slog.String("remote", remote), slog.String("error", rr.err.Error()))
				}
				return
			}

			resp := s.handler(ctx, rr.req)
			if err := wire.WriteMessage(conn, resp); err != nil {
				s.log.WarnContext(ctx, "write response failed", slog.String("remote", remote), slog.String("error", err.Error()))
				return
			}
		}
	}
}

func (*Server) writeError(conn net.Conn, msg string) {
	resp := &pb.Response{Ok: false, Error: msg}
	_ = wire.WriteMessage(conn, resp)
}

// versionUpgradeTarget returns "server" or "client" depending on which side
// is running the older protocol and needs to be upgraded.
func versionUpgradeTarget(serverVersion, clientVersion uint32) string {
	if serverVersion < clientVersion {
		return "server"
	}
	return "client"
}
