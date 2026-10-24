// Command cargo-dep-graph is a CLI tool for visualizing Cargo dependency graphs.
package main

import (
	"errors"
	"fmt"
	"os"

	"cargo-dep-graph/internal/cargo"
	"cargo-dep-graph/internal/config"
	"cargo-dep-graph/internal/fetch"
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

	source := cfg.RepositoryURL
	if cfg.TestMode {
		source = cfg.TestRepoPath
	}

	tomlText, err := fetch.FetchCargoToml(source)
	if err != nil {
		return fmt.Errorf("fetch manifest (%s): %w", source, err)
	}

	doc, err := toml.Parse(tomlText)
	if err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}

	deps, err := cargo.ExtractDirectDependencies(doc)
	if errors.Is(err, cargo.ErrNoDependencies) {
		fmt.Printf("\n=== Direct dependencies of %s ===\n(none)\n", cfg.PackageName)
		return nil
	}
	if err != nil {
		return fmt.Errorf("extract dependencies: %w", err)
	}

	printDependencies(cfg.PackageName, deps)
	return nil
}

// printConfig writes every user-configurable parameter as "key = value".
func printConfig(cfg *config.Config) {
	fmt.Println("=== Configuration parameters ===")
	for _, kv := range cfg.AsMap() {
		fmt.Printf("%-*s = %s\n", keyColumnWidth, kv[0], kv[1])
	}
}

func printDependencies(pkg string, deps []cargo.Dependency) {
	fmt.Printf("\n=== Direct dependencies of %s ===\n", pkg)
	if len(deps) == 0 {
		fmt.Println("(none)")
		return
	}
	for _, d := range deps {
		fmt.Printf("- %s\n", d.Name)
	}
}
