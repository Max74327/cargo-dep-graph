package graph_test

import (
	"path/filepath"
	"testing"

	"cargo-dep-graph/internal/graph"
)

func TestReverse(t *testing.T) {
	g := graph.New()
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")
	g.AddEdge("B", "D")

	rev := g.Reverse()

	// In the reversed graph, D should have dependencies B.
	if !rev.HasNode("D") {
		t.Fatal("D missing in reversed graph")
	}
	if got := rev.Node("D").Dependencies; len(got) != 1 || got[0] != "B" {
		t.Errorf("D.Dependencies = %v, want [B]", got)
	}

	// B should have dependency A.
	if got := rev.Node("B").Dependencies; len(got) != 1 || got[0] != "A" {
		t.Errorf("B.Dependencies = %v, want [A]", got)
	}
}

func TestFindReverseDependenciesSimple(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": nil,
	}
	// Find who depends on C. Expected: B and A.
	g := graph.FindReverseDependencies("C", graph.BuildOptions{}, deps)

	if g.Len() != 3 {
		t.Fatalf("Len() = %d, want 3 (C, B, A)", g.Len())
	}
	if !g.HasNode("C") || !g.HasNode("B") || !g.HasNode("A") {
		t.Errorf("Nodes() = %v, want C, B, A", g.Nodes())
	}
	// C should depend on B in the reverse graph.
	if got := g.Node("C").Dependencies; len(got) != 1 || got[0] != "B" {
		t.Errorf("C.Dependencies = %v, want [B]", got)
	}
}

func TestFindReverseDependenciesMaxDepth(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": nil,
	}
	// Find who depends on C with depth 1. Expected: C and B (no A).
	g := graph.FindReverseDependencies("C", graph.BuildOptions{MaxDepth: 1}, deps)

	if g.HasNode("A") {
		t.Error("A should be cut off by MaxDepth=1")
	}
	if !g.HasNode("C") || !g.HasNode("B") {
		t.Errorf("Nodes() = %v, want C, B", g.Nodes())
	}
}

func TestFindReverseDependenciesFilter(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": nil,
	}
	// Find who depends on C, filtering out B.
	g := graph.FindReverseDependencies("C", graph.BuildOptions{FilterSubstring: "B"}, deps)

	if g.HasNode("A") {
		t.Error("A should not be traversed through hidden B")
	}
	if !g.IsHidden("B") {
		t.Error("B should be marked hidden")
	}
	// C still has edge to B (hidden).
	if got := g.Node("C").Dependencies; len(got) != 1 || got[0] != "B" {
		t.Errorf("C.Dependencies = %v, want [B]", got)
	}
}

func TestFindReverseDependenciesFixture(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "fixtures", "graph_reverse.txt")
	deps, err := graph.LoadTestRepo(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Find who depends on D.
	g := graph.FindReverseDependencies("D", graph.BuildOptions{}, deps)

	// Expected: D (root), E, C, B, A.
	wantNodes := []string{"A", "B", "C", "D", "E"}
	if len(g.Nodes()) != len(wantNodes) {
		t.Fatalf("Nodes() = %v, want %v", g.Nodes(), wantNodes)
	}
	for _, name := range wantNodes {
		if !g.HasNode(name) {
			t.Errorf("missing node %s", name)
		}
	}
}
