#!/usr/bin/env bash
# ==============================================================================
# spotumn uninstaller
# Removes spotumn binary, configuration, cache, and login credentials from device.
# Keeps the source code repository completely intact.
# ==============================================================================

set -euo pipefail

# Adaptive ANSI color codes (High contrast & readable in both light and dark terminal modes)
CLR_RESET="\033[0m"
CLR_BOLD="\033[1m"
CLR_LAVENDER="\033[1;34m"
CLR_GREEN="\033[1;32m"
CLR_PEACH="\033[1;33m"
CLR_RED="\033[1;31m"
CLR_SUBTEXT="\033[2m"

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
echo -e "${CLR_RESET}${CLR_SUBTEXT}   Spotify TUI Client (Uninstaller)${CLR_RESET}\n"

# Parse command-line arguments
AUTO_CONFIRM=false
DELETE_CONFIG=""
DELETE_CACHE=""

for arg in "$@"; do
    case "$arg" in
        -y|--yes|-f|--force)
            AUTO_CONFIRM=true
            ;;
        --keep-config)
            DELETE_CONFIG=false
            if [ -z "$DELETE_CACHE" ]; then DELETE_CACHE=true; fi
            ;;
        --keep-cache)
            DELETE_CACHE=false
            if [ -z "$DELETE_CONFIG" ]; then DELETE_CONFIG=true; fi
            ;;
        --keep-both)
            DELETE_CONFIG=false
            DELETE_CACHE=false
            ;;
        --all|--purge)
            DELETE_CONFIG=true
            DELETE_CACHE=true
            ;;
        -h|--help)
            echo "Usage: ./uninstall.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --keep-config            Keep configuration & credentials (~/.config/spotumn)"
            echo "  --keep-cache             Keep cache directory (~/.cache/spotumn)"
            echo "  --keep-both              Keep both configuration and cache"
            echo "  --all, --purge           Remove everything (binary, config, and cache)"
            echo "  -y, --yes, -f, --force   Skip confirmation prompts"
            echo "  -h, --help               Show this help message"
            echo ""
            echo "The source directory (${SCRIPT_DIR}) is always kept intact."
            exit 0
            ;;
        *)
            log_error "Unknown argument: $arg"
            echo "Run './uninstall.sh --help' for usage."
            exit 1
            ;;
    esac
done

# Interactive choice if not specified via CLI flags
if [ -z "$DELETE_CONFIG" ] || [ -z "$DELETE_CACHE" ]; then
    if [ -t 0 ] && [ "$AUTO_CONFIRM" = false ]; then
        echo -e "${CLR_BOLD}Choose what to remove:${CLR_RESET}"
        echo -e "  ${CLR_LAVENDER}${CLR_BOLD}1)${CLR_RESET} ${CLR_BOLD}Remove everything${CLR_RESET} (binary, config, credentials, cache)  ${CLR_SUBTEXT}[Default]${CLR_RESET}"
        echo -e "  ${CLR_LAVENDER}${CLR_BOLD}2)${CLR_RESET} ${CLR_BOLD}Keep config & credentials${CLR_RESET}, remove cache & binary"
        echo -e "  ${CLR_LAVENDER}${CLR_BOLD}3)${CLR_RESET} ${CLR_BOLD}Keep cache${CLR_RESET}, remove config & binary"
        echo -e "  ${CLR_LAVENDER}${CLR_BOLD}4)${CLR_RESET} ${CLR_BOLD}Keep both config & cache${CLR_RESET}, remove binary only"
        echo ""
        read -r -p "Select option [1-4] (default: 1): " sel
        case "$sel" in
            2)
                DELETE_CONFIG=false
                DELETE_CACHE=true
                ;;
            3)
                DELETE_CONFIG=true
                DELETE_CACHE=false
                ;;
            4)
                DELETE_CONFIG=false
                DELETE_CACHE=false
                ;;
            *)
                DELETE_CONFIG=true
                DELETE_CACHE=true
                ;;
        esac
    else
        DELETE_CONFIG=true
        DELETE_CACHE=true
    fi
fi

# Confirmation prompt
if [ "$AUTO_CONFIRM" = false ]; then
    if [ -t 0 ]; then
        echo ""
        echo -e "${CLR_PEACH}${CLR_BOLD}Summary of actions:${CLR_RESET}"
        echo -e "  • ${CLR_BOLD}spotumn binary${CLR_RESET}: Remove"
        if [ "$DELETE_CONFIG" = true ]; then
            echo -e "  • ${CLR_BOLD}Config & credentials${CLR_RESET} (~/.config/spotumn): ${CLR_RED}Delete${CLR_RESET}"
        else
            echo -e "  • ${CLR_BOLD}Config & credentials${CLR_RESET} (~/.config/spotumn): ${CLR_GREEN}Keep${CLR_RESET}"
        fi
        if [ "$DELETE_CACHE" = true ]; then
            echo -e "  • ${CLR_BOLD}Cache${CLR_RESET} (~/.cache/spotumn): ${CLR_RED}Delete${CLR_RESET}"
        else
            echo -e "  • ${CLR_BOLD}Cache${CLR_RESET} (~/.cache/spotumn): ${CLR_GREEN}Keep${CLR_RESET}"
        fi
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

# 3. Remove configuration, credentials, state, and caches conditionally
log_info "Cleaning data traces..."

# Main config directory (~/.config/spotumn)
SPOTUMN_CONFIG_DIR="${HOME}/.config/spotumn"
if [ "$DELETE_CONFIG" = true ]; then
    if [ -d "$SPOTUMN_CONFIG_DIR" ]; then
        rm -rf "$SPOTUMN_CONFIG_DIR"
        log_success "Removed configuration & credentials: ${SPOTUMN_CONFIG_DIR}"
    fi
    SPOTUMN_LEGACY_DIR="${HOME}/.spotumn"
    if [ -d "$SPOTUMN_LEGACY_DIR" ]; then
        rm -rf "$SPOTUMN_LEGACY_DIR"
        log_success "Removed legacy directory: ${SPOTUMN_LEGACY_DIR}"
    fi
else
    if [ -d "$SPOTUMN_CONFIG_DIR" ]; then
        log_info "Preserved configuration & credentials: ${SPOTUMN_CONFIG_DIR}"
    fi
fi

# Cache directory (~/.cache/spotumn)
SPOTUMN_CACHE_DIR="${HOME}/.cache/spotumn"
if [ "$DELETE_CACHE" = true ]; then
    if [ -d "$SPOTUMN_CACHE_DIR" ]; then
        rm -rf "$SPOTUMN_CACHE_DIR"
        log_success "Removed cache directory: ${SPOTUMN_CACHE_DIR}"
    fi
else
    if [ -d "$SPOTUMN_CACHE_DIR" ]; then
        log_info "Preserved cache directory: ${SPOTUMN_CACHE_DIR}"
    fi
fi

# Clean any temporary files (/tmp/spotumn* /tmp/art-*)
rm -rf /tmp/spotumn* /tmp/art-* 2>/dev/null || true

echo ""
echo -e "${CLR_GREEN}${CLR_BOLD}Uninstallation Complete!${CLR_RESET}"
echo -e "${CLR_SUBTEXT}──────────────────────────────────────────────────────────────────${CLR_RESET}"
if [ "$DELETE_CONFIG" = true ] && [ "$DELETE_CACHE" = true ]; then
    echo -e "All spotumn traces, login credentials, and cached files have been removed."
else
    echo -e "Binary removed. Preserved files remain accessible."
fi
echo -e "Your source folder (${CLR_LAVENDER}${SCRIPT_DIR}${CLR_RESET}) remains intact."
echo ""
