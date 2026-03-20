package client_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OskarLeirvaag/Lootsheet/src/net/client"
)

func TestTokenRoundTrip(t *testing.T) {
	dir := t.TempDir()

	// Initially no token.
	_, found, err := client.LookupToken(dir, "localhost:7547")
	if err != nil {
		t.Fatalf("LookupToken: %v", err)
	}
	if found {
		t.Error("expected no token initially")
	}

	// Save a token.
	if err := client.SaveToken(dir, "localhost:7547", "abc123"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	// Look it up.
	token, found, err := client.LookupToken(dir, "localhost:7547")
	if err != nil {
		t.Fatalf("LookupToken: %v", err)
	}
	if !found {
		t.Fatal("expected token to be found")
	}
	if token != "abc123" {
		t.Errorf("token = %q, want %q", token, "abc123")
	}

	// Different address has no token.
	_, found, err = client.LookupToken(dir, "other:1234")
	if err != nil {
		t.Fatalf("LookupToken: %v", err)
	}
	if found {
		t.Error("expected no token for different address")
	}
}

func TestTokenFilePermissions(t *testing.T) {
	dir := t.TempDir()

	if err := client.SaveToken(dir, "localhost:7547", "secret"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "tokens.json"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0o600 {
		t.Errorf("tokens.json mode = %o, want 0600", mode)
	}
}

func TestSaveTokenOverwrite(t *testing.T) {
	dir := t.TempDir()

	if err := client.SaveToken(dir, "localhost:7547", "first"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	if err := client.SaveToken(dir, "localhost:7547", "second"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	token, found, err := client.LookupToken(dir, "localhost:7547")
	if err != nil {
		t.Fatalf("LookupToken: %v", err)
	}
	if !found {
		t.Fatal("expected token")
	}
	if token != "second" {
		t.Errorf("token = %q, want %q", token, "second")
	}
}

func TestMultipleAddresses(t *testing.T) {
	dir := t.TempDir()

	if err := client.SaveToken(dir, "host1:7547", "token1"); err != nil {
		t.Fatalf("SaveToken host1: %v", err)
	}
	if err := client.SaveToken(dir, "host2:7547", "token2"); err != nil {
		t.Fatalf("SaveToken host2: %v", err)
	}

	tok1, _, _ := client.LookupToken(dir, "host1:7547")
	tok2, _, _ := client.LookupToken(dir, "host2:7547")

	if tok1 != "token1" {
		t.Errorf("host1 token = %q, want %q", tok1, "token1")
	}
	if tok2 != "token2" {
		t.Errorf("host2 token = %q, want %q", tok2, "token2")
	}
}
