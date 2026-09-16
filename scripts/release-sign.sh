#!/bin/bash
#
# Signs release.json with the offline release keys. Run through `make release-sign`.
#
# release.json is signed with an ed25519 key and an ML-DSA-65 key; the app
# requires both (see hivepaas_app/pkg/releasesig). The result is
# release.signed.json - release.json and its signatures in one file - which is
# what installations fetch. release.json stays as the readable, reviewed copy.
#
# The signing tool is not run from the working tree. It is built from
# tools/releasesign/main.go as of RELEASESIGN_SHA - a commit that was reviewed -
# outside any module, with the toolchain and module downloads switched off. A
# change to the tool on a branch therefore does not reach the private keys until
# someone reviews it and moves the pin, and an import from outside the standard
# library fails the build instead of being fetched.
#
# The envelope is then checked a second time without the tool: its payload must be
# release.json byte for byte, and every signature must verify with openssl against
# a public key openssl derives from the private key itself. A tool that signed or
# packed something other than the file it was given, or used another key, passes
# its own check but not this one.
#
# Environment:
#   RELEASESIGN_SHA  full commit sha to build the tool from (required)
#   KEYS             space-separated private key files, <key-id>.key, one per
#                    algorithm (required)
#   IN               file to sign (default: release.json)
#   OUT              envelope to write (default: release.signed.json)

set -euo pipefail

IN="${IN:-release.json}"
OUT="${OUT:-${IN%.json}.signed.json}"
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
"$TMP/releasesign" sign "${SIGN_ARGS[@]}" -in "$IN" -out "$OUT"

echo "---------------------------------------------------------------"
echo "Independent check with $("$OPENSSL" version | cut -d' ' -f1-2):"
echo "---------------------------------------------------------------"

# releasesign writes the payload and each signature on a line of its own:
#   "payload": "<base64>",
#   {"keyId":"<id>","alg":"<alg>","sig":"<base64>"}
payload_b64="$(sed -n 's/^ *"payload": *"\([^"]*\)",*$/\1/p' "$OUT")"
[[ -n "$payload_b64" ]] || fail "$OUT has no payload"
printf '%s' "$payload_b64" | "$OPENSSL" base64 -d -A >"$TMP/payload"
cmp -s "$TMP/payload" "$IN" || fail "$OUT does not carry $IN byte for byte; do not publish it"
echo "payload:                 identical to $IN"

for key in "${KEY_FILES[@]}"; do
  key_id="$(basename "$key" .key)"
  key_type="$("$OPENSSL" pkey -in "$key" -noout -text | head -1)"
  case "$key_type" in
  ED25519*) opts=(-pkeyopt instance:Ed25519ctx -pkeyopt "context-string:$SIGN_CONTEXT") ;;
  ML-DSA-65*) opts=(-pkeyopt "context-string:$SIGN_CONTEXT") ;;
  *) fail "$key: unexpected key type '$key_type'" ;;
  esac

  sig_b64="$(sed -n "s/.*{\"keyId\":\"${key_id}\",\"alg\":\"[^\"]*\",\"sig\":\"\([^\"]*\)\"}.*/\1/p" "$OUT")"
  [[ -n "$sig_b64" ]] || fail "$OUT has no signature by key $key_id"
  printf '%s' "$sig_b64" | "$OPENSSL" base64 -d -A >"$TMP/$key_id.sig"
  "$OPENSSL" pkey -in "$key" -pubout -out "$TMP/$key_id.pub.pem"

  printf '%-24s ' "$key_id (${key_type%% *}):"
  "$OPENSSL" pkeyutl -verify -pubin -inkey "$TMP/$key_id.pub.pem" -rawin -in "$TMP/payload" \
    -sigfile "$TMP/$key_id.sig" "${opts[@]}" ||
    fail "openssl does not accept the signature by $key_id; do not publish $OUT"
done

echo "sha256 (openssl): $("$OPENSSL" dgst -sha256 -r "$IN" | cut -d' ' -f1)"
echo
echo "Done. Commit $OUT together with $IN."
