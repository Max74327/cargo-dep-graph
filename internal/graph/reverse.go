package graph

// FindReverseDependencies returns a graph of packages that depend on root.
//
// It builds a full forward graph from all known edges, reverses it, and then
// runs the same iterative DFS as Build on the reversed graph.
func FindReverseDependencies(root string, opts BuildOptions, deps map[string][]string) *Graph {
	full := FromMap(deps)
	reversed := full.Reverse()

	revDepsFn := func(name string) []string {
		if n := reversed.Node(name); n != nil {
			return n.Dependencies
		}
		return nil
	}
	return Build(root, opts, revDepsFn)
}
