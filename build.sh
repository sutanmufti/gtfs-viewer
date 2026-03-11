#!/usr/bin/env bash
set -e

# Build the Svelte viewer
cd viewer
npm ci
npm run build
cd ..

# Build Go binaries
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

mkdir -p dist

for PLATFORM in "${PLATFORMS[@]}"; do
  OS="${PLATFORM%/*}"
  ARCH="${PLATFORM#*/}"
  OUTPUT="dist/gtfs-viewer-${OS}-${ARCH}"
  if [ "$OS" = "windows" ]; then
    OUTPUT="${OUTPUT}.exe"
  fi
  echo "Building $OUTPUT..."
  GOOS=$OS GOARCH=$ARCH go build -o "$OUTPUT" .
done

echo "Done. Binaries in ./dist"
