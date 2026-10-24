package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cargo-dep-graph/internal/config"
)

// writeTemp writes content to a temporary YAML file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

// --- Load: happy path ---

func TestLoadValidConfig(t *testing.T) {
	path := writeTemp(t, `
package_name: serde
repository_url: https://github.com/serde-rs/serde
test_mode: false
test_repo_path: ""
ascii_tree: true
filter_substring: ""
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PackageName != "serde" {
		t.Errorf("PackageName = %q, want %q", cfg.PackageName, "serde")
	}
	if cfg.RepositoryURL != "https://github.com/serde-rs/serde" {
		t.Errorf("RepositoryURL = %q", cfg.RepositoryURL)
	}
	if cfg.TestMode {
		t.Errorf("TestMode = true, want false")
	}
	if !cfg.ASCIITree {
		t.Errorf("ASCIITree = false, want true")
	}
}

func TestLoadTestModeConfig(t *testing.T) {
	path := writeTemp(t, `
package_name: demo
test_mode: true
test_repo_path: tests/fixtures/graph.txt
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.TestMode {
		t.Errorf("TestMode = false, want true")
	}
	if cfg.TestRepoPath != "tests/fixtures/graph.txt" {
		t.Errorf("TestRepoPath = %q", cfg.TestRepoPath)
	}
}

// --- Load: error paths ---

func TestLoadMissingFile(t *testing.T) {
	_, err := config.Load("no-such-file.yaml")
	if !errors.Is(err, config.ErrFileNotFound) {
		t.Fatalf("err = %v, want ErrFileNotFound", err)
	}
}

func TestLoadEmptyPackage(t *testing.T) {
	path := writeTemp(t, "package_name: \"\"\nrepository_url: x\n")
	_, err := config.Load(path)
	if !errors.Is(err, config.ErrEmptyPackage) {
		t.Fatalf("err = %v, want ErrEmptyPackage", err)
	}
}

func TestLoadEmptyRepoWithoutTestMode(t *testing.T) {
	path := writeTemp(t, "package_name: a\ntest_mode: false\n")
	_, err := config.Load(path)
	if !errors.Is(err, config.ErrEmptyRepo) {
		t.Fatalf("err = %v, want ErrEmptyRepo", err)
	}
}

func TestLoadTestModeWithoutPath(t *testing.T) {
	path := writeTemp(t, "package_name: a\ntest_mode: true\n")
	_, err := config.Load(path)
	if !errors.Is(err, config.ErrMissingTestPath) {
		t.Fatalf("err = %v, want ErrMissingTestPath", err)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	path := writeTemp(t, "package_name without colon\n")
	_, err := config.Load(path)
	if !errors.Is(err, config.ErrInvalidYAML) {
		t.Fatalf("err = %v, want ErrInvalidYAML", err)
	}
}

// --- ParseYAML: scalars and comments ---

func TestParseYAMLSimple(t *testing.T) {
	n, err := config.ParseYAML("a: 1\nb: two\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := n.String("a"); got != "1" {
		t.Errorf("a = %q, want %q", got, "1")
	}
	if got := n.String("b"); got != "two" {
		t.Errorf("b = %q, want %q", got, "two")
	}
}

func TestParseYAMLBool(t *testing.T) {
	n, err := config.ParseYAML("flag: true\noff: false\nempty: \"\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !n.Bool("flag") {
		t.Errorf("flag = false, want true")
	}
	if n.Bool("off") {
		t.Errorf("off = true, want false")
	}
	if n.Bool("empty") {
		t.Errorf("empty = true, want false")
	}
}

func TestParseYAMLHashInDoubleQuotes(t *testing.T) {
	n, err := config.ParseYAML(`filter_substring: "abvgd/uhgurbg/#dfbd"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := n.String("filter_substring"); got != "abvgd/uhgurbg/#dfbd" {
		t.Errorf("got %q, want %q", got, "abvgd/uhgurbg/#dfbd")
	}
}

func TestParseYAMLHashInSingleQuotes(t *testing.T) {
	n, err := config.ParseYAML(`filter_substring: 'a#b'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := n.String("filter_substring"); got != "a#b" {
		t.Errorf("got %q, want %q", got, "a#b")
	}
}

func TestParseYAMLHashOutsideQuotes(t *testing.T) {
	n, err := config.ParseYAML(`filter_substring: abvgd/uhgurbg/#dfbd`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := n.String("filter_substring"); got != "abvgd/uhgurbg/" {
		t.Errorf("got %q, want %q", got, "abvgd/uhgurbg/")
	}
}

func TestParseYAMLTrailingCommentWithSpaces(t *testing.T) {
	n, err := config.ParseYAML(`filter_substring: value   # comment`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := n.String("filter_substring"); got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}
}

func TestParseYAMLSingleQuoteEscape(t *testing.T) {
	// In YAML, '' inside single quotes is a literal '.
	n, err := config.ParseYAML(`name: 'it''s fine'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := n.String("name"); got != "it's fine" {
		t.Errorf("got %q, want %q", got, "it's fine")
	}
}

func TestParseYAMLRejectsLists(t *testing.T) {
	_, err := config.ParseYAML("items: [a, b]\n")
	if err == nil {
		t.Fatal("expected error for list value")
	}
}

func TestParseYAMLRejectsMaps(t *testing.T) {
	_, err := config.ParseYAML("items: {a: 1}\n")
	if err == nil {
		t.Fatal("expected error for inline map value")
	}
}

// --- AsMap ---

func TestAsMapOrder(t *testing.T) {
	cfg := &config.Config{
		PackageName:     "serde",
		RepositoryURL:   "https://github.com/serde-rs/serde",
		TestMode:        false,
		TestRepoPath:    "",
		ASCIITree:       true,
		FilterSubstring: "",
		MaxDepth:        2,
	}
	got := cfg.AsMap()
	want := [][2]string{
		{"package_name", "serde"},
		{"repository_url", "https://github.com/serde-rs/serde"},
		{"test_mode", "false"},
		{"test_repo_path", ""},
		{"ascii_tree", "true"},
		{"filter_substring", ""},
		{"max_depth", "2"},
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AsMap()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// --- MaxDepth ---

func TestLoadMaxDepth(t *testing.T) {
	path := writeTemp(t, "package_name: a\nrepository_url: x\nmax_depth: 3\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", cfg.MaxDepth)
	}
}

func TestLoadMaxDepthAbsent(t *testing.T) {
	path := writeTemp(t, "package_name: a\nrepository_url: x\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxDepth != 0 {
		t.Errorf("MaxDepth = %d, want 0 (default)", cfg.MaxDepth)
	}
}

func TestLoadInvalidMaxDepth(t *testing.T) {
	path := writeTemp(t, "package_name: a\nrepository_url: x\nmax_depth: -1\n")
	_, err := config.Load(path)
	if !errors.Is(err, config.ErrInvalidMaxDepth) {
		t.Fatalf("err = %v, want ErrInvalidMaxDepth", err)
	}
}

func TestLoadNonNumericMaxDepth(t *testing.T) {
	// Нечисловое значение должно молча замениться на дефолт (0), а не падать.
	path := writeTemp(t, "package_name: a\nrepository_url: x\nmax_depth: abc\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxDepth != 0 {
		t.Errorf("MaxDepth = %d, want 0 (default for invalid)", cfg.MaxDepth)
	}
}

// --- Node.Int ---

func TestNodeInt(t *testing.T) {
	n, err := config.ParseYAML("a: 42\nb: notanumber\nc: \"\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := n.Int("a", -1); got != 42 {
		t.Errorf("a = %d, want 42", got)
	}
	if got := n.Int("b", -1); got != -1 {
		t.Errorf("b = %d, want -1 (def)", got)
	}
	if got := n.Int("missing", -1); got != -1 {
		t.Errorf("missing = %d, want -1 (def)", got)
	}
}
