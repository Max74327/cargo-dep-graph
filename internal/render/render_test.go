package render_test

import (
	"strings"
	"testing"

	"cargo-dep-graph/internal/graph"
	"cargo-dep-graph/internal/render"
)

// buildTestGraph creates a small graph: A -> B, A -> C, B -> D, C -> D.
func buildTestGraph() *graph.Graph {
	g := graph.New()
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")
	g.AddEdge("B", "D")
	g.AddEdge("C", "D")
	return g
}

// --- Mermaid tests ---

func TestMermaidSimple(t *testing.T) {
	g := buildTestGraph()
	got := render.Mermaid(g, "A")

	if !strings.HasPrefix(got, "flowchart TD\n") {
		t.Errorf("output must start with 'flowchart TD', got %q", got[:20])
	}
	// Root must be declared first.
	if !strings.HasPrefix(got, "flowchart TD\n    A\n") {
		t.Errorf("root A must be declared first, got:\n%s", got)
	}
	for _, want := range []string{
		"    A --> B\n",
		"    A --> C\n",
		"    B --> D\n",
		"    C --> D\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, got)
		}
	}
}

func TestMermaidLeafDeclaration(t *testing.T) {
	g := graph.New()
	g.AddNode("Leaf")
	got := render.Mermaid(g, "Leaf")

	if !strings.Contains(got, "    Leaf\n") {
		t.Errorf("leaf must be declared explicitly, got:\n%s", got)
	}
}

func TestMermaidSkipsHidden(t *testing.T) {
	g := graph.New()
	g.AddEdge("A", "B")
	g.AddEdge("B", "C")
	g.MarkHidden("B")

	got := render.Mermaid(g, "A")
	if strings.Contains(got, "B") {
		t.Errorf("hidden node B must not appear, got:\n%s", got)
	}
	if !strings.Contains(got, "    A\n") {
		t.Errorf("A must be declared as a leaf (its only dep B is hidden), got:\n%s", got)
	}
}

func TestMermaidDeterministic(t *testing.T) {
	g := graph.New()
	g.AddEdge("Z", "A")
	g.AddEdge("Z", "M")
	got := render.Mermaid(g, "Z")

	// Edges are sorted by dep name, so A comes before M.
	aIdx := strings.Index(got, "Z --> A")
	mIdx := strings.Index(got, "Z --> M")
	if aIdx < 0 || mIdx < 0 || aIdx > mIdx {
		t.Errorf("edges must be sorted (A before M), got:\n%s", got)
	}
}

// --- ASCIITree tests (без изменений) ---

func TestASCIITreeSimple(t *testing.T) {
	g := buildTestGraph()
	got := render.ASCIITree(g, "A")

	want := "" +
		"A\n" +
		"├── B\n" +
		"│   └── D\n" +
		"└── C\n" +
		"    └── D\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestASCIITreeCycle(t *testing.T) {
	g := graph.New()
	g.AddEdge("A", "B")
	g.AddEdge("B", "A")
	got := render.ASCIITree(g, "A")

	if !strings.Contains(got, "(cycle)") {
		t.Errorf("cycle must be marked, got:\n%s", got)
	}
	if strings.Count(got, "A") > 2 {
		t.Errorf("A appears too many times (infinite loop?), got:\n%s", got)
	}
}

func TestASCIITreeSkipsHidden(t *testing.T) {
	g := graph.New()
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")
	g.MarkHidden("B")
	got := render.ASCIITree(g, "A")

	if strings.Contains(got, "B") {
		t.Errorf("hidden B must not appear, got:\n%s", got)
	}
	if !strings.Contains(got, "C") {
		t.Errorf("visible C must appear, got:\n%s", got)
	}
}

func TestASCIITreeHiddenRoot(t *testing.T) {
	g := graph.New()
	g.AddEdge("A", "B")
	g.MarkHidden("A")
	if got := render.ASCIITree(g, "A"); got != "" {
		t.Errorf("hidden root must yield empty output, got %q", got)
	}
}

func TestASCIITreeMissingRoot(t *testing.T) {
	g := graph.New()
	if got := render.ASCIITree(g, "X"); got != "" {
		t.Errorf("missing root must yield empty output, got %q", got)
	}
}
