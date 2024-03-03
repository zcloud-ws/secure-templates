#!/usr/bin/env sh

export INSTALL_VERSION=${INSTALL_VERSION:-"v0.0.1-alpha.2"}
export INSTALL_OS=${INSTALL_OS}
export INSTALL_ARCH=${INSTALL_ARCH}
export INSTALL_DEST_DIR=${INSTALL_DEST_DIR}
export INSTALL_EXT=""

if [ "$INSTALL_DEST_DIR" != "" ]; then
  INSTALL_DEST_DIR="${INSTALL_DEST_DIR}/"
  INSTALL_DEST_DIR="$(echo "$INSTALL_DEST_DIR" | sed 's|//|/|g')"
fi

if [ "$INSTALL_DEST_DIR" = "" ]; then
  if [ "$(id -u)" = "0" ]; then
    INSTALL_DEST_DIR="/usr/local/bin"
  else
    INSTALL_DEST_DIR="$PWD"
  fi
fi

# Detect OS
if [ "${INSTALL_OS}" = "" ]; then
  case $(uname | tr '[:upper:]' '[:lower:]') in
  linux*)
    INSTALL_OS=Linux
    ;;
  bsd*)
    INSTALL_OS=Linux
    ;;
  darwin*)
    INSTALL_OS=Darwin
    ;;
  msys*)
    INSTALL_OS=Windows
    ;;
  *)
    INSTALL_OS=Linux
    ;;
  esac
fi
if [ "$INSTALL_OS" != "Linux" ] && [ "$INSTALL_OS" != "Darwin" ]; then
  echo "Script don't support installation for OS $INSTALL_OS."
  exit 1
fi
# Detect arch
if [ "${INSTALL_ARCH}" = "" ]; then
  case $(arch | tr '[:upper:]' '[:lower:]') in
  x86_64*)
    INSTALL_ARCH=x86_64
    ;;
  i386*)
    INSTALL_ARCH=i386
    ;;
  aarch64*)
    INSTALL_ARCH=arm64
    ;;
  arm64*)
    INSTALL_ARCH=arm64
    ;;
  *)
    INSTALL_ARCH=x86_64
    ;;
  esac
fi

if [ "${INSTALL_OS}" = "win" ]; then
  INSTALL_EXT=".exe"
fi

export FILE_NAME="secure-templates_${INSTALL_OS}_${INSTALL_ARCH}.tar.gz"

export DOWNLOAD_LINK="https://github.com/edimarlnx/secure-templates/releases/download/v0.0.1-alpha.1/${FILE_NAME}"

echo "Download from: $DOWNLOAD_LINK"
echo "Install to: $INSTALL_DEST_DIR"
curl -sL "$DOWNLOAD_LINK" | tar -xz -C "$INSTALL_DEST_DIR" secure-templates
