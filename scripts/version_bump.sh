#!/bin/sh
set -e

# Makes sure we're in this script's directory (avoid symlinks and escape special chars)
cd "$(cd "$(dirname "$0")"; pwd -P)"

set +e
semver="$(echo "$1" | grep '^v[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}$')"
set -e
if [ -z "$semver" ]; then printf "Version must be provided and valid semvar\\nex: version_bump.sh v1.2.3\\n" && exit 1; fi
printf "%s" "$semver" > ../.version
sed -i "s/VERSION=\"v.*/VERSION=\"$semver\"/" get_installer.bash
