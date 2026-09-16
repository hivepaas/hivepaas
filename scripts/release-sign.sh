#!/bin/bash
#
# Signs release.json with the offline release keys. Run through `make release-sign`.
#
# release.json is signed with an ed25519 key and an ML-DSA-65 key; the app
# requires both (see hivepaas_app/pkg/releasesig).
#
# The signing tool is not run from the working tree. It is built from
# tools/releasesign/main.go as of RELEASESIGN_SHA - a commit that was reviewed -
# outside any module, with the toolchain and module downloads switched off. A
# change to the tool on a branch therefore does not reach the private keys until
# someone reviews it and moves the pin, and an import from outside the standard
# library fails the build instead of being fetched.
#
# Every signature is then checked a second time with openssl, against a public key
# openssl derives from the private key itself. A tool that signed something other
# than the file it was given, or with a key other than the one it was given, passes
# its own check but not this one.
#
# Environment:
#   RELEASESIGN_SHA  full commit sha to build the tool from (required)
#   KEYS             space-separated private key files, <key-id>.key, one per
#                    algorithm (required)
#   IN               file to sign (default: release.json); signatures go to $IN.sig

set -euo pipefail

IN="${IN:-release.json}"
SIG="$IN.sig"
# Must match signContext in tools/releasesign/main.go.
SIGN_CONTEXT="hivepaas-release-v1"

fail() {
  echo "release-sign: $*" >&2
  exit 1
}

[[ "${RELEASESIGN_SHA:-}" =~ ^[0-9a-f]{40}$ ]] ||
  fail "RELEASESIGN_SHA must be a full 40-character commit sha (got '${RELEASESIGN_SHA:-}'); set it in the Makefile"
[[ -n "${KEYS:-}" ]] ||
  fail 'KEYS is required: make release-sign KEYS="/offline/2026-ed.key /offline/2026-ml.key"'
read -r -a KEY_FILES <<<"$KEYS"
for key in "${KEY_FILES[@]}"; do
  [[ -f "$key" ]] || fail "key file $key not found"
  [[ "$(basename "$key")" == *.key ]] || fail "$key: key files are named <key-id>.key"
done
[[ -f "$IN" ]] || fail "$IN not found"

git cat-file -e "${RELEASESIGN_SHA}^{commit}" 2>/dev/null ||
  fail "commit $RELEASESIGN_SHA is not in this clone (git fetch first)"

# OpenSSL 3.5 or later: ML-DSA arrived in 3.5. LibreSSL, which macOS ships as
# openssl, has neither ML-DSA nor Ed25519ctx.
OPENSSL=""
for candidate in openssl /opt/homebrew/opt/openssl@3/bin/openssl /usr/local/opt/openssl@3/bin/openssl; do
  if command -v "$candidate" >/dev/null 2>&1 &&
    "$candidate" version 2>/dev/null | grep -Eq '^OpenSSL (3\.([5-9]|[1-9][0-9])|[4-9])\.'; then
    OPENSSL="$candidate"
    break
  fi
done
[[ -n "$OPENSSL" ]] || fail "OpenSSL 3.5+ not found (brew install openssl@3); it is required for the independent check"

if ! git diff --quiet HEAD -- "$IN" 2>/dev/null; then
  echo "WARNING: $IN has uncommitted changes. What gets signed is the file on disk, not the commit." >&2
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "---------------------------------------------------------------"
echo "Building releasesign from:"
git log -1 --format='  %H%n  %an, %ad%n  %s' "$RELEASESIGN_SHA"
echo "---------------------------------------------------------------"

git show "${RELEASESIGN_SHA}:tools/releasesign/main.go" >"$TMP/main.go" ||
  fail "tools/releasesign/main.go does not exist at $RELEASESIGN_SHA"

(
  cd "$TMP"
  env GOTOOLCHAIN=local GOFLAGS= GOWORK=off GOPROXY=off \
    go build -trimpath -o releasesign main.go
) || fail "building the tool failed (it must import only the standard library)"

SIGN_ARGS=()
for key in "${KEY_FILES[@]}"; do
  SIGN_ARGS+=(-key "$key")
done
"$TMP/releasesign" sign "${SIGN_ARGS[@]}" -in "$IN" -out "$SIG"

echo "---------------------------------------------------------------"
echo "Independent check with $("$OPENSSL" version | cut -d' ' -f1-2):"
echo "---------------------------------------------------------------"

for key in "${KEY_FILES[@]}"; do
  key_id="$(basename "$key" .key)"
  key_type="$("$OPENSSL" pkey -in "$key" -noout -text | head -1)"
  case "$key_type" in
  ED25519*) opts=(-pkeyopt instance:Ed25519ctx -pkeyopt "context-string:$SIGN_CONTEXT") ;;
  ML-DSA-65*) opts=(-pkeyopt "context-string:$SIGN_CONTEXT") ;;
  *) fail "$key: unexpected key type '$key_type'" ;;
  esac

  # releasesign writes one signature per line: {"keyId":"<id>","alg":"<alg>","sig":"<base64>"}
  sig_b64="$(sed -n "s/.*{\"keyId\":\"${key_id}\",\"alg\":\"[^\"]*\",\"sig\":\"\([^\"]*\)\"}.*/\1/p" "$SIG")"
  [[ -n "$sig_b64" ]] || fail "$SIG has no signature by key $key_id"
  printf '%s' "$sig_b64" | "$OPENSSL" base64 -d -A >"$TMP/$key_id.sig"
  "$OPENSSL" pkey -in "$key" -pubout -out "$TMP/$key_id.pub.pem"

  printf '%-24s ' "$key_id (${key_type%% *}):"
  "$OPENSSL" pkeyutl -verify -pubin -inkey "$TMP/$key_id.pub.pem" -rawin -in "$IN" \
    -sigfile "$TMP/$key_id.sig" "${opts[@]}" ||
    fail "openssl does not accept the signature by $key_id; do not publish $SIG"
done

echo "sha256 (openssl): $("$OPENSSL" dgst -sha256 -r "$IN" | cut -d' ' -f1)"
echo
echo "Done. Commit $SIG together with $IN."
