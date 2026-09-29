#!/usr/bin/env bash
#
# bake-install-ref.sh IN OUT TAG: writes IN (deployment/release/install.sh) to
# OUT with TAG as the ref it downloads the stack files from. The release
# workflow runs it on the installer it attaches to a release, so that installer
# deploys the stack of its own release rather than whatever main holds.
#
# It replaces the one INSTALL_REF_DEFAULT=main line and fails when there is not
# exactly one, rather than attach an installer that still reads main.

set -euo pipefail

if [ $# -ne 3 ]; then
  printf 'usage: %s IN OUT TAG\n' "$0" >&2
  exit 2
fi
in=$1 out=$2 tag=$3

# A tag lands in a shell assignment in a script run as root: only what a
# release tag is made of.
if ! [[ $tag =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z]+[0-9]*)?$ ]]; then
  printf '%s is not a release tag (vMAJOR.MINOR.PATCH or vMAJOR.MINOR.PATCH-beta<N>)\n' "$tag" >&2
  exit 1
fi

count=$(grep -c '^INSTALL_REF_DEFAULT=main$' "$in" || true)
if [ "$count" != 1 ]; then
  printf '%s has %s INSTALL_REF_DEFAULT=main lines, not one\n' "$in" "$count" >&2
  exit 1
fi

sed "s/^INSTALL_REF_DEFAULT=main\$/INSTALL_REF_DEFAULT=$tag/" "$in" >"$out.tmp"
chmod 755 "$out.tmp"
mv -f "$out.tmp" "$out"
