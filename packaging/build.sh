#!/usr/bin/env bash
set -euo pipefail

# packaging/build.sh [version]
VER="${1:-0.0.0}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/app"
DIST_DIR="$ROOT_DIR/dist"

mkdir -p "$DIST_DIR"

echo "Building linux binary via go (simple build)..."
cd "$APP_DIR"
GOOS=linux GOARCH=amd64 go build -o "$DIST_DIR/versaDumps-$VER"

echo "Creating tar.gz..."
tar -C "$DIST_DIR" -czf "$DIST_DIR/versaDumps-$VER-linux-amd64.tar.gz" "versaDumps-$VER"

if command -v rpmbuild >/dev/null 2>&1; then
  echo "rpmbuild found: creating RPM..."
  RPM_TOP="$DIST_DIR/rpmbuild"
  mkdir -p "$RPM_TOP"/SOURCES
  cp "$DIST_DIR/versaDumps-$VER-linux-amd64.tar.gz" "$RPM_TOP/SOURCES/"
  rpmbuild -bb "$ROOT_DIR/packaging/fedora/versaDumps.spec" --define "_topdir $RPM_TOP" --define "version $VER"
  echo "RPM created in $RPM_TOP/RPMS"
else
  echo "rpmbuild not found; skipping RPM creation. Place tar.gz into rpmbuild SOURCES and run rpmbuild -ba packaging/fedora/versaDumps.spec" 
fi

echo "Done. Artifacts in $DIST_DIR"
