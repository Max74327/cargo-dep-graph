# cargo-dep-graph

CLI tool in Zig 0.17 that builds and visualizes the dependency graph
of Rust (Cargo) packages. Parses `Cargo.toml` directly, walks the graph
with an iterative DFS, and renders the result as a Mermaid diagram or
an ASCII tree. No package managers or third-party libraries are used
to obtain dependency information.

**Current status:** Stage 1 — minimal prototype with YAML configuration.

## Overview

The tool is developed in five stages:

1. **Stage 1 — minimal prototype with configuration.** Reads a YAML
   configuration file, validates every parameter, and prints the
   parsed values. *(current)*
2. **Stage 2 — data collection.** Extracts direct dependencies from
   `Cargo.toml` fetched from a repository URL.
3. **Stage 3 — graph construction.** Builds the transitive dependency
   graph with an iterative DFS and handles cycles.
4. **Stage 4 — reverse dependencies.** Lists packages that depend on
   a given package.
5. **Stage 5 — visualization.** Renders the graph as Mermaid and as
   an ASCII tree, and compares the output with the standard tooling.

## Functions and Settings

User-configurable parameters are provided through a YAML file
(`config.yaml` by default):

| Parameter | Type | Description |
|---|---|---|
| `package_name` | string | Name of the package to analyze |
| `repository_url` | string | URL of the package repository |
| `test_mode` | bool | Enable test repository mode |
| `test_repo_path` | string | Path to the test repository file |
| `ascii_tree` | bool | Output dependencies as an ASCII tree |
| `filter_substring` | string | Substring used to filter packages |

### Error Handling

| Error | Cause |
|---|---|
| `FileNotFound` | Configuration file not found |
| `FileTooBig` | File exceeds 1 MiB |
| `InvalidYaml` | Line does not contain `:` |
| `InvalidBool` | Value is neither `true` nor `false` |
| `EmptyPackageName` | Empty `package_name` value |
| `InvalidUrl` | Empty `repository_url` |
| `MissingField` | `package_name` is not set |

## Module Layout

| File | Responsibility |
|---|---|
| `src/main.zig` | CLI entry point, argument handling, output |
| `src/config_parser.zig` | YAML parsing, `Config` type, error handling |
| `tests/test_config.zig` | Unit tests for the configuration parser |

## Build and Run

```bash
# Build
zig build

# Run tests
zig build test

# Run with the default configuration
./run.sh

# Run with an explicit configuration path
./run.sh path/to/config.yaml
```

`run.sh` checks that Zig 0.17 or newer is installed and prints
installation instructions if it is not.

## Usage Examples

`config.yaml`:

```yaml
package_name: "serde"
repository_url: "https://github.com/serde-rs/serde"
test_mode: false
test_repo_path: "tests/fixtures/acyclic.txt"
ascii_tree: true
filter_substring: ""
```

Output:

```
=== Configuration parameters ===
package_name     = serde
repository_url   = https://github.com/serde-rs/serde
test_mode        = false
test_repo_path   = tests/fixtures/acyclic.txt
ascii_tree       = true
filter_substring =
```

Example error when the file is missing:

```
$ ./run.sh missing.yaml
Configuration error (missing.yaml): FileNotFound
```

