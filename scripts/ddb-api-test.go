//go:build ignore

// Quick validation script for D&D Beyond direct API access.
// Usage: go run scripts/ddb-api-test.go
// Then paste your cobalt cookie when prompted.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	fmt.Println("=== DDB Direct API Test ===")
	fmt.Println()

	// Step 1: Config endpoint (no auth)
	fmt.Println("[1/5] Testing config endpoint (no auth)...")
	configBody := testConfig()

	// Step 2: Get cobalt cookie
	fmt.Print("\nPaste your cobalt cookie: ")
	var cobalt string
	fmt.Scanln(&cobalt)
	cobalt = strings.TrimSpace(cobalt)
	if cobalt == "" {
		fmt.Println("No cookie provided, skipping authenticated tests.")
		return
	}

	// Step 3: Auth - exchange cobalt for bearer token
	fmt.Println("\n[2/5] Testing auth (cobalt → bearer token)...")
	token := testAuth(cobalt)
	if token == "" {
		fmt.Println("FAILED: Could not obtain bearer token. Stopping.")
		return
	}

	// Step 4: Monsters
	fmt.Println("\n[3/5] Testing monsters endpoint...")
	testMonsters(token)

	// Step 5: Items
	fmt.Println("\n[4/5] Testing items endpoint...")
	testItems(token)

	// Step 6: Spells
	fmt.Println("\n[5/5] Testing spells endpoint (class=Wizard, id=8)...")
	testSpells(token)

	// Step 7: Dump samples for field inspection
	fmt.Println("\n[bonus] Dumping single monster, item, and spell for field inspection...")
	dumpSample(token, "monster", monsterURL+"?search=goblin&skip=0&take=1")
	dumpSample(token, "item", itemsURL+"?sharingSetting=2")
	dumpSample(token, "spell", spellsURL+"?classId=8&classLevel=20&sharingSetting=2")

	// Step 8: Dump config sections for rules/conditions
	fmt.Println("\n[bonus] Dumping config sections (conditions, actions, rules)...")
	dumpConfigSections(configBody)

	fmt.Println("\n=== Done ===")
}

func testConfig() []byte {
	resp, err := http.Get(configURL)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return nil
	}
	defer resp.Body.Close()
	fmt.Printf("  Status: %d\n", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("  Parse error: %v\n", err)
		fmt.Printf("  Body preview: %.200s\n", string(body))
		return nil
	}

	// Check for sources
	if sources, ok := data["sources"]; ok {
		if arr, ok := sources.([]any); ok {
			fmt.Printf("  OK: Got %d sources\n", len(arr))
			for i, s := range arr {
				if i >= 3 {
					fmt.Printf("  ... and %d more\n", len(arr)-3)
					break
				}
				if m, ok := s.(map[string]any); ok {
					fmt.Printf("  - ID: %v, Name: %v\n", m["id"], m["description"])
				}
			}
		}
	} else {
		fmt.Printf("  Response keys: %v\n", mapKeys(data))
	}

	// Summarize all top-level keys
	fmt.Printf("  Config top-level keys:\n")
	for k, v := range data {
		switch val := v.(type) {
		case []any:
			fmt.Printf("    %s: array[%d]\n", k, len(val))
		case map[string]any:
			fmt.Printf("    %s: object\n", k)
		default:
			fmt.Printf("    %s: %T\n", k, v)
		}
	}

	return body
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
	fmt.Printf("  Status: %d\n", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("  Parse error: %v\n", err)
		fmt.Printf("  Body preview: %.200s\n", string(body))
		return ""
	}

	if token, ok := data["token"].(string); ok && token != "" {
		fmt.Printf("  OK: Got bearer token (%d chars)\n", len(token))
		return token
	}

	fmt.Printf("  FAILED: Response keys: %v\n", mapKeys(data))
	fmt.Printf("  Body: %.300s\n", string(body))
	return ""
}

func testMonsters(token string) {
	client := &http.Client{Timeout: 15 * time.Second}

	url := monsterURL + "?search=goblin&skip=0&take=5"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("  Status: %d\n", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("  Body preview: %.500s\n", string(body))
		return
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("  Parse error: %v\n", err)
		fmt.Printf("  Body preview: %.300s\n", string(body))
		return
	}

	prettyPrintResults("monster", body)
}

func testItems(token string) {
	client := &http.Client{Timeout: 15 * time.Second}

	url := itemsURL + "?sharingSetting=2"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("  Status: %d\n", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("  Body preview: %.500s\n", string(body))
		return
	}

	prettyPrintResults("item", body)
}

