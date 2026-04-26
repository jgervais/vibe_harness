#!/bin/sh
set -eu

REPO="jgervais/vibe_harness"
BIN="vibe-harness"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) echo "unsupported arch: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "unsupported os: $OS"; exit 1 ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required"; exit 1
fi

ARCHIVE_URL="https://github.com/$REPO/releases/latest/download/${BIN}_${OS}_${ARCH}.tar.gz"

TMP_DIR="$(mktemp -d)"
curl -fsSL "$ARCHIVE_URL" | tar xz -C "$TMP_DIR" "$BIN" 2>/dev/null || {
  echo "failed to download $ARCHIVE_URL"
  exit 1
}

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
mkdir -p "$INSTALL_DIR"
mv "$TMP_DIR/$BIN" "$INSTALL_DIR/$BIN"
chmod +x "$INSTALL_DIR/$BIN"
rm -rf "$TMP_DIR"

echo "Installed vibe-harness to $INSTALL_DIR/$BIN"
