#!/usr/bin/env bash
set -euo pipefail

MIN_GO_MAJOR=1
MIN_GO_MINOR=26
DEFAULT_CONFIG="config.yaml"
BIN_PATH="bin/cargo-dep-graph"

# Проверяет, что go в PATH есть и его версия >= MIN_GO_MAJOR.MIN_GO_MINOR.
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

# Собирает проект и запускает бинарник с указанным конфигом.
build_and_run() {
  local config="${1:-$DEFAULT_CONFIG}"
  if command -v make >/dev/null 2>&1; then
    make build
  else
    go build -o "$BIN_PATH" ./cmd/cargo-dep-graph
  fi
  exec "$BIN_PATH" "$config"
}

CONFIG="${1:-$DEFAULT_CONFIG}"

if command -v nix >/dev/null 2>&1; then
  # shellcheck disable=SC2016
  exec nix develop --command bash -c '
		set -euo pipefail
		if command -v make >/dev/null 2>&1; then
			make build
		else
			go build -o bin/cargo-dep-graph ./cmd/cargo-dep-graph
		fi
		exec ./bin/cargo-dep-graph "$1"
	' _ "$CONFIG"
fi

require_go
build_and_run "$CONFIG"
