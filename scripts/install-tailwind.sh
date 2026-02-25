#!/usr/bin/env bash
# Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
# Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
# All rights reserved. See LICENSE for details.
# install-tailwind.sh — Download the Tailwind CSS standalone CLI.
#
# The standalone CLI bundles the Tailwind compiler + all first-party plugins
# (@tailwindcss/forms, @tailwindcss/typography) in a single binary — no
# Node.js or npm required.
#
# Usage:
#   ./scripts/install-tailwind.sh          # installs to ./bin/tailwindcss
#   ./scripts/install-tailwind.sh 3.4.19   # specific version

set -euo pipefail

VERSION="${1:-3.4.19}"
INSTALL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/bin"

# Detect platform and architecture.
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
    linux)  PLATFORM="linux" ;;
    darwin) PLATFORM="macos" ;;
    *)      echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH_SUFFIX="x64" ;;
    aarch64|arm64) ARCH_SUFFIX="arm64" ;;
    *)             echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

BINARY="tailwindcss-${PLATFORM}-${ARCH_SUFFIX}"
URL="https://github.com/tailwindlabs/tailwindcss/releases/download/v${VERSION}/${BINARY}"

mkdir -p "$INSTALL_DIR"
echo "==> Downloading Tailwind CSS v${VERSION} (${PLATFORM}/${ARCH_SUFFIX})..."
curl -sLo "$INSTALL_DIR/tailwindcss" "$URL"
chmod +x "$INSTALL_DIR/tailwindcss"

echo "==> Installed: $INSTALL_DIR/tailwindcss"
"$INSTALL_DIR/tailwindcss" --help | head -1
echo ""
echo "Add to PATH:  export PATH=\"$INSTALL_DIR:\$PATH\""
