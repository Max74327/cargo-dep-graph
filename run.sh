#!/usr/bin/env bash
# Запуск cargo-dep-graph. Если доступен Nix — использует flake-окружение.
set -euo pipefail

MIN_GO_MAJOR=1
MIN_GO_MINOR=23

require_go() {
  if ! command -v go >/dev/null 2>&1; then
    echo "Go не установлен. Используйте 'nix develop' или установите Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+" >&2
    exit 1
  fi

  local version major minor
  version="$(go env GOVERSION 2>/dev/null | sed 's/^go//')"
  major="${version%%.*}"
  minor="${version#*.}"
  minor="${minor%%.*}"

  if ((major < MIN_GO_MAJOR)) || { ((major == MIN_GO_MAJOR)) && ((minor < MIN_GO_MINOR)); }; then
    echo "Нужен Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+, найден go${version}" >&2
    exit 1
  fi
}

if command -v nix >/dev/null 2>&1; then
  exec nix develop --command bash -c '
		set -euo pipefail
		go build -o bin/cargo-dep-graph ./cmd/cargo-dep-graph
		exec ./bin/cargo-dep-graph "${1:-config.yaml}"
	' _ "${1:-config.yaml}"
fi

require_go
go build -o bin/cargo-dep-graph ./cmd/cargo-dep-graph
exec ./bin/cargo-dep-graph "${1:-config.yaml}"
