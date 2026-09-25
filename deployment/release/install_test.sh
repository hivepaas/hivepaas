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

# ----------------------------------------------------------- Settings files

test_kv_quote_round_trip() {
  local v
  for v in plain "it's" "a'b'c" "with space" "" '$x `y` $(z)' '"dq"' 'back\slash'; do
    check "round trip of $v" "$v" "$(kv_unquote "$(kv_quote "$v")")"
  done
}

test_kv_unquote() {
  check "single quotes" "a b" "$(kv_unquote "'a b'")"
  check "double quotes" "a b" "$(kv_unquote '"a b"')"
  check "bare" "a b" "$(kv_unquote 'a b')"
  check "an escaped quote" "it's" "$(kv_unquote "'it'\\''s'")"
  check "nothing runs" '$(touch x)' "$(kv_unquote '$(touch x)')"
}

collect() {
  COLLECTED="${COLLECTED}$1=$2|"
}

test_read_kv_file() {
  local f="$TMP/kv.conf" out
  printf '%s\n' '# a comment' '' '  HIVEPAAS_A=one' 'export HIVEPAAS_B="two words"' \
    "HIVEPAAS_C = 'it'\\''s'  " 'not a setting' >"$f"
  printf 'HIVEPAAS_D=crlf\r\nHIVEPAAS_E=no newline' >>"$f"
  COLLECTED=''
  out=$(read_kv_file "$f" collect 2>&1)
  read_kv_file "$f" collect 2>/dev/null
  check "lines read" "HIVEPAAS_A=one|HIVEPAAS_B=two words|HIVEPAAS_C=it's|HIVEPAAS_D=crlf|HIVEPAAS_E=no newline|" \
    "$COLLECTED"
  check_contains "a line that is not KEY=VALUE is reported" "$out" "line 6 is not KEY=VALUE"
}

# ------------------------------------------------------------------ Release

REPO=$(cd "$HERE/../.." && pwd)

# envelope FILE JSON: a release.signed.json around JSON, as releasesign writes one.
envelope() {
  local payload sum
  payload=$(printf '%s' "$2" | base64 | tr -d '\n')
  sum=$(printf '%s' "$2" | sha256_of)
  printf '{"payload":"%s","sha256":"%s","signatures":[]}' "$payload" "$sum" >"$1"
}

RELEASE_JSON='{"beta":{"appVersion":"v0.1.1-beta3","appImage":"hivepaas/hivepaas-dev:0.1.0","dbImage":"postgres:18.3-alpine","redisImage":"redis:8.6-alpine","traefikImage":"traefik:v3.7"},"stable":{"appImage":"hivepaas/hivepaas:1.0.0"}}'

test_release_payload() {
  envelope "$TMP/ok.json" "$RELEASE_JSON"
  check "payload" "$RELEASE_JSON" "$(release_payload "$TMP/ok.json")"
  jq '.sha256 = "0000"' "$TMP/ok.json" >"$TMP/tampered.json"
  release_payload "$TMP/tampered.json" >/dev/null
  check "a checksum mismatch returns 2" 2 "$?"
  printf 'not json' >"$TMP/junk.json"
  release_payload "$TMP/junk.json" >/dev/null
  check "not an envelope returns 1" 1 "$?"
}

test_release_payload_of_this_repo() {
  release_payload "$REPO/release.signed.json" >"$TMP/repo-release.json"
  check "release.signed.json verifies" 0 "$?"
  check_ok "its beta channel has an app image" test -n "$(release_field "$TMP/repo-release.json" beta appImage)"
}

test_release_field() {
  printf '%s' "$RELEASE_JSON" >"$TMP/release.json"
  check "app image" hivepaas/hivepaas-dev:0.1.0 "$(release_field "$TMP/release.json" beta appImage)"
  check "missing field" "" "$(release_field "$TMP/release.json" stable dbImage)"
  check "missing channel" "" "$(release_field "$TMP/release.json" nightly appImage)"
  check "channels" "beta, stable" "$(release_channels "$TMP/release.json")"
}

test_app_env_of_channel() {
  check "beta" beta "$(app_env_of_channel beta)"
  check "stable" production "$(app_env_of_channel stable)"
  check_fails "anything else" app_env_of_channel nightly
}

