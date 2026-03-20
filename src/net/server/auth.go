package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	tokenFileName = "token"
	tokenBytes    = 32
	dirPerm       = 0o755
	filePerm      = 0o600
)

// LoadOrGenerateToken returns the bearer token stored in dir, or generates a
// new 32-byte hex-encoded token if none exists. The token file is created with
// mode 0600.
func LoadOrGenerateToken(dir string) (string, error) {
	tokenPath := filepath.Join(dir, tokenFileName)

	data, err := os.ReadFile(tokenPath)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token != "" {
			return token, nil
		}
	}

	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return "", fmt.Errorf("create token directory: %w", err)
	}

	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	token := hex.EncodeToString(raw)
	if err := os.WriteFile(tokenPath, []byte(token+"\n"), filePerm); err != nil {
		return "", fmt.Errorf("write token file: %w", err)
	}

	return token, nil
}

// ValidateToken performs a constant-time comparison of got and want.
func ValidateToken(got, want string) bool {
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
