#!/usr/bin/env bash
set -euo pipefail

# Most of this script is taken from https://github.com/mitsuhiko/rye/blob/main/scripts/install.sh

# Wrap everything in a function so that a truncated script
# does not have the chance to cause issues.
__wrap__() {

# allow overriding the version
VERSION=${LEETGO_VERSION:-latest}
PREFIX=${LEETGO_PREFIX:-${HOME}/.local}

REPO=j178/leetgo
PLATFORM=`uname -s`
ARCH=`uname -m`

if [[ $PLATFORM == "Darwin" ]]; then
  PLATFORM="macOS"
elif [[ $PLATFORM == "Linux" ]]; then
  PLATFORM="linux"
fi

if [[ $ARCH == armv8* ]] || [[ $ARCH == arm64* ]] || [[ $ARCH == aarch64* ]]; then
  ARCH="arm64"
elif [[ $ARCH == i686* ]]; then
  ARCH="x86_64"
fi

BINARY="leetgo_${PLATFORM}_${ARCH}"

# Oddly enough GitHub has different URLs for latest vs specific version
if [[ $VERSION == "latest" ]]; then
  DOWNLOAD_URL=https://github.com/${REPO}/releases/latest/download/${BINARY}.tar.gz
else
  DOWNLOAD_URL=https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}.tar.gz
fi

echo "This script will automatically download and install leetgo (${VERSION}) for you."
echo "leetgo will be installed to \${PREFIX}/bin/leetgo, which is ${PREFIX}/bin/leetgo"
echo "You may install leetgo in a different location by setting the PREFIX environment variable."

if [ "x$(id -u)" == "x0" ]; then
  echo "warning: this script is running as root.  This is dangerous and unnecessary!"
fi

if ! hash curl 2> /dev/null; then
  echo "error: you do not have 'curl' installed which is required for this script."
  exit 1
fi

if ! hash tar 2> /dev/null; then
  echo "error: you do not have 'tar' installed which is required for this script."
  exit 1
fi

TEMP_INSTALL_DIR=`mktemp "${TMPDIR:-/tmp/}.leetgoinstall.XXXXXXXX"`
TEMP_FILE_GZ="${TEMP_INSTALL_DIR}.tar.gz"

rm -rf "$TEMP_INSTALL_DIR"
mkdir -p "$TEMP_INSTALL_DIR"

cleanup() {
  rm -rf "$TEMP_INSTALL_DIR"
  rm -f "$TEMP_FILE_GZ"
}

trap cleanup EXIT
echo "Downloading $DOWNLOAD_URL"
HTTP_CODE=$(curl -SL --progress-bar "$DOWNLOAD_URL" --output "$TEMP_FILE_GZ" --write-out "%{http_code}")
if [[ ${HTTP_CODE} -lt 200 || ${HTTP_CODE} -gt 299 ]]; then
  echo "error: platform ${PLATFORM} (${ARCH}) is unsupported."
  exit 1
fi

tar -xzf "$TEMP_FILE_GZ" -C "$TEMP_INSTALL_DIR"

DEST="$PREFIX/bin/leetgo"

chmod +x "$TEMP_INSTALL_DIR/leetgo"
mv "$TEMP_INSTALL_DIR/leetgo" "$DEST"

echo "leetgo was installed successfully to $DEST"

}; __wrap__
