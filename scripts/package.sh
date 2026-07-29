#!/usr/bin/env bash
# Builds release binaries and packages each into a tar.gz under dist/
# (gitignored), bundling the binary with LICENSE.
#
# Go doesn't have a separate musl "target" the way Rust does: the same
# source builds either dynamically or fully static depending on
# CGO_ENABLED, which only matters here because net/http's DNS resolver
# switches to glibc's NSS resolver when cgo is enabled.
#   - gnu build:   CGO_ENABLED=1 (needs gcc), dynamically linked to the
#     host's glibc, uses the system resolver (NSS, /etc/hosts, etc.)
#   - musl build:  CGO_ENABLED=0, fully static, pure-Go resolver, runs
#     unmodified on musl distros (Alpine) and scratch containers.
#
# Usage: scripts/package.sh

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

version="$(git describe --tags --abbrev=0)"
binary="weathertui"
targets=(gnu:1 musl:0)

mkdir -p dist

for entry in "${targets[@]}"; do
  suffix="${entry%%:*}"
  cgo="${entry##*:}"
  name="$binary-$version-linux-amd64-$suffix"
  stage="dist/$name"

  echo "==> building $suffix (CGO_ENABLED=$cgo)"
  CGO_ENABLED="$cgo" GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o "dist/$binary" "./cmd/$binary"

  mkdir -p "$stage"
  mv "dist/$binary" "$stage/"
  cp LICENSE "$stage/"
  tar -C dist -czf "dist/$name.tar.gz" "$name"
  rm -rf "$stage"
  echo "==> dist/$name.tar.gz"
done
