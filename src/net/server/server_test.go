package server_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/net/server"
)

func TestLoadOrGenerateToken(t *testing.T) {
	dir := t.TempDir()

	token1, err := server.LoadOrGenerateToken(dir)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if len(token1) != 64 {
		t.Errorf("token length = %d, want 64", len(token1))
	}

	token2, err := server.LoadOrGenerateToken(dir)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if token1 != token2 {
		t.Errorf("tokens differ on reload: %q vs %q", token1, token2)
	}
}

func TestValidateToken(t *testing.T) {
	if !server.ValidateToken("abc", "abc") {
		t.Error("expected match for identical tokens")
	}
	if server.ValidateToken("abc", "xyz") {
		t.Error("expected mismatch for different tokens")
	}
}

func TestLoadOrGenerateTLS(t *testing.T) {
	dir := t.TempDir()

	cfg, err := server.LoadOrGenerateTLS(dir)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if len(cfg.Certificates) == 0 {
		t.Fatal("no certificates")
	}

	if _, err := os.Stat(filepath.Join(dir, "cert.pem")); err != nil {
		t.Errorf("cert.pem missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "key.pem")); err != nil {
		t.Errorf("key.pem missing: %v", err)
	}

	cfg2, err := server.LoadOrGenerateTLS(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(cfg2.Certificates) == 0 {
		t.Fatal("no certificates on reload")
	}
}

func TestTokenFilePermissions(t *testing.T) {
	dir := t.TempDir()

	if _, err := server.LoadOrGenerateToken(dir); err != nil {
		t.Fatalf("LoadOrGenerateToken: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "token"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("mode = %o, want 0600", mode)
	}
}

func TestTokenNotEmpty(t *testing.T) {
	dir := t.TempDir()

	token, err := server.LoadOrGenerateToken(dir)
	if err != nil {
		t.Fatalf("LoadOrGenerateToken: %v", err)
	}
	if strings.TrimSpace(token) == "" {
		t.Error("token is empty")
	}
}
