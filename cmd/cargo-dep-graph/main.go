// Command cargo-dep-graph is a CLI tool for visualizing Cargo dependency graphs.
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"cargo-dep-graph/internal/cargo"
	"cargo-dep-graph/internal/config"
	"cargo-dep-graph/internal/fetch"
	"cargo-dep-graph/internal/graph"
	"cargo-dep-graph/internal/registry"
	"cargo-dep-graph/internal/render"
	"cargo-dep-graph/internal/toml"
)

// defaultConfigPath is used when no positional argument is given.
const defaultConfigPath = "config.serde.yaml"

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

	start := time.Now()
	g, err := buildGraph(cfg)
	elapsed := time.Since(start)
	if err != nil {
		return err
	}

	// Заголовок с информацией о режиме и времени.
	printHeader(cfg, elapsed)

	if cfg.ASCIITree {
		fmt.Print(render.ASCIITree(g, cfg.PackageName))
	} else {
		fmt.Println(render.Mermaid(g, cfg.PackageName))
	}
	return nil
}

// buildGraph creates the dependency graph based on the configuration.
func buildGraph(cfg *config.Config) (*graph.Graph, error) {
	opts := graph.BuildOptions{
		MaxDepth:        cfg.MaxDepth,
		FilterSubstring: cfg.FilterSubstring,
	}

	// Recursive mode (via crates.io) takes priority.
	if cfg.Recursive {
		if cfg.TestMode {
			return nil, fmt.Errorf("recursive mode is not supported in test mode")
		}
		client := registry.NewClient()
		return graph.Resolve(cfg.PackageName, opts, client)
	}

	if cfg.TestMode {
		return buildFromTestRepo(cfg, opts)
	}
	return buildFromCargo(cfg, opts)
}

// buildFromTestRepo builds a graph from a test repository file.
func buildFromTestRepo(cfg *config.Config, opts graph.BuildOptions) (*graph.Graph, error) {
	deps, err := graph.LoadTestRepo(cfg.TestRepoPath)
	if err != nil {
		return nil, fmt.Errorf("load test repo (%s): %w", cfg.TestRepoPath, err)
	}

	if cfg.Reverse {
		return graph.FindReverseDependencies(cfg.PackageName, opts, deps), nil
	}

	depsFn := func(name string) []string { return deps[name] }
	return graph.Build(cfg.PackageName, opts, depsFn), nil
}

// buildFromCargo builds a graph from the direct dependencies of Cargo.toml.
func buildFromCargo(cfg *config.Config, opts graph.BuildOptions) (*graph.Graph, error) {
	tomlText, err := fetch.CargoToml(cfg.RepositoryURL)
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

	directNames := make([]string, 0, len(deps))
	for _, d := range deps {
		directNames = append(directNames, d.Name)
	}
	depsMap := make(map[string][]string)
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

// printHeader writes a short summary of the mode and timing.
func printHeader(cfg *config.Config, elapsed time.Duration) {
	mode := "direct (Cargo.toml)"
	switch {
	case cfg.TestMode:
		mode = "test repo"
	case cfg.Recursive:
		mode = "recursive (crates.io)"
	}

	reverse := ""
	if cfg.Reverse {
		reverse = " [reverse]"
	}

	fmt.Printf("\n=== Result (%s%s) ===\n", mode, reverse)
	if cfg.Recursive {
		fmt.Printf("Resolution time: %s\n", elapsed.Round(time.Millisecond))
	}
	fmt.Println()
}