test_image_name_and_tag() {
  check "name" hivepaas/hivepaas-dev "$(image_name hivepaas/hivepaas-dev:0.1.0)"
  check "tag" 0.1.0 "$(image_tag hivepaas/hivepaas-dev:0.1.0)"
  check "name with a registry port" registry.example.com:5000/hivepaas/hivepaas \
    "$(image_name registry.example.com:5000/hivepaas/hivepaas:1.2)"
  check "tag with a registry port" 1.2 "$(image_tag registry.example.com:5000/hivepaas/hivepaas:1.2)"
  check "no tag" "" "$(image_tag registry.example.com:5000/hivepaas/hivepaas)"
  check "digest dropped" 18.3-alpine "$(image_tag postgres:18.3-alpine@sha256:abc)"
}

test_derive_agent_image() {
  check "dev" hivepaas/hivepaas-agent-dev:0.1.0 "$(derive_agent_image hivepaas/hivepaas-dev:0.1.0)"
  check "stable" hivepaas/hivepaas-agent:1.0.0 "$(derive_agent_image hivepaas/hivepaas:1.0.0)"
  check "registry and digest" registry.example.com:5000/hivepaas/hivepaas-agent-dev:0.1.0 \
    "$(derive_agent_image registry.example.com:5000/hivepaas/hivepaas-dev:0.1.0@sha256:abc)"
  check "no tag" hivepaas/hivepaas-agent-dev "$(derive_agent_image hivepaas/hivepaas-dev)"
  check_fails "not a HivePaaS image" derive_agent_image nginx:1
}

test_pg_major_of_image() {
  check "minor tag" 18 "$(pg_major_of_image postgres:18.3-alpine)"
  check "major tag" 17 "$(pg_major_of_image postgres:17)"
  check "with digest" 18 "$(pg_major_of_image postgres:18-alpine@sha256:abc)"
  check_fails "latest" pg_major_of_image postgres:latest
  check_fails "no tag" pg_major_of_image postgres
  check_fails "a registry port is no tag" pg_major_of_image registry:5000/postgres
}

test_docker_static_arch() {
  check "x86_64" x86_64 "$(docker_static_arch x86_64)"
  check "aarch64" aarch64 "$(docker_static_arch aarch64)"
  check "arm64" aarch64 "$(docker_static_arch arm64)"
  check "armv7l" armhf "$(docker_static_arch armv7l)"
  check_fails "riscv64" docker_static_arch riscv64
}

test_latest_docker_from_index() {
  local index='<a href="docker-28.5.2.tgz">docker-28.5.2.tgz</a>
<a href="docker-29.10.0.tgz">docker-29.10.0.tgz</a>
<a href="docker-29.9.1.tgz">docker-29.9.1.tgz</a>
<a href="docker-rootless-extras-30.0.0.tgz">docker-rootless-extras-30.0.0.tgz</a>'
  check "newest" 29.10.0 "$(printf '%s\n' "$index" | latest_docker_from_index)"
  check_fails "an empty page" latest_docker_from_index </dev/null
}

# ------------------------------------------------------------------ Network

test_route_src() {
  check "src" 172.31.5.10 "$(printf '1.1.1.1 via 172.31.0.1 dev eth0 src 172.31.5.10 uid 0\n' | route_src)"
  check "none" "" "$(printf 'unreachable 1.1.1.1\n' | route_src)"
}

test_cidr_overlaps() {
  check_ok "a smaller network inside" cidr_overlaps 10.11.0.0/16 10.11.5.0/24
  check_ok "a larger network around" cidr_overlaps 10.11.0.0/16 10.0.0.0/8
  check_ok "an address inside" cidr_overlaps 10.11.0.0/16 10.11.3.4
  check_ok "everything" cidr_overlaps 10.11.0.0/16 0.0.0.0/0
  check_fails "a neighbour" cidr_overlaps 10.11.0.0/16 10.12.0.0/16
  check_fails "docker0" cidr_overlaps 10.11.0.0/16 172.17.0.0/16
}

test_routes_overlap() {
  local routes='default via 172.31.0.1 dev eth0
172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 linkdown
172.31.0.0/20 dev eth0 proto kernel scope link src 172.31.5.10'
  check_fails "a cloud VM" routes_overlap 10.11.0.0/16 <<<"$routes"
  check_ok "a VPN route over 10/8" routes_overlap 10.11.0.0/16 <<<"$routes
10.0.0.0/8 via 172.31.0.1 dev eth0"
  check_ok "a typed route" routes_overlap 10.11.0.0/16 <<<"$routes
unreachable 10.11.0.0/24"
}

