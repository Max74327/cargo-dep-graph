#!/usr/bin/env bash
#
# run.sh - build and run cargo-dep-graph.
#
# Verifies that Zig 0.17 or newer is installed. If it is not,
# prints installation instructions for several package managers
# and exits without trying to build.

set -euo pipefail

readonly REQUIRED_MAJOR=0
readonly REQUIRED_MINOR=17

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Print an error message to stderr and exit with status 1.
die() {
    printf 'run.sh: %s\n' "$*" >&2
    exit 1
}

# Print the installation instructions for every supported package manager.
print_install_help() {
    cat >&2 <<'EOF'
Zig 0.17 or newer is required, but was not found in PATH.

Installation options:

  Arch Linux / Manjaro (AUR)
    yay -S zig-bin          # prebuilt binary
    yay -S zig              # build from source
    paru -S zig-bin         # same, using paru

  From source (any distribution)
    git clone https://github.com/ziglang/zig
    cd zig
    git checkout 0.17.0
    cmake -B build -DCMAKE_BUILD_TYPE=Release
    cmake --build build
    # then add ./build/stage1/bin to PATH

  Prebuilt tarball from ziglang.org
    https://ziglang.org/download/
    Download the archive for your platform, unpack it, and add
    the resulting directory to PATH.

  Note
    Zig 0.17 is a very recent release. At the time of writing,
    most distribution repositories (apt, dnf, pacman, apk, brew)
    ship an older version. Prefer the AUR package on Arch, or the
    official tarball / source build elsewhere.
EOF
}

# ---------------------------------------------------------------------------
# Locate zig
# ---------------------------------------------------------------------------

if ! command -v zig >/dev/null 2>&1; then
    printf 'run.sh: zig not found in PATH.\n\n' >&2
    print_install_help
    exit 1
fi

# ---------------------------------------------------------------------------
# Check version
# ---------------------------------------------------------------------------

# `zig version` prints something like "0.17.0" or "0.17.0-dev.1234+abc".
version_raw="$(zig version)"

# Strip anything after '-' or '+', keep the dotted prefix.
version_core="${version_raw%%-*}"
version_core="${version_core%%+*}"

# Split into major and minor. `IFS=.` makes read stop at the first dot
# for major and take the second field for minor.
IFS='.' read -r major minor _rest <<<"$version_core"

# Guard against a non-numeric version string.
if ! [[ "$major" =~ ^[0-9]+$ && "$minor" =~ ^[0-9]+$ ]]; then
    die "cannot parse zig version: '$version_raw'"
fi

if (( major < REQUIRED_MAJOR )) ||
   (( major == REQUIRED_MAJOR && minor < REQUIRED_MINOR )); then
    printf 'run.sh: zig %s is too old (need %d.%d or newer).\n\n' \
        "$version_raw" "$REQUIRED_MAJOR" "$REQUIRED_MINOR" >&2
    print_install_help
    exit 1
fi

printf 'run.sh: using zig %s\n' "$version_raw"

# ---------------------------------------------------------------------------
# Build and run
# ---------------------------------------------------------------------------

exec zig build run -- "$@"
