package graph_test

import (
	"os"
	"path/filepath"
	"testing"

	"cargo-dep-graph/internal/graph"
)

// --- Graph structure tests ---

func TestNewGraphIsEmpty(t *testing.T) {
	g := graph.New()
	if g.Len() != 0 {
		t.Errorf("Len() = %d, want 0", g.Len())
	}
	if len(g.Nodes()) != 0 {
		t.Errorf("Nodes() = %v, want empty", g.Nodes())
	}
}

func TestAddNodeIsIdempotent(t *testing.T) {
	g := graph.New()
	g.AddNode("A")
	g.AddNode("A")
	if g.Len() != 1 {
		t.Errorf("Len() = %d, want 1", g.Len())
	}
}

func TestAddEdgeCreatesNodes(t *testing.T) {
	g := graph.New()
	g.AddEdge("A", "B")
	if !g.HasNode("A") || !g.HasNode("B") {
		t.Error("AddEdge should create both nodes")
	}
	if got := g.Node("A").Dependencies; len(got) != 1 || got[0] != "B" {
		t.Errorf("A.Dependencies = %v, want [B]", got)
	}
}

func TestNodesSorted(t *testing.T) {
	g := graph.New()
	g.AddNode("Z")
	g.AddNode("A")
	g.AddNode("M")
	got := g.Nodes()
	want := []string{"A", "M", "Z"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Nodes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestMarkHidden(t *testing.T) {
	g := graph.New()
	g.AddNode("A")
	g.MarkHidden("B")
	if !g.IsHidden("B") {
		t.Error("B should be hidden")
	}
	if g.IsHidden("A") {
		t.Error("A should not be hidden")
	}
	if g.IsHidden("C") {
		t.Error("C (absent) should not be hidden")
	}
	if got := g.Nodes(); len(got) != 1 || got[0] != "A" {
		t.Errorf("Nodes() = %v, want [A]", got)
	}
	if got := g.HiddenNodes(); len(got) != 1 || got[0] != "B" {
		t.Errorf("HiddenNodes() = %v, want [B]", got)
	}
	if g.Len() != 1 {
		t.Errorf("Len() = %d, want 1 (visible only)", g.Len())
	}
}

// --- Build tests ---

// depsMap returns a depsFn backed by a map, returning nil for unknown keys.
func depsMap(m map[string][]string) func(string) []string {
	return func(name string) []string { return m[name] }
}

func TestBuildAcyclic(t *testing.T) {
	deps := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D", "E"},
		"D": {"F"},
		"E": {"F"},
		"F": nil,
	}
	g := graph.Build("A", graph.BuildOptions{}, depsMap(deps))

	if g.Len() != 6 {
		t.Fatalf("Len() = %d, want 6", g.Len())
	}
	for _, name := range []string{"A", "B", "C", "D", "E", "F"} {
		if !g.HasNode(name) {
			t.Errorf("%s should be present", name)
		}
	}
	if got := g.Node("A").Dependencies; len(got) != 2 {
		t.Errorf("A.Dependencies = %v, want 2 elements", got)
	}
}

func TestBuildCycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"A", "D"},
		"D": nil,
	}
	g := graph.Build("A", graph.BuildOptions{}, depsMap(deps))

	if g.Len() != 4 {
		t.Fatalf("Len() = %d, want 4", g.Len())
	}
	if !g.HasNode("D") {
		t.Error("D should be reachable through the cycle")
	}
	// The edge C -> A must be preserved (graph reflects real deps).
	found := false
	for _, dep := range g.Node("C").Dependencies {
		if dep == "A" {
			found = true
		}
	}
	if !found {
		t.Error("edge C -> A should be preserved")
	}
}

func TestBuildMaxDepth(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"D"},
		"D": {"E"},
		"E": nil,
	}
	g := graph.Build("A", graph.BuildOptions{MaxDepth: 2}, depsMap(deps))

	if g.Len() != 3 {
		t.Fatalf("Len() = %d, want 3 (A, B, C)", g.Len())
	}
	if g.HasNode("D") || g.HasNode("E") {
		t.Error("D and E must be cut off by MaxDepth=2")
	}
}

func TestBuildFilterSubstring(t *testing.T) {
	deps := map[string][]string{
		"A": {"B", "D"},
		"B": {"D"},
		"D": {"E"},
	}
	g := graph.Build("A", graph.BuildOptions{FilterSubstring: "D"}, depsMap(deps))

	if g.HasNode("E") {
		t.Error("E must not be traversed through hidden D")
	}
	if !g.IsHidden("D") {
		t.Error("D should be marked hidden")
	}
	// Edges to D are preserved.
	if got := g.Node("A").Dependencies; len(got) != 2 || got[1] != "D" {
		t.Errorf("A.Dependencies = %v, want [B D]", got)
	}
	if got := g.Nodes(); len(got) != 2 {
		t.Errorf("Nodes() = %v, want 2 visible nodes", got)
	}
}

func TestBuildRootFiltered(t *testing.T) {
	deps := map[string][]string{"D": {"E"}}
	g := graph.Build("D", graph.BuildOptions{FilterSubstring: "D"}, depsMap(deps))

	if !g.IsHidden("D") {
		t.Error("root D should be hidden")
	}
	if g.Len() != 0 {
		t.Errorf("Len() = %d, want 0 visible", g.Len())
	}
}

// --- LoadTestRepo tests ---

func writeTempGraph(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "graph.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp graph: %v", err)
	}
	return path
}

func TestLoadTestRepoSimple(t *testing.T) {
	path := writeTempGraph(t, "A B\nA C\nB D\n")
	deps, err := graph.LoadTestRepo(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := deps["A"]; len(got) != 2 || got[0] != "B" || got[1] != "C" {
		t.Errorf("A = %v, want [B C]", got)
	}
	if _, ok := deps["D"]; !ok {
		t.Error("leaf D must be present in the map")
	}
}

func TestLoadTestRepoTolerantFormat(t *testing.T) {
	// Colons and lowercase must be normalized.
	path := writeTempGraph(t, "a: b\nB:c\n")
	deps, err := graph.LoadTestRepo(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := deps["A"]; len(got) != 1 || got[0] != "B" {
		t.Errorf("A = %v, want [B]", got)
	}
	if got := deps["B"]; len(got) != 1 || got[0] != "C" {
		t.Errorf("B = %v, want [C]", got)
	}
}

func TestLoadTestRepoComments(t *testing.T) {
	path := writeTempGraph(t, "# comment\nA B\n\n# another\n")
	deps, err := graph.LoadTestRepo(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := deps["A"]; !ok {
		t.Error("A missing")
	}
}

func TestLoadTestRepoBadLine(t *testing.T) {
	path := writeTempGraph(t, "A B C\n")
	_, err := graph.LoadTestRepo(path)
	if err == nil {
		t.Fatal("expected error for bad line")
	}
}

func TestLoadTestRepoMissingFile(t *testing.T) {
	_, err := graph.LoadTestRepo("no-such-graph.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadTestRepoIntegration(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "fixtures", "graph_acyclic.txt")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture not found: %v", err)
	}
	deps, err := graph.LoadTestRepo(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	g := graph.Build("A", graph.BuildOptions{}, depsMap(deps))
	if g.Len() != 6 {
		t.Errorf("Len() = %d, want 6 from graph_acyclic.txt", g.Len())
	}
}
