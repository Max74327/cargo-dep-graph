package graph_test

import (
	"errors"
	"testing"

	"cargo-dep-graph/internal/graph"
)

// fakeRegistry implements graph.Registry for tests.
type fakeRegistry struct {
	deps map[string][]string
	errs map[string]error
}

func (f *fakeRegistry) FetchDependencies(name string) ([]string, error) {
	if err, ok := f.errs[name]; ok {
		return nil, err
	}
	return f.deps[name], nil
}

// --- Basic resolution ---

func TestResolveSimple(t *testing.T) {
	reg := &fakeRegistry{deps: map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
		"D": nil,
	}}
	g, err := graph.Resolve("A", graph.BuildOptions{}, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Len() != 4 {
		t.Fatalf("Len() = %d, want 4", g.Len())
	}
	for _, name := range []string{"A", "B", "C", "D"} {
		if !g.HasNode(name) {
			t.Errorf("%s should be present", name)
		}
	}
}

func TestResolveCycle(t *testing.T) {
	reg := &fakeRegistry{deps: map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"A"},
	}}
	g, err := graph.Resolve("A", graph.BuildOptions{}, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Len() != 3 {
		t.Fatalf("Len() = %d, want 3 (cycle must be broken)", g.Len())
	}
}

func TestResolveMaxDepth(t *testing.T) {
	reg := &fakeRegistry{deps: map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"D"},
		"D": nil,
	}}
	g, err := graph.Resolve("A", graph.BuildOptions{MaxDepth: 2}, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Len() != 3 {
		t.Fatalf("Len() = %d, want 3 (A, B, C)", g.Len())
	}
	if g.HasNode("D") {
		t.Error("D must be cut off by MaxDepth=2")
	}
}

func TestResolveFilter(t *testing.T) {
	reg := &fakeRegistry{deps: map[string][]string{
		"A": {"B", "D"},
		"B": {"D"},
		"D": {"E"},
	}}
	g, err := graph.Resolve("A", graph.BuildOptions{FilterSubstring: "D"}, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.IsHidden("D") {
		t.Error("D should be hidden")
	}
	if g.HasNode("E") {
		t.Error("E must not be traversed through hidden D")
	}
	// Edge to D is preserved.
	found := false
	for _, dep := range g.Node("A").Dependencies {
		if dep == "D" {
			found = true
		}
	}
	if !found {
		t.Error("edge A -> D should be preserved")
	}
}

// --- Error handling ---

func TestResolveError(t *testing.T) {
	reg := &fakeRegistry{
		deps: map[string][]string{
			"A": {"B"},
		},
		errs: map[string]error{
			"B": errors.New("network failure"),
		},
	}
	_, err := graph.Resolve("A", graph.BuildOptions{}, reg)
	if err == nil {
		t.Fatal("expected error when registry fails")
	}
}

func TestResolveRootError(t *testing.T) {
	reg := &fakeRegistry{
		errs: map[string]error{
			"A": errors.New("not found"),
		},
	}
	_, err := graph.Resolve("A", graph.BuildOptions{}, reg)
	if err == nil {
		t.Fatal("expected error when root cannot be resolved")
	}
}
