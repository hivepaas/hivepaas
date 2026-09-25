#!/usr/bin/env bash
#
# Installs HivePaaS on this server.
#
#   curl -fsSL https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/release/install.sh | sudo bash
#
# It installs Docker when it is missing, makes the server a single-node swarm
# and deploys the HivePaaS stack of a release channel, beta for now. Every
# answer is saved in /etc/hivepaas/install.env. Running it again finishes an
# interrupted install or re-checks the host: it never resets a swarm, removes a
# service, network or volume, or generates a secret a second time.
# `install.sh --help` lists the settings of a silent install.
#
# For tests only:
#   HIVEPAAS_INSTALL_LIB=1       define the functions and stop (install_test.sh)
#   HIVEPAAS_INSTALL_FILES_DIR   copy the stack files from this directory
#   HIVEPAAS_RELEASE_URL         read the release info from this URL
#   HIVEPAAS_INSTALL_ENV_FILE    where the answers are saved
#   HIVEPAAS_TTY                 read answers from this file, not the terminal
#   HIVEPAAS_WAIT_SECONDS        how long to wait for the dashboard (300)

# ------------------------------------------------------------------- Output

STEP_NO=0
STEP_TOTAL=9
C_RESET='' C_BOLD='' C_RED='' C_GREEN='' C_YELLOW='' C_BLUE='' C_LOGO=''

setup_colors() {
  if [ -t 1 ] && [ -z "${NO_COLOR:-}" ] && [ "${TERM:-}" != dumb ]; then
    C_RESET=$'\033[0m' C_BOLD=$'\033[1m' C_RED=$'\033[31m' C_GREEN=$'\033[32m'
    C_YELLOW=$'\033[33m' C_BLUE=$'\033[34m' C_LOGO=$'\033[1;33m'
  fi
}

info() { printf '  %s\n' "$*"; }
ok() { printf '  %s✔%s %s\n' "$C_GREEN" "$C_RESET" "$*"; }
warn() { printf '  %s!%s %s\n' "$C_YELLOW" "$C_RESET" "$*" >&2; }
die() {
  printf '\n%s✘ %s%s\n' "$C_RED" "$*" "$C_RESET" >&2
  exit 1
}
step() {
  STEP_NO=$((STEP_NO + 1))
  printf '\n%s%s[%d/%d] %s%s\n' "$C_BOLD" "$C_BLUE" "$STEP_NO" "$STEP_TOTAL" "$*" "$C_RESET"
}

# on_error runs for a command that failed where nothing expected it to. Inside
# a subshell it stays quiet: the caller sees the status and reports it there.
on_error() {
  [ "${BASH_SUBSHELL:-0}" -eq 0 ] || return 0
  printf '\n%s✘ The installer stopped at line %s (exit status %s).%s\n' "$C_RED" "$2" "$1" "$C_RESET" >&2
  printf '  Fix what the output above says, then run it again: it picks up where it stopped.\n' >&2
  exit "$1"
}

print_logo() {
  local cols
  cols=$(stty size 2>/dev/null </dev/tty | awk '{print $2}') || cols=
  if [ -n "$cols" ] && [ "$cols" -lt 94 ]; then
    printf '\n%sHivePaaS%s\n' "$C_LOGO" "$C_RESET"
    return 0
  fi
  printf '\n%s' "$C_LOGO"
  cat <<'LOGO'
       ▄████▄          ██      ██ ██                  ███████                       ████████
     ▄██▀░░▀██▄       ░██     ░██░░                  ░██░░░░██                     ██░░░░░░
  ▄████▄    ▄████▄    ░██     ░██ ██ ██    ██  █████ ░██   ░██  ██████    ██████  ░██
▄██▀░░▀██▄▄██▀░░▀██▄  ░██████████░██░██   ░██ ██░░░██░███████  ░░░░░░██  ░░░░░░██ ░█████████
██░ ░░ ░████░ ░░ ░██  ░██░░░░░░██░██░░██ ░██ ░███████░██░░░░    ███████   ███████ ░░░░░░░░██
▀██▄░░▄██▀▀██▄░░▄██▀  ░██     ░██░██ ░░████  ░██░░░░ ░██       ██░░░░██  ██░░░░██        ░██
  ▀████▀    ▀████▀    ░██     ░██░██  ░░██   ░░██████░██      ░░████████░░████████ ████████
     ▀██████▀         ░░      ░░ ░░    ░░     ░░░░░░ ░░        ░░░░░░░░  ░░░░░░░░ ░░░░░░░░
LOGO
  printf '%s' "$C_RESET"
}

# ------------------------------------------------------------------- Checks

# Suffixes under which a registrable domain has three labels, not two. No list
# is complete, which is why the derived root domain is shown for correction.
TWO_LEVEL_SUFFIXES="co.uk org.uk ac.uk gov.uk me.uk ltd.uk plc.uk com.vn net.vn org.vn edu.vn gov.vn
com.au net.au org.au co.jp ne.jp or.jp com.br com.cn com.tw com.hk com.sg com.my co.nz co.za co.in
co.id co.kr com.tr com.mx com.ar"

lowercase() { printf '%s' "$1" | tr '[:upper:]' '[:lower:]'; }

valid_email() {
  [[ $1 =~ ^[^@[:space:]]+@[^@[:space:]]+\.[^@[:space:]]+$ ]]
}

valid_password() {
  [ "${#1}" -ge 10 ]
}

# valid_hostname: a lowercase domain name with at least two labels, the last
# one starting with a letter - which rules out an IP address.
valid_hostname() {
  [ "${#1}" -le 253 ] &&
    [[ $1 =~ ^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]([a-z0-9-]{0,61}[a-z0-9])?$ ]]
}

# valid_domain: a hostname as typed, in any case.
valid_domain() {
  valid_hostname "$(lowercase "$1")"
}

# root_domain_of: the registrable part of a domain - its last two labels, or
# three under a known two-level suffix.
root_domain_of() {
  local domain="$1" last2 suffix
  last2=$(printf '%s' "$domain" | awk -F. 'NF >= 2 {print $(NF-1) "." $NF}')
  for suffix in $TWO_LEVEL_SUFFIXES; do
    if [ "$last2" = "$suffix" ]; then
      printf '%s' "$domain" | awk -F. 'NF >= 3 {print $(NF-2) "." $(NF-1) "." $NF; exit} {print}'
      return 0
    fi
  done
  printf '%s\n' "${last2:-$domain}"
}

# valid_root_domain ROOT: a hostname the app domain is, or is under.
valid_root_domain() {
  local root
  root=$(lowercase "$1")
  valid_hostname "$root" || return 1
  case "$(lowercase "${HIVEPAAS_APP_DOMAIN:-}")" in
    "$root" | *".$root") return 0 ;;
  esac
  return 1
}

# normalize_dir: one slash between names and none at the end.
normalize_dir() {
  local dir
  dir=$(printf '%s' "$1" | tr -s /)
  if [ "$dir" != / ]; then
    dir=${dir%/}
  fi
  printf '%s' "$dir"
}

