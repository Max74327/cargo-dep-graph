# cargo-dep-graph

CLI tool for visualizing Cargo dependency graphs.

## Status

**Stage 1 — minimal prototype with configuration.** Implemented:

- loading and validation of a YAML configuration file;
- printing all user-configurable parameters as `key = value`;
- error handling for every parameter.

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
