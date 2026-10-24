#!/usr/bin/env bash
# Run cargo-dep-graph. Uses Nix when available, otherwise falls back to make.
set -euo pipefail

DEFAULT_CONFIG="config.yaml"
CONFIG="${1:-$DEFAULT_CONFIG}"

if command -v nix >/dev/null 2>&1; then
  exec nix develop --command make run CONFIG="$CONFIG"
fi

if ! command -v make >/dev/null 2>&1; then
  echo "make is required. Install make, or use Nix: https://nixos.org/download" >&2
  exit 1
fi

exec make run CONFIG="$CONFIG"