test_ip_host_rule() {
  check "two addresses" 'Host(`1.2.3.4`) || Host(`10.0.0.5`)' "$(ip_host_rule 1.2.3.4 10.0.0.5)"
  check "the same twice" 'Host(`1.2.3.4`)' "$(ip_host_rule 1.2.3.4 1.2.3.4)"
  check "no public address" 'Host(`10.0.0.5`)' "$(ip_host_rule "" 10.0.0.5)"
  check "none at all" 'Host(`127.0.0.1`)' "$(ip_host_rule "" "")"
}

test_addresses_line() {
  HIVEPAAS_APP_DOMAIN=app.example.com PUBLIC_IP=1.2.3.4 HOST_IP=10.0.0.5
  check "all three" "https://app.example.com, https://1.2.3.4, https://10.0.0.5" "$(addresses_line)"
  HOST_IP=1.2.3.4
  check "an address once" "https://app.example.com, https://1.2.3.4" "$(addresses_line)"
  PUBLIC_IP='' HOST_IP=''
  check "the domain alone" "https://app.example.com" "$(addresses_line)"
}

# -------------------------------------------------------------------- Stack

# stack_env: what install.sh exports for docker stack deploy, with sample values.
stack_env() {
  export HIVEPAAS_APP_ENV=beta HIVEPAAS_ROOT_DOMAIN=mydomain.com HIVEPAAS_APP_DOMAIN=hivepaas.dev.mydomain.com \
    HIVEPAAS_APP_SECRET=s3cret HIVEPAAS_JWT_SECRET=jwt HIVEPAAS_DATA_DIR=/var/lib/hivepaas \
    HIVEPAAS_PROJECT_DATA_DIR=/data/projects HIVEPAAS_DB_PASSWORD=dbpw HIVEPAAS_REDIS_PASSWORD=redispw \
    HIVEPAAS_AGENT_TOKEN=agenttok HIVEPAAS_IP_RULE='Host(`1.2.3.4`) || Host(`10.0.0.5`)' \
    HIVEPAAS_ADMIN_EMAIL=you@example.com HIVEPAAS_ADMIN_PASSWORD='p$ss "w0rd"' HP_DB_MAJOR=18 \
    HIVEPAAS_IMAGE_APP=hivepaas/hivepaas-dev:0.1.0 HIVEPAAS_IMAGE_WORKER=hivepaas/hivepaas-dev:0.1.0 \
    HIVEPAAS_IMAGE_UPDATER=hivepaas/hivepaas-dev:0.1.0 HIVEPAAS_IMAGE_AGENT=hivepaas/hivepaas-agent-dev:0.1.0 \
    HIVEPAAS_IMAGE_DB=postgres:18.3-alpine HIVEPAAS_IMAGE_REDIS=redis:8.6-alpine HIVEPAAS_IMAGE_TRAEFIK=traefik:v3.7
}

test_stack_renders() {
  local out status
  if ! command -v docker >/dev/null 2>&1; then
    printf 'skip test_stack_renders: no docker CLI\n'
    return 0
  fi
  stack_env
  out=$(cd "$HERE" && docker stack config -c hivepaas.yaml 2>&1)
  check "renders" 0 "$?"
  check_contains "the router by domain" "$out" 'traefik.http.routers.app-router.rule: Host(`hivepaas.dev.mydomain.com`)'
  check_contains "the router by address" "$out" \
    'traefik.http.routers.x-custom-router-ip.rule: Host(`1.2.3.4`) || Host(`10.0.0.5`)'
  check_contains "the Postgres major" "$out" 'PGDATA: /var/lib/postgresql/18/docker'
  check_contains "project data seen from the app" "$out" 'target: /host/data/projects'
  check "every HivePaaS process gets the secret" 5 "$(printf '%s\n' "$out" | grep -c 'HP_APP_SECRET: s3cret')"
  check_lacks "no admin without the first-boot file" "$out" HP_USER_ADMIN_PASSWORD
  out=$(cd "$HERE" && docker stack config -c hivepaas.yaml -c hivepaas.first-boot.yaml 2>&1)
  check "renders with the first-boot file" 0 "$?"
  check "the admin, literally, on app and worker" 2 \
    "$(printf '%s\n' "$out" | grep -c 'HP_USER_ADMIN_PASSWORD: p$ss "w0rd"')"
  unset HIVEPAAS_APP_SECRET
  out=$(cd "$HERE" && docker stack config -c hivepaas.yaml 2>&1)
  status=$?
  check_fails "a missing setting fails" test "$status" -eq 0
  check_contains "and is named" "$out" HIVEPAAS_APP_SECRET
}

