// Package graph implements dependency graph construction and traversal.
package graph

import "strings"

// BuildOptions controls how a dependency graph is built.
type BuildOptions struct {
	// MaxDepth limits the depth of traversal. 0 means "no limit".
	MaxDepth int
	// FilterSubstring skips nodes whose name contains this substring.
	// Empty string disables filtering.
	FilterSubstring string
}

// Build walks the dependency tree starting at root and returns a graph
// containing all reachable nodes (respecting options).
//
// depsFn returns the direct dependencies of a given package name.
// It uses iterative DFS with an explicit stack, handles cycles via a
// visited set, skips nodes matching FilterSubstring (marking them hidden),
// and limits traversal depth to MaxDepth.
func Build(root string, opts BuildOptions, depsFn func(string) []string) *Graph {
	g := New()

	type frame struct {
		name  string
		depth int
	}
	stack := []frame{{name: root, depth: 0}}
	visited := make(map[string]bool)

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[cur.name] {
			continue
		}
		visited[cur.name] = true

		if opts.FilterSubstring != "" && strings.Contains(cur.name, opts.FilterSubstring) {
			g.MarkHidden(cur.name)
			continue
		}

		g.AddNode(cur.name)

		if opts.MaxDepth > 0 && cur.depth >= opts.MaxDepth {
			continue
		}

		for _, dep := range depsFn(cur.name) {
			g.AddEdge(cur.name, dep)
			if !visited[dep] {
				stack = append(stack, frame{name: dep, depth: cur.depth + 1})
			}
		}
	}

	return g
}
