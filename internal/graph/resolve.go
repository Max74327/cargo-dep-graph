package graph

import (
	"fmt"
	"strings"
)

// Registry is the interface for fetching dependencies of a package.
type Registry interface {
	FetchDependencies(name string) ([]string, error)
}

// Resolve builds a dependency graph by recursively fetching dependencies
// from the registry, starting at root. It uses an iterative DFS with an
// explicit stack, handling cycles and respecting MaxDepth and FilterSubstring.
func Resolve(root string, opts BuildOptions, reg Registry) (*Graph, error) {
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

		// Fetch dependencies from the registry.
		deps, err := reg.FetchDependencies(cur.name)
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", cur.name, err)
		}

		for _, dep := range deps {
			g.AddEdge(cur.name, dep)
			if !visited[dep] {
				stack = append(stack, frame{name: dep, depth: cur.depth + 1})
			}
		}
	}

	return g, nil
}
