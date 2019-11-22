#!/usr/bin/env bash
set -e

# Default version (change with new tagged releases)
VERSION="v0.5.0"

if [ -z "$HOME" ]; then
    echo "The \$HOME environment variable must be present"
    exit 1
fi

OS="$(uname | tr '[:upper:]' '[:lower:]')"
if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
    echo "This script only works on linux and mac. If running on windows, download and run the installer executable directly."
    exit 1
fi

if [ "$(uname -m)" != "x86_64" ]; then
    echo "The installer only works on 64 bit machines"
    exit 1
fi

if command -v curl > /dev/null 2>&1; then
    DL_CMD=curl
else
    if command -v wget > /dev/null 2>&1; then
        DL_CMD=wget
    else
        echo "Either wget or curl is required to be installed"
        exit 1
    fi
fi

# Parse input args
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--version'|-v)
        shift
        if [[ $# -ne 0 ]]; then
            export VERSION="${1}"
        else
            echo -e "Please provide the desired version. e.g. --version v2.4.0"
            exit 1
        fi
        ;;
    *)
        printf "Downloads the dragonchain installer (if necessary) and runs it.\\nUse --version <desired_version> to use a specific version of the dragonchain installer\\n"
        exit 1
        ;;
  esac
  shift
done
set +u

# Ensure download directory is created
mkdir -p "$HOME/.local/bin/"

# Download the installer if necessary
LOCAL_FILE="$HOME/.local/bin/dc-installer-$VERSION"
if [ ! -f "$LOCAL_FILE" ]; then
    DOWNLOAD_URL="https://github.com/dragonchain/dragonchain-installer/releases/download/$VERSION/dc-installer-$OS-amd64"
    echo "Downloading installer at $DOWNLOAD_URL"
    if [ "$DL_CMD" = "curl" ]; then
        curl -Lf "$DOWNLOAD_URL" -o "$LOCAL_FILE"
    else
        # Clean up if wget fails
        if ! wget -O "$LOCAL_FILE" "$DOWNLOAD_URL"; then
            rm "$LOCAL_FILE"
            exit 1
        fi
    fi
fi
# Make sure file is executable
chmod +x "$LOCAL_FILE"
# If macos, remove extended attributes to ensure no macos security issues running the installer
if [ "$OS" = "darwin" ]; then
    xattr -c "$LOCAL_FILE"
fi

echo "Installer is available at $LOCAL_FILE"
echo "Running installer now"

# Run the actual installer
$LOCAL_FILE
