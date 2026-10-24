// Package graph models a dependency graph of Cargo packages.
package graph

import "sort"

// Node represents a single package in the dependency graph.
type Node struct {
	// Name is always equal to the key under which the node is stored.
	Name string
	// Dependencies lists direct dependency names (may include hidden nodes).
	Dependencies []string
	// hidden marks the node as filtered out during analysis.
	hidden bool
}

// Graph is a directed graph of package dependencies.
type Graph struct {
	nodes map[string]*Node
}

// New creates an empty graph.
func New() *Graph {
	return &Graph{nodes: make(map[string]*Node)}
}

// AddNode adds a node with the given name if it does not exist yet.
func (g *Graph) AddNode(name string) *Node {
	if n, ok := g.nodes[name]; ok {
		return n
	}
	n := &Node{Name: name}
	g.nodes[name] = n
	return n
}

// AddEdge adds a dependency edge from -> to.
// Both nodes are created if they do not exist.
func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	g.nodes[from].Dependencies = append(g.nodes[from].Dependencies, to)
}

// Node returns the node with the given name, or nil if absent.
func (g *Graph) Node(name string) *Node {
	return g.nodes[name]
}

// HasNode reports whether the graph contains a node with the given name.
func (g *Graph) HasNode(name string) bool {
	_, ok := g.nodes[name]
	return ok
}

// MarkHidden marks the named node as hidden.
// The node is created if it does not exist yet.
func (g *Graph) MarkHidden(name string) {
	g.AddNode(name).hidden = true
}

// IsHidden reports whether the named node is marked as hidden.
func (g *Graph) IsHidden(name string) bool {
	n, ok := g.nodes[name]
	return ok && n.hidden
}

// HiddenNodes returns all hidden node names sorted alphabetically.
func (g *Graph) HiddenNodes() []string {
	names := make([]string, 0, len(g.nodes))
	for name, n := range g.nodes {
		if n.hidden {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// Nodes returns all visible node names sorted alphabetically.
func (g *Graph) Nodes() []string {
	names := make([]string, 0, len(g.nodes))
	for name, n := range g.nodes {
		if !n.hidden {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// Len returns the number of visible nodes in the graph.
func (g *Graph) Len() int {
	count := 0
	for _, n := range g.nodes {
		if !n.hidden {
			count++
		}
	}
	return count
}

// Reverse returns a new graph with all edges reversed.
// Hidden flags are preserved.
func (g *Graph) Reverse() *Graph {
	rev := New()
	for name, n := range g.nodes {
		rev.AddNode(name)
		if n.hidden {
			rev.MarkHidden(name)
		}
		for _, dep := range n.Dependencies {
			rev.AddEdge(dep, name)
		}
	}
	return rev
}

// FromMap builds a complete graph from a dependency map.
// Every key and every value becomes a node; every pair becomes an edge.
func FromMap(deps map[string][]string) *Graph {
	g := New()
	for from, tos := range deps {
		g.AddNode(from)
		for _, to := range tos {
			g.AddEdge(from, to)
		}
	}
	return g
}
