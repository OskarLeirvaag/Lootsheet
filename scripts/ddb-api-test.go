//go:build ignore

// Quick validation script for D&D Beyond direct API access.
// Usage: go run scripts/ddb-api-test.go
// Paste your cobalt cookie when prompted.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	authURL    = "https://auth-service.dndbeyond.com/v1/cobalt-token"
	configURL  = "https://www.dndbeyond.com/api/config/json"
	monsterURL = "https://monster-service.dndbeyond.com/v1/Monster"
	itemsURL   = "https://character-service.dndbeyond.com/character/v5/game-data/items"
	spellsURL  = "https://character-service.dndbeyond.com/character/v5/game-data/spells"
)

func main() {
	fmt.Println("=== DDB Direct API Test ===\n")

	fmt.Println("[1/5] Config endpoint (no auth)...")
	testConfig()

	fmt.Print("\nPaste your cobalt cookie (or Enter to skip): ")
	var cobalt string
	fmt.Scanln(&cobalt)
	cobalt = strings.TrimSpace(cobalt)
	if cobalt == "" {
		fmt.Println("Skipped authenticated tests.")
		return
	}

	fmt.Println("\n[2/5] Auth (cobalt → bearer)...")
	token := testAuth(cobalt)
	if token == "" {
		fmt.Println("FAILED: no bearer token.")
		return
	}

	fmt.Println("\n[3/5] Monsters (search=goblin, take=5)...")
	testEndpoint(token, monsterURL+"?search=goblin&skip=0&take=5", "monster")

	fmt.Println("\n[4/5] Items...")
	testEndpoint(token, itemsURL+"?sharingSetting=2", "item")

	fmt.Println("\n[5/5] Spells (Wizard, id=8)...")
	testEndpoint(token, spellsURL+"?classId=8&classLevel=20&sharingSetting=2", "spell")

	fmt.Println("\n=== Done ===")
}

func testConfig() {
	resp, err := http.Get(configURL)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Status: %d\n", resp.StatusCode)

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("  Parse error: %v\n", err)
		return
	}

	for _, key := range []string{"sources", "conditions", "basicActions", "rules", "weaponProperties", "challengeRatings", "creatureSizes", "monsterTypes"} {
		if arr, ok := data[key].([]any); ok {
			fmt.Printf("  %s: %d entries\n", key, len(arr))
		}
	}
}

func testAuth(cobalt string) string {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("POST", authURL, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "CobaltSession="+cobalt)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Status: %d\n", resp.StatusCode)

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("  Parse error: %v\n", err)
		return ""
	}

	if token, ok := data["token"].(string); ok && token != "" {
		fmt.Printf("  OK: bearer token (%d chars)\n", len(token))
		return token
	}

	fmt.Printf("  FAILED: %s\n", truncate(string(body), 200))
	return ""
}

func testEndpoint(token, url, label string) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  Status: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Printf("  Body: %s\n", truncate(string(body), 200))
		return
	}

	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		if arr, ok := obj["data"].([]any); ok {
			fmt.Printf("  OK: %d %ss\n", len(arr), label)
			for i, item := range arr {
				if i >= 3 {
					fmt.Printf("  ... and %d more\n", len(arr)-3)
					break
				}
				printPreview(item)
			}
			return
		}
	}

	fmt.Printf("  Body: %s\n", truncate(string(body), 300))
}

func printPreview(v any) {
	m, ok := v.(map[string]any)
	if !ok {
		return
	}

	name, _ := m["name"].(string)
	if name == "" {
		if def, ok := m["definition"].(map[string]any); ok {
			name, _ = def["name"].(string)
		}
	}

	fmt.Printf("  - ID: %v  Name: %s\n", m["id"], name)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
