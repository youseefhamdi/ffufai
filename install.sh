#!/bin/bash

# ffufai installation script
# This script builds and installs ffufai with proper dependencies

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${HOME}/.local/bin"
VERSION="1.0.0"

echo "🔧 ffufai v${VERSION} Installation Script"
echo "=========================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    echo "   Visit: https://golang.org/doc/install"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
MIN_VERSION="1.21"

if ! printf '%s\n%s' "${MIN_VERSION}" "${GO_VERSION}" | sort -C -V; then
    echo "❌ Go version ${GO_VERSION} is too old. Please install Go ${MIN_VERSION} or later."
    exit 1
fi

echo "✅ Go version ${GO_VERSION} detected"

# Check if ffuf is installed
if ! command -v ffuf &> /dev/null; then
    echo "⚠️  ffuf is not installed. Installing..."
    go install github.com/ffuf/ffuf@latest
    echo "✅ ffuf installed successfully"
else
    echo "✅ ffuf already installed"
fi

# Create installation directory if it doesn't exist
mkdir -p "${INSTALL_DIR}"

# Build ffufai
echo "🔨 Building ffufai..."
cd "${SCRIPT_DIR}"

# Choose the source file (improved version by default)
if [ -f "ffufai-improved.go" ]; then
    SOURCE_FILE="ffufai-improved.go"
elif [ -f "ffufai.go" ]; then
    SOURCE_FILE="ffufai.go"
else
    echo "❌ No source file found (ffufai.go or ffufai-improved.go)"
    exit 1
fi

# Build the binary
go build -ldflags "-X main.Version=${VERSION}" -o "${INSTALL_DIR}/ffufai" "${SOURCE_FILE}"

# Make it executable
chmod +x "${INSTALL_DIR}/ffufai"

echo "✅ ffufai built successfully"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    echo "⚠️  ${INSTALL_DIR} is not in your PATH"
    echo "   Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "   export PATH=\"\$PATH:${INSTALL_DIR}\""
    echo ""
    echo "   Or run ffufai with full path: ${INSTALL_DIR}/ffufai"
fi

# Check for API key
if [ -z "${PERPLEXITY_API_KEY}" ]; then
    echo ""
    echo "⚠️  PERPLEXITY_API_KEY environment variable is not set"
    echo "   1. Get your API key from: https://www.perplexity.ai/settings/api"
    echo "   2. Set the environment variable:"
    echo "      export PERPLEXITY_API_KEY=\"your_api_key_here\""
    echo "   3. Add it to your shell profile for persistence"
fi

echo ""
echo "🎉 Installation complete!"
echo ""
echo "Usage examples:"
echo "  ffufai -u https://example.com/FUZZ -w /path/to/wordlist.txt"
echo "  ffufai --help"
echo "  ffufai --version"
echo ""
echo "For more information, see README.md"