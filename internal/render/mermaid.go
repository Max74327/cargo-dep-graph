package render

import (
	"fmt"
	"strings"

	"cargo-dep-graph/internal/graph"
)

// Mermaid returns a Mermaid flowchart representation of the graph.
// The root node is declared first, then all edges and leaves.
// Hidden nodes and edges leading to them are omitted.
func Mermaid(g *graph.Graph, root string) string {
	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	// Declare root first (if visible) and its edges.
	if g.HasNode(root) && !g.IsHidden(root) {
		fmt.Fprintf(&sb, "    %s\n", root)
		for _, dep := range visibleDeps(g, root) {
			fmt.Fprintf(&sb, "    %s --> %s\n", root, dep)
		}
	}

	// Then edges and leaves for all other visible nodes.
	for _, name := range g.Nodes() {
		if name == root {
			continue
		}
		deps := visibleDeps(g, name)
		if len(deps) == 0 {
			fmt.Fprintf(&sb, "    %s\n", name)
			continue
		}
		for _, dep := range deps {
			fmt.Fprintf(&sb, "    %s --> %s\n", name, dep)
		}
	}

	return sb.String()
}
