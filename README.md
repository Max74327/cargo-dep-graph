# cargo-dep-graph

CLI tool for visualizing Cargo dependency graphs.

The tool does **not** use package managers or third-party libraries to obtain
dependency information. `Cargo.toml` is fetched and parsed by the tool itself.

## Status

**Stage 2 — data collection.** Implemented:

- **Stage 1:** Loading and validation of a YAML configuration file, printing
  all user-configurable parameters as `key = value`, error handling.
- **Stage 2:** Fetching `Cargo.toml` from a GitHub repository URL or a local
  file, parsing TOML with a minimal in-repo parser, extracting direct
  dependencies from the `[dependencies]` section, and printing them.

## Requirements

- Go 1.26+
- (optional) Nix, for reproducible builds
- (optional) `golangci-lint`, for linting

## Build

With Nix (recommended, pins Go 1.26 and `golangci-lint`):

```bash
nix develop
make build
```

Without Nix:

```bash
make build
```

The binary is written to `bin/cargo-dep-graph`.

## Run

```bash
./run.sh                 # uses config.yaml by default
./run.sh custom.yaml     # or an explicit config path
```

`run.sh` uses `nix develop` when Nix is available, and falls back to a local
Go toolchain otherwise.

## Test

```bash
make test
```

## Lint

```bash
make lint
```

## Configuration

All parameters are read from a YAML file (`config.yaml` by default).

| Parameter          | Type   | Description                                                |
| ------------------ | ------ | ---------------------------------------------------------- |
| `package_name`     | string | Name of the Cargo package to analyze                       |
| `repository_url`   | string | Repository URL (used when `test_mode` is `false`)          |
| `test_mode`        | bool   | Read from `test_repo_path` instead of `repository_url`     |
| `test_repo_path`   | string | Path to a local test repository file                       |
| `ascii_tree`       | bool   | Render dependencies as an ASCII tree (stage 5)             |
| `filter_substring` | string | Skip packages whose name contains this substring (stage 3) |

Example:

```yaml
package_name: serde_json
repository_url: https://github.com/serde-rs/json
test_mode: false
test_repo_path: ""
ascii_tree: false
filter_substring: ""
```

### Data Sources

- If `test_mode` is `false`, the tool fetches `Cargo.toml` from `repository_url`.
  GitHub URLs are normalized to `raw.githubusercontent.com` automatically, so
  `https://github.com/serde-rs/json` is fetched as
  `https://raw.githubusercontent.com/serde-rs/json/HEAD/Cargo.toml`.
- If `test_mode` is `true`, the tool reads `Cargo.toml` from `test_repo_path`
  (a local file path or a `file://` URL).

## Error handling

Every configuration parameter and every stage of the pipeline is validated.
Errors are reported with a context that shows where the failure occurred:

```
error: load config (config.yaml): configuration file not found: config.yaml
error: load config (config.yaml): package_name must not be empty
error: fetch manifest (https://github.com/user/repo): unexpected HTTP status: 404 for ...
error: parse manifest: toml: line 3: expected key = value
```

If a `Cargo.toml` has no `[dependencies]` section (for example, a workspace
manifest or a leaf package), the tool prints `(none)` instead of failing.

## Project layout

```
cmd/cargo-dep-graph/   CLI entry point
internal/config/       YAML loader, minimal parser, validation
internal/fetch/        Cargo.toml fetcher (HTTP + local files)
internal/toml/         Minimal TOML parser
internal/cargo/        [dependencies] extraction
tests/fixtures/        Cargo.toml fixtures
```

## License

MIT
