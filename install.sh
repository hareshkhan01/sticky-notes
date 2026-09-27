#!/bin/sh

set -eu

REPO="hareshkhan01/sticky-notes"
INSTALL_DIR="${HOME}/.local/bin"

echo "Installing stick..."

# Detect operating system
OS="$(uname -s)"

case "$OS" in
    Linux)
        OS="linux"
        ;;
    Darwin)
        OS="darwin"
        ;;
    *)
        echo "Error: Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Detect CPU architecture
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64 | amd64)
        ARCH="amd64"
        ;;
    aarch64 | arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get the latest release tag from GitHub
echo "Checking the latest release..."

RELEASE_JSON="$(curl -fsSL \
    "https://api.github.com/repos/${REPO}/releases/latest")"

VERSION="$(printf '%s' "$RELEASE_JSON" |
    sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' |
    head -n 1)"

if [ -z "$VERSION" ]; then
    echo "Error: Could not find the latest release."
    exit 1
fi

# Remove the leading v from the version
VERSION="${VERSION#v}"

ARCHIVE="stick_${VERSION}_${OS}_${ARCH}.tar.gz"

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

echo "Version: ${VERSION}"
echo "Platform: ${OS}/${ARCH}"
echo "Downloading: ${ARCHIVE}"

# Create a temporary directory
TMP_DIR="$(mktemp -d)"

trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

# Download the archive
curl -fL "$DOWNLOAD_URL" -o "${TMP_DIR}/${ARCHIVE}"

# Extract the archive
tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"

# Verify the binary exists
if [ ! -f "${TMP_DIR}/stick" ]; then
    echo "Error: The stick binary was not found in the archive."
    exit 1
fi

# Create the installation directory
mkdir -p "$INSTALL_DIR"

# Install the binary
install -m 0755 "${TMP_DIR}/stick" "${INSTALL_DIR}/stick"

echo ""
echo "Successfully installed stick!"
echo ""
echo "Location: ${INSTALL_DIR}/stick"
echo ""

# Check whether the installation directory is in PATH
case ":${PATH}:" in
    *":${INSTALL_DIR}:"*)
        echo "Run 'stick --help' to get started."
        ;;
    *)
        echo "Add the following line to your shell configuration:"
        echo ""
        echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        echo "Then restart your terminal or reload your shell."
        ;;
esac