# valid_data_dir: an absolute path Docker can bind as it is written, outside
# the directories the system owns or clears. Checked as normalize_dir leaves it.
valid_data_dir() {
  local dir
  dir=$(normalize_dir "$1")
  [[ $dir =~ ^(/[A-Za-z0-9._-]+)+$ ]] || return 1
  case "$dir/" in
    */./* | */../*) return 1 ;;
  esac
  case "$dir/" in
    /bin/* | /boot/* | /dev/* | /etc/* | /lib/* | /lib32/* | /lib64/* | /libx32/* | /proc/* | \
      /run/* | /sbin/* | /sys/* | /tmp/* | /usr/* | /var/run/* | /var/tmp/* | /var/lib/docker/*)
      return 1
      ;;
    /var/ | /home/ | /root/)
      return 1
      ;;
  esac
  return 0
}

# valid_project_data_dir: a data directory that is not the app data directory
# or one above it, where project directories would land among HivePaaS's own.
valid_project_data_dir() {
  local dir
  valid_data_dir "$1" || return 1
  dir=$(normalize_dir "$1")
  case "$(normalize_dir "${HIVEPAAS_DATA_DIR:-}")/" in
    "$dir"/*) return 1 ;;
  esac
  return 0
}

valid_app_secret() {
  [ "${#1}" -ge 32 ] && [[ ! $1 =~ [[:space:]] ]]
}

valid_ipv4() {
  local octet='(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])'
  [[ $1 =~ ^$octet\.$octet\.$octet\.$octet$ ]]
}

# rand_alnum N: N random letters and digits.
rand_alnum() {
  local out=''
  while [ "${#out}" -lt "$1" ]; do
    out="$out$(openssl rand -base64 48 | tr -dc 'A-Za-z0-9')"
  done
  printf '%s' "${out:0:$1}"
}

# version_ge A B: A is B or later. Compares numbers part by part and ignores
# anything after a - or +.
version_ge() {
  local a="${1%%[-+]*}" b="${2%%[-+]*}" i x y
  local -a pa pb
  IFS=. read -r -a pa <<<"$a"
  IFS=. read -r -a pb <<<"$b"
  for i in 0 1 2 3; do
    x=${pa[i]:-0}
    y=${pb[i]:-0}
    x=${x//[!0-9]/}
    y=${y//[!0-9]/}
    x=$((10#${x:-0}))
    y=$((10#${y:-0}))
    if [ "$x" -gt "$y" ]; then return 0; fi
    if [ "$x" -lt "$y" ]; then return 1; fi
  done
  return 0
}

# ----------------------------------------------------------- Settings files

# kv_quote: a value in single quotes, as install.env stores it.
kv_quote() {
  local q="'"
  printf "'%s'" "${1//$q/$q\\$q$q}"
}

# kv_unquote: the value of a KEY=VALUE line - taken literally, or from single
# or double quotes. Nothing in it is expanded or executed.
kv_unquote() {
  local v="$1" q="'" esc="'\\''"
  case "$v" in
    "'"*"'") v=${v#\'}; v=${v%\'}; v=${v//"$esc"/$q} ;;
    '"'*'"') v=${v#\"}; v=${v%\"} ;;
  esac
  printf '%s' "$v"
}

# read_kv_file FILE FN: calls FN KEY VALUE for each KEY=VALUE line of FILE.
# Blank lines and # comments are skipped, and so is an `export ` in front.
read_kv_file() {
  local file="$1" fn="$2" line key value n=0
  while IFS= read -r line || [ -n "$line" ]; do
    n=$((n + 1))
    line=${line%$'\r'}
    line=${line#"${line%%[![:space:]]*}"}
    case "$line" in
      '' | '#'*) continue ;;
    esac
    line=${line#export }
    case "$line" in
      *=*) ;;
      *)
        warn "$file line $n is not KEY=VALUE; ignored."
        continue
        ;;
    esac
    key=${line%%=*}
    key=${key%"${key##*[![:space:]]}"}
    value=${line#*=}
    value=${value#"${value%%[![:space:]]*}"}
    value=${value%"${value##*[![:space:]]}"}
    "$fn" "$key" "$(kv_unquote "$value")"
  done <"$file"
}

# ------------------------------------------------------------------ Release

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum | awk '{print $1}'
  else
    shasum -a 256 | awk '{print $1}'
  fi
}

# release_payload FILE: the release info inside release.signed.json. Returns 2
# when the payload does not match its sha256, 1 when the file is not an
# envelope at all.
release_payload() {
  local payload want got
  payload=$(jq -r '.payload // empty' "$1" 2>/dev/null) || return 1
  want=$(jq -r '.sha256 // empty' "$1" 2>/dev/null) || return 1
  if [ -z "$payload" ] || [ -z "$want" ]; then return 1; fi
  got=$(printf '%s' "$payload" | base64 -d | sha256_of) || return 1
  if [ "$got" != "$want" ]; then return 2; fi
  printf '%s' "$payload" | base64 -d
}

# release_field FILE CHANNEL FIELD: one field of a channel of the release info.
release_field() {
  jq -r --arg ch "$2" --arg f "$3" '.[$ch][$f] // empty' "$1"
}

release_channels() {
  jq -r 'keys | join(", ")' "$1"
}

# app_env_of_channel CHANNEL: the app's env, and config file, for a channel.
app_env_of_channel() {
  case "$1" in
    beta) printf 'beta' ;;
    stable) printf 'production' ;;
    *) return 1 ;;
  esac
}

# image_name IMAGE / image_tag IMAGE: an image reference without, and its tag
# alone. A digest is dropped, and a registry's port is not taken for a tag.
image_name() {
  local image="${1%%@*}" last
  last=${image##*/}
  case "$last" in
    *:*) printf '%s' "${image%:*}" ;;
    *) printf '%s' "$image" ;;
  esac
}

image_tag() {
  local image="${1%%@*}" last
  last=${image##*/}
  case "$last" in
    *:*) printf '%s' "${last##*:}" ;;
  esac
}

# derive_agent_image APP_IMAGE: the agent image built with an app image -
# `-agent` after `hivepaas`, the same tag.
derive_agent_image() {
  local name tag prefix='' repo
  name=$(image_name "$1")
  tag=$(image_tag "$1")
  case "$name" in
    */*) prefix="${name%/*}/" ;;
  esac
  repo=${name##*/}
  case "$repo" in
    hivepaas) repo=hivepaas-agent ;;
    hivepaas-*) repo="hivepaas-agent-${repo#hivepaas-}" ;;
    *) return 1 ;;
  esac
  printf '%s%s%s' "$prefix" "$repo" "${tag:+:$tag}"
}

# pg_major_of_image IMAGE: the Postgres major an image's tag names.
pg_major_of_image() {
  local tag major
  tag=$(image_tag "$1")
  major=${tag%%[!0-9]*}
  if [ -z "$major" ]; then return 1; fi
  printf '%s' "$major"
}

# docker_static_arch MACHINE: the directory of Docker's static builds for an
# architecture `uname -m` names.
docker_static_arch() {
  case "$1" in
    x86_64 | amd64) printf 'x86_64' ;;
    aarch64 | arm64) printf 'aarch64' ;;
    armv6l | armv7l | armhf) printf 'armhf' ;;
    ppc64le | s390x) printf '%s' "$1" ;;
    *) return 1 ;;
  esac
}

# latest_docker_from_index: the newest version an index page of Docker's
# static builds lists, read from stdin.
latest_docker_from_index() {
  local v
  v=$(grep -oE 'docker-[0-9]+\.[0-9]+\.[0-9]+\.tgz' | sed -e 's/^docker-//' -e 's/\.tgz$//' |
    sort -t. -k1,1n -k2,2n -k3,3n | tail -n 1) || true
  if [ -z "$v" ]; then return 1; fi
  printf '%s\n' "$v"
}

# ------------------------------------------------------------------ Network

HOST_IP='' PUBLIC_IP=''

# route_src: the source address in `ip route get` output, read from stdin.
route_src() {
  awk '{for (i = 1; i < NF; i++) if ($i == "src") {print $(i + 1); exit}}'
}

# host_ip: this host's address on its default route.
host_ip() {
  local ip
  command -v ip >/dev/null 2>&1 || return 1
  ip=$(ip route get 1.1.1.1 2>/dev/null | route_src) || return 1
  valid_ipv4 "$ip" || return 1
  printf '%s' "$ip"
}

# public_ip: the address the internet sees this host at.
public_ip() {
  local ip
  ip=$(curl -4 -fsS --max-time 5 https://ifconfig.io 2>/dev/null | tr -d '[:space:]') || return 1
  valid_ipv4 "$ip" || return 1
  printf '%s' "$ip"
}

ip4_to_int() {
  local a b c d
  IFS=. read -r a b c d <<<"$1"
  printf '%s' $(((a << 24) + (b << 16) + (c << 8) + d))
}

# cidr_overlaps A B: two IPv4 networks share an address. An address without a
# prefix length is a network of one.
cidr_overlaps() {
  local a="${1%/*}" b="${2%/*}" an=32 bn=32 n mask
  case "$1" in */*) an=${1#*/} ;; esac
  case "$2" in */*) bn=${2#*/} ;; esac
  if ! valid_ipv4 "$a" || ! valid_ipv4 "$b"; then return 1; fi
  n=$an
  if [ "$bn" -lt "$n" ]; then n=$bn; fi
  if [ "$n" -eq 0 ]; then return 0; fi
  mask=$(((0xFFFFFFFF << (32 - n)) & 0xFFFFFFFF))
  [ $(($(ip4_to_int "$a") & mask)) -eq $(($(ip4_to_int "$b") & mask)) ]
}

# routes_overlap CIDR: a route in `ip route show` output, read from stdin,
# reaches into CIDR.
routes_overlap() {
  local first second dest
  while read -r first second _; do
    case "$first" in
      default) continue ;;
      [0-9]*) dest=$first ;;
      *) dest=$second ;;
    esac
    if cidr_overlaps "$1" "$dest"; then return 0; fi
  done
  return 1
}

# ip_host_rule IP...: the Traefik rule matching requests made to these
# addresses. Empty and repeated ones are skipped; with none left the rule
# matches only 127.0.0.1, so the router still has one.
ip_host_rule() {
  local ip rule='' seen=' '
  for ip in "$@"; do
    if [ -z "$ip" ]; then continue; fi
    case "$seen" in *" $ip "*) continue ;; esac
    seen="$seen$ip "
    rule="${rule:+$rule || }Host(\`$ip\`)"
  done
  printf '%s' "${rule:-Host(\`127.0.0.1\`)}"
}

addresses_line() {
  local line="https://$HIVEPAAS_APP_DOMAIN"
  if [ -n "$PUBLIC_IP" ]; then line="$line, https://$PUBLIC_IP"; fi
  if [ -n "$HOST_IP" ] && [ "$HOST_IP" != "$PUBLIC_IP" ]; then line="$line, https://$HOST_IP"; fi
  printf '%s' "$line"
}

# ---------------------------------------------------------------- Questions

ADMIN_USERNAME='admin'
INSTALL_ENV=${HIVEPAAS_INSTALL_ENV_FILE:-/etc/hivepaas/install.env}

# The answers install.env keeps, so that a second run asks nothing twice and
# generates no secret twice.
SAVED_KEYS="HIVEPAAS_CHANNEL HIVEPAAS_ADMIN_EMAIL HIVEPAAS_ADMIN_PASSWORD HIVEPAAS_APP_DOMAIN
HIVEPAAS_ROOT_DOMAIN HIVEPAAS_APP_SECRET HIVEPAAS_DATA_DIR HIVEPAAS_PROJECT_DATA_DIR
HIVEPAAS_JWT_SECRET HIVEPAAS_DB_PASSWORD HIVEPAAS_REDIS_PASSWORD HIVEPAAS_AGENT_TOKEN
HIVEPAAS_INSTALLED"

