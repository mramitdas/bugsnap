#!/bin/sh
# Install bugsnap from GitHub releases.
#   curl -fsSL https://raw.githubusercontent.com/mramitdas/bugsnap/main/scripts/install.sh | sh
# Optional:
#   VERSION=v0.1.1        install a specific tag
#   INSTALL_DIR=~/bin     install somewhere else

set -eu

REPO="mramitdas/bugsnap"
VERSION="${VERSION:-latest}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin | linux) ;;
  *) echo "bugsnap: unsupported OS: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "bugsnap: unsupported architecture: $arch" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | head -1 | sed 's/.*: *"\(.*\)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "bugsnap: could not determine latest version" >&2
    exit 1
  fi
fi

asset="bugsnap_${VERSION#v}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$VERSION/$asset"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "downloading $url"
curl -fsSL "$url" -o "$tmp/bugsnap.tar.gz"
tar -xzf "$tmp/bugsnap.tar.gz" -C "$tmp"

if [ -n "${INSTALL_DIR:-}" ]; then
  dir="$INSTALL_DIR"
elif [ -w /usr/local/bin ]; then
  dir=/usr/local/bin
else
  dir="$HOME/.local/bin"
fi
mkdir -p "$dir"
install -m 755 "$tmp/bugsnap" "$dir/bugsnap"

echo "installed bugsnap $VERSION to $dir/bugsnap"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "note: add $dir to your PATH" ;;
esac
