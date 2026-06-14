#!/bin/bash
set -e

REPO="neko233-com/bytemsg233"
BINARY="bytemsg233"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case $OS in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

VERSION=${1:-latest}
if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
fi

URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY}_${OS}_${ARCH}.tar.gz"

echo "Downloading $BINARY $VERSION for $OS/$ARCH..."
curl -fsSL "$URL" | tar -xz -C /tmp

echo "Installing to $INSTALL_DIR..."
sudo mv /tmp/$BINARY $INSTALL_DIR/
sudo chmod +x $INSTALL_DIR/$BINARY

echo "$BINARY $VERSION installed successfully!"
