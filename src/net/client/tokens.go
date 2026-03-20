package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	tokensFile = "tokens.json"
	dirPerm    = 0o755
	filePerm   = 0o600
)

// LookupToken returns the saved token for addr, if any.
func LookupToken(configDir, addr string) (string, bool, error) {
	tokens, err := loadTokens(configDir)
	if err != nil {
		return "", false, err
	}

	token, ok := tokens[addr]
	return token, ok, nil
}

// SaveToken persists the token for addr in the config directory.
func SaveToken(configDir, addr, token string) error {
	tokens, err := loadTokens(configDir)
	if err != nil {
		return err
	}

	tokens[addr] = token

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tokens: %w", err)
	}

	if err := os.MkdirAll(configDir, dirPerm); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(configDir, tokensFile)
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("write tokens file: %w", err)
	}

	return nil
}

func loadTokens(configDir string) (map[string]string, error) {
	path := filepath.Join(configDir, tokensFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read tokens file: %w", err)
	}

	var tokens map[string]string
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("parse tokens file: %w", err)
	}

	if tokens == nil {
		tokens = make(map[string]string)
	}

	return tokens, nil
}