# ---------------------------------------------------------------- Questions

# answers LINE...: a terminal that answers these, one line each.
answers() {
  printf '%s\n' "$@" >"$TMP/tty"
  HIVEPAAS_TTY=$TMP/tty
  open_tty
}

test_ask_questions_interactive() {
  answers not-an-email you@example.com short password123 password124 password123 password123 \
    HivePaaS.Dev.MyDomain.com '' '' /var/lib/hivepaas/ ''
  ask_questions 2>/dev/null
  check "email, asked again" you@example.com "$HIVEPAAS_ADMIN_EMAIL"
  check "password, typed twice alike" password123 "$HIVEPAAS_ADMIN_PASSWORD"
  check "app domain, lowercased" hivepaas.dev.mydomain.com "$HIVEPAAS_APP_DOMAIN"
  check "root domain, derived" mydomain.com "$HIVEPAAS_ROOT_DOMAIN"
  check_ok "app secret, generated" matches "$HIVEPAAS_APP_SECRET" '^[0-9a-f]{64}$'
  check "data dir, normalized" /var/lib/hivepaas "$HIVEPAAS_DATA_DIR"
  check "project data dir, defaulted" /var/lib/hivepaas/project_data "$HIVEPAAS_PROJECT_DATA_DIR"
}

test_ask_questions_silent() {
  HIVEPAAS_ADMIN_EMAIL=you@example.com HIVEPAAS_ADMIN_PASSWORD=password123 HIVEPAAS_APP_DOMAIN=app.example.co.uk
  ask_questions
  check "root domain" example.co.uk "$HIVEPAAS_ROOT_DOMAIN"
  check "data dir" /var/lib/hivepaas "$HIVEPAAS_DATA_DIR"
  check "project data dir" /var/lib/hivepaas/project_data "$HIVEPAAS_PROJECT_DATA_DIR"
  check_ok "app secret" valid_app_secret "$HIVEPAAS_APP_SECRET"
}

test_ask_questions_silent_missing() {
  local out
  out=$( (ask_questions) 2>&1)
  check "stops" 1 "$?"
  check_contains "lists the missing settings" "$out" \
    "HIVEPAAS_ADMIN_EMAIL HIVEPAAS_ADMIN_PASSWORD HIVEPAAS_APP_DOMAIN"
}

test_ask_questions_given_invalid() {
  local out
  HIVEPAAS_ADMIN_EMAIL=you@example.com HIVEPAAS_ADMIN_PASSWORD=short HIVEPAAS_APP_DOMAIN=app.example.com
  out=$( (ask_questions) 2>&1)
  check "a short password stops" 1 "$?"
  check_contains "naming it" "$out" HIVEPAAS_ADMIN_PASSWORD
  check_lacks "without printing it" "$out" short
  HIVEPAAS_ADMIN_PASSWORD=password123 HIVEPAAS_DATA_DIR=/etc/hivepaas
  out=$( (ask_questions) 2>&1)
  check "a system directory stops" 1 "$?"
  HIVEPAAS_DATA_DIR=/data/hp HIVEPAAS_PROJECT_DATA_DIR=/data/hp
  out=$( (ask_questions) 2>&1)
  check "project data on app data stops" 1 "$?"
  check_contains "naming it" "$out" HIVEPAAS_PROJECT_DATA_DIR
}

test_ask_questions_installed() {
  HIVEPAAS_INSTALLED=true HIVEPAAS_APP_DOMAIN=app.example.com HIVEPAAS_APP_SECRET=0123456789abcdef0123456789abcdef
  ask_questions
  check "no admin asked for once installed" "" "${HIVEPAAS_ADMIN_PASSWORD:-}"
}

test_load_settings_precedence() {
  printf '%s\n' HIVEPAAS_ADMIN_EMAIL=file@example.com HIVEPAAS_ADMIN_PASSWORD=file-password >"$TMP/conf"
  printf '%s\n' "HIVEPAAS_ADMIN_PASSWORD='saved-password'" "HIVEPAAS_DATA_DIR='/srv/hivepaas'" >"$TMP/install.env"
  CONFIG_FILE=$TMP/conf INSTALL_ENV=$TMP/install.env HIVEPAAS_ADMIN_EMAIL=env@example.com
  load_settings
  check "the environment wins" env@example.com "$HIVEPAAS_ADMIN_EMAIL"
  check "the file wins over install.env" file-password "$HIVEPAAS_ADMIN_PASSWORD"
  check "install.env fills the rest" /srv/hivepaas "$HIVEPAAS_DATA_DIR"
  check "the channel defaults to beta" beta "$HIVEPAAS_CHANNEL"
}

