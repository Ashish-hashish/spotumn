#!/usr/bin/env bash
# ==============================================================================
# spotumn uninstaller
# Removes spotumn binary, configuration, cache, and login credentials from device.
# Keeps the source code repository completely intact.
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
REAL_SCRIPT_DIR="$(realpath "$SCRIPT_DIR" 2>/dev/null || echo "$SCRIPT_DIR")"

echo -e "${CLR_LAVENDER}${CLR_BOLD}"
cat << 'EOF'
  ___ _ __   ___ | |_ _   _ _ __ ___  _ __  
 / __| '_ \ / _ \| __| | | | '_ ` _ \| '_ \ 
 \__ \ |_) | (_) | |_| |_| | | | | | | | | |
 |___/ .__/ \___/ \__|\__,_|_| |_| |_|_| |_|
     |_|                                    
EOF
echo -e "${CLR_SUBTEXT}   Spotify TUI Client (Uninstaller)${CLR_RESET}\n"

# Parse command-line arguments
AUTO_CONFIRM=false
for arg in "$@"; do
    case "$arg" in
        -y|--yes|-f|--force)
            AUTO_CONFIRM=true
            ;;
        -h|--help)
            echo "Usage: ./uninstall.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -y, --yes, -f, --force   Skip confirmation prompt"
            echo "  -h, --help               Show this help message"
            echo ""
            echo "This script uninstalls the spotumn binary, removes all configuration,"
            echo "clears all cached data, and wipes all saved login credentials."
            echo "The source directory (${SCRIPT_DIR}) is kept intact."
            exit 0
            ;;
        *)
            log_error "Unknown argument: $arg"
            echo "Run './uninstall.sh --help' for usage."
            exit 1
            ;;
    esac
done

# Confirm with user if interactive
if [ "$AUTO_CONFIRM" = false ]; then
    if [ -t 0 ]; then
        echo -e "${CLR_PEACH}${CLR_BOLD}Warning:${CLR_RESET} This will remove:"
        echo -e "  • ${CLR_BOLD}spotumn binary${CLR_RESET} from your system"
        echo -e "  • ${CLR_BOLD}Login credentials & tokens${CLR_RESET} (~/.config/spotumn/credentials.json)"
        echo -e "  • ${CLR_BOLD}Configuration & saved state${CLR_RESET} (~/.config/spotumn)"
        echo -e "  • ${CLR_BOLD}Cached album art & librespot cache${CLR_RESET}"
        echo ""
        echo -e "${CLR_GREEN}Note:${CLR_RESET} Your source repository (${SCRIPT_DIR}) will remain ${CLR_BOLD}untouched${CLR_RESET}."
        echo ""
        read -r -p "Are you sure you want to proceed? [y/N] " response
        case "$response" in
            [yY][eE][sS]|[yY])
                ;;
            *)
                echo -e "\n${CLR_SUBTEXT}Uninstallation cancelled.${CLR_RESET}"
                exit 0
                ;;
        esac
    fi
fi

echo ""
# 1. Stop running processes
log_info "Stopping active spotumn processes..."
if pgrep -x spotumn >/dev/null 2>&1; then
    pkill -x spotumn 2>/dev/null || true
    sleep 0.5
    log_success "Stopped running spotumn process"
else
    echo -e "  • No running spotumn process found"
fi

if pgrep -f "librespot.*--name spotumn" >/dev/null 2>&1; then
    pkill -f "librespot.*--name spotumn" 2>/dev/null || true
    log_success "Stopped spotumn librespot background daemon"
fi

# 2. Remove binary from system install paths
log_info "Removing spotumn binary..."
REMOVED_ANY_BIN=false

# Collect possible binary locations
CANDIDATE_BINS=()
if [ -n "${PREFIX:-}" ]; then
    CANDIDATE_BINS+=("${PREFIX}/bin/spotumn")
fi
CANDIDATE_BINS+=("${HOME}/.local/bin/spotumn")
CANDIDATE_BINS+=("/usr/local/bin/spotumn")

# Add location from PATH if found
PATH_BIN="$(command -v spotumn 2>/dev/null || true)"
if [ -n "$PATH_BIN" ]; then
    CANDIDATE_BINS+=("$PATH_BIN")
fi

# Deduplicate and remove binaries safely
SEEN_PATHS=()
for bin in "${CANDIDATE_BINS[@]}"; do
    if [ -f "$bin" ] || [ -L "$bin" ]; then
        REAL_BIN="$(realpath "$bin" 2>/dev/null || echo "$bin")"
        
        # NEVER delete binary inside source directory
        if [[ "$REAL_BIN" == "${REAL_SCRIPT_DIR}"/* ]] || [[ "$REAL_BIN" == "$REAL_SCRIPT_DIR" ]]; then
            log_warn "Skipping binary inside source folder: $bin"
            continue
        fi

        # Skip if already processed
        ALREADY_SEEN=false
        for seen in "${SEEN_PATHS[@]}"; do
            if [ "$seen" = "$REAL_BIN" ]; then
                ALREADY_SEEN=true
                break
            fi
        done
        if [ "$ALREADY_SEEN" = true ]; then
            continue
        fi
        SEEN_PATHS+=("$REAL_BIN")

        if [ -w "$bin" ] || [ -w "$(dirname "$bin")" ]; then
            rm -f "$bin"
            log_success "Removed binary: $bin"
            REMOVED_ANY_BIN=true
        elif command -v sudo >/dev/null 2>&1; then
            log_info "Elevated permissions required to remove $bin"
            sudo rm -f "$bin"
            log_success "Removed binary with sudo: $bin"
            REMOVED_ANY_BIN=true
        else
            log_error "Permission denied: unable to remove $bin (run with sudo or check permissions)"
        fi
    fi
done

if [ "$REMOVED_ANY_BIN" = false ]; then
    echo -e "  • No installed spotumn binary found in standard system locations"
fi

# 3. Remove configuration, credentials, state, and caches
log_info "Removing configuration, credentials, and data traces..."

# Main config directory (~/.config/spotumn)
SPOTUMN_CONFIG_DIR="${HOME}/.config/spotumn"
if [ -d "$SPOTUMN_CONFIG_DIR" ]; then
    rm -rf "$SPOTUMN_CONFIG_DIR"
    log_success "Removed configuration & credentials: ${SPOTUMN_CONFIG_DIR}"
else
    echo -e "  • ${SPOTUMN_CONFIG_DIR} does not exist"
fi

# Cache directory (~/.cache/spotumn)
SPOTUMN_CACHE_DIR="${HOME}/.cache/spotumn"
if [ -d "$SPOTUMN_CACHE_DIR" ]; then
    rm -rf "$SPOTUMN_CACHE_DIR"
    log_success "Removed cache directory: ${SPOTUMN_CACHE_DIR}"
fi

# Legacy/fallback directory (~/.spotumn)
SPOTUMN_LEGACY_DIR="${HOME}/.spotumn"
if [ -d "$SPOTUMN_LEGACY_DIR" ]; then
    rm -rf "$SPOTUMN_LEGACY_DIR"
    log_success "Removed legacy directory: ${SPOTUMN_LEGACY_DIR}"
fi

# Clean any temporary files (/tmp/spotumn* /tmp/art-*)
rm -rf /tmp/spotumn* /tmp/art-* 2>/dev/null || true

echo ""
echo -e "${CLR_GREEN}${CLR_BOLD}Uninstallation Complete!${CLR_RESET}"
echo -e "${CLR_SUBTEXT}──────────────────────────────────────────────────────────────────${CLR_RESET}"
echo -e "All spotumn traces, login credentials, and config files have been removed."
echo -e "Your source folder (${CLR_LAVENDER}${SCRIPT_DIR}${CLR_RESET}) remains intact."
echo ""
