// Package render implements visualization formats for the dependency graph.
package render

import (
	"fmt"
	"sort"
	"strings"

	"cargo-dep-graph/internal/graph"
)

// ASCIITree returns an ASCII tree representation of the graph starting at root.
// Hidden nodes are skipped. Cycles are marked with "(cycle)" instead of
// being expanded indefinitely.
func ASCIITree(g *graph.Graph, root string) string {
	if !g.HasNode(root) || g.IsHidden(root) {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(root)
	sb.WriteByte('\n')

	path := map[string]bool{root: true}
	children := visibleDeps(g, root)
	for i, child := range children {
		last := i == len(children)-1
		renderChild(&sb, g, child, "", last, path)
	}
	return sb.String()
}

// visibleDeps returns the sorted list of visible dependencies of name.
func visibleDeps(g *graph.Graph, name string) []string {
	n := g.Node(name)
	if n == nil {
		return nil
	}
	var deps []string
	for _, dep := range n.Dependencies {
		if !g.IsHidden(dep) {
			deps = append(deps, dep)
		}
	}
	sort.Strings(deps)
	return deps
}

// renderChild writes a single subtree using the given prefix and marker.
func renderChild(sb *strings.Builder, g *graph.Graph, name, prefix string, last bool, path map[string]bool) {
	marker := "├── "
	childPrefix := prefix + "│   "
	if last {
		marker = "└── "
		childPrefix = prefix + "    "
	}

	sb.WriteString(prefix)
	sb.WriteString(marker)
	sb.WriteString(name)

	if path[name] {
		sb.WriteString(" (cycle)\n")
		return
	}
	sb.WriteByte('\n')

	path[name] = true
	defer delete(path, name)

	children := visibleDeps(g, name)
	for i, child := range children {
		isLast := i == len(children)-1
		renderChild(sb, g, child, childPrefix, isLast, path)
	}
}

// Example output:
//
// A
// ├── B
// │   └── D
// └── C
//
//	└── D
//
// If the graph has a cycle A -> B -> A:
//
// A
// └── B
//
//	└── A (cycle)
func ExampleASCIITree() {
	g := graph.New()
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")
	g.AddEdge("B", "D")
	g.AddEdge("C", "D")

	fmt.Print(ASCIITree(g, "A"))
	// Output:
	// A
	// ├── B
	// │   └── D
	// └── C
	//     └── D
}
