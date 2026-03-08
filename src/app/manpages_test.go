package app

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestGenerateManPagesWritesLeafAndGroupDocs(t *testing.T) {
	outputDir := t.TempDir()

	if err := GenerateManPages(outputDir); err != nil {
		t.Fatalf("generate man pages: %v", err)
	}

	for _, name := range []string{
		"lootsheet.1",
		"lootsheet-account.1",
		"lootsheet-account-create.1",
		"lootsheet-journal-post.1",
		"lootsheet-report-writeoff-candidates.1",
	} {
		path := filepath.Join(outputDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read generated man page %q: %v", name, err)
		}
		if !strings.Contains(string(content), "LOOTSHEET") {
			t.Fatalf("generated man page %q missing title", name)
		}
	}
}

func TestCheckedInManPagesMatchGeneratedOutput(t *testing.T) {
	outputDir := t.TempDir()

	if err := GenerateManPages(outputDir); err != nil {
		t.Fatalf("generate man pages: %v", err)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}

	wantDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "docs", "man")
	wantFiles := readManPageTree(t, wantDir)
	gotFiles := readManPageTree(t, outputDir)

	if len(wantFiles) == 0 {
		t.Fatalf("checked-in man page directory %q is empty", wantDir)
	}

	wantNames := sortedKeys(wantFiles)
	gotNames := sortedKeys(gotFiles)
	if !slices.Equal(wantNames, gotNames) {
		t.Fatalf("checked-in man page files differ\nwant: %v\ngot:  %v", wantNames, gotNames)
	}

	for _, name := range wantNames {
		if wantFiles[name] != gotFiles[name] {
			t.Fatalf("checked-in man page %q is out of date", name)
		}
	}
}

func readManPageTree(t testing.TB, root string) map[string]string {
	t.Helper()

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read man page directory %q: %v", root, err)
	}

	files := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatalf("read man page %q: %v", entry.Name(), err)
		}
		files[entry.Name()] = string(content)
	}

	return files
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
