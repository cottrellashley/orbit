#!/bin/sh
# Install orbit - role-based launcher for AI environments
# Usage: curl -fsSL https://raw.githubusercontent.com/cottrellashley/orbit/main/install.sh | sh

set -e

REPO="cottrellashley/orbit"
BINARY="orbit"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  darwin|linux) ;;
  *)
    echo "Error: unsupported OS: $OS"
    exit 1
    ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"

# Get latest release tag
echo "Fetching latest release..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "$TAG" ]; then
  echo "Error: could not determine latest release"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

echo "Downloading ${BINARY} ${TAG} for ${OS}/${ARCH}..."
curl -fsSL -o "/tmp/${BINARY}" "$URL"
chmod +x "/tmp/${BINARY}"

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "/tmp/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "/tmp/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed ${BINARY} ${TAG} to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Run 'orbit --help' to get started."
