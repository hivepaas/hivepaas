#!/usr/bin/env bash
#
# Tests for install.sh: `make test-installer`, or `bash install_test.sh`.
#
# install.sh is sourced with HIVEPAAS_INSTALL_LIB=1, which defines its
# functions and runs nothing. Each test_ function runs in a subshell, so what
# one sets cannot leak into the next. The stack render test needs the docker
# CLI and is skipped without it. Runs under bash 3.2 (macOS) as under bash 5.
#
# The tests feed install.sh literal $, `...` and $(...) on purpose, and set
# the settings it reads:
# shellcheck disable=SC2016,SC2031,SC2153

set -u

HERE=$(cd "$(dirname "$0")" && pwd)
TMP=$(mktemp -d)
RESULTS=$TMP/results
trap 'rm -rf "$TMP"' EXIT

# shellcheck disable=SC2034 # read by install.sh
HIVEPAAS_INSTALL_LIB=1
# shellcheck source=install.sh
. "$HERE/install.sh" || {
  printf 'cannot load %s\n' "$HERE/install.sh" >&2
  exit 1
}

record() {
  printf '%s\n' "$1" >>"$RESULTS"
}

# check NAME WANT GOT
check() {
  if [ "$2" = "$3" ]; then
    record pass
  else
    record fail
    printf 'FAIL %s\n  want: %s\n  got:  %s\n' "$1" "$2" "$3"
  fi
}

# check_ok NAME CMD...: CMD succeeds. check_fails NAME CMD...: it fails.
check_ok() {
  local name="$1"
  shift
  if "$@"; then record pass; else
    record fail
    printf 'FAIL %s: expected success\n' "$name"
  fi
}

check_fails() {
  local name="$1"
  shift
  if "$@"; then
    record fail
    printf 'FAIL %s: expected failure\n' "$name"
  else record pass; fi
}

matches() {
  [[ $1 =~ $2 ]]
}

# check_contains NAME TEXT PART / check_lacks NAME TEXT PART
check_contains() {
  case "$2" in
    *"$3"*) record pass ;;
    *)
      record fail
      printf 'FAIL %s: missing %s\n  in: %s\n' "$1" "$3" "$2"
      ;;
  esac
}

check_lacks() {
  case "$2" in
    *"$3"*)
      record fail
      printf 'FAIL %s: should not contain %s\n' "$1" "$3"
      ;;
    *) record pass ;;
  esac
}

# ------------------------------------------------------------------- Checks

test_valid_email() {
  local e
  for e in you@example.com a.b+c@sub.example.co.uk; do
    check_ok "email $e" valid_email "$e"
  done
  for e in "" you you@ @example.com you@example "you @example.com" a@b@c.com; do
    check_fails "email '$e'" valid_email "$e"
  done
}

test_valid_password() {
  check_ok "10 characters" valid_password 1234567890
  check_ok "spaces count" valid_password "pass word!!"
  check_fails "9 characters" valid_password 123456789
}

test_valid_hostname() {
  local h
  for h in hivepaas.dev.mydomain.com a.io x-1.example.com xn--80ak6aa92e.com; do
    check_ok "hostname $h" valid_hostname "$h"
  done
  for h in "" localhost Example.com 1.2.3.4 -a.example.com a-.example.com a..b.com a.b.c. "exa mple.com" \
    a_b.example.com "$(printf 'a%.0s' $(seq 64)).com"; do
    check_fails "hostname '$h'" valid_hostname "$h"
  done
  check_ok "a typed domain in capitals" valid_domain HivePaaS.Example.COM
}

test_root_domain_of() {
  check "subdomain" mydomain.com "$(root_domain_of hivepaas.dev.mydomain.com)"
  check "two labels" mydomain.com "$(root_domain_of mydomain.com)"
  check "co.uk" example.co.uk "$(root_domain_of app.example.co.uk)"
  check "co.uk itself" example.co.uk "$(root_domain_of example.co.uk)"
  check "com.vn" example.com.vn "$(root_domain_of a.b.example.com.vn)"
  check "a bare suffix" co.uk "$(root_domain_of co.uk)"
}

