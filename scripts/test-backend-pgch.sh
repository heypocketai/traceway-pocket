#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
IMAGE_NAME="traceway-test-pgch"

echo "Building test image..."
docker build -f "$REPO_ROOT/Dockerfile.test-pgch" -t "$IMAGE_NAME" "$REPO_ROOT"

echo ""
echo "Running pgch tests..."
docker run --rm --privileged "$IMAGE_NAME"
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo ""
    echo "All pgch tests passed."
else
    echo ""
    echo "pgch tests FAILED (exit code $EXIT_CODE)."
    exit $EXIT_CODE
fi