# The saved answers a later run may not change: the secrets open data already
# written with them, and the rest says where that data is and what serves it.
FIXED_KEYS="HIVEPAAS_CHANNEL HIVEPAAS_APP_DOMAIN HIVEPAAS_ROOT_DOMAIN HIVEPAAS_APP_SECRET
HIVEPAAS_DATA_DIR HIVEPAAS_PROJECT_DATA_DIR HIVEPAAS_JWT_SECRET HIVEPAAS_DB_PASSWORD
HIVEPAAS_REDIS_PASSWORD HIVEPAAS_AGENT_TOKEN"

ASSUME_YES=0
CONFIG_FILE=''
HAVE_TTY=0
MISSING=''
RELEASE_FILE=''

# open_tty: questions are read from the terminal, not from stdin, which
# `curl | bash` makes the script itself. fd 3 reads answers, fd 4 shows the
# questions.
open_tty() {
  if [ -n "${HIVEPAAS_TTY:-}" ]; then
    exec 3<"$HIVEPAAS_TTY" 4>/dev/null
    HAVE_TTY=1
  elif (exec </dev/tty) 2>/dev/null; then
    exec 3</dev/tty 4>/dev/tty
    HAVE_TTY=1
  fi
}

# ask VAR PROMPT: one line from the terminal into VAR.
ask() {
  printf '  %s' "$2" >&4
  IFS= read -r -u 3 "$1" || die "No answer to '$2'."
}

# ask_secret VAR PROMPT: the same, not echoed.
ask_secret() {
  printf '  %s' "$2" >&4
  IFS= read -r -s -u 3 "$1" || die "No answer to '$2'."
  printf '\n' >&4
}

# confirm QUESTION: yes unless answered no. --yes answers it; with no terminal
# and no --yes the installer stops rather than guess.
confirm() {
  local reply
  if [ "$ASSUME_YES" = 1 ]; then return 0; fi
  if [ "$HAVE_TTY" != 1 ]; then die "$1 Nothing to answer on; run again with --yes to answer yes."; fi
  ask reply "$1 [Y/n] "
  case "$reply" in
    '' | [Yy] | [Yy][Ee][Ss]) return 0 ;;
  esac
  return 1
}

# offer QUESTION: no unless answered yes. --yes and a missing terminal both
# leave it at no.
offer() {
  local reply
  if [ "$ASSUME_YES" = 1 ] || [ "$HAVE_TTY" != 1 ]; then return 1; fi
  ask reply "$1 [y/N] "
  case "$reply" in
    [Yy] | [Yy][Ee][Ss]) return 0 ;;
  esac
  return 1
}

# answer VAR PROMPT CHECK ERROR [DEFAULT [DEFAULT_LABEL]]: VAR as given, as
# answered, or its default. A given value that fails CHECK stops the install;
# an answer that fails it is asked again. With no terminal and no default the
# variable goes on the MISSING list.
answer() {
  local var="$1" prompt="$2" check="$3" error="$4" default="${5:-}" label="${6:-${5:-}}" value
  value=${!var:-}
  if [ -n "$value" ]; then
    "$check" "$value" || die "$var: $error"
    return 0
  fi
  if [ "$HAVE_TTY" != 1 ]; then
    if [ -n "$default" ]; then
      printf -v "$var" '%s' "$default"
    else
      MISSING="$MISSING $var"
    fi
    return 0
  fi
  while :; do
    ask value "$prompt${label:+ [$label]}: "
    value=${value:-$default}
    if [ -n "$value" ] && "$check" "$value"; then break; fi
    warn "$error"
  done
  printf -v "$var" '%s' "$value"
}

answer_password() {
  local p1 p2
  if [ -n "${HIVEPAAS_ADMIN_PASSWORD:-}" ]; then
    valid_password "$HIVEPAAS_ADMIN_PASSWORD" || die "HIVEPAAS_ADMIN_PASSWORD: it needs at least 10 characters."
    return 0
  fi
  if [ "$HAVE_TTY" != 1 ]; then
    MISSING="$MISSING HIVEPAAS_ADMIN_PASSWORD"
    return 0
  fi
  while :; do
    ask_secret p1 "Admin password (at least 10 characters): "
    if ! valid_password "$p1"; then
      warn "The password needs at least 10 characters."
      continue
    fi
    ask_secret p2 "Admin password, again: "
    if [ "$p1" = "$p2" ]; then break; fi
    warn "The two passwords differ; try again."
  done
  HIVEPAAS_ADMIN_PASSWORD=$p1
}

# take_setting KEY VALUE: a line of a --config file. The environment wins over
# the file; anything but a HIVEPAAS_ setting is ignored, and the name is
# checked before it is assigned, so the file cannot run anything.
take_setting() {
  if [[ ! $1 =~ ^HIVEPAAS_[A-Z0-9_]+$ ]]; then
    warn "$CONFIG_FILE: $1 is not a HivePaaS setting; ignored."
    return 0
  fi
  if [ -n "${!1:-}" ] || [ -z "$2" ]; then return 0; fi
  printf -v "$1" '%s' "$2"
}

# take_saved KEY VALUE: a line of install.env, for a setting the environment
# and --config left unset. One they set to another value stops the install
# when the setting is fixed; otherwise theirs wins.
take_saved() {
  case " $SAVED_KEYS " in
    *[[:space:]]"$1"[[:space:]]*) ;;
    *) return 0 ;;
  esac
  if [ -z "${!1:-}" ]; then
    printf -v "$1" '%s' "$2"
    return 0
  fi
  if [ "${!1}" != "$2" ]; then
    case " $FIXED_KEYS " in
      *[[:space:]]"$1"[[:space:]]*)
        die "$1 is saved in $INSTALL_ENV with another value, and it cannot change once HivePaaS" \
          "has been installed with it. Leave it out of the environment and of --config."
        ;;
    esac
  fi
}

# normalize_settings: settings put in the form they are saved and compared in.
normalize_settings() {
  if [ -n "${HIVEPAAS_APP_DOMAIN:-}" ]; then HIVEPAAS_APP_DOMAIN=$(lowercase "$HIVEPAAS_APP_DOMAIN"); fi
  if [ -n "${HIVEPAAS_ROOT_DOMAIN:-}" ]; then HIVEPAAS_ROOT_DOMAIN=$(lowercase "$HIVEPAAS_ROOT_DOMAIN"); fi
  if [ -n "${HIVEPAAS_DATA_DIR:-}" ]; then HIVEPAAS_DATA_DIR=$(normalize_dir "$HIVEPAAS_DATA_DIR"); fi
  if [ -n "${HIVEPAAS_PROJECT_DATA_DIR:-}" ]; then
    HIVEPAAS_PROJECT_DATA_DIR=$(normalize_dir "$HIVEPAAS_PROJECT_DATA_DIR")
  fi
}

# load_settings: the environment, then --config, then install.env - each only
# for what the ones before it left unset.
load_settings() {
  if [ -n "$CONFIG_FILE" ]; then
    [ -r "$CONFIG_FILE" ] || die "Cannot read $CONFIG_FILE."
    read_kv_file "$CONFIG_FILE" take_setting
  fi
  normalize_settings
  if [ -f "$INSTALL_ENV" ]; then
    read_kv_file "$INSTALL_ENV" take_saved
  fi
  HIVEPAAS_CHANNEL=${HIVEPAAS_CHANNEL:-beta}
  app_env_of_channel "$HIVEPAAS_CHANNEL" >/dev/null ||
    die "HIVEPAAS_CHANNEL: '$HIVEPAAS_CHANNEL' is not a channel; it is beta or stable."
}

# ask_questions: every setting install.env does not hold yet. The admin is
# asked for only until HivePaaS has created them.
ask_questions() {
  if [ "${HIVEPAAS_INSTALLED:-}" != true ]; then
    answer HIVEPAAS_ADMIN_EMAIL "Admin email" valid_email "Enter an email address, like you@example.com."
    answer_password
  fi
  answer HIVEPAAS_APP_DOMAIN "App domain, the dashboard's address (e.g. hivepaas.example.com)" valid_domain \
    "Enter a domain name of two labels or more, like hivepaas.example.com."
  normalize_settings
  if [ -n "${HIVEPAAS_APP_DOMAIN:-}" ]; then
    answer HIVEPAAS_ROOT_DOMAIN "Root domain, which apps get subdomains of" valid_root_domain \
      "Enter the app domain or a domain it is under, like example.com." "$(root_domain_of "$HIVEPAAS_APP_DOMAIN")"
  fi
  answer HIVEPAAS_APP_SECRET "App secret, which encrypts stored secrets" valid_app_secret \
    "The app secret needs 32 characters or more, and no spaces." "$(openssl rand -hex 32)" "Enter to generate"
  answer HIVEPAAS_DATA_DIR "App data directory" valid_data_dir \
    "Enter an absolute path of letters, digits, '.', '_' and '-', outside the system's directories." \
    /var/lib/hivepaas
  normalize_settings
  answer HIVEPAAS_PROJECT_DATA_DIR "Project data directory" valid_project_data_dir \
    "Enter an absolute path outside the system's directories, and not the app data directory or one above it." \
    "$HIVEPAAS_DATA_DIR/project_data"
  normalize_settings
  if [ -n "$MISSING" ]; then
    die "No terminal to ask on, and these settings are missing:$MISSING. Set them in the environment" \
      "or in a --config file; install.sh --help lists them."
  fi
}

generate_secrets() {
  HIVEPAAS_JWT_SECRET=${HIVEPAAS_JWT_SECRET:-$(rand_alnum 32)}
  HIVEPAAS_DB_PASSWORD=${HIVEPAAS_DB_PASSWORD:-$(rand_alnum 32)}
  HIVEPAAS_REDIS_PASSWORD=${HIVEPAAS_REDIS_PASSWORD:-$(rand_alnum 32)}
  HIVEPAAS_AGENT_TOKEN=${HIVEPAAS_AGENT_TOKEN:-$(rand_alnum 32)}
}