test_valid_root_domain() {
  HIVEPAAS_APP_DOMAIN=hivepaas.dev.mydomain.com
  local r
  for r in mydomain.com dev.mydomain.com hivepaas.dev.mydomain.com MyDomain.com; do
    check_ok "root $r" valid_root_domain "$r"
  done
  for r in other.com ydomain.com com ""; do
    check_fails "root '$r'" valid_root_domain "$r"
  done
}

test_normalize_dir() {
  check "trailing slash" /var/lib/hivepaas "$(normalize_dir /var/lib/hivepaas/)"
  check "repeated slashes" /data/hp "$(normalize_dir //data///hp//)"
  check "root" / "$(normalize_dir /)"
}

test_valid_data_dir() {
  local d
  for d in /var/lib/hivepaas /var/lib/hivepaas/ /data /srv/hive-paas_1.0 /home/me/hivepaas /root/hivepaas \
    /opt/hivepaas; do
    check_ok "data dir $d" valid_data_dir "$d"
  done
  for d in "" / relative "/var/lib/hive paas" /data:/x /data/../etc /data/./x /etc /etc/hivepaas \
    /usr/local/hp /tmp/hp /var /home /var/lib/docker/x /run/hp /proc; do
    check_fails "data dir '$d'" valid_data_dir "$d"
  done
}

test_valid_project_data_dir() {
  HIVEPAAS_DATA_DIR=/var/lib/hivepaas
  local d
  for d in /var/lib/hivepaas/project_data /data/projects /var/lib/hive; do
    check_ok "project data dir $d" valid_project_data_dir "$d"
  done
  for d in /var/lib/hivepaas /var/lib/hivepaas/ /var/lib /etc/projects; do
    check_fails "project data dir '$d'" valid_project_data_dir "$d"
  done
}

test_valid_app_secret() {
  check_ok "32 characters" valid_app_secret 0123456789abcdef0123456789abcdef
  check_ok "64 hex" valid_app_secret "$(openssl rand -hex 32)"
  check_fails "31 characters" valid_app_secret 0123456789abcdef0123456789abcde
  check_fails "a space" valid_app_secret "0123456789abcdef 0123456789abcdef"
}

test_valid_ipv4() {
  local ip
  for ip in 1.2.3.4 255.255.255.255 10.0.0.1 0.0.0.0; do
    check_ok "ip $ip" valid_ipv4 "$ip"
  done
  for ip in 256.1.1.1 1.2.3 1.2.3.4.5 01.2.3.4 a.b.c.d ""; do
    check_fails "ip '$ip'" valid_ipv4 "$ip"
  done
}

test_rand_alnum() {
  local a b
  a=$(rand_alnum 32)
  b=$(rand_alnum 32)
  check "length" 32 "${#a}"
  check_ok "letters and digits only" matches "$a" '^[A-Za-z0-9]+$'
  check_fails "two draws differ" test "$a" = "$b"
}

test_version_ge() {
  check_ok "equal after padding" version_ge 29.5.0 29.5
  check_fails "a patch below" version_ge 29.4.9 29.5
  check_ok "10 after 9" version_ge 29.10.0 29.9.1
  check_ok "API equal" version_ge 1.54 1.54
  check_fails "API below" version_ge 1.53 1.54
  check_ok "API above" version_ge 1.56 1.54
  check_ok "next major" version_ge 30.0.0 29.5
  check_ok "a suffix is ignored" version_ge 29.5.0-rc.1 29.5
  check_fails "a major below" version_ge 28.5.2 29.5
  check_ok "a leading zero" version_ge 29.08 29.5
}

test_lowercase() {
  check "lowercase" hivepaas.example.com "$(lowercase HivePaaS.Example.COM)"
}

# ------------------------------------------------------------------- Runner

for t in $(declare -F | awk '$3 ~ /^test_/ {print $3}'); do
  if ! ("$t"); then
    record fail
    printf 'FAIL %s: stopped early\n' "$t"
  fi
done
pass=$(grep -c '^pass$' "$RESULTS" 2>/dev/null) || pass=0
fail=$(grep -c '^fail$' "$RESULTS" 2>/dev/null) || fail=0
printf '%s passed, %s failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
