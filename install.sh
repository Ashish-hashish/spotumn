#!/usr/bin/env bash
# ==============================================================================
# spotumn installer
# Compiles spotumn from source and installs the binary.
# ==============================================================================

set -euo pipefail

# ANSI color codes (Catppuccin Mocha aesthetic)
CLR_RESET="\033[0m"
CLR_BOLD="\033[1m"
CLR_LAVENDER="\033[38;2;180;190;254m"
CLR_GREEN="\033[38;2;166;227;161m"
CLR_PEACH="\033[38;2;250;179;135m"
CLR_RED="\033[38;2;243;139;168m"
CLR_SUBTEXT="\033[38;2;166;173;200m"

log_info() {
    echo -e "${CLR_LAVENDER}${CLR_BOLD}==>${CLR_RESET} ${CLR_BOLD}$1${CLR_RESET}"
}

log_success() {
    echo -e "${CLR_GREEN}${CLR_BOLD}✔${CLR_RESET} $1"
}

log_warn() {
    echo -e "${CLR_PEACH}${CLR_BOLD}▲${CLR_RESET} $1"
}

log_error() {
    echo -e "${CLR_RED}${CLR_BOLD}✖${CLR_RESET} $1" >&2
}

# Determine script directory (root of repository)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${CLR_LAVENDER}${CLR_BOLD}"
cat << 'EOF'
  ___ _ __   ___ | |_ _   _ _ __ ___  _ __  
 / __| '_ \ / _ \| __| | | | '_ ` _ \| '_ \ 
 \__ \ |_) | (_) | |_| |_| | | | | | | | | |
 |___/ .__/ \___/ \__|\__,_|_| |_| |_|_| |_|
     |_|                                    
EOF
echo -e "${CLR_SUBTEXT}   Spotify TUI Client (Compiling from Source)${CLR_RESET}\n"

# 1. Verify Go toolchain
log_info "Checking prerequisites..."
if ! command -v go >/dev/null 2>&1; then
    log_error "Go compiler ('go') is not installed or not in PATH."
    log_error "Please install Go (>= 1.20) from https://go.dev/dl/ and try again."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
log_success "Found Go compiler: ${GO_VERSION}"

# Check librespot availability (optional but recommended)
if command -v librespot >/dev/null 2>&1 || [ -f "/usr/sbin/librespot" ]; then
    log_success "Found librespot daemon on system"
else
    log_warn "librespot not found in PATH or /usr/sbin. (Optional: install librespot for integrated Spotify Connect playback)"
fi

# 2. Determine target install directory
PREFIX="${PREFIX:-}"
if [ -n "$PREFIX" ]; then
    INSTALL_DIR="${PREFIX}/bin"
elif [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="${HOME}/.local/bin"
fi

mkdir -p "$INSTALL_DIR"

# 3. Build from source (strictly compiled, no pre-built binaries)
BUILD_TMP="$(mktemp -d -t spotumn-build-XXXXXX)"
cleanup() {
    rm -rf "$BUILD_TMP"
}
trap cleanup EXIT

log_info "Downloading Go dependencies..."
go mod download

log_info "Compiling spotumn from source..."
TARGET_BIN="${BUILD_TMP}/spotumn"
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$TARGET_BIN" ./cmd/spotumn
log_success "Compilation successful"

# 4. Install binary
log_info "Installing spotumn to ${INSTALL_DIR}..."
install -m 0755 "$TARGET_BIN" "${INSTALL_DIR}/spotumn"
log_success "Installed binary: ${INSTALL_DIR}/spotumn"

# 5. Initialize config directory with secure permissions
CONFIG_DIR="${HOME}/.config/spotumn"
mkdir -p "$CONFIG_DIR"
chmod 0700 "$CONFIG_DIR"

CONFIG_FILE="${CONFIG_DIR}/config.yml"
if [ ! -f "$CONFIG_FILE" ]; then
    cat > "$CONFIG_FILE" << 'EOF'
# spotumn configuration (Optional - works out of the box with public client ID)
# client_id: "[your_client_id]"
# port: 8989
EOF
    chmod 0600 "$CONFIG_FILE"
    log_success "Created config template at ${CONFIG_FILE} (mode 0600)"
fi

echo ""
echo -e "${CLR_GREEN}${CLR_BOLD}Installation Complete!${CLR_RESET}"
echo -e "${CLR_SUBTEXT}──────────────────────────────────────────────────────────────────${CLR_RESET}"

# Verify PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    log_warn "${INSTALL_DIR} is not currently in your \$PATH."
    echo -e "  Add it to your shell configuration (e.g. ~/.bashrc or ~/.zshrc):"
    echo -e "    ${CLR_LAVENDER}export PATH=\"${INSTALL_DIR}:\$PATH\"${CLR_RESET}\n"
fi

echo -e "Ready to use!"
echo -e "  • Works out of the box with public Spotify Connect client ID"
echo -e "  • (Optional) Override client ID: ${CLR_LAVENDER}client_id: \"[your_client_id]\"${CLR_RESET} in ${CONFIG_FILE}"
echo -e "  • Launch ${CLR_LAVENDER}spotumn${CLR_RESET} to start listening!"
echo ""
