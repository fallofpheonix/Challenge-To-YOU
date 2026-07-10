#!/bin/bash
set -e

# Resolve script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Initializing build pipeline in: $PROJECT_ROOT"

# Create clean distribution directory
DIST_DIR="$PROJECT_ROOT/dist"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Navigate to backend directory to resolve packages
cd "$PROJECT_ROOT/backend"

# Compilation target definitions
targets=(
    "darwin/amd64/server_darwin_amd64"
    "darwin/arm64/server_darwin_arm64"
    "linux/amd64/server_linux_amd64"
    "windows/amd64/server_windows_amd64.exe"
)

echo "Compiling optimized Go backend server binaries..."

for target in "${targets[@]}"; do
    IFS="/" read -r goos goarch binary_name <<< "$target"
    output_path="$DIST_DIR/$binary_name"
    
    echo "  -> Building $goos/$goarch [$binary_name]..."
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
        -ldflags="-s -w" \
        -o "$output_path" \
        cmd/sandbox/main.go
done

# Copy challenges folder into dist to ensure self-contained levels
echo "Packaging static challenge assets..."
cp -R challenges "$DIST_DIR/challenges"

echo "Build pipeline execution complete. Artifacts stored in: $DIST_DIR"
ls -la "$DIST_DIR"
