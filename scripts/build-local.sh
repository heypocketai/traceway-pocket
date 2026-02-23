#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$ROOT_DIR/frontend"

unset BILLING_PATH
unset CLOUD_MODE
unset PUBLIC_TURNSTILE_SITE_KEY
unset TRACEWAY_URL

npm install
npm run build

rm -rf "$ROOT_DIR/backend/static/dist"
mkdir -p "$ROOT_DIR/backend/static/dist"
cp -r "$ROOT_DIR/frontend/build/"* "$ROOT_DIR/backend/static/dist/"

echo "Frontend built and bundled into backend/static/dist/"
