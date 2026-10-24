// Command cargo-dep-graph is a CLI tool for visualizing Cargo dependency graphs.
package main

import (
	"errors"
	"fmt"
	"os"

	"cargo-dep-graph/internal/cargo"
	"cargo-dep-graph/internal/config"
	"cargo-dep-graph/internal/fetch"
	"cargo-dep-graph/internal/graph"
	"cargo-dep-graph/internal/toml"
)

// defaultConfigPath is used when no positional argument is given.
const defaultConfigPath = "config.yaml"

// keyColumnWidth aligns "key = value" output; longest key is filter_substring (16).
const keyColumnWidth = 16

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the CLI logic and returns an error instead of calling os.Exit,
// which keeps it testable.
func run(args []string) error {
	configPath := defaultConfigPath
	if len(args) > 0 {
		configPath = args[0]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config (%s): %w", configPath, err)
	}

	printConfig(cfg)

	g, err := buildGraph(cfg)
	if err != nil {
		return err
	}

	printGraph(cfg.PackageName, g)
	return nil
}

// buildGraph creates the dependency graph based on the configuration.
func buildGraph(cfg *config.Config) (*graph.Graph, error) {
	opts := graph.BuildOptions{
		MaxDepth:        cfg.MaxDepth,
		FilterSubstring: cfg.FilterSubstring,
	}

	// Stage 3: test mode uses a local graph description file.
	if cfg.TestMode {
		deps, err := graph.LoadTestRepo(cfg.TestRepoPath)
		if err != nil {
			return nil, fmt.Errorf("load test repo (%s): %w", cfg.TestRepoPath, err)
		}
		depsFn := func(name string) []string { return deps[name] }
		return graph.Build(cfg.PackageName, opts, depsFn), nil
	}

	// Stage 2: fetch Cargo.toml and use its direct dependencies only.
	tomlText, err := fetch.FetchCargoToml(cfg.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest (%s): %w", cfg.RepositoryURL, err)
	}

	doc, err := toml.Parse(tomlText)
	if err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	deps, err := cargo.ExtractDirectDependencies(doc)
	if err != nil && !errors.Is(err, cargo.ErrNoDependencies) {
		return nil, fmt.Errorf("extract dependencies: %w", err)
	}

	// Build a one-level graph from the direct dependencies.
	depsMap := make(map[string][]string)
	for _, d := range deps {
		depsMap[d.Name] = nil
	}
	directNames := make([]string, 0, len(deps))
	for _, d := range deps {
		directNames = append(directNames, d.Name)
	}
	depsMap[cfg.PackageName] = directNames

	depsFn := func(name string) []string { return depsMap[name] }
	return graph.Build(cfg.PackageName, opts, depsFn), nil
}

// printConfig writes every user-configurable parameter as "key = value".
func printConfig(cfg *config.Config) {
	fmt.Println("=== Configuration parameters ===")
	for _, kv := range cfg.AsMap() {
		fmt.Printf("%-*s = %s\n", keyColumnWidth, kv[0], kv[1])
	}
}

// printGraph writes the dependency graph in a simple textual form.
func printGraph(pkg string, g *graph.Graph) {
	fmt.Printf("\n=== Dependency graph of %s ===\n", pkg)

	if g.HasNode(pkg) && !g.IsHidden(pkg) {
		printNode(pkg, g)
	}

	for _, name := range g.Nodes() {
		if name == pkg {
			continue
		}
		printNode(name, g)
	}
}

func printNode(name string, g *graph.Graph) {
	deps := g.Node(name).Dependencies
	if len(deps) == 0 {
		fmt.Printf("- %s\n", name)
		return
	}
	fmt.Printf("- %s -> %v\n", name, deps)
}
