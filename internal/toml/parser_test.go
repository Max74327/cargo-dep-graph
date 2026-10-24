package toml_test

import (
	"strings"
	"testing"

	"cargo-dep-graph/internal/toml"
)

// --- Basic parsing ---

func TestParseSimpleSection(t *testing.T) {
	doc, err := toml.Parse(`
[package]
name = "serde"
version = "1.0.0"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tbl, ok := doc.GetTable("package")
	if !ok {
		t.Fatal("package table not found")
	}
	if got := tbl["name"].Str; got != "serde" {
		t.Errorf("name = %q, want %q", got, "serde")
	}
	if got := tbl["version"].Str; got != "1.0.0" {
		t.Errorf("version = %q, want %q", got, "1.0.0")
	}
}

func TestParseMultipleSections(t *testing.T) {
	doc, err := toml.Parse(`
[package]
name = "demo"

[dependencies]
serde = "1.0"
tokio = "1.0"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := doc.GetTable("package"); !ok {
		t.Error("package table not found")
	}
	deps, ok := doc.GetTable("dependencies")
	if !ok {
		t.Fatal("dependencies table not found")
	}
	if len(deps) != 2 {
		t.Errorf("len(deps) = %d, want 2", len(deps))
	}
}

// --- Comments ---

func TestParseComments(t *testing.T) {
	doc, err := toml.Parse(`
# top comment
[dependencies]
serde = "1.0" # trailing comment
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deps, _ := doc.GetTable("dependencies")
	if got := deps["serde"].Str; got != "1.0" {
		t.Errorf("serde = %q, want %q", got, "1.0")
	}
}

func TestParseHashInsideString(t *testing.T) {
	doc, err := toml.Parse(`
[dependencies]
color = "red#ff0000"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deps, _ := doc.GetTable("dependencies")
	if got := deps["color"].Str; got != "red#ff0000" {
		t.Errorf("color = %q, want %q", got, "red#ff0000")
	}
}

// --- Inline tables ---

func TestParseInlineTable(t *testing.T) {
	doc, err := toml.Parse(`
[dependencies]
tokio = { version = "1", features = ["full"] }
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deps, _ := doc.GetTable("dependencies")
	v := deps["tokio"]
	if !v.IsTbl {
		t.Fatal("tokio value is not a table")
	}
	if got := v.Table["version"].Str; got != "1" {
		t.Errorf("version = %q, want %q", got, "1")
	}
	features := v.Table["features"]
	if !features.IsArr {
		t.Fatal("features is not an array")
	}
	if len(features.Arr) != 1 || features.Arr[0].Str != "full" {
		t.Errorf("features = %v, want [\"full\"]", features.Arr)
	}
}

// --- Arrays ---

func TestParseArray(t *testing.T) {
	doc, err := toml.Parse(`
[package]
keywords = ["a", "b", "c"]
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tbl, _ := doc.GetTable("package")
	v := tbl["keywords"]
	if !v.IsArr {
		t.Fatal("keywords is not an array")
	}
	if len(v.Arr) != 3 {
		t.Fatalf("len = %d, want 3", len(v.Arr))
	}
	for i, want := range []string{"a", "b", "c"} {
		if v.Arr[i].Str != want {
			t.Errorf("Arr[%d] = %q, want %q", i, v.Arr[i].Str, want)
		}
	}
}

func TestParseEmptyArray(t *testing.T) {
	doc, err := toml.Parse(`[package]
keywords = []
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tbl, _ := doc.GetTable("package")
	v := tbl["keywords"]
	if !v.IsArr || len(v.Arr) != 0 {
		t.Errorf("keywords = %v, want empty array", v)
	}
}

// --- Strings ---

func TestParseSingleQuotedString(t *testing.T) {
	doc, err := toml.Parse(`[package]
name = 'literal\nstring'
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tbl, _ := doc.GetTable("package")
	if got := tbl["name"].Str; got != `literal\nstring` {
		t.Errorf("name = %q, want literal backslash-n", got)
	}
}

func TestParseDoubleQuotedEscape(t *testing.T) {
	doc, err := toml.Parse(`[package]
name = "line\nbreak"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tbl, _ := doc.GetTable("package")
	if got := tbl["name"].Str; got != "line\nbreak" {
		t.Errorf("name = %q, want real newline", got)
	}
}

// --- Error cases ---

func TestParseUnterminatedSection(t *testing.T) {
	_, err := toml.Parse("[dependencies\nserde = \"1.0\"\n")
	if err == nil {
		t.Fatal("expected error for unterminated section")
	}
	if !strings.Contains(err.Error(), "unterminated section") {
		t.Errorf("err = %v, want it to mention unterminated section", err)
	}
}

func TestParseMissingEquals(t *testing.T) {
	_, err := toml.Parse("[dependencies]\nserde\n")
	if err == nil {
		t.Fatal("expected error for missing '='")
	}
}

func TestParseEmptyKey(t *testing.T) {
	_, err := toml.Parse("[dependencies]\n = \"1.0\"\n")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestParseUnterminatedString(t *testing.T) {
	_, err := toml.Parse("[dependencies]\nserde = \"1.0\n")
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
}

func TestParseUnterminatedArray(t *testing.T) {
	_, err := toml.Parse("[dependencies]\nserde = [\"a\"\n")
	if err == nil {
		t.Fatal("expected error for unterminated array")
	}
}

func TestParseUnterminatedInlineTable(t *testing.T) {
	_, err := toml.Parse("[dependencies]\nserde = { version = \"1\"\n")
	if err == nil {
		t.Fatal("expected error for unterminated inline table")
	}
}

// --- Root-level keys (no section) ---

func TestParseRootLevelKey(t *testing.T) {
	doc, err := toml.Parse(`top = "value"
[package]
name = "demo"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	root, ok := doc.GetTable("")
	if !ok {
		t.Fatal("root table not found")
	}
	if got := root["top"].Str; got != "value" {
		t.Errorf("top = %q, want %q", got, "value")
	}
}