test_load_settings_fixed() {
  local out saved=0123456789abcdef0123456789abcdef
  printf '%s\n' "HIVEPAAS_APP_SECRET='$saved'" "HIVEPAAS_DATA_DIR='/srv/hivepaas'" >"$TMP/install.env"
  INSTALL_ENV=$TMP/install.env
  out=$(HIVEPAAS_APP_SECRET=${saved}X; (load_settings) 2>&1)
  check "a different app secret stops" 1 "$?"
  check_contains "naming it" "$out" HIVEPAAS_APP_SECRET
  check_lacks "without printing it" "$out" "$saved"
  out=$(HIVEPAAS_APP_SECRET=$saved HIVEPAAS_DATA_DIR=/srv//hivepaas/; (load_settings) 2>&1)
  check "the same values, written differently, go on" 0 "$?"
  out=$(HIVEPAAS_CHANNEL=nightly; (load_settings) 2>&1)
  check "an unknown channel stops" 1 "$?"
}

test_load_settings_runs_nothing() {
  printf '%s\n' 'HIVEPAAS_X[$(touch '"$TMP"'/ran)]=1' 'PATH=/nowhere' 'HIVEPAAS_Y=$(touch '"$TMP"'/ran2)' >"$TMP/conf"
  CONFIG_FILE=$TMP/conf INSTALL_ENV=$TMP/none.env
  load_settings 2>/dev/null
  check_fails "a key is not evaluated" test -e "$TMP/ran"
  check_fails "a value is not evaluated" test -e "$TMP/ran2"
  check "a value is taken literally" '$(touch '"$TMP"'/ran2)' "$HIVEPAAS_Y"
  check_fails "only HIVEPAAS_ settings are taken" test "$PATH" = /nowhere
}

test_save_settings() {
  INSTALL_ENV=$TMP/etc/hivepaas/install.env
  HIVEPAAS_ADMIN_PASSWORD="it's \$x \"quoted\"" HIVEPAAS_APP_DOMAIN=app.example.com HIVEPAAS_INSTALLED=''
  save_settings
  check "file mode" 600 "$(stat -c %a "$INSTALL_ENV" 2>/dev/null || stat -f %Lp "$INSTALL_ENV")"
  check "dir mode" 700 "$(stat -c %a "${INSTALL_ENV%/*}" 2>/dev/null || stat -f %Lp "${INSTALL_ENV%/*}")"
  check_lacks "an empty setting is not written" "$(cat "$INSTALL_ENV")" HIVEPAAS_INSTALLED
  unset HIVEPAAS_ADMIN_PASSWORD HIVEPAAS_APP_DOMAIN
  load_settings
  check "a password reads back" "it's \$x \"quoted\"" "$HIVEPAAS_ADMIN_PASSWORD"
  check "a domain reads back" app.example.com "$HIVEPAAS_APP_DOMAIN"
}

test_confirm_and_offer() {
  ASSUME_YES=1
  check_ok "--yes confirms" confirm "Go?"
  check_fails "--yes declines an offer" offer "Upgrade?"
  ASSUME_YES=0
  out=$( (confirm "Go?") 2>&1)
  check "no terminal and no --yes stops" 1 "$?"
  check_contains "saying how to go on" "$out" "--yes"
  check_fails "no terminal declines an offer" offer "Upgrade?"
  answers '' n y ''
  check_ok "Enter confirms" confirm "Go?"
  check_fails "n does not" confirm "Go?"
  check_ok "y takes an offer" offer "Upgrade?"
  check_fails "Enter declines it" offer "Upgrade?"
}

test_print_summary_hides_secrets() {
  local out
  HIVEPAAS_ADMIN_EMAIL=you@example.com HIVEPAAS_ADMIN_PASSWORD=password123 HIVEPAAS_APP_DOMAIN=app.example.com
  HIVEPAAS_ROOT_DOMAIN=example.com HIVEPAAS_APP_SECRET=0123456789abcdef0123456789abcdef
  HIVEPAAS_DATA_DIR=/var/lib/hivepaas HIVEPAAS_PROJECT_DATA_DIR=/var/lib/hivepaas/project_data PUBLIC_IP=1.2.3.4
  out=$(print_summary)
  check_lacks "no password" "$out" password123
  check_lacks "no secret" "$out" 0123456789abcdef0123456789abcdef
  check_contains "the secret's end, to recognize it" "$out" "****cdef"
  check_contains "the addresses" "$out" "https://app.example.com, https://1.2.3.4"
}

# --------------------------------------------------------------------- Host

test_read_os_release() {
  printf '%s\n' 'PRETTY_NAME="Linux Mint 22.1"' 'NAME="Linux Mint"' 'ID=linuxmint' 'ID_LIKE="ubuntu debian"' \
    'VERSION_CODENAME=xia' 'UBUNTU_CODENAME=noble' >"$TMP/os-release"
  read_os_release "$TMP/os-release"
  check "id" linuxmint "$OS_ID"
  check "id like" "ubuntu debian" "$OS_ID_LIKE"
  check "name" "Linux Mint 22.1" "$OS_NAME"
  check "ubuntu codename" noble "$OS_UBUNTU_CODENAME"
  printf 'ID="opensuse-leap"\nID_LIKE="suse opensuse"\n' >"$TMP/os-release2"
  OS_NAME=''
  read_os_release "$TMP/os-release2"
  check "a name falls back to the id" opensuse-leap "$OS_NAME"
}

test_package_manager_of() {
  check "ubuntu" apt "$(package_manager_of ubuntu '')"
  check "mint" apt "$(package_manager_of linuxmint 'ubuntu debian')"
  check "kali" apt "$(package_manager_of kali debian)"
  check "rocky" dnf "$(package_manager_of rocky 'rhel centos fedora')"
  check "oracle" dnf "$(package_manager_of ol fedora)"
  check "amazon" dnf "$(package_manager_of amzn fedora)"
  check "leap" zypper "$(package_manager_of opensuse-leap 'suse opensuse')"
  check "sles" zypper "$(package_manager_of sles suse)"
  check "arch" pacman "$(package_manager_of arch '')"
  check "endeavour" pacman "$(package_manager_of endeavouros arch)"
  check "alpine" apk "$(package_manager_of alpine '')"
  check_fails "gentoo" package_manager_of gentoo ''
  check_fails "nixos" package_manager_of nixos ''
}

test_docker_install_method() {
  local id
  for id in ubuntu debian raspbian fedora centos rhel rocky; do
    check "$id" getdocker "$(docker_install_method "$id" '')"
  done
  check "alma" dnf-rhel "$(docker_install_method almalinux 'rhel centos fedora')"
  check "oracle" dnf-rhel "$(docker_install_method ol fedora)"
  check "amazon" dnf "$(docker_install_method amzn fedora)"
  check "mint" apt-repo "$(docker_install_method linuxmint 'ubuntu debian')"
  check "kali" apt-repo "$(docker_install_method kali debian)"
  check "tumbleweed" zypper "$(docker_install_method opensuse-tumbleweed 'opensuse suse')"
  check "sles" zypper "$(docker_install_method sles suse)"
  check "arch" pacman "$(docker_install_method arch '')"
  check "manjaro" pacman "$(docker_install_method manjaro arch)"
  check "alpine" apk "$(docker_install_method alpine '')"
  check_fails "gentoo" docker_install_method gentoo ''
}

test_earlyoom_config() {
  local conf
  check "apt" /etc/default/earlyoom "$(earlyoom_config_file apt)"
  check "dnf" /etc/default/earlyoom "$(earlyoom_config_file dnf)"
  check "pacman" /etc/default/earlyoom "$(earlyoom_config_file pacman)"
  check "zypper" /etc/sysconfig/earlyoom "$(earlyoom_config_file zypper)"
  check "apk" /etc/conf.d/earlyoom "$(earlyoom_config_file apk)"
  check "systemd arguments" "EARLYOOM_ARGS=\"-m 5 -s 20 -r 3600 --avoid $EARLYOOM_AVOID\"" "$(earlyoom_config apt)"
  check_fails "the regex has no whitespace for systemd to split on" matches "$EARLYOOM_AVOID" '[[:space:]]'
  conf=$(earlyoom_config apk)
  check_contains "openrc memory" "$conf" "mem_min_percent=5"
  check_contains "openrc avoid" "$conf" "avoid_cmds='^(hivepaas|"
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
