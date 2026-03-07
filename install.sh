#!/usr/bin/env bash
set -euo pipefail

REPO="Ooooze/batctl"
BINARY="batctl"
INSTALL_DIR="/usr/bin"
SYSTEMD_DIR="/etc/systemd/system"
UDEV_DIR="/etc/udev/rules.d"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${CYAN}[info]${NC}  $*"; }
ok()    { echo -e "${GREEN}[ok]${NC}    $*"; }
warn()  { echo -e "${YELLOW}[warn]${NC}  $*"; }
error() { echo -e "${RED}[error]${NC} $*" >&2; }
die()   { error "$@"; exit 1; }

usage() {
    cat <<EOF
${BOLD}batctl installer${NC}

Usage: $0 [OPTIONS]

Options:
  --uninstall    Remove batctl and all associated files
  --no-systemd   Skip systemd service and udev rule installation
  --version VER  Install a specific version (e.g. 2025.06.1)
  --help         Show this help

Install:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/master/install.sh | sudo bash

Uninstall:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/master/install.sh | sudo bash -s -- --uninstall
EOF
    exit 0
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        die "This script must be run as root. Use: sudo bash $0"
    fi
}

detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) echo "x86_64" ;;
        aarch64|arm64) echo "aarch64" ;;
        *) die "Unsupported architecture: $arch" ;;
    esac
}

detect_os() {
    local os
    os=$(uname -s)
    case "$os" in
        Linux) ;;
        *) die "batctl only supports Linux (detected: $os)" ;;
    esac
}

get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"tag_name":\s*"v?([^"]+)".*/\1/')

    if [[ -z "$version" ]]; then
        die "Failed to determine latest version. Check your internet connection."
    fi
    echo "$version"
}

download_and_install() {
    local version="$1"
    local arch="$2"
    local tarball="batctl-${version}-linux-${arch}.tar.gz"
    local url="https://github.com/${REPO}/releases/download/v${version}/${tarball}"
    local tmpdir

    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    info "Downloading batctl v${version} for linux/${arch}..."
    if ! curl -fsSL -o "${tmpdir}/${tarball}" "$url"; then
        die "Download failed. Version v${version} may not have a binary for ${arch}.\n       Check: https://github.com/${REPO}/releases"
    fi

    info "Extracting..."
    tar xzf "${tmpdir}/${tarball}" -C "$tmpdir"

    if [[ ! -f "${tmpdir}/${BINARY}" ]]; then
        die "Archive does not contain '${BINARY}' binary"
    fi

    info "Installing binary to ${INSTALL_DIR}/${BINARY}..."
    install -Dm755 "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    ok "Binary installed: ${INSTALL_DIR}/${BINARY}"
}

install_systemd() {
    info "Installing systemd service..."
    cat > "${SYSTEMD_DIR}/batctl.service" <<'SERVICE'
[Unit]
Description=Apply battery charge thresholds (batctl)
After=multi-user.target

[Service]
Type=oneshot
ExecStart=/usr/bin/batctl apply
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
SERVICE
    ok "Service installed: ${SYSTEMD_DIR}/batctl.service"

    info "Installing udev rule..."
    cat > "${UDEV_DIR}/99-batctl-resume.rules" <<'UDEV'
# Restore battery charge thresholds after resume from suspend
ACTION=="change", SUBSYSTEM=="power_supply", ATTR{type}=="Battery", RUN+="/usr/bin/batctl apply"
UDEV
    ok "Udev rule installed: ${UDEV_DIR}/99-batctl-resume.rules"

    if command -v systemctl &>/dev/null; then
        systemctl daemon-reload
        ok "systemd daemon reloaded"
    fi

    if command -v udevadm &>/dev/null; then
        udevadm control --reload-rules 2>/dev/null || true
        ok "udev rules reloaded"
    fi
}

do_uninstall() {
    check_root
    info "Uninstalling batctl..."

    if command -v systemctl &>/dev/null; then
        systemctl disable batctl.service 2>/dev/null || true
        systemctl stop batctl.service 2>/dev/null || true
    fi

    local files=(
        "${INSTALL_DIR}/${BINARY}"
        "${SYSTEMD_DIR}/batctl.service"
        "${UDEV_DIR}/99-batctl-resume.rules"
        "/etc/batctl.conf"
    )

    for f in "${files[@]}"; do
        if [[ -f "$f" ]]; then
            rm -f "$f"
            ok "Removed: $f"
        fi
    done

    if command -v systemctl &>/dev/null; then
        systemctl daemon-reload 2>/dev/null || true
    fi
    if command -v udevadm &>/dev/null; then
        udevadm control --reload-rules 2>/dev/null || true
    fi

    ok "batctl has been uninstalled"
}

do_install() {
    local version="$1"
    local skip_systemd="$2"

    check_root
    detect_os

    local arch
    arch=$(detect_arch)

    if [[ -z "$version" ]]; then
        version=$(get_latest_version)
    fi

    echo ""
    echo -e "${BOLD}  ⚡ batctl installer${NC}"
    echo -e "  Version: ${CYAN}${version}${NC}  Arch: ${CYAN}${arch}${NC}"
    echo ""

    download_and_install "$version" "$arch"

    if [[ "$skip_systemd" == "false" ]]; then
        install_systemd
    else
        warn "Skipping systemd/udev installation (--no-systemd)"
    fi

    echo ""
    ok "${BOLD}batctl v${version} installed successfully!${NC}"
    echo ""
    echo -e "  Get started:"
    echo -e "    ${CYAN}sudo batctl${NC}              # Launch TUI"
    echo -e "    ${CYAN}batctl status${NC}            # Show battery info"
    echo -e "    ${CYAN}sudo batctl set --stop 80${NC} # Set charge limit"
    echo -e "    ${CYAN}sudo batctl persist enable${NC} # Survive reboots"
    echo ""
}

main() {
    local version=""
    local uninstall=false
    local skip_systemd=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --uninstall)  uninstall=true; shift ;;
            --no-systemd) skip_systemd=true; shift ;;
            --version)    version="$2"; shift 2 ;;
            --help|-h)    usage ;;
            *)            die "Unknown option: $1. Use --help for usage." ;;
        esac
    done

    if [[ "$uninstall" == "true" ]]; then
        do_uninstall
    else
        do_install "$version" "$skip_systemd"
    fi
}

main "$@"
