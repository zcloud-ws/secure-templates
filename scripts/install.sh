#!/usr/bin/env bash
set -e

# Config
REPO="zcloud-ws/secure-templates"
BINARY="secure-templates"
ALIAS_NAME="${STPL_ALIAS_NAME:-stpl}"
INSTALL_DIR="${STPL_INSTALL_DIR:-}"
VERSION="${STPL_VERSION:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Detect OS
detect_os() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$OS" in
    linux) ;;
    darwin) ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) error "Unsupported OS: $OS" ;;
  esac
}

# Detect architecture
detect_arch() {
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
  esac
}

# Get version (use STPL_VERSION if set, otherwise fetch latest)
get_version() {
  if [ -n "$VERSION" ]; then
    # Ensure version starts with 'v'
    case "$VERSION" in
      v*) ;;
      *) VERSION="v${VERSION}" ;;
    esac
    info "Using specified version: ${VERSION}"
  else
    VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | cut -d'"' -f4)
    if [ -z "$VERSION" ]; then
      error "Failed to get latest version. Check your internet connection."
    fi
    info "Latest version: ${VERSION}"
  fi
}

# Determine install directory
get_install_dir() {
  if [ -n "$INSTALL_DIR" ]; then
    : # Custom directory specified
  elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  elif [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    INSTALL_DIR="$HOME/.local/bin"
  else
    INSTALL_DIR="/usr/local/bin"
  fi
}

# Download, verify and install
install_binary() {
  local VERSION_NUM="${VERSION#v}"
  local FILE_NAME="${BINARY}-${VERSION_NUM}-${OS}-${ARCH}"
  [ "$OS" = "windows" ] && FILE_NAME="${FILE_NAME}.exe"

  local DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILE_NAME}"
  local CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}_${VERSION_NUM}_checksums.txt"

  info "Downloading ${BINARY} ${VERSION} for ${OS}/${ARCH}..."

  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  if ! curl -sL "$DOWNLOAD_URL" -o "${TMP_DIR}/${BINARY}"; then
    error "Failed to download from ${DOWNLOAD_URL}"
  fi

  # Verify checksum if available
  if curl -sL "$CHECKSUM_URL" -o "${TMP_DIR}/checksums.txt" 2>/dev/null; then
    EXPECTED=$(grep "${FILE_NAME}" "${TMP_DIR}/checksums.txt" | awk '{print $1}')
    if [ -n "$EXPECTED" ]; then
      ACTUAL=$(sha256sum "${TMP_DIR}/${BINARY}" 2>/dev/null | awk '{print $1}' || shasum -a 256 "${TMP_DIR}/${BINARY}" | awk '{print $1}')
      if [ "$EXPECTED" != "$ACTUAL" ]; then
        error "Checksum verification failed!"
      fi
      info "Checksum verified"
    fi
  fi

  chmod +x "${TMP_DIR}/${BINARY}"
  mkdir -p "${INSTALL_DIR}"

  if [ -w "${INSTALL_DIR}" ]; then
    mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/"
  else
    warn "Need elevated permissions for ${INSTALL_DIR}"
    sudo mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/"
  fi

  # Create alias symlink
  if [ -n "${ALIAS_NAME}" ] && [ "${ALIAS_NAME}" != "${BINARY}" ]; then
    local alias_path="${INSTALL_DIR}/${ALIAS_NAME}"
    if [ -w "${INSTALL_DIR}" ]; then
      ln -sf "${INSTALL_DIR}/${BINARY}" "${alias_path}"
    else
      sudo ln -sf "${INSTALL_DIR}/${BINARY}" "${alias_path}"
    fi
    info "Created alias: ${alias_path}"
  fi

  info "Installed: ${INSTALL_DIR}/${BINARY}"

  # Check if install directory is in PATH
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      warn "${INSTALL_DIR} is not in your PATH"
      echo "  Add with: export PATH=\"\$PATH:${INSTALL_DIR}\""
      ;;
  esac
}

# Main
main() {
  detect_os
  detect_arch
  get_version
  get_install_dir

  if [ "${OS}" = "windows" ]; then
    error "Windows is not supported by this installer. Download manually from: https://github.com/${REPO}/releases"
  fi

  echo ""
  echo "  ${BINARY} installer"
  echo "  ─────────────────────────────"
  echo "  Version:  ${VERSION}"
  echo "  Platform: ${OS}/${ARCH}"
  echo "  Target:   ${INSTALL_DIR}"
  echo ""

  install_binary

  echo ""
  info "Installation complete! Run '${BINARY} --help' to get started."
}

main
