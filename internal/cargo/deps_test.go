package cargo_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cargo-dep-graph/internal/cargo"
	"cargo-dep-graph/internal/toml"
)

// loadFixture reads a TOML fixture and parses it.
func loadFixture(t *testing.T, name string) *toml.Document {
	t.Helper()
	path := filepath.Join("..", "..", "tests", "fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	doc, err := toml.Parse(string(data))
	if err != nil {
		t.Fatalf("parse fixture %s: %v", name, err)
	}
	return doc
}

func TestExtractDirectDependenciesSimple(t *testing.T) {
	doc := loadFixture(t, "simple.toml")
	deps, err := cargo.ExtractDirectDependencies(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 3 {
		t.Fatalf("len(deps) = %d, want 3", len(deps))
	}
	want := []string{"anyhow", "serde", "tokio"}
	for i, w := range want {
		if deps[i].Name != w {
			t.Errorf("deps[%d].Name = %q, want %q", i, deps[i].Name, w)
		}
	}
}

func TestExtractDirectDependenciesInlineTable(t *testing.T) {
	doc := loadFixture(t, "inline_table.toml")
	deps, err := cargo.ExtractDirectDependencies(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("len(deps) = %d, want 1", len(deps))
	}
	if deps[0].Name != "reqwest" {
		t.Errorf("deps[0].Name = %q, want %q", deps[0].Name, "reqwest")
	}
}

func TestExtractDirectDependenciesNoSection(t *testing.T) {
	doc := loadFixture(t, "no_deps.toml")
	_, err := cargo.ExtractDirectDependencies(doc)
	if !errors.Is(err, cargo.ErrNoDependencies) {
		t.Fatalf("err = %v, want ErrNoDependencies", err)
	}
}

func TestExtractDirectDependenciesMalformed(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "fixtures", "malformed.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if _, err := toml.Parse(string(data)); err == nil {
		t.Fatal("expected Parse to fail on malformed TOML")
	}
}

func TestExtractDirectDependenciesEmptySection(t *testing.T) {
	src := `
[package]
name = "empty"

[dependencies]
`
	doc, err := toml.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	deps, err := cargo.ExtractDirectDependencies(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("len(deps) = %d, want 0", len(deps))
	}
}

func TestExtractDirectDependenciesSorted(t *testing.T) {
	src := `
[dependencies]
zeta = "1"
alpha = "1"
mu = "1"
`
	doc, err := toml.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	deps, err := cargo.ExtractDirectDependencies(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"alpha", "mu", "zeta"}
	for i, w := range want {
		if deps[i].Name != w {
			t.Errorf("deps[%d].Name = %q, want %q", i, deps[i].Name, w)
		}
	}
}