# save_settings: install.env, replaced whole: readable by root alone, and never
# half written.
save_settings() {
  local dir="${INSTALL_ENV%/*}" tmp key
  mkdir -p "$dir"
  chmod 700 "$dir"
  tmp=$(mktemp "$dir/.install.env.XXXXXX")
  {
    printf '# HivePaaS install settings, written by install.sh.\n'
    printf '# HIVEPAAS_APP_SECRET is the only key to the encrypted data: keep a copy of\n'
    printf '# this file somewhere other than this server.\n'
    for key in $SAVED_KEYS; do
      if [ -n "${!key:-}" ]; then printf '%s=%s\n' "$key" "$(kv_quote "${!key}")"; fi
    done
  } >"$tmp"
  chmod 600 "$tmp"
  mv -f "$tmp" "$INSTALL_ENV"
}

mask() {
  printf '****%s' "${1: -4}"
}

print_summary() {
  printf '\n'
  if [ -n "$RELEASE_FILE" ]; then
    info "Release        HivePaaS $(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" appVersion) ($HIVEPAAS_CHANNEL)"
  fi
  if [ "${HIVEPAAS_INSTALLED:-}" != true ]; then
    info "Admin          $ADMIN_USERNAME, $HIVEPAAS_ADMIN_EMAIL, a password of ${#HIVEPAAS_ADMIN_PASSWORD} characters"
  fi
  info "App domain     $HIVEPAAS_APP_DOMAIN"
  info "Root domain    $HIVEPAAS_ROOT_DOMAIN"
  info "App secret     $(mask "$HIVEPAAS_APP_SECRET")"
  info "App data       $HIVEPAAS_DATA_DIR"
  info "Project data   $HIVEPAAS_PROJECT_DATA_DIR"
  info "Addresses      $(addresses_line)"
  printf '\n'
}

# --------------------------------------------------------------------- Host

# earlyoom never kills these. postgres and redis-server are HivePaaS's own, but
# the names also cover user apps' databases: earlyoom sees process names, not
# services. The kernel cuts a name at 15 characters, hence the wildcards.
EARLYOOM_AVOID='^(hivepaas|hivepaas-agent|traefik|postgres|redis-server|dockerd|containerd|containerd-shim|sshd|systemd|systemd-.*'
EARLYOOM_AVOID+='|vlagent.*|victoria-logs.*|zot-linux-.*)$'

OS_ID='' OS_ID_LIKE='' OS_CODENAME='' OS_UBUNTU_CODENAME='' OS_DEBIAN_CODENAME='' OS_NAME=''
PKG_MANAGER=''
PKG_INDEX_FRESH=0

take_os_release() {
  case "$1" in
    ID) OS_ID=$(lowercase "$2") ;;
    ID_LIKE) OS_ID_LIKE=$(lowercase "$2") ;;
    VERSION_CODENAME) OS_CODENAME=$2 ;;
    UBUNTU_CODENAME) OS_UBUNTU_CODENAME=$2 ;;
    DEBIAN_CODENAME) OS_DEBIAN_CODENAME=$2 ;;
    PRETTY_NAME) OS_NAME=$2 ;;
  esac
}

# read_os_release FILE: the distribution, read as data rather than sourced.
read_os_release() {
  read_kv_file "$1" take_os_release
  OS_NAME=${OS_NAME:-$OS_ID}
}

# package_manager_of ID ID_LIKE: the package manager of a distribution, by its
# /etc/os-release ID or, for a derivative, the IDs it is like.
package_manager_of() {
  local id
  for id in $1 $2; do
    case "$id" in
      debian | ubuntu | raspbian) printf 'apt' && return 0 ;;
      fedora | rhel | centos | rocky | almalinux | ol | amzn) printf 'dnf' && return 0 ;;
      sles | suse | opensuse | opensuse-*) printf 'zypper' && return 0 ;;
      arch | manjaro) printf 'pacman' && return 0 ;;
      alpine) printf 'apk' && return 0 ;;
    esac
  done
  return 1
}

# docker_install_method ID ID_LIKE: how Docker is installed on a distribution.
#   getdocker  Docker's convenience script, for the distributions it supports
#   apt-repo   Docker's apt repository, for derivatives of Ubuntu and Debian
#   dnf-rhel   Docker's RHEL repository, for the RHEL rebuilds it leaves out
#   dnf, zypper, pacman, apk   the distribution's own package
docker_install_method() {
  case "$1" in
    ubuntu | debian | raspbian | centos | fedora | rhel | rocky) printf 'getdocker' && return 0 ;;
    amzn) printf 'dnf' && return 0 ;;
  esac
  case " $2 " in
    *" ubuntu "* | *" debian "*) printf 'apt-repo' && return 0 ;;
    *" rhel "* | *" centos "* | *" fedora "*) printf 'dnf-rhel' && return 0 ;;
  esac
  case "$(package_manager_of "$1" "$2")" in
    zypper) printf 'zypper' ;;
    pacman) printf 'pacman' ;;
    apk) printf 'apk' ;;
    *) return 1 ;;
  esac
}

# earlyoom_config_file PKG_MANAGER: where the distribution's earlyoom service
# reads its arguments.
earlyoom_config_file() {
  case "$1" in
    zypper) printf '/etc/sysconfig/earlyoom' ;;
    apk) printf '/etc/conf.d/earlyoom' ;;
    *) printf '/etc/default/earlyoom' ;;
  esac
}

# earlyoom_config PKG_MANAGER: act when available memory is under 5% and free
# swap under 20% - the swap is there to be used first. Under systemd the regex
# is unquoted: $EARLYOOM_ARGS is split on whitespace, and it has none.
earlyoom_config() {
  if [ "$1" = apk ]; then
    printf 'mem_min_percent=5\nswap_min_percent=20\navoid_cmds=%s\ncommand_args="-r 3600"\n' \
      "$(kv_quote "$EARLYOOM_AVOID")"
  else
    printf 'EARLYOOM_ARGS="-m 5 -s 20 -r 3600 --avoid %s"\n' "$EARLYOOM_AVOID"
  fi
}

pm_install() {
  case "$PKG_MANAGER" in
    apt)
      if [ "$PKG_INDEX_FRESH" != 1 ]; then
        apt-get update -qq </dev/null || return 1
        PKG_INDEX_FRESH=1
      fi
      DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "$@" </dev/null
      ;;
    dnf) dnf install -y -q "$@" </dev/null ;;
    zypper) zypper --non-interactive --quiet install "$@" </dev/null ;;
    pacman)
      # Arch has no partial upgrades: the index is refreshed with the system.
      if [ "$PKG_INDEX_FRESH" != 1 ]; then
        pacman -Syu --noconfirm </dev/null || return 1
        PKG_INDEX_FRESH=1
      fi
      pacman -S --noconfirm --needed "$@" </dev/null
      ;;
    apk) apk add --no-cache "$@" </dev/null ;;
    *) return 1 ;;
  esac
}

has_systemd() {
  command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]
}

# service_start NAME: started now and at every boot.
service_start() {
  if has_systemd; then
    systemctl enable --now "$1" >/dev/null 2>&1
  elif command -v rc-service >/dev/null 2>&1; then
    rc-update add "$1" default >/dev/null 2>&1 || true
    rc-service "$1" status >/dev/null 2>&1 || rc-service "$1" start >/dev/null 2>&1
  else
    return 1
  fi
}

# service_restart NAME: restarted now, so it reads its configuration, and
# started at every boot.
service_restart() {
  if has_systemd; then
    systemctl enable "$1" >/dev/null 2>&1 && systemctl restart "$1" >/dev/null 2>&1
  elif command -v rc-service >/dev/null 2>&1; then
    rc-update add "$1" default >/dev/null 2>&1 || true
    rc-service "$1" restart >/dev/null 2>&1
  else
    return 1
  fi
}

ca_bundle_present() {
  [ -s /etc/ssl/certs/ca-certificates.crt ] || [ -s /etc/pki/tls/certs/ca-bundle.crt ] ||
    [ -s /etc/ssl/ca-bundle.pem ] || [ -s /etc/ssl/cert.pem ]
}

install_tools() {
  local tool
  local -a missing=()
  for tool in curl openssl jq; do
    if ! command -v "$tool" >/dev/null 2>&1; then missing+=("$tool"); fi
  done
  if ! ca_bundle_present; then missing+=(ca-certificates); fi
  if [ "${#missing[@]}" -eq 0 ]; then
    ok "curl, openssl and jq are here."
    return 0
  fi
  info "Installing ${missing[*]}..."
  pm_install "${missing[@]}" || die "Could not install ${missing[*]}. Install them, then run the installer again."
  ok "Installed ${missing[*]}."
}

check_resources() {
  local mem_mb disk_mb
  mem_mb=$(awk '/^MemTotal:/ {print int($2 / 1024)}' /proc/meminfo 2>/dev/null) || mem_mb=
  if [ -n "$mem_mb" ] && [ "$mem_mb" -lt 1900 ]; then
    warn "This server has ${mem_mb} MB of memory; HivePaaS wants 2 GB or more."
  fi
  disk_mb=$(df -Pm /var/lib 2>/dev/null | awk 'NR == 2 {print $4}') || disk_mb=
  if [ -n "$disk_mb" ] && [ "$disk_mb" -lt 20480 ]; then
    warn "/var/lib has $((disk_mb / 1024)) GB free; HivePaaS wants 20 GB or more."
  fi
}

