#!/usr/bin/env bash
# Uninstallation script - removes Spotumn binaries, desktop entries, icons, and configuration files.
set -euo pipefail

CLR_RESET="\033[0m"
CLR_BOLD="\033[1m"
CLR_BLUE="\033[34m"
CLR_GREEN="\033[32m"
CLR_YELLOW="\033[33m"
CLR_RED="\033[31m"
CLR_DIM="\033[2m"

log_info() {
    echo -e "${CLR_BLUE}==>${CLR_RESET} ${CLR_BOLD}$1${CLR_RESET}"
}

log_success() {
    echo -e "${CLR_GREEN}✔${CLR_RESET} $1"
}

log_warn() {
    echo -e "${CLR_YELLOW}▲${CLR_RESET} $1"
}

log_error() {
    echo -e "${CLR_RED}✖${CLR_RESET} $1" >&2
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REAL_SCRIPT_DIR="$(realpath "$SCRIPT_DIR" 2>/dev/null || echo "$SCRIPT_DIR")"

echo -e "${CLR_BLUE}${CLR_BOLD}"
cat << 'EOF'
  ___ _ __   ___ | |_ _   _ _ __ ___  _ __  
 / __| '_ \ / _ \| __| | | | '_ ` _ \| '_ \ 
 \__ \ |_) | (_) | |_| |_| | | | | | | | | |
 |___/ .__/ \___/ \__|\__,_|_| |_| |_|_| |_|
     |_|                                    
EOF
echo -e "${CLR_RESET}${CLR_DIM}   Spotify TUI Client (Uninstaller)${CLR_RESET}\n"

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
            echo "  --all, --purge           Remove binary, configuration, and cache"
            echo "  -y, --yes, -f, --force   Skip confirmation prompt"
            echo "  -h, --help               Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown argument: $arg"
            echo "Run './uninstall.sh --help' for usage."
            exit 1
            ;;
    esac
done

if [ -z "$DELETE_CONFIG" ] || [ -z "$DELETE_CACHE" ]; then
    if [ -t 0 ] && [ "$AUTO_CONFIRM" = false ]; then
        echo -e "${CLR_BOLD}Select uninstallation scope:${CLR_RESET}"
        echo -e "  ${CLR_BLUE}1)${CLR_RESET} ${CLR_BOLD}Keep settings & cache${CLR_RESET}  ${CLR_DIM}(Remove binary only - Default)${CLR_RESET}"
        echo -e "  ${CLR_BLUE}2)${CLR_RESET} ${CLR_BOLD}Keep settings only${CLR_RESET}      ${CLR_DIM}(Remove binary and cache)${CLR_RESET}"
        echo -e "  ${CLR_BLUE}3)${CLR_RESET} ${CLR_BOLD}Complete purge${CLR_RESET}          ${CLR_DIM}(Remove binary, settings, and cache)${CLR_RESET}"
        echo ""
        read -r -p "Enter choice [1/2/3] (default: 1): " choice
        case "$choice" in
            2)
                DELETE_CONFIG=false
                DELETE_CACHE=true
                ;;
            3)
                DELETE_CONFIG=true
                DELETE_CACHE=true
                ;;
            *)
                DELETE_CONFIG=false
                DELETE_CACHE=false
                ;;
        esac
    else
        DELETE_CONFIG=false
        DELETE_CACHE=false
    fi
fi

if pgrep -x "spotumn" >/dev/null 2>&1; then
    log_info "Stopping active spotumn process..."
    pkill -x "spotumn" 2>/dev/null || true
    sleep 0.5
fi

log_info "Removing spotumn binary..."
CANDIDATE_BINS=()
if [ -n "${PREFIX:-}" ]; then
    CANDIDATE_BINS+=("${PREFIX}/bin/spotumn")
fi
CANDIDATE_BINS+=("${HOME}/.local/bin/spotumn")
CANDIDATE_BINS+=("/usr/local/bin/spotumn")

PATH_BIN="$(command -v spotumn 2>/dev/null || true)"
if [ -n "$PATH_BIN" ]; then
    CANDIDATE_BINS+=("$PATH_BIN")
fi

SEEN_PATHS=()
REMOVED_ANY_BIN=false

for bin in "${CANDIDATE_BINS[@]}"; do
    if [ -f "$bin" ] || [ -L "$bin" ]; then
        REAL_BIN="$(realpath "$bin" 2>/dev/null || echo "$bin")"

        if [[ "$REAL_BIN" == "${REAL_SCRIPT_DIR}"/* ]] || [[ "$REAL_BIN" == "$REAL_SCRIPT_DIR" ]]; then
            continue
        fi

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
            sudo rm -f "$bin"
            log_success "Removed binary: $bin (sudo)"
            REMOVED_ANY_BIN=true
        else
            log_error "Permission denied: unable to remove $bin"
        fi
    fi
done

if [ "$REMOVED_ANY_BIN" = false ]; then
    log_info "No installed binary found in system locations"
fi

CONFIG_DIR="${HOME}/.config/spotumn"
if [ "$DELETE_CONFIG" = true ]; then
    if [ -d "$CONFIG_DIR" ]; then
        rm -rf "$CONFIG_DIR"
        log_success "Removed configuration: ${CONFIG_DIR}"
    fi
    if [ -d "${HOME}/.spotumn" ]; then
        rm -rf "${HOME}/.spotumn"
    fi
else
    if [ -d "$CONFIG_DIR" ]; then
        log_info "Preserved configuration: ${CONFIG_DIR}"
    fi
fi

CACHE_DIR="${HOME}/.cache/spotumn"
if [ "$DELETE_CACHE" = true ]; then
    if [ -d "$CACHE_DIR" ]; then
        rm -rf "$CACHE_DIR"
        log_success "Removed cache: ${CACHE_DIR}"
    fi
else
    if [ -d "$CACHE_DIR" ]; then
        log_info "Preserved cache: ${CACHE_DIR}"
    fi
fi

rm -rf /tmp/spotumn* /tmp/art-* 2>/dev/null || true

echo ""
echo -e "${CLR_GREEN}${CLR_BOLD}Spotumn uninstallation complete!${CLR_RESET}\n"
