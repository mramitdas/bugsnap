#!/usr/bin/env bash
# Records well on camera: a bug is captured, inspected, and replayed in a clean clone.
# Usage: scripts/demo.sh   (requires bugsnap on PATH and Docker optional)
set -eu

BIN="${BUGSNAP:-bugsnap}"
demo=$(mktemp -d)
demo_clean="${demo}-clean"
trap 'rm -rf "$demo" "$demo_clean"' EXIT
cd "$demo"

git init -q
git config user.email demo@example.com
git config user.name demo
printf 'export function total(items) {\n  return items.length;\n}\n' > cart.js
git add -A
git commit -qm "initial"

# introduce the bug
printf 'export function total(items) {\n  return items.reduce((s, i) => s + i.price, 0) + 10;\n}\n' > cart.js
export CART_API_KEY=demo-secret-key

echo "==> bugsnap capture"
$BIN capture --note "totals are off by 10" -- sh -c 'echo "Error: expected 30, got 40"; exit 1'

bundle=$(ls bugsnap-*.zip | head -1)

echo
echo "==> bugsnap inspect"
$BIN inspect "$bundle"

echo
echo "==> bugsnap replay --apply (in a clean clone)"
git clone -q "$demo" "$demo_clean"
cd "$demo_clean"
$BIN replay "$demo/$bundle" --apply
git diff --stat