preflight() {
  if [ "$(id -u)" -ne 0 ]; then die "Run the installer as root, e.g. with sudo."; fi
  if [ "$(uname -s)" != Linux ]; then die "HivePaaS runs on Linux."; fi
  if [ ! -r /etc/os-release ]; then die "Cannot tell which Linux this is: /etc/os-release is missing."; fi
  read_os_release /etc/os-release
  PKG_MANAGER=$(package_manager_of "$OS_ID" "$OS_ID_LIKE") ||
    die "$OS_NAME is not supported. The installer knows Debian, Ubuntu and their derivatives;" \
      "Fedora, RHEL, CentOS, Rocky, AlmaLinux, Oracle Linux and Amazon Linux; SLES and openSUSE;" \
      "Arch and Manjaro; and Alpine."
  ok "$OS_NAME ($PKG_MANAGER)"
  check_resources
  install_tools
}

make_swap_file() {
  { fallocate -l "${2}M" "$1" 2>/dev/null || dd if=/dev/zero of="$1" bs=1M count="$2" 2>/dev/null; } &&
    chmod 600 "$1" && mkswap "$1" >/dev/null 2>&1 && swapon "$1" 2>/dev/null || return 1
  if ! grep -q "^$1 " /etc/fstab 2>/dev/null; then
    printf '%s none swap sw 0 0\n' "$1" >>/etc/fstab
  fi
}

# setup_swap: without swap, a host out of memory can only drop program code
# and file cache, and everything on it crawls - healthchecks included - long
# before the kernel's OOM killer acts. A failure here is a warning.
setup_swap() {
  local size="${HIVEPAAS_SWAP_SIZE_MB:-2048}" file=/swapfile free_mb
  if [ "${HIVEPAAS_SWAP:-true}" = false ]; then
    info "Swap: skipped (HIVEPAAS_SWAP=false)."
    return 0
  fi
  if [ "$(awk 'NR > 1' /proc/swaps 2>/dev/null | wc -l)" -gt 0 ]; then
    ok "Swap is on already."
  elif [ -e "$file" ]; then
    warn "Swap: $file exists but is not in use; left as it is."
  else
    free_mb=$(df -Pm / | awk 'NR == 2 {print $4}') || free_mb=0
    if [ "${free_mb:-0}" -lt $((size + 1024)) ]; then
      warn "Swap: not enough free disk for a ${size} MB swap file; skipped."
    elif make_swap_file "$file" "$size"; then
      ok "Swap: ${size} MB at $file."
    else
      rm -f "$file"
      warn "Swap: this host would not take a swap file; skipped."
    fi
  fi
  # Swap only what is really idle: at the default of 60 the kernel also swaps
  # out memory that services touch regularly.
  if mkdir -p /etc/sysctl.d 2>/dev/null &&
    printf 'vm.swappiness = 10\n' 2>/dev/null >/etc/sysctl.d/99-hivepaas-memory.conf &&
    sysctl -w vm.swappiness=10 >/dev/null 2>&1; then
    ok "vm.swappiness = 10"
  else
    warn "Could not set vm.swappiness to 10."
  fi
}

# setup_earlyoom: kills the largest process that is not a system one before a
# host out of memory stalls. A failure here is a warning.
setup_earlyoom() {
  local conf
  local -a packages=(earlyoom)
  if [ "${HIVEPAAS_EARLYOOM:-true}" = false ]; then
    info "earlyoom: skipped (HIVEPAAS_EARLYOOM=false)."
    return 0
  fi
  if ! command -v earlyoom >/dev/null 2>&1; then
    if [ "$PKG_MANAGER" = apk ]; then packages+=(earlyoom-openrc); fi
    if ! pm_install "${packages[@]}" >/dev/null 2>&1; then
      warn "earlyoom: $OS_NAME does not package it (RHEL-like systems have it in EPEL); skipped."
      return 0
    fi
  fi
  conf=$(earlyoom_config_file "$PKG_MANAGER")
  if mkdir -p "${conf%/*}" && earlyoom_config "$PKG_MANAGER" >"$conf" && service_restart earlyoom; then
    ok "earlyoom is on, and leaves HivePaaS's own processes alone."
  else
    warn "earlyoom: installed, but its service would not start."
  fi
}

# ------------------------------------------------------------------- Docker

MIN_DOCKER_VERSION=29.5
MIN_DOCKER_API=1.54
DOCKER_VERSION='' DOCKER_API=''
WORK_DIR=''

read_docker_versions() {
  DOCKER_VERSION=$(docker version --format '{{.Server.Version}}' 2>/dev/null) || DOCKER_VERSION=
  DOCKER_API=$(docker version --format '{{.Server.APIVersion}}' 2>/dev/null) || DOCKER_API=
}

docker_new_enough() {
  [ -n "$DOCKER_VERSION" ] && version_ge "$DOCKER_VERSION" "$MIN_DOCKER_VERSION" &&
    version_ge "$DOCKER_API" "$MIN_DOCKER_API"
}

wait_for_docker() {
  for _ in $(seq 30); do
    if docker info >/dev/null 2>&1; then return 0; fi
    sleep 1
  done
  return 1
}

# latest_docker_version: the newest Docker release, from the index of Docker's
# static builds for this architecture.
latest_docker_version() {
  local arch
  arch=$(docker_static_arch "$(uname -m)") || return 1
  curl -fsSL --max-time 10 "https://download.docker.com/linux/static/stable/$arch/" 2>/dev/null |
    latest_docker_from_index
}

install_docker_apt_repo() {
  local dist=debian codename
  codename=${OS_DEBIAN_CODENAME:-$OS_CODENAME}
  case " $OS_ID_LIKE " in
    *" ubuntu "*)
      dist=ubuntu
      codename=$OS_UBUNTU_CODENAME
      ;;
  esac
  if [ -z "$codename" ]; then return 1; fi
  pm_install ca-certificates curl &&
    install -m 0755 -d /etc/apt/keyrings &&
    curl -fsSL "https://download.docker.com/linux/$dist/gpg" -o /etc/apt/keyrings/docker.asc &&
    chmod a+r /etc/apt/keyrings/docker.asc &&
    printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/%s %s stable\n' \
      "$(dpkg --print-architecture)" "$dist" "$codename" >/etc/apt/sources.list.d/docker.list &&
    apt-get update -qq </dev/null &&
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq docker-ce docker-ce-cli containerd.io \
      docker-buildx-plugin docker-compose-plugin </dev/null
}

install_docker_rhel_repo() {
  local repo=https://download.docker.com/linux/rhel/docker-ce.repo
  dnf install -y -q dnf-plugins-core </dev/null &&
    { dnf config-manager --add-repo "$repo" >/dev/null 2>&1 ||
      dnf config-manager addrepo --overwrite --from-repofile="$repo" >/dev/null; } &&
    dnf install -y -q docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin </dev/null
}

# install_docker: installs Docker, or upgrades it to the latest release, the
# way the distribution takes it; then starts it at boot and now.
install_docker() {
  local method
  method=$(docker_install_method "$OS_ID" "$OS_ID_LIKE") ||
    die "The installer cannot install Docker on $OS_NAME. Install Docker $MIN_DOCKER_VERSION or newer" \
      "(https://docs.docker.com/engine/install/), then run it again."
  info "Installing Docker ($method)..."
  case "$method" in
    getdocker)
      curl -fsSL https://get.docker.com -o "$WORK_DIR/get-docker.sh" && sh "$WORK_DIR/get-docker.sh" </dev/null
      ;;
    apt-repo) install_docker_apt_repo ;;
    dnf-rhel) install_docker_rhel_repo ;;
    dnf | zypper) pm_install docker ;;
    pacman) pm_install docker ;;
    apk) apk add --no-cache --upgrade docker </dev/null ;;
  esac || die "Installing Docker failed; see the output above."
  service_start docker || true
  wait_for_docker || die "Docker is installed but does not answer. Look at 'systemctl status docker', then run the installer again."
  read_docker_versions
}

ensure_docker() {
  local latest just_installed=0
  if ! command -v docker >/dev/null 2>&1; then
    confirm "Docker is not installed. Install it now?" ||
      die "HivePaaS needs Docker $MIN_DOCKER_VERSION or newer. Install it, then run the installer again."
    install_docker
    just_installed=1
  elif ! docker info >/dev/null 2>&1; then
    service_start docker || true
    wait_for_docker || die "Docker does not answer. Start it ('systemctl start docker'), then run the installer again."
  fi
  read_docker_versions
  if ! docker_new_enough; then
    if [ "$just_installed" = 1 ]; then
      die "$OS_NAME's Docker is $DOCKER_VERSION (API $DOCKER_API); HivePaaS needs $MIN_DOCKER_VERSION" \
        "(API $MIN_DOCKER_API) or newer. Install a newer Docker (https://docs.docker.com/engine/install/)," \
        "then run the installer again."
    fi
    confirm "Docker $DOCKER_VERSION (API $DOCKER_API) is older than HivePaaS needs ($MIN_DOCKER_VERSION, API $MIN_DOCKER_API). Upgrade it now? Its containers restart." ||
      die "HivePaaS needs Docker $MIN_DOCKER_VERSION or newer. Upgrade it, then run the installer again."
    install_docker
    if ! docker_new_enough; then
      die "Docker is still $DOCKER_VERSION after the upgrade: $OS_NAME does not package a newer one." \
        "Install Docker $MIN_DOCKER_VERSION or newer (https://docs.docker.com/engine/install/), then run the installer again."
    fi
  elif [ "$just_installed" = 0 ] && latest=$(latest_docker_version) && ! version_ge "$DOCKER_VERSION" "$latest"; then
    if [ "${HIVEPAAS_UPGRADE_DOCKER:-}" = true ] ||
      offer "Docker $DOCKER_VERSION is installed and $latest is out. Upgrade? Its containers restart."; then
      install_docker
    fi
  fi
  ok "Docker $DOCKER_VERSION (API $DOCKER_API)"
}

# ---------------------------------------------------------- Swarm and files

SUBNET=10.11.0.0/16
SUBNET_GATEWAY=10.11.0.1
REPO_RAW=https://raw.githubusercontent.com/hivepaas/hivepaas
CONFIG_DIR=${INSTALL_ENV%/*}

ensure_swarm() {
  local state node
  state=$(docker info --format '{{.Swarm.LocalNodeState}}')
  case "$state" in
    active)
      if [ "$(docker info --format '{{.Swarm.ControlAvailable}}')" != true ]; then
        die "This server is a worker in a swarm, and HivePaaS runs on a manager. Run the installer on a manager."
      fi
      ok "This server is a swarm manager."
      ;;
    inactive)
      # A host with more than one address has swarm init refuse to guess which
      # to advertise; the default route's is the one other nodes would reach.
      if [ -n "$HOST_IP" ]; then
        docker swarm init --advertise-addr "$HOST_IP" >/dev/null || die "docker swarm init failed; see above."
      else
        docker swarm init >/dev/null || die "docker swarm init failed; see above."
      fi
      ok "Swarm started${HOST_IP:+, advertising $HOST_IP}."
      ;;
    *)
      die "The swarm on this server is '$state'. Bring it back to active ('docker swarm unlock', if it is" \
        "locked), then run the installer again."
      ;;
  esac
  node=$(docker info --format '{{.Swarm.NodeID}}')
  docker node update --label-add hivepaas.role=control-plane "$node" >/dev/null
  ok "Node labeled hivepaas.role=control-plane."
}

# ensure_network: the network Traefik finds services on. Its subnet is dev's
# when no route of this host reaches into it; nothing depends on the subnet.
ensure_network() {
  if docker network inspect hivepaas_net >/dev/null 2>&1; then
    ok "Network hivepaas_net is there."
    return 0
  fi
  if ! (command -v ip >/dev/null 2>&1 && ip route show 2>/dev/null | routes_overlap "$SUBNET") &&
    docker network create --driver overlay --attachable --subnet "$SUBNET" --gateway "$SUBNET_GATEWAY" \
      --opt com.docker.network.driver.mtu=1380 hivepaas_net >/dev/null 2>&1; then
    ok "Network hivepaas_net ($SUBNET)."
    return 0
  fi
  docker network create --driver overlay --attachable --opt com.docker.network.driver.mtu=1380 \
    hivepaas_net >/dev/null || die "Could not create the network hivepaas_net; see above."
  ok "Network hivepaas_net, on a subnet Docker picked: $SUBNET is in use on this host."
}

# write_self_signed_cert DIR ROOT APP: the certificate the app would otherwise
# make on its first boot, and adopts when it finds one. Traefik serves it from
# its first second instead of a default of its own.
write_self_signed_cert() {
  local san="DNS:$2,DNS:*.$2"
  if [ "$3" != "$2" ]; then san="$san,DNS:$3"; fi
  openssl req -x509 -days 365 -nodes -sha256 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
    -keyout "$1/self-signed.key" -out "$1/self-signed.crt" -subj "/CN=$2" \
    -addext "subjectAltName=$san" >/dev/null 2>&1 && chmod 600 "$1/self-signed.key"
}

# fetch_install_file NAME DEST: a file of deployment/release, from the ref the
# installer came from.
fetch_install_file() {
  local tmp="$2.tmp"
  mkdir -p "${2%/*}"
  if [ -n "${HIVEPAAS_INSTALL_FILES_DIR:-}" ]; then
    cp "$HIVEPAAS_INSTALL_FILES_DIR/$1" "$tmp"
  else
    curl -fsSL --retry 3 "$REPO_RAW/${HIVEPAAS_INSTALL_REF:-main}/deployment/release/$1" -o "$tmp"
  fi || {
    rm -f "$tmp"
    die "Could not get $1."
  }
  mv -f "$tmp" "$2"
}

prepare_files() {
  local data="$HIVEPAAS_DATA_DIR" certs="$HIVEPAAS_DATA_DIR/ssl/certs"
  mkdir -p "$certs" "$data/traefik/etc/dynamic" "$data/traefik/var/log" "$HIVEPAAS_PROJECT_DATA_DIR"
  ok "Directories: $data, $HIVEPAAS_PROJECT_DATA_DIR."
  if [ -s "$certs/self-signed.crt" ] && [ -s "$certs/self-signed.key" ]; then
    ok "Self-signed certificate: kept."
  else
    write_self_signed_cert "$certs" "$HIVEPAAS_ROOT_DOMAIN" "$HIVEPAAS_APP_DOMAIN" ||
      die "Could not make the self-signed certificate."
    ok "Self-signed certificate for $HIVEPAAS_ROOT_DOMAIN, *.$HIVEPAAS_ROOT_DOMAIN and $HIVEPAAS_APP_DOMAIN."
  fi
  fetch_install_file hivepaas.yaml "$CONFIG_DIR/hivepaas.yaml"
  fetch_install_file hivepaas.first-boot.yaml "$CONFIG_DIR/hivepaas.first-boot.yaml"
  # The app writes Traefik's configuration from here on; a run after the first
  # leaves it alone.
  if [ ! -f "$data/traefik/etc/dynamic/dynamic_conf.yml" ]; then
    fetch_install_file traefik/dynamic_conf.yml "$data/traefik/etc/dynamic/dynamic_conf.yml"
  fi
  ok "Stack files in $CONFIG_DIR."
}

# ------------------------------------------------------------------- Deploy

STACK=hivepaas
REDEPLOY=0
INSTALL_STATE=''
HP_DB_VOLUME='' HP_DB_MAJOR=''
UPDATE_ARGS=()
HIVEPAAS_IP_RULE=''

fetch_release() {
  local url="${HIVEPAAS_RELEASE_URL:-$REPO_RAW/${HIVEPAAS_RELEASE_BRANCH:-release}/release.signed.json}" status=0
  curl -fsSL --retry 3 "$url" -o "$WORK_DIR/release.signed.json" 2>/dev/null ||
    die "Could not read the release info from $url. HIVEPAAS_RELEASE_BRANCH names the branch it is read from."
  release_payload "$WORK_DIR/release.signed.json" >"$WORK_DIR/release.json" || status=$?
  case "$status" in
    0) ;;
    2) die "The release info from $url does not match its checksum; nothing is installed from it." ;;
    *) die "$url is not release info." ;;
  esac
  RELEASE_FILE=$WORK_DIR/release.json
  if [ -z "$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" appImage)" ]; then
    die "The release info has no '$HIVEPAAS_CHANNEL' channel. It has: $(release_channels "$RELEASE_FILE")."
  fi
}

gather_addresses() {
  HOST_IP=$(host_ip) || HOST_IP=
  PUBLIC_IP=$(public_ip) || PUBLIC_IP=
  HIVEPAAS_IP_RULE=$(ip_host_rule "$PUBLIC_IP" "$HOST_IP")
}

service_exists() {
  docker service inspect "${STACK}_$1" >/dev/null 2>&1
}

db_volume_exists() {
  [ -n "$(docker volume ls -q --filter name=hivepaas_db 2>/dev/null)" ]
}

# detect_install_state: fresh, unfinished (deployed, the admin's password not
# yet removed) or installed.
detect_install_state() {
  if [ ! -f "$INSTALL_ENV" ] && { service_exists app || db_volume_exists; }; then
    die "HivePaaS is on this server already, but $INSTALL_ENV is not, and the secrets in it are what open" \
      "its database. Put your copy of the file back, then run the installer again."
  fi
  if [ "${HIVEPAAS_INSTALLED:-}" = true ]; then
    INSTALL_STATE=installed
    if ! service_exists app && [ "$REDEPLOY" != 1 ]; then
      die "HivePaaS was installed here, but its services are gone. Run the installer with --redeploy to deploy them again."
    fi
  elif service_exists app; then
    INSTALL_STATE=unfinished
  else
    INSTALL_STATE=fresh
  fi
}

running_image() {
  docker service inspect --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}' "${STACK}_$1" 2>/dev/null
}

# image_for SERVICE RELEASE_IMAGE: a running service keeps its image. Images
# are the updater's to move, with whatever a move needs, which a deploy does
# not do.
image_for() {
  local image
  image=$(running_image "$1") || image=
  printf '%s' "${image:-$2}"
}

take_db_volume() {
  case "$1" in
    HP_DB_VOLUME) if [[ $2 =~ ^[A-Za-z0-9_.-]+$ ]]; then HP_DB_VOLUME=$2; fi ;;
    HP_DB_MAJOR) if [[ $2 =~ ^[0-9]+$ ]]; then HP_DB_MAJOR=$2; fi ;;
  esac
}

resolve_images() {
  local app agent db_env="$HIVEPAAS_DATA_DIR/system/update/db-volume.env"
  app=$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" appImage)
  agent=${HIVEPAAS_AGENT_IMAGE:-$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" agentImage)}
  if [ -z "$agent" ]; then
    agent=$(derive_agent_image "$app") ||
      die "Cannot tell the agent image of $app; set HIVEPAAS_AGENT_IMAGE."
  fi
  HIVEPAAS_IMAGE_APP=$(image_for app "$app")
  HIVEPAAS_IMAGE_WORKER=$(image_for worker "$app")
  HIVEPAAS_IMAGE_UPDATER=$(image_for updater "$app")
  HIVEPAAS_IMAGE_AGENT=$(image_for agent "$agent")
  HIVEPAAS_IMAGE_DB=$(image_for db "$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" dbImage)")
  HIVEPAAS_IMAGE_REDIS=$(image_for redis "$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" redisImage)")
  HIVEPAAS_IMAGE_TRAEFIK=$(image_for traefik "$(release_field "$RELEASE_FILE" "$HIVEPAAS_CHANNEL" traefikImage)")
  # A Postgres major upgrade moves the database to a volume of its own, and
  # the updater records where. Deploying without it would put the database
  # back as it was before the upgrade.
  if [ -f "$db_env" ]; then read_kv_file "$db_env" take_db_volume; fi
  if [ -z "$HP_DB_MAJOR" ]; then
    HP_DB_MAJOR=$(pg_major_of_image "$HIVEPAAS_IMAGE_DB") ||
      die "Cannot tell the Postgres major version of $HIVEPAAS_IMAGE_DB."
  fi
}

# deploy_stack FIRST_BOOT: docker stack deploy with the settings in the
# environment it interpolates. FIRST_BOOT=1 adds the admin's account, which the
# app reads on its first boot only.
deploy_stack() {
  local -a files=(-c "$CONFIG_DIR/hivepaas.yaml")
  if [ "$1" = 1 ]; then files+=(-c "$CONFIG_DIR/hivepaas.first-boot.yaml"); fi
  (
    HIVEPAAS_APP_ENV=$(app_env_of_channel "$HIVEPAAS_CHANNEL")
    export HIVEPAAS_APP_ENV HIVEPAAS_APP_DOMAIN HIVEPAAS_ROOT_DOMAIN HIVEPAAS_APP_SECRET \
      HIVEPAAS_JWT_SECRET HIVEPAAS_DATA_DIR HIVEPAAS_PROJECT_DATA_DIR HIVEPAAS_DB_PASSWORD \
      HIVEPAAS_REDIS_PASSWORD HIVEPAAS_AGENT_TOKEN HIVEPAAS_IP_RULE HIVEPAAS_ADMIN_EMAIL \
      HIVEPAAS_ADMIN_PASSWORD HP_DB_MAJOR HIVEPAAS_IMAGE_APP HIVEPAAS_IMAGE_WORKER \
      HIVEPAAS_IMAGE_UPDATER HIVEPAAS_IMAGE_AGENT HIVEPAAS_IMAGE_DB HIVEPAAS_IMAGE_REDIS \
      HIVEPAAS_IMAGE_TRAEFIK
    if [ -n "$HP_DB_VOLUME" ]; then export HP_DB_VOLUME; fi
    docker stack deploy --with-registry-auth --detach=true "${files[@]}" "$STACK"
  ) || die "docker stack deploy failed; see the output above."
}

# run_migrations IMAGE: the app does not migrate its database on boot; the
# updater does, on an update. The first time is here: retried until the
# database takes connections, while the app restarts until the tables exist.
run_migrations() {
  local log="$WORK_DIR/migrate.log" start=$SECONDS limit="${HIVEPAAS_WAIT_SECONDS:-300}"
  info "Migrating the database, once it is up..."
  until HP_DB_PASSWORD=$HIVEPAAS_DB_PASSWORD docker run --rm --network "${STACK}_local_net" \
    -e HP_DB_HOST=db -e HP_DB_PORT=5432 -e HP_DB_USER=hivepaas -e HP_DB_PASSWORD -e HP_DB_DB_NAME=hivepaas \
    "$1" sql-migrate up -config=hivepaas_app/db/dbconfig.yml -env=main >"$log" 2>&1; do
    if [ $((SECONDS - start)) -ge "$limit" ]; then
      tail -n 20 "$log" >&2
      die "The database migrations did not run within $limit seconds. See 'docker service ps ${STACK}_db'" \
        "and 'docker service logs ${STACK}_db', then run the installer again."
    fi
    sleep 5
  done
  ok "Database migrated: $(tail -n 1 "$log")"
}

update_state() {
  docker service inspect "${STACK}_$1" --format '{{if .UpdateStatus}}{{.UpdateStatus.State}}{{end}}' 2>/dev/null
}

# wait_for_update SERVICE: until swarm is done updating the service, or has
# given up on the update.
wait_for_update() {
  local start=$SECONDS limit="${HIVEPAAS_WAIT_SECONDS:-300}"
  while [ $((SECONDS - start)) -lt "$limit" ]; do
    case "$(update_state "$1")" in
      updating | rollback_started) sleep 3 ;;
      *) return 0 ;;
    esac
  done
  return 1
}

task_template() {
  docker service inspect "${STACK}_$1" --format '{{json .Spec.TaskTemplate}}' 2>/dev/null
}

# redeploy_stack: docker stack deploy over a running installation. The deploy
# resets the OOM priority tune_services gives every service, so the database
# restarts along with the app; an app started while the database is down
# cannot connect, exits, and has swarm roll its update back. The stack is then
# deployed once more, with the database back.
redeploy_stack() {
  local old_task before after attempt
  for attempt in 1 2; do
    old_task=$(running_task app) || old_task=''
    before=$(task_template app) || before=''
    deploy_stack 0
    # Read at once: a rollback would put the old template back.
    after=$(task_template app) || after=''
    if ! { wait_for_update db && wait_for_update redis; }; then
      die "The database did not come back within ${HIVEPAAS_WAIT_SECONDS:-300} seconds. See" \
        "'docker service ps ${STACK}_db --no-trunc', then run the installer again."
    fi
    # A deploy that left the app's tasks as they were has nothing to wait for.
    if [ "$after" = "$before" ]; then return 0; fi
    wait_for_new_task app "$old_task" || die_waiting
    case "$(update_state app)" in
      rollback_*) ;;
      *) return 0 ;;
    esac
    if [ "$attempt" = 2 ]; then
      die "Swarm rolled back the app's update twice. See 'docker service ps ${STACK}_app --no-trunc'" \
        "and 'docker service logs ${STACK}_app', then run the installer again."
    fi
    info "The app's update was rolled back while the database restarted; deploying again..."
  done
}

deploy() {
  case "$INSTALL_STATE" in
    fresh)
      resolve_images
      deploy_stack 1
      run_migrations "$HIVEPAAS_IMAGE_APP"
      ;;
    unfinished)
      if [ "$REDEPLOY" = 1 ]; then
        resolve_images
        if [ -n "${HIVEPAAS_ADMIN_PASSWORD:-}" ]; then deploy_stack 1; else deploy_stack 0; fi
      else
        info "Deployed by an earlier run; finishing what it started."
      fi
      run_migrations "$(running_image app)"
      ;;
    installed)
      if [ "$REDEPLOY" = 1 ]; then
        resolve_images
        redeploy_stack
      else
        ok "Deployed already, and left as it is (--redeploy deploys it again)."
        update_ip_routes
      fi
      ;;
  esac
}

# update_ip_routes: the routers by address follow an address that changed.
# They are service labels, so changing them restarts nothing. Without the
# public address this run, they are left alone rather than lose it.
update_ip_routes() {
  local current
  if [ -z "$PUBLIC_IP" ]; then
    warn "Could not look up this server's public address; the routes by address are left as they are."
    return 0
  fi
  current=$(docker service inspect "${STACK}_app" \
    --format '{{index .Spec.Labels "traefik.http.routers.x-custom-router-ip.rule"}}' 2>/dev/null) || return 0
  if [ "$current" = "$HIVEPAAS_IP_RULE" ]; then return 0; fi
  docker service update --detach --quiet \
    --label-add "traefik.http.routers.x-custom-router-ip.rule=$HIVEPAAS_IP_RULE" \
    --label-add "traefik.http.routers.x-custom-router-ip-http.rule=$HIVEPAAS_IP_RULE" \
    "${STACK}_app" >/dev/null || die "Could not update the routes by address."
  ok "Routes by address now: $HIVEPAAS_IP_RULE"
}

# -------------------------------------------------------------------- Waits

dashboard_answers() {
  curl -fsSk --max-time 5 --resolve "$HIVEPAAS_APP_DOMAIN:443:127.0.0.1" \
    "https://$HIVEPAAS_APP_DOMAIN/_/ping" >/dev/null 2>&1
}

wait_for_dashboard() {
  local start=$SECONDS limit="${HIVEPAAS_WAIT_SECONDS:-300}" said=0
  until dashboard_answers; do
    if [ $((SECONDS - start)) -ge "$limit" ]; then return 1; fi
    if [ $((SECONDS - start)) -ge $((said + 30)) ]; then
      said=$((SECONDS - start))
      info "Still waiting (${said}s)..."
    fi
    sleep 5
  done
}

die_waiting() {
  die "HivePaaS did not answer within ${HIVEPAAS_WAIT_SECONDS:-300} seconds. See what it is doing with" \
    "'docker service ps ${STACK}_app --no-trunc' and 'docker service logs ${STACK}_app', then run the" \
    "installer again."
}

# running_task SERVICE: the ID of the service's task that is running, if any.
running_task() {
  docker service ps "${STACK}_$1" --filter desired-state=running --format '{{.ID}} {{.CurrentState}}' \
    2>/dev/null | awk '$2 == "Running" {print $1; exit}'
}

# wait_for_new_task SERVICE OLD_TASK: until a task other than OLD_TASK runs.
# A task with a healthcheck only counts as running once it is healthy.
wait_for_new_task() {
  local start=$SECONDS limit="${HIVEPAAS_WAIT_SECONDS:-300}" task
  while [ $((SECONDS - start)) -lt "$limit" ]; do
    task=$(running_task "$1") || task=
    if [ -n "$task" ] && [ "$task" != "$2" ]; then return 0; fi
    sleep 3
  done
  return 1
}

# service_update_args SERVICE: what the service still needs, into UPDATE_ARGS:
# the OOM priority of a system service, and the admin's account out of its
# environment.
service_update_args() {
  local svc="${STACK}_$1" oom env key nl=$'\n'
  UPDATE_ARGS=()
  oom=$(docker service inspect --format '{{.Spec.TaskTemplate.ContainerSpec.OomScoreAdj}}' "$svc" 2>/dev/null) ||
    return 1
  if [ "$oom" != -500 ]; then UPDATE_ARGS+=(--oom-score-adj -500); fi
  env=$(docker service inspect --format '{{range .Spec.TaskTemplate.ContainerSpec.Env}}{{println .}}{{end}}' "$svc")
  for key in HP_USER_ADMIN_USERNAME HP_USER_ADMIN_EMAIL HP_USER_ADMIN_PASSWORD; do
    case "$nl$env" in *"$nl$key="*) UPDATE_ARGS+=(--env-rm "$key") ;; esac
  done
}

# tune_services: the system services get the kernel's OOM priority -500, so
# that a user app is what the kernel kills when memory runs out - `docker stack
# deploy` drops oom_score_adj - and the admin's password leaves the app and
# the worker once the app has used it.
#
# Only what differs is changed, one service at a time, each waited for: an app
# restarted while the database restarts cannot connect, exits, and has swarm
# roll its update back. The database goes first and the app last. An update
# swarm rolled back anyway stops the install rather than pass for done.
tune_services() {
  local svc old_task
  for svc in db redis traefik agent worker updater app; do
    service_update_args "$svc" || continue
    if [ "${#UPDATE_ARGS[@]}" -eq 0 ]; then continue; fi
    old_task=$(running_task "$svc") || old_task=''
    info "Updating ${STACK}_$svc..."
    docker service update --detach --quiet "${UPDATE_ARGS[@]}" "${STACK}_$svc" >/dev/null ||
      die "Could not update ${STACK}_$svc."
    if [ -n "$old_task" ]; then
      wait_for_new_task "$svc" "$old_task" ||
        die "${STACK}_$svc did not come back within ${HIVEPAAS_WAIT_SECONDS:-300} seconds. See" \
          "'docker service ps ${STACK}_$svc --no-trunc', then run the installer again."
    fi
    service_update_args "$svc" || true
    if [ "${#UPDATE_ARGS[@]}" -gt 0 ]; then
      die "Swarm rolled back the update of ${STACK}_$svc: its new task did not stay up. See" \
        "'docker service ps ${STACK}_$svc --no-trunc', then run the installer again."
    fi
  done
  wait_for_dashboard || die_waiting
}

# ------------------------------------------------------------------- Finish

finish_install() {
  HIVEPAAS_ADMIN_PASSWORD=
  HIVEPAAS_INSTALLED=true
  save_settings
  ok "The admin's password is out of $INSTALL_ENV and of the services' settings."
}

print_done() {
  local ips=''
  print_logo
  printf '\n  %s%sHivePaaS is running.%s\n\n' "$C_BOLD" "$C_GREEN" "$C_RESET"
  info "Dashboard   https://$HIVEPAAS_APP_DOMAIN"
  info "            once DNS points $HIVEPAAS_APP_DOMAIN at this server; point *.$HIVEPAAS_ROOT_DOMAIN"
  info "            at it too, for your apps."
  if [ -n "$PUBLIC_IP" ]; then ips=" https://$PUBLIC_IP"; fi
  if [ -n "$HOST_IP" ] && [ "$HOST_IP" != "$PUBLIC_IP" ]; then ips="$ips https://$HOST_IP"; fi
  if [ -n "$ips" ]; then info "            Until then:$ips"; fi
  if [ "$1" = 1 ]; then
    info "Sign in     as $ADMIN_USERNAME ($HIVEPAAS_ADMIN_EMAIL), with the password you chose."
  fi
  printf '\n'
  warn "The certificate is self-signed until you get one in the setup, so the browser warns"
  info "  about it: accept the warning to go on."
  warn "After a restart, HivePaaS can take 30 to 60 seconds to answer."
  warn "$INSTALL_ENV holds the app secret, the only key to your encrypted data."
  info "  Keep a copy of it somewhere other than this server."
  printf '\n'
}

# --------------------------------------------------------------------- Main

usage() {
  cat <<'USAGE'
Usage: install.sh [--yes] [--config FILE] [--redeploy]

Installs HivePaaS on this server, as root. Every answer is saved in
/etc/hivepaas/install.env; running it again finishes an interrupted install
or re-checks the host.

Options:
  -y, --yes        answer yes to installing or upgrading Docker and to the
                   final confirmation; with every setting given, nothing is asked
  --config FILE    read settings from FILE: KEY=VALUE lines, never executed
  --redeploy       deploy the stack again on an installed server. What HivePaaS
                   changed on its own services since goes back to the stack file.
  -h, --help       show this help

Settings, from the environment or --config (the environment wins):
  HIVEPAAS_ADMIN_EMAIL         the admin's email
  HIVEPAAS_ADMIN_PASSWORD      the admin's password, 10 characters or more
  HIVEPAAS_APP_DOMAIN          the dashboard's domain, e.g. hivepaas.example.com
  HIVEPAAS_ROOT_DOMAIN         the domain apps get subdomains of (default: from the app domain)
  HIVEPAAS_APP_SECRET          the key stored secrets are encrypted with, 32 characters
                               or more (default: generated)
  HIVEPAAS_DATA_DIR            HivePaaS's data (default: /var/lib/hivepaas)
  HIVEPAAS_PROJECT_DATA_DIR    projects' data (default: <data dir>/project_data)
  HIVEPAAS_CHANNEL             beta or stable (default: beta)
  HIVEPAAS_SWAP=false          do not add a swap file
  HIVEPAAS_SWAP_SIZE_MB        the swap file's size (default: 2048)
  HIVEPAAS_EARLYOOM=false      do not install earlyoom
  HIVEPAAS_UPGRADE_DOCKER=true upgrade Docker to its latest release without asking
  HIVEPAAS_AGENT_IMAGE         the agent's image (default: from the release)
  HIVEPAAS_RELEASE_BRANCH      the branch the release info is read from (default: release)
  HIVEPAAS_INSTALL_REF         the ref the stack files are downloaded from (default: main)

Silent install:
  curl -fsSL .../install.sh | sudo HIVEPAAS_ADMIN_EMAIL=... HIVEPAAS_ADMIN_PASSWORD=... \
    HIVEPAAS_APP_DOMAIN=... bash -s -- --yes
USAGE
}

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      -y | --yes) ASSUME_YES=1 ;;
      --redeploy) REDEPLOY=1 ;;
      --config)
        if [ $# -lt 2 ]; then die "--config needs a file."; fi
        CONFIG_FILE=$2
        shift
        ;;
      --config=*) CONFIG_FILE=${1#--config=} ;;
      -h | --help)
        usage
        exit 0
        ;;
      *) die "Unknown option: $1. See --help." ;;
    esac
    shift
  done
}

cleanup() {
  if [ -n "$WORK_DIR" ]; then rm -rf "$WORK_DIR"; fi
}

main() {
  local just_installed=0
  set -Eeuo pipefail
  trap 'on_error $? $LINENO' ERR
  trap cleanup EXIT
  parse_args "$@"
  setup_colors
  print_logo
  open_tty
  WORK_DIR=$(mktemp -d)

  step "Preflight"
  preflight

  step "Docker"
  ensure_docker

  step "Settings"
  load_settings
  detect_install_state
  ask_questions
  generate_secrets
  gather_addresses
  if [ "$INSTALL_STATE" = fresh ] || [ "$REDEPLOY" = 1 ]; then fetch_release; fi
  if [ "$INSTALL_STATE" = installed ]; then
    ok "HivePaaS is installed here, with the settings in $INSTALL_ENV."
    if [ "$REDEPLOY" = 1 ]; then
      warn "--redeploy puts HivePaaS's services back as the stack file has them: Traefik's settings, the"
      info "  worker and updater replicas and the app's routing labels, if HivePaaS changed them since."
      confirm "Redeploy the stack?" || die "Stopped; the stack was not redeployed."
    fi
  else
    print_summary
    confirm "Install HivePaaS with these settings?" || die "Stopped; HivePaaS was not installed."
  fi
  save_settings
  ok "Settings saved in $INSTALL_ENV."

  step "Host memory"
  setup_swap || true
  setup_earlyoom || true

  step "Swarm"
  ensure_swarm
  ensure_network

  step "Files"
  prepare_files

  step "Deploy"
  deploy

  step "Waiting for HivePaaS"
  wait_for_dashboard || die_waiting
  tune_services
  ok "HivePaaS answers."

  step "Finish"
  if [ "${HIVEPAAS_INSTALLED:-}" != true ]; then
    finish_install
    just_installed=1
  fi
  print_done "$just_installed"
}

if [ "${HIVEPAAS_INSTALL_LIB:-}" != 1 ]; then
  main "$@"
fi