func testSpells(token string) {
	client := &http.Client{Timeout: 15 * time.Second}

	// Wizard classId=8, classLevel=20 to get all spells
	url := spellsURL + "?classId=8&classLevel=20&sharingSetting=2"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("  Status: %d\n", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("  Body preview: %.500s\n", string(body))
		return
	}

	prettyPrintResults("spell", body)
}

func prettyPrintResults(label string, body []byte) {
	// Try as array
	var arr []any
	if err := json.Unmarshal(body, &arr); err == nil {
		fmt.Printf("  OK: Got %d %ss\n", len(arr), label)
		for i, item := range arr {
			if i >= 3 {
				fmt.Printf("  ... and %d more\n", len(arr)-3)
				break
			}
			if m, ok := item.(map[string]any); ok {
				printEntryPreview(m)
			}
		}
		return
	}

	// Try as object with nested data
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		// Check for pagination wrapper
		if data, ok := obj["data"]; ok {
			if dataArr, ok := data.([]any); ok {
				fmt.Printf("  OK: Got %d %ss (wrapped)\n", len(dataArr), label)
				for i, item := range dataArr {
					if i >= 3 {
						fmt.Printf("  ... and %d more\n", len(dataArr)-3)
						break
					}
					if m, ok := item.(map[string]any); ok {
						printEntryPreview(m)
					}
				}
				return
			}
		}
		fmt.Printf("  Response keys: %v\n", mapKeys(obj))

		// Dump top-level structure to understand format
		for k, v := range obj {
			switch val := v.(type) {
			case []any:
				fmt.Printf("  key %q: array[%d]\n", k, len(val))
			case map[string]any:
				fmt.Printf("  key %q: object{%v}\n", k, mapKeys(val))
			default:
				fmt.Printf("  key %q: %T = %v\n", k, v, truncate(fmt.Sprintf("%v", v), 80))
			}
		}
		return
	}

	fmt.Printf("  Body preview: %.500s\n", string(body))
}

func printEntryPreview(m map[string]any) {
	name := ""
	if n, ok := m["name"].(string); ok {
		name = n
	}
	if name == "" {
		// Try definition.name pattern
		if def, ok := m["definition"].(map[string]any); ok {
			if n, ok := def["name"].(string); ok {
				name = n
			}
		}
	}

	id := m["id"]
	tp := ""
	if t, ok := m["type"].(string); ok {
		tp = t
	}

	if name != "" {
		fmt.Printf("  - ID: %v  Name: %s  Type: %s\n", id, name, tp)
	} else {
		fmt.Printf("  - Keys: %v\n", mapKeys(m))
	}
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Dump saves full JSON response to a file for inspection.
func dump(label string, body []byte) {
	fname := fmt.Sprintf("ddb_%s_response.json", label)
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		body = pretty.Bytes()
	}
	if err := os.WriteFile(fname, body, 0o644); err != nil {
		fmt.Printf("  (could not dump to %s: %v)\n", fname, err)
	} else {
		fmt.Printf("  Full response saved to %s\n", fname)
	}
}

// dumpConfigSections extracts and dumps conditions, basicActions, and rules from the config response.
func dumpConfigSections(body []byte) {
	if body == nil {
		fmt.Println("  Skipped (no config body)")
		return
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("  Parse error: %v\n", err)
		return
	}

	sections := []string{"conditions", "basicActions", "rules", "senses", "damageTypes", "weaponProperties"}
	for _, key := range sections {
		if arr, ok := data[key].([]any); ok {
			fmt.Printf("  %s: %d entries\n", key, len(arr))
			// Dump first entry
			if len(arr) > 0 {
				first, _ := json.MarshalIndent(arr[0], "    ", "  ")
				fmt.Printf("    sample: %s\n", string(first))
			}
			// Dump full section
			full, _ := json.MarshalIndent(arr, "", "  ")
			dump("config_"+key, full)
		} else {
			fmt.Printf("  %s: not found or not array\n", key)
		}
	}
}

// dumpSample fetches and saves the first entry from a wrapped response for field inspection.
func dumpSample(token, label, url string) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  %s: ERROR %v\n", label, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("  %s: HTTP %d\n", label, resp.StatusCode)
		return
	}

	// Extract first entry from wrapped response
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		if data, ok := obj["data"].([]any); ok && len(data) > 0 {
			first, _ := json.MarshalIndent(data[0], "", "  ")
			dump(label+"_sample", first)
			return
		}
	}

	// Fallback: dump full response
	dump(label+"_full", body)
}
