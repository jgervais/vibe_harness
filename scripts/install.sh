#!/usr/bin/env bash
set -euo pipefail

REPO="jgervais/vibe_harness"
INSTALL_PATH="/usr/local/bin/vibe-harness"

detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*)  echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)      echo "unknown" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)             echo "unknown" ;;
    esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
    echo "ERROR: Unsupported platform (OS=$OS, ARCH=$ARCH)"
    exit 1
fi

BINARY="vibe-harness-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY="${BINARY}.exe"
fi

echo "Detected platform: ${OS}/${ARCH}"

TAG="${1:-latest}"
if [ "$TAG" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${BINARY}"
fi

CHECKSUM_URL="${DOWNLOAD_URL}.sha256"

echo "Downloading vibe-harness ${TAG}..."
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$DOWNLOAD_URL" -o "${TMPDIR}/${BINARY}"

echo "Verifying checksum..."
curl -fsSL "$CHECKSUM_URL" -o "${TMPDIR}/checksum.txt" 2>/dev/null || true

if [ -f "${TMPDIR}/checksum.txt" ]; then
    cd "$TMPDIR"
    echo "$(cat checksum.txt)  ${BINARY}" | sha256sum -c --status 2>/dev/null || {
        echo "ERROR: Checksum verification failed"
        exit 1
    }
    echo "Checksum verified."
else
    echo "WARNING: No checksum available, skipping verification"
fi

chmod +x "${TMPDIR}/${BINARY}"

echo "Installing to ${INSTALL_PATH}..."
if [ -w "$(dirname "$INSTALL_PATH")" ]; then
    mv "${TMPDIR}/${BINARY}" "$INSTALL_PATH"
else
    sudo mv "${TMPDIR}/${BINARY}" "$INSTALL_PATH"
fi

echo "vibe-harness installed successfully at ${INSTALL_PATH}"
vibe-harness --version 2>/dev/null || true