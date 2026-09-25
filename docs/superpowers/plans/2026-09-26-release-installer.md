# Release Installer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `deployment/release/install.sh`, the installer a person runs on their own Linux server (`curl -fsSL .../install.sh | sudo bash`) to get HivePaaS's beta channel - Docker, a single-node swarm, the stack - interactively or silently, and to run again safely.

**Architecture:** One bash script of functions, with `main` called on the last line, so a test can source it (`HIVEPAAS_INSTALL_LIB=1`) and call the functions alone. It saves every answer in `/etc/hivepaas/install.env`, reads the images from the signed release info, deploys `deployment/release/hivepaas.yaml` - with `hivepaas.first-boot.yaml` and the admin's account on the first install only - migrates the database, waits for the dashboard, then sets each system service's OOM priority and removes the admin's password, one service at a time. The pure functions and the questions are tested by `install_test.sh`; the Docker paths by `install_e2e.sh`, a whole install in docker:dind.

**Tech Stack:** bash (the tests run under macOS's bash 3.2 and bash 5; servers have 4.4 or later), Docker Engine 29.5+ in swarm mode with `docker stack deploy`, jq, openssl, curl; shellcheck from `koalaman/shellcheck:stable`; `docker:dind` and `bash:5.2` images for the tests.

**Spec:** `docs/superpowers/specs/2026-09-25-release-installer-design.md` - amended in the same commit as this plan with what planning found (§2 steps 7-8, §3, §5, §6, §7, §10, §11).

## Global Constraints

- Docker Engine `>= 29.5` and API `>= 1.54`. Missing or older: ask to install or upgrade. At least that but older than the latest release: ask, default no.
- Channel `beta` by default (`HIVEPAAS_CHANNEL`); `beta` runs `config/config.beta.toml` (`HP_ENV=beta`), `stable` runs `config/config.production.toml` (`HP_ENV=production`).
- Release info from `https://raw.githubusercontent.com/hivepaas/hivepaas/<HIVEPAAS_RELEASE_BRANCH:-release>/release.signed.json`; the `sha256` of the decoded payload is checked, the signatures are not.
- Admin username `admin`; email checked; password at least 10 characters, typed twice, not echoed.
- App secret: given, at least 32 characters without spaces; Enter generates 64 hex. JWT secret: 32 generated alphanumeric characters. Database password, Redis password, agent token: 32 generated alphanumeric characters each.
- App data defaults to `/var/lib/hivepaas`; project data to `<app data>/project_data`.
- `/etc/hivepaas/install.env`: directory `0700`, file `0600`, replaced atomically, values in single quotes, read as data and never sourced. The environment wins over `--config`, `--config` over `install.env`; the fixed settings cannot change once saved.
- Nothing destructive: never leave or reset a swarm, never remove a service, network or volume, never generate a secret a second time.
- Swap, `vm.swappiness` and earlyoom never fail the install.
- Questions are read from `/dev/tty` (fd 3), never from stdin.
- Colour only when stdout is a terminal, `NO_COLOR` is unset and `TERM` is not `dumb`.
- shellcheck clean on `install.sh`, `install_test.sh` and `install_e2e.sh` (the two test files carry file-level `disable` lines for intentional literal `$`).
- The local Docker Desktop runs the user's own HivePaaS swarm: nothing in this plan deploys to it, updates its services or touches its swarm. Every Docker-level test runs inside the `hivepaas-install-e2e` docker:dind container, removed afterwards. The installer itself is never run on the Mac.
- Work on branch `feat/release-installer`; merge into `main` locally with `--no-ff`; never push. Every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- **`install.env` lost while the stack or its database volume is still there.** The installer must stop and ask for the file back, never generate new secrets the database would refuse. Pinned by the e2e check "a run without install.env stops" (Task 7).
- **`--redeploy` while the database restarts.** A deploy resets every service's OOM priority, so the database restarts with the app, the app cannot connect, exits, and swarm rolls its update back; the installer must deploy again once the database is back, and must not wait for an app a deploy left alone. Pinned by the e2e `--redeploy` checks (Task 7); the run that planning made took the rolled-back branch.
- **A setting given again in another form** - a trailing slash, doubled slashes, capitals in a domain. It must not pass for a change of a fixed setting. Pinned by `test_load_settings_fixed` (Task 4).
- **A config or `db-volume.env` file that tries to run something** - `$(...)` in a key or a value, `PATH=` or any name the installer does not own. Everything must be taken as data. Pinned by `test_load_settings_runs_nothing` (Task 4) and `test_take_db_volume` (Task 6).
- **A server whose routes cover `10.11.0.0/16`** (a VPN, a `10.0.0.0/8` VPC route). The network must fall back to a subnet Docker picks. Pinned by `test_routes_overlap` and `test_cidr_overlaps` (Task 3); the fallback call itself runs only on such a server.

## Before you start

- **Repository:** `/Users/tnt/go/src/github.com/hivepaas/hivepaas`, on `main`. Every command below runs from there.
- **Why the file is built in sections.** Each task appends whole sections to `install.sh`, each headed `# ---...--- <Name>` and declaring the globals it owns at its top, so every task leaves a file shellcheck passes: a global declared before the task that reads it would be reported unused. Tests go into `install_test.sh` the same way, above its `Runner` section, which stays last.
- **The tests** source `install.sh` with `HIVEPAAS_INSTALL_LIB=1` and run each `test_` function in a subshell; a check writes `pass` or `fail` to a results file, and the runner prints `N passed, M failed` and exits non-zero on a failure. They must pass under macOS's `/bin/bash` 3.2 (no associative arrays, no `mapfile`, no `${var,,}`) and under bash 5.
- **What was verified while planning** (the code below is that code): shellcheck clean after every task; each task's new tests fail on the previous task's `install.sh` and pass on its own; the final suite passes under bash 3.2 and under `bash:5.2`; and `install_e2e.sh` passes end to end on this Mac, the HivePaaS images (amd64 only) running emulated inside an arm64 docker:dind.
- **Found while planning, and handled in the code:** the app panics at boot with no CORS origin list (the stack sets `["*"]`, as dev does); the app exits at boot when the database is unreachable, which makes swarm roll back an app update that runs while the database restarts (the service updates are sequenced, and a redeploy retries); `docker stack deploy` drops `oom_score_adj` in Docker 29.8 too.

## Files

| file | task | what |
|---|---|---|
| `deployment/release/install.sh` | 1-6 | the installer |
| `deployment/release/install_test.sh` | 1-6 | tests of its functions and questions |
| `deployment/release/hivepaas.yaml` | 3 | the stack |
| `deployment/release/hivepaas.first-boot.yaml` | 3 | the admin's account, first install only |
| `deployment/release/traefik/dynamic_conf.yml` | 3 | Traefik's default certificate, as dev's |
| `deployment/release/install_e2e.sh` | 7 | a whole install in docker:dind |
| `Makefile` | 1, 7 | `test-installer`, `test-installer-e2e` |

---

### Task 1: The installer's skeleton, its checks, and the test harness

**Files:**
- Create: `deployment/release/install.sh`
- Create: `deployment/release/install_test.sh`
- Modify: `Makefile` (a new target after `test-cover`)

**Interfaces:**
- Consumes: nothing.
- Produces, in `install.sh`:
  - output: `setup_colors`; `info MSG...`, `ok MSG...` (stdout); `warn MSG...` (stderr); `die MSG...` (stderr, `exit 1`); `step TITLE` (`[n/9] TITLE`, globals `STEP_NO`, `STEP_TOTAL=9`); `on_error STATUS LINE` (the ERR trap: prints the line, exits; quiet in a subshell); `print_logo`; colour globals `C_RESET C_BOLD C_RED C_GREEN C_YELLOW C_BLUE C_LOGO`, empty until `setup_colors`.
  - checks, each returning 0/1: `valid_email S`, `valid_password S`, `valid_hostname S` (lowercase only), `valid_domain S` (any case), `valid_root_domain ROOT` (reads `HIVEPAAS_APP_DOMAIN`), `valid_data_dir DIR`, `valid_project_data_dir DIR` (reads `HIVEPAAS_DATA_DIR`), `valid_app_secret S`, `valid_ipv4 S`, `version_ge A B`.
  - values on stdout: `lowercase S`, `root_domain_of DOMAIN`, `normalize_dir DIR`, `rand_alnum N`.
- Produces, in `install_test.sh`: `record pass|fail`, `check NAME WANT GOT`, `check_ok NAME CMD...`, `check_fails NAME CMD...`, `check_contains NAME TEXT PART`, `check_lacks NAME TEXT PART`, `matches TEXT REGEX`; globals `HERE` (the test's directory) and `TMP` (a scratch directory, removed on exit).

- [ ] **Step 1: Branch**

```bash
git checkout main && git checkout -b feat/release-installer
```

- [ ] **Step 2: Write the failing tests**

Create `deployment/release/install_test.sh`:

```bash
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
```

- [ ] **Step 3: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - `install.sh: No such file or directory`, then `cannot load .../deployment/release/install.sh`, exit status 1.

- [ ] **Step 4: Write the skeleton and the checks**

Create `deployment/release/install.sh`, and `chmod +x deployment/release/install.sh deployment/release/install_test.sh`. The logo is the user's, character for character (trailing spaces do not matter):

```bash
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
```

- [ ] **Step 5: Run the tests**

Run: `bash deployment/release/install_test.sh`
Expected: `103 passed, 0 failed`

- [ ] **Step 6: Add the Makefile target**

In `Makefile`, after the `test-cover` target (its recipe is `@./scripts/test.sh`), add - recipes are indented with a tab:

```make
# The release installer (deployment/release): shellcheck, then the tests of
# its functions under this machine's bash.
test-installer:
	@docker run --rm -v "$(PWD)/deployment/release":/mnt:ro -w /mnt koalaman/shellcheck:stable -x install.sh install_test.sh
	@bash deployment/release/install_test.sh
```

Run: `make test-installer`
Expected: no shellcheck output, then `103 passed, 0 failed`.

- [ ] **Step 7: Commit**

```bash
git add deployment/release/install.sh deployment/release/install_test.sh Makefile
git commit -m "feat(installer): the release installer's checks, and its test harness" -m "The start of deployment/release/install.sh: output, and the checks of what a person types. install_test.sh sources it and runs under bash 3.2 as under bash 5; make test-installer runs it after shellcheck." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: Settings files and the release info

**Files:**
- Modify: `deployment/release/install.sh` (append two sections)
- Modify: `deployment/release/install_test.sh` (two sections above `Runner`)

**Interfaces:**
- Consumes: `warn` (Task 1).
- Produces:
  - `kv_quote VALUE` - VALUE in single quotes, `'` escaped as `'\''`; `kv_unquote RAW` - the value of a `KEY=VALUE` line, from single or double quotes or bare, never expanded.
  - `read_kv_file FILE FN` - calls `FN KEY VALUE` for each `KEY=VALUE` line; skips blank lines and `#` comments, drops `export ` and a trailing `\r`, trims spaces around the key and value, and warns `FILE line N is not KEY=VALUE; ignored.` for anything else.
  - `sha256_of` (stdin to hex); `release_payload FILE` - the decoded payload of a `release.signed.json` on stdout, return 2 when it does not match its `sha256`, 1 when the file is no envelope; `release_field FILE CHANNEL FIELD` (empty when missing); `release_channels FILE` (`beta, stable`).
  - `app_env_of_channel CHANNEL` - `beta` or `production`, return 1 otherwise.
  - `image_name IMAGE`, `image_tag IMAGE` (digest dropped, a registry port is no tag); `derive_agent_image APP_IMAGE` (return 1 for an image that is not `hivepaas...`); `pg_major_of_image IMAGE` (return 1 without a numeric tag).
  - `docker_static_arch MACHINE` (`x86_64`, `aarch64`, `armhf`, `ppc64le`, `s390x`, else return 1); `latest_docker_from_index` - the newest `docker-X.Y.Z.tgz` of an index page on stdin, return 1 when none.

- [ ] **Step 1: Write the failing tests**

In `deployment/release/install_test.sh`, insert above the `# ---...--- Runner` line, with one blank line after it:

```bash
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
```

- [ ] **Step 2: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - the new checks fail with `command not found` for `kv_quote`, `release_payload` and the rest (`114 passed, 40 failed` when planning).

- [ ] **Step 3: Write the code**

Append to `deployment/release/install.sh`, after one blank line:

```bash
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
```

- [ ] **Step 4: Run the tests**

Run: `make test-installer`
Expected: no shellcheck output, then `154 passed, 0 failed`.

- [ ] **Step 5: Commit**

```bash
git add deployment/release/install.sh deployment/release/install_test.sh
git commit -m "feat(installer): settings files, and the signed release info" -m "KEY=VALUE files are read as data, never sourced. The release info is decoded and its sha256 checked; the images, the agent image derived from the app image, and the Postgres major come from it." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The stack, and the routes by the server's addresses

**Files:**
- Create: `deployment/release/hivepaas.yaml`
- Create: `deployment/release/hivepaas.first-boot.yaml`
- Create: `deployment/release/traefik/dynamic_conf.yml`
- Modify: `deployment/release/install.sh` (append the `Network` section)
- Modify: `deployment/release/install_test.sh` (two sections above `Runner`)

**Interfaces:**
- Consumes: `valid_ipv4` (Task 1).
- Produces:
  - `route_src` (the `src` of `ip route get` output on stdin); `host_ip` (the default route's source address, return 1 without one); `public_ip` (from `https://ifconfig.io`, return 1 on failure); `ip4_to_int A.B.C.D`; `cidr_overlaps A B` (an address without `/n` is a `/32`); `routes_overlap CIDR` (`ip route show` output on stdin; routes of a type such as `unreachable` count); `ip_host_rule IP...` (``Host(`a`) || Host(`b`)``, empty and repeated addresses skipped, ``Host(`127.0.0.1`)`` when none is left); `addresses_line` (reads `HIVEPAAS_APP_DOMAIN`, `PUBLIC_IP`, `HOST_IP`); globals `HOST_IP`, `PUBLIC_IP`.
  - The stack files interpolate exactly these, which Task 6's `deploy_stack` exports: `HIVEPAAS_APP_ENV`, `HIVEPAAS_ROOT_DOMAIN`, `HIVEPAAS_APP_DOMAIN`, `HIVEPAAS_APP_SECRET`, `HIVEPAAS_JWT_SECRET`, `HIVEPAAS_DATA_DIR`, `HIVEPAAS_PROJECT_DATA_DIR`, `HIVEPAAS_DB_PASSWORD`, `HIVEPAAS_REDIS_PASSWORD`, `HIVEPAAS_AGENT_TOKEN`, `HIVEPAAS_IP_RULE`, `HIVEPAAS_ADMIN_EMAIL`, `HIVEPAAS_ADMIN_PASSWORD`, `HP_DB_MAJOR`, `HP_DB_VOLUME` (optional), `HIVEPAAS_IMAGE_APP`, `HIVEPAAS_IMAGE_WORKER`, `HIVEPAAS_IMAGE_UPDATER`, `HIVEPAAS_IMAGE_AGENT`, `HIVEPAAS_IMAGE_DB`, `HIVEPAAS_IMAGE_REDIS`, `HIVEPAAS_IMAGE_TRAEFIK`.

The stack is `deployment/dev/hivepaas.yaml` with the spec's §5 differences. Two of them came out of planning and are not optional: `HP_HTTP_SERVER_CORS_ALLOW_ORIGINS: '["*"]'` (without an origin list the app panics at boot: `conflict settings: all origins disabled`), and `.service=x-custom-svc-app` on every `x-custom` router (with a second HTTP service on the app, Traefik will not pick one for a router that names none). `docker stack config` renders it without touching any swarm.

- [ ] **Step 1: Write the failing tests**

In `deployment/release/install_test.sh`, insert above the `# ---...--- Runner` line, with one blank line after it:

```bash
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
```

- [ ] **Step 2: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - `route_src`, `cidr_overlaps` and the rest are not found, and `docker stack config` cannot open `hivepaas.yaml`.

- [ ] **Step 3: Write the stack files**

Create `deployment/release/hivepaas.yaml`:

```yaml
# The HivePaaS stack as install.sh deploys it.
#
# Every value that differs between servers is a ${VAR} the installer exports
# before `docker stack deploy`: the answers saved in /etc/hivepaas/install.env
# and the images of the release it installs. `docker stack config -c
# hivepaas.yaml` shows what a set of them renders to.
#
# The healthchecks, update_config and restart policies are dev's, and
# deployment/dev/hivepaas.yaml says why each number is what it is. Keep the two
# in step.

# The environment every HivePaaS process loads its configuration from: app,
# worker, updater and agent all read the database, the cache and the secrets.
x-hivepaas-env: &hivepaas-env
  HP_ENV: ${HIVEPAAS_APP_ENV:?}
  HP_CONFIG_FILE: config/config.${HIVEPAAS_APP_ENV:?}.toml
  HP_ROOT_DOMAIN: ${HIVEPAAS_ROOT_DOMAIN:?}
  HP_APP_DOMAIN: ${HIVEPAAS_APP_DOMAIN:?}
  HP_APP_SECRET: ${HIVEPAAS_APP_SECRET:?}
  HP_SESSION_JWT_SECRET: ${HIVEPAAS_JWT_SECRET:?}
  HP_STORAGE_HOST_DIR: ${HIVEPAAS_DATA_DIR:?}
  HP_STORAGE_PROJECT_DATA_HOST_DIR: ${HIVEPAAS_PROJECT_DATA_DIR:?}
  HP_DB_HOST: db
  HP_DB_PORT: "5432"
  HP_DB_USER: hivepaas
  HP_DB_PASSWORD: ${HIVEPAAS_DB_PASSWORD:?}
  HP_DB_DB_NAME: hivepaas
  # The database is on the stack's own overlay network, not reachable from outside.
  HP_DB_SSL_MODE: disable
  HP_CACHE_URL: redis://default:${HIVEPAAS_REDIS_PASSWORD:?}@redis:6379/0
  HP_AGENT_SECRET_TOKEN: ${HIVEPAAS_AGENT_TOKEN:?}
  # The app will not start with no origin list. The dashboard is served by the
  # app itself, at the app domain and at the server's addresses, and its
  # session cookies are SameSite=Lax, which a cross-site page cannot send.
  HP_HTTP_SERVER_CORS_ALLOW_ORIGINS: '["*"]'

x-logging: &logging
  driver: json-file
  options:
    max-size: 50m
    max-file: "5"
    compress: "true"

services:
  traefik:
    image: ${HIVEPAAS_IMAGE_TRAEFIK:?}
    command:
      - "--providers.swarm=true"
      - "--providers.swarm.watch=true"
      - "--providers.swarm.network=hivepaas_net"
      - "--providers.swarm.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.web.allowacmebypass=true"
      - "--entrypoints.websecure.address=:443"
      - "--entrypoints.websecure.http.tls=true"
      - "--entrypoints.ping.address=127.0.0.1:8082"
      - "--ping=true"
      - "--ping.entrypoint=ping"
      - "--providers.file.directory=/etc/traefik/dynamic"
      - "--providers.file.watch=true"
      - "--log=true"
      - "--log.level=INFO"
      - "--accesslog=true"
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://127.0.0.1:8082/ping"]
      start_period: 60s
      start_interval: 5s
      interval: 30s
      timeout: 5s
      retries: 3
    networks:
      hivepaas_net:
        aliases:
          - hivepaas_traefik
    ports:
      - target: 80
        published: 80
        protocol: tcp
        mode: host
      - target: 443
        published: 443
        protocol: tcp
        mode: host
    volumes:
      - ${HIVEPAAS_DATA_DIR:?}/traefik/etc:/etc/traefik
      - ${HIVEPAAS_DATA_DIR:?}/traefik/var/log:/var/log/traefik
      - ${HIVEPAAS_DATA_DIR:?}/ssl/certs:/etc/traefik/ssl/certs
      - /var/run/docker.sock:/var/run/docker.sock:ro
    deploy:
      labels:
        - "hivepaas.app.info={\"name\":\"traefik\",\"key\":\"traefik\"}"
      mode: replicated
      replicas: 1
      placement:
        constraints:
          - node.role == manager
          - node.labels.hivepaas.role == control-plane
      restart_policy:
        condition: any
        delay: 5s
        window: 120s
      update_config:
        failure_action: rollback
        monitor: 180s
        max_failure_ratio: 0.5
    logging: *logging

  db:
    image: ${HIVEPAAS_IMAGE_DB:?}
    networks:
      - local_net
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U hivepaas"]
      interval: 10s
      timeout: 10s
      retries: 6
      start_period: 60s
    environment:
      POSTGRES_DB: hivepaas
      POSTGRES_USER: hivepaas
      POSTGRES_PASSWORD: ${HIVEPAAS_DB_PASSWORD:?}
      PGDATA: /var/lib/postgresql/${HP_DB_MAJOR:?}/docker
    volumes:
      - db:/var/lib/postgresql
    deploy:
      labels:
        - "hivepaas.app.info={\"name\":\"db\",\"key\":\"db\"}"
        - "traefik.enable=false"
      mode: replicated
      replicas: 1
      placement:
        constraints:
          - node.role == manager
          - node.labels.hivepaas.role == control-plane
      restart_policy:
        condition: any
        delay: 5s
        window: 120s
    logging: *logging

  redis:
    image: ${HIVEPAAS_IMAGE_REDIS:?}
    command: ["redis-server", "--save", "", "--appendonly", "no", "--requirepass", "${HIVEPAAS_REDIS_PASSWORD:?}"]
    networks:
      - local_net
    deploy:
      labels:
        - "hivepaas.app.info={\"name\":\"redis\",\"key\":\"redis\"}"
        - "traefik.enable=false"
      mode: replicated
      replicas: 1
      placement:
        constraints:
          - node.role == manager
          - node.labels.hivepaas.role == control-plane
      restart_policy:
        condition: any
        delay: 5s
        window: 120s
    logging: *logging

  app:
    image: ${HIVEPAAS_IMAGE_APP:?}
    stop_grace_period: 10m
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://127.0.0.1:10000/_/ping"]
      start_period: 120s
      start_interval: 5s
      interval: 30s
      timeout: 5s
      retries: 3
    networks:
      hivepaas_net:
        aliases:
          - hivepaas_app
      local_net:
    environment:
      <<: *hivepaas-env
    volumes:
      - ${HIVEPAAS_DATA_DIR:?}:/var/lib/hivepaas
      - ${HIVEPAAS_DATA_DIR:?}:/host${HIVEPAAS_DATA_DIR:?}
      - ${HIVEPAAS_PROJECT_DATA_DIR:?}:/host${HIVEPAAS_PROJECT_DATA_DIR:?}
      - /var/run/docker.sock:/var/run/docker.sock
    deploy:
      labels:
        - "hivepaas.app.info={\"name\":\"app\",\"key\":\"app\"}"
        - "traefik.enable=true"
        - "traefik.swarm.network=hivepaas_net"
        # By domain, as dev. The app rewrites these when its routing settings
        # are saved.
        - "traefik.http.services.app-svc.loadbalancer.server.port=10000"
        - "traefik.http.routers.app-router.rule=Host(`${HIVEPAAS_APP_DOMAIN:?}`)"
        - "traefik.http.routers.app-router.entrypoints=websecure"
        - "traefik.http.routers.app-router.tls=true"
        - "traefik.http.routers.app-router.service=app-svc"
        - "traefik.http.middlewares.app-router-compress.compress=true"
        - "traefik.http.middlewares.app-router-compress.compress.defaultencoding=br"
        - "traefik.http.middlewares.app-router-compress.compress.minresponsebodybytes=1024"
        - "traefik.http.routers.app-router.middlewares=app-router-compress@swarm"
        # Everything named x-custom- survives that rewrite (updateSwarmServiceLabels).
        # Each router names its service: with two on the service, Traefik
        # would not pick one for a router that leaves it out.
        - "traefik.http.services.x-custom-svc-app.loadbalancer.server.port=10000"
        - "traefik.http.routers.x-custom-router-acme.rule=PathPrefix(`/.well-known/acme-challenge/`)"
        - "traefik.http.routers.x-custom-router-acme.entrypoints=web"
        - "traefik.http.routers.x-custom-router-acme.tls=false"
        - "traefik.http.routers.x-custom-router-acme.service=x-custom-svc-app"
        - "traefik.http.middlewares.x-custom-addprefix-acme.addprefix.prefix=/acme"
        - "traefik.http.routers.x-custom-router-acme.middlewares=x-custom-addprefix-acme@swarm"
        # By this server's addresses, until DNS points the app domain at it.
        # install.sh updates the rule when an address changes.
        - "traefik.http.routers.x-custom-router-ip.rule=${HIVEPAAS_IP_RULE:?}"
        - "traefik.http.routers.x-custom-router-ip.entrypoints=websecure"
        - "traefik.http.routers.x-custom-router-ip.tls=true"
        - "traefik.http.routers.x-custom-router-ip.service=x-custom-svc-app"
        - "traefik.http.routers.x-custom-router-ip-http.rule=${HIVEPAAS_IP_RULE:?}"
        - "traefik.http.routers.x-custom-router-ip-http.entrypoints=web"
        - "traefik.http.routers.x-custom-router-ip-http.service=x-custom-svc-app"
        - "traefik.http.routers.x-custom-router-ip-http.middlewares=x-custom-redirect-https@swarm"
        - "traefik.http.middlewares.x-custom-redirect-https.redirectscheme.scheme=https"
      mode: replicated
      replicas: 1
      placement:
        constraints:
          - node.role == manager
          - node.labels.hivepaas.role == control-plane
      update_config:
        failure_action: rollback
        monitor: 240s
        max_failure_ratio: 0.5
      restart_policy:
        condition: any
        delay: 5s
        window: 120s
    logging: *logging

  worker:
    image: ${HIVEPAAS_IMAGE_WORKER:?}
    stop_grace_period: 10m
    healthcheck:
      test: ["CMD-SHELL", "HB=$$(cat /tmp/hivepaas-worker.alive 2>/dev/null); [ $$(( $$(date +%s) - $${HB:-0} )) -lt 60 ]"]
      start_period: 120s
      start_interval: 5s
      interval: 30s
      timeout: 5s
      retries: 3
    networks:
      - local_net
    environment:
      <<: *hivepaas-env
      HP_RUN_MODE: worker
    volumes:
      - ${HIVEPAAS_DATA_DIR:?}:/var/lib/hivepaas
      - ${HIVEPAAS_DATA_DIR:?}:/host${HIVEPAAS_DATA_DIR:?}
      - ${HIVEPAAS_PROJECT_DATA_DIR:?}:/host${HIVEPAAS_PROJECT_DATA_DIR:?}
      - /var/run/docker.sock:/var/run/docker.sock
    deploy:
      labels:
        - "hivepaas.app.info={\"name\":\"worker\",\"key\":\"worker\"}"
        - "traefik.enable=false"
      mode: replicated
      replicas: 0
      placement:
        constraints:
          - node.role == manager
          - node.labels.hivepaas.role == control-plane
      restart_policy:
        condition: any
        delay: 5s
        window: 120s
      update_config:
        failure_action: rollback
        monitor: 240s
        max_failure_ratio: 0.5
    logging: *logging

  updater:
    image: ${HIVEPAAS_IMAGE_UPDATER:?}
    stop_grace_period: 10m
    networks:
      - local_net
    environment:
      <<: *hivepaas-env
      HP_RUN_MODE: updater
    volumes:
      - ${HIVEPAAS_DATA_DIR:?}:/var/lib/hivepaas
      - ${HIVEPAAS_DATA_DIR:?}:/host${HIVEPAAS_DATA_DIR:?}
      - /var/run/docker.sock:/var/run/docker.sock
    deploy:
      labels:
        - "hivepaas.app.info={\"name\":\"updater\",\"key\":\"updater\"}"
        - "traefik.enable=false"
      mode: replicated
      replicas: 0
      placement:
        constraints:
          - node.role == manager
          - node.labels.hivepaas.role == control-plane
      restart_policy:
        condition: on-failure
        delay: 5s
        window: 120s
    logging: *logging

  agent:
    image: ${HIVEPAAS_IMAGE_AGENT:?}
    stop_grace_period: 10m
    environment:
      <<: *hivepaas-env
      HP_RUN_MODE: agent
      HP_AGENT_PORT: "10001"
      HP_AGENT_NODE_ID: "{{.Node.ID}}"
      HP_AGENT_NODE_NAME: "{{.Node.Hostname}}"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /:/host
    networks:
      - local_net
    deploy:
      mode: global
      endpoint_mode: dnsrr
      labels:
        - "hivepaas.app.info={\"name\":\"agent\",\"key\":\"agent\"}"
        - "traefik.enable=false"
      restart_policy:
        condition: any
        delay: 5s
        window: 120s
    logging:
      driver: json-file
      options:
        max-size: 20m
        max-file: "5"
        compress: "true"

networks:
  hivepaas_net:
    external: true
  local_net:
    driver: overlay
    attachable: true
    driver_opts:
      com.docker.network.driver.mtu: 1380

volumes:
  db:
    # A Postgres major upgrade moves the database to a volume of its own and
    # records it in <data dir>/system/update/db-volume.env, which install.sh
    # reads before a deploy. See deployment/dev/hivepaas.yaml.
    name: ${HP_DB_VOLUME:-hivepaas_db}
    driver: local
```

Create `deployment/release/hivepaas.first-boot.yaml`:

```yaml
# Deployed with hivepaas.yaml on the first install only.
#
# The app reads the admin's account on its first boot, when it creates them,
# and never again. Once the dashboard answers, install.sh removes these from
# the services (docker service update --env-rm) and the password from
# install.env, so it does not stay readable in `docker service inspect`. A later
# deploy leaves this file out, and so does not bring them back.
services:
  app:
    environment:
      HP_USER_ADMIN_USERNAME: admin
      HP_USER_ADMIN_EMAIL: ${HIVEPAAS_ADMIN_EMAIL:?}
      HP_USER_ADMIN_PASSWORD: ${HIVEPAAS_ADMIN_PASSWORD:?}
  # The worker runs the same first-boot step when it starts first.
  worker:
    environment:
      HP_USER_ADMIN_USERNAME: admin
      HP_USER_ADMIN_EMAIL: ${HIVEPAAS_ADMIN_EMAIL:?}
      HP_USER_ADMIN_PASSWORD: ${HIVEPAAS_ADMIN_PASSWORD:?}
```

Create `deployment/release/traefik/dynamic_conf.yml`, a copy of dev's:

```yaml
tls:
  certificates:
    - certFile: /etc/traefik/ssl/certs/self-signed.crt
      keyFile: /etc/traefik/ssl/certs/self-signed.key
```

- [ ] **Step 4: Write the network helpers**

Append to `deployment/release/install.sh`, after one blank line:

```bash
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
```

- [ ] **Step 5: Run the tests**

Run: `make test-installer`
Expected: no shellcheck output, then `183 passed, 0 failed`. (Without a docker CLI, `test_stack_renders` prints `skip` and fewer checks run.)

- [ ] **Step 6: Commit**

```bash
git add deployment/release/hivepaas.yaml deployment/release/hivepaas.first-boot.yaml deployment/release/traefik/dynamic_conf.yml deployment/release/install.sh deployment/release/install_test.sh
git commit -m "feat(installer): the release stack, and routes by the server's addresses" -m "hivepaas.yaml is dev's with every host value a variable, images from the release, the configuration in the environment of every HivePaaS process, and x-custom routers by address that the app's label rewrite keeps. The admin's account is a file of its own, deployed on the first install only." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: Questions, and the settings install.env keeps

**Files:**
- Modify: `deployment/release/install.sh` (append the `Questions` section)
- Modify: `deployment/release/install_test.sh` (one section above `Runner`)

**Interfaces:**
- Consumes: the checks, `lowercase`, `normalize_dir`, `root_domain_of`, `rand_alnum`, `die`, `warn`, `info` (Task 1); `read_kv_file`, `kv_quote`, `app_env_of_channel`, `release_field` (Task 2); `addresses_line` (Task 3).
- Produces:
  - globals: `ADMIN_USERNAME='admin'`; `INSTALL_ENV` (`$HIVEPAAS_INSTALL_ENV_FILE` or `/etc/hivepaas/install.env`); `SAVED_KEYS`, `FIXED_KEYS`; `ASSUME_YES=0`, `CONFIG_FILE=''`, `HAVE_TTY=0`, `MISSING=''`, `RELEASE_FILE=''`.
  - `open_tty` (fd 3 reads answers, fd 4 shows questions; `HIVEPAAS_TTY` for tests); `ask VAR PROMPT`, `ask_secret VAR PROMPT` (die on end of input); `confirm QUESTION` (yes by default; `--yes` answers it; no terminal without `--yes` dies); `offer QUESTION` (no by default; `--yes` and no terminal both mean no).
  - `answer VAR PROMPT CHECK ERROR [DEFAULT [LABEL]]`; `answer_password`; `take_setting KEY VALUE` (a `--config` line); `take_saved KEY VALUE` (an `install.env` line); `normalize_settings`; `load_settings`; `ask_questions` (dies listing `MISSING`); `generate_secrets`; `save_settings`; `mask S` (`****` and the last four); `print_summary`.

- [ ] **Step 1: Write the failing tests**

In `deployment/release/install_test.sh`, insert above the `# ---...--- Runner` line, with one blank line after it:

```bash
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
```

- [ ] **Step 2: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - `open_tty`, `ask_questions`, `load_settings` and the rest are not found (`196 passed, 25 failed` when planning).

- [ ] **Step 3: Write the code**

Append to `deployment/release/install.sh`, after one blank line:

```bash
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
```

- [ ] **Step 4: Run the tests**

Run: `make test-installer`
Expected: no shellcheck output, then `234 passed, 0 failed`.

- [ ] **Step 5: Commit**

```bash
git add deployment/release/install.sh deployment/release/install_test.sh
git commit -m "feat(installer): the questions, and the settings install.env keeps" -m "Questions come from /dev/tty, a missing one stops a silent install with the variables to set, and a given value that fails its check stops it without printing the value. The environment wins over --config and --config over install.env; the secrets, domains and directories cannot change once saved." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 5: The host - distribution, Docker, swap and earlyoom

**Files:**
- Modify: `deployment/release/install.sh` (append the `Host` and `Docker` sections)
- Modify: `deployment/release/install_test.sh` (one section above `Runner`)

**Interfaces:**
- Consumes: `version_ge`, `lowercase`, `info`, `ok`, `warn`, `die` (Task 1); `read_kv_file`, `kv_quote`, `docker_static_arch`, `latest_docker_from_index` (Task 2); `confirm`, `offer` (Task 4).
- Produces:
  - globals: `EARLYOOM_AVOID`; `OS_ID OS_ID_LIKE OS_CODENAME OS_UBUNTU_CODENAME OS_DEBIAN_CODENAME OS_NAME`; `PKG_MANAGER` (`apt|dnf|zypper|pacman|apk`), `PKG_INDEX_FRESH`; `MIN_DOCKER_VERSION=29.5`, `MIN_DOCKER_API=1.54`; `DOCKER_VERSION`, `DOCKER_API`; `WORK_DIR` (set by `main` in Task 6).
  - `read_os_release FILE` (as data); `package_manager_of ID ID_LIKE`; `docker_install_method ID ID_LIKE` (`getdocker|apt-repo|dnf-rhel|dnf|zypper|pacman|apk`); `earlyoom_config_file PM`; `earlyoom_config PM`; `pm_install PKG...`; `has_systemd`; `service_start NAME`; `service_restart NAME`; `install_tools`; `check_resources`; `preflight`; `setup_swap`; `setup_earlyoom` (both warn instead of failing); `read_docker_versions`; `docker_new_enough`; `wait_for_docker`; `latest_docker_version`; `install_docker`; `ensure_docker`.

Only the distribution tables are tested here; installing Docker and earlyoom runs in Task 7's container (Alpine) and, for the other families, on the user's server.

- [ ] **Step 1: Write the failing tests**

In `deployment/release/install_test.sh`, insert above the `# ---...--- Runner` line, with one blank line after it:

```bash
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
```

- [ ] **Step 2: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - `read_os_release`, `package_manager_of` and the rest are not found (`237 passed, 35 failed` when planning).

- [ ] **Step 3: Write the code**

Append to `deployment/release/install.sh`, after one blank line:

```bash
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
```

- [ ] **Step 4: Run the tests**

Run: `make test-installer`
Expected: no shellcheck output, then `279 passed, 0 failed`.

- [ ] **Step 5: Commit**

```bash
git add deployment/release/install.sh deployment/release/install_test.sh
git commit -m "feat(installer): the host - its distribution, Docker, swap and earlyoom" -m "Docker is installed the way each family takes it - get.docker.com, Docker's apt or RHEL repository, or the distribution's package - and upgraded on request; the latest release is read from Docker's static build index. Swap and earlyoom only ever warn." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 6: Swarm, files, deploy, the waits, and main

**Files:**
- Modify: `deployment/release/install.sh` (append the last five sections)
- Modify: `deployment/release/install_test.sh` (two sections above `Runner`)

**Interfaces:**
- Consumes: everything above - in particular `release_payload`, `release_field`, `derive_agent_image`, `pg_major_of_image`, `read_kv_file` (Task 2); `host_ip`, `public_ip`, `ip_host_rule`, `routes_overlap` (Task 3); `load_settings`, `ask_questions`, `generate_secrets`, `save_settings`, `print_summary`, `confirm`, `open_tty` (Task 4); `preflight`, `ensure_docker`, `setup_swap`, `setup_earlyoom`, `WORK_DIR` (Task 5).
- Produces:
  - globals: `SUBNET=10.11.0.0/16`, `SUBNET_GATEWAY=10.11.0.1`, `REPO_RAW`, `CONFIG_DIR` (the directory of `INSTALL_ENV`); `STACK=hivepaas`, `REDEPLOY=0`, `INSTALL_STATE` (`fresh|unfinished|installed`), `HP_DB_VOLUME`, `HP_DB_MAJOR`, `UPDATE_ARGS`, `HIVEPAAS_IP_RULE`; the `HIVEPAAS_IMAGE_*` set by `resolve_images`.
  - `ensure_swarm`, `ensure_network`, `write_self_signed_cert DIR ROOT APP`, `fetch_install_file NAME DEST`, `prepare_files`; `fetch_release`, `gather_addresses`, `service_exists SVC`, `db_volume_exists`, `detect_install_state`, `running_image SVC`, `image_for SVC RELEASE_IMAGE`, `take_db_volume`, `resolve_images`, `deploy_stack FIRST_BOOT`, `run_migrations IMAGE`, `update_state SVC`, `wait_for_update SVC`, `task_template SVC`, `redeploy_stack`, `deploy`, `update_ip_routes`; `dashboard_answers`, `wait_for_dashboard`, `die_waiting`, `running_task SVC`, `wait_for_new_task SVC OLD`, `service_update_args SVC`, `tune_services`; `finish_install`, `print_done JUST_INSTALLED`; `usage`, `parse_args`, `cleanup`, `main`; and the last line, which runs `main` unless `HIVEPAAS_INSTALL_LIB=1`.

The order in `main` is the spec's nine steps. Three things in it are there because of what planning saw, and must stay: the migrations run after the deploy (the app does not migrate on boot); `tune_services` runs after the dashboard answers and updates one service at a time, database first and app last, checking afterwards that swarm did not roll an update back; and `redeploy_stack` deploys a second time when the app's update was rolled back while the database restarted.

- [ ] **Step 1: Write the failing tests**

In `deployment/release/install_test.sh`, insert above the `# ---...--- Runner` line, with one blank line after it:

```bash
# ------------------------------------------------------------------- Deploy

test_write_self_signed_cert() {
  local text
  mkdir -p "$TMP/certs"
  write_self_signed_cert "$TMP/certs" mydomain.com hivepaas.dev.mydomain.com
  check "written" 0 "$?"
  text=$(openssl x509 -in "$TMP/certs/self-signed.crt" -noout -text)
  check_contains "common name" "$text" "CN=mydomain.com"
  check_contains "names" "$text" "DNS:mydomain.com, DNS:*.mydomain.com, DNS:hivepaas.dev.mydomain.com"
  check_contains "an EC P-256 key" "$text" "prime256v1"
  check "the key is root's alone" 600 "$(stat -c %a "$TMP/certs/self-signed.key" 2>/dev/null ||
    stat -f %Lp "$TMP/certs/self-signed.key")"
  write_self_signed_cert "$TMP/certs" mydomain.com mydomain.com
  text=$(openssl x509 -in "$TMP/certs/self-signed.crt" -noout -text)
  check_contains "the app on the root domain is named once" "$text" "DNS:mydomain.com, DNS:*.mydomain.com"$'\n'
}

test_take_db_volume() {
  HP_DB_VOLUME='' HP_DB_MAJOR=''
  printf '%s\n' HP_DB_VOLUME=hivepaas_db_19 HP_DB_MAJOR=19 >"$TMP/db-volume.env"
  read_kv_file "$TMP/db-volume.env" take_db_volume
  check "volume" hivepaas_db_19 "$HP_DB_VOLUME"
  check "major" 19 "$HP_DB_MAJOR"
  HP_DB_VOLUME='' HP_DB_MAJOR=''
  printf '%s\n' 'HP_DB_VOLUME=x;rm -rf /' HP_DB_MAJOR=abc 'PATH=/nowhere' >"$TMP/db-volume.env"
  read_kv_file "$TMP/db-volume.env" take_db_volume
  check "a volume name that is not one is ignored" "" "$HP_DB_VOLUME"
  check "a major that is not a number is ignored" "" "$HP_DB_MAJOR"
  check_fails "nothing else is taken" test "$PATH" = /nowhere
}

# stack_variables: the ${VAR}s the stack files interpolate - not those in
# comments, nor the $${VAR}s escaped for the container's shell.
stack_variables() {
  cat "$HERE/hivepaas.yaml" "$HERE/hivepaas.first-boot.yaml" | grep -v '^[[:space:]]*#' |
    grep -oE '(^|[^$])\$\{[A-Z_]+' | sed 's/.*\${//' | sort -u
}

test_deploy_exports_every_stack_variable() {
  local var body
  body=$(declare -f deploy_stack)
  for var in $(stack_variables); do
    check_contains "deploy_stack exports $var" "$body" "$var"
  done
  check_ok "the stack files were read" test "$(stack_variables | wc -l)" -ge 20
}

# --------------------------------------------------------------------- Main

test_parse_args() {
  parse_args --yes --config /tmp/a.conf --redeploy
  check "--yes" 1 "$ASSUME_YES"
  check "--config" /tmp/a.conf "$CONFIG_FILE"
  check "--redeploy" 1 "$REDEPLOY"
  parse_args --config=/tmp/b.conf -y
  check "--config=" /tmp/b.conf "$CONFIG_FILE"
  out=$( (parse_args --config) 2>&1)
  check "--config without a file stops" 1 "$?"
  out=$( (parse_args --nope) 2>&1)
  check "an unknown option stops" 1 "$?"
  check_contains "naming it" "$out" --nope
}

test_help() {
  local out key
  out=$(bash "$HERE/install.sh" --help)
  check "--help exits 0" 0 "$?"
  check_contains "usage" "$out" "Usage: install.sh"
  for key in ADMIN_EMAIL ADMIN_PASSWORD APP_DOMAIN ROOT_DOMAIN APP_SECRET DATA_DIR PROJECT_DATA_DIR CHANNEL \
    SWAP SWAP_SIZE_MB EARLYOOM UPGRADE_DOCKER AGENT_IMAGE RELEASE_BRANCH INSTALL_REF; do
    check_contains "--help lists HIVEPAAS_$key" "$out" "HIVEPAAS_$key"
  done
}

test_refuses_without_root() {
  local out
  if [ "$(id -u)" -eq 0 ]; then
    printf 'skip test_refuses_without_root: running as root\n'
    return 0
  fi
  out=$(HIVEPAAS_TTY=/dev/null bash "$HERE/install.sh" 2>&1)
  check "stops" 1 "$?"
  check_contains "saying why" "$out" "as root"
}
```

- [ ] **Step 2: Run them to see them fail**

Run: `bash deployment/release/install_test.sh`
Expected: FAIL - `write_self_signed_cert`, `parse_args` and the rest are not found, and `bash install.sh --help` prints nothing (`284 passed, 51 failed` when planning).

- [ ] **Step 3: Write the code**

Append to `deployment/release/install.sh`, after one blank line:

```bash
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
```

- [ ] **Step 4: Run the tests, under both shells**

Run: `make test-installer`
Expected: no shellcheck output, then `339 passed, 0 failed`.

Run: `docker run --rm -v "$PWD":/repo -w /repo/deployment/release bash:5.2 bash -c 'apk add -q jq openssl >/dev/null 2>&1; bash install_test.sh'`
Expected: `skip test_refuses_without_root: running as root`, `skip test_stack_renders: no docker CLI`, then `326 passed, 0 failed`.

- [ ] **Step 5: Commit**

```bash
git add deployment/release/install.sh deployment/release/install_test.sh
git commit -m "feat(installer): swarm, deploy, the wait for the dashboard, and main" -m "A first install deploys the stack with the admin's account, migrates the database and waits for the dashboard; then each system service gets its OOM priority and the admin's password leaves the app and worker, one service at a time. A second run redeploys nothing unless asked, and keeps the routes by address current." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 7: A whole install in docker:dind, the final review, and the merge

**Files:**
- Create: `deployment/release/install_e2e.sh`
- Modify: `Makefile` (`test-installer` shellchecks it too; a new `test-installer-e2e`)

**Interfaces:**
- Consumes: `install.sh` as a program (`--yes`, `--redeploy`), its test-only variables `HIVEPAAS_RELEASE_URL` and `HIVEPAAS_INSTALL_FILES_DIR`, and, sourced, `release_payload`, `release_field`, `derive_agent_image`.
- Produces: `make test-installer-e2e`, which prints `ok: ...` per check and `all end-to-end checks passed`, or stops at the first `FAIL: ...` with exit status 1.

The e2e never touches the local swarm: it starts a privileged `docker:dind` container named `hivepaas-install-e2e` (native platform), installs HivePaaS inside it with this checkout's `release.signed.json` and stack files, and removes it at the end (`HIVEPAAS_E2E_KEEP=1` keeps it). The HivePaaS images are amd64 only; on another architecture the script pulls them for amd64, where they run emulated, and puts a `docker` wrapper in the container that adds `--resolve-image never` to `docker stack deploy` - resolving would pin the amd64 platform, which no node there has. Every image is pulled before the install because the container's daemon runs with `DOCKER_SERVICE_PREFER_OFFLINE_IMAGE=1`. `HIVEPAAS_SWAP=false` keeps the install from setting `vm.swappiness`, which is the kernel's, shared with Docker Desktop's VM.

- [ ] **Step 1: Write the e2e script**

Create `deployment/release/install_e2e.sh`, and `chmod +x` it:

```bash
#!/usr/bin/env bash
#
# End-to-end test of install.sh, in a docker:dind container: a silent install
# with this checkout's release info and stack files, then what a person would
# check - the dashboard by domain and by address, signing in, the admin's
# password gone - and the runs after it: one that must change nothing, one
# with another app secret and one without install.env that must stop, one
# after an address changed, and a --redeploy.
#
#   make test-installer-e2e     (bash deployment/release/install_e2e.sh)
#
# It pulls the release's images and takes several minutes. The container is
# privileged and runs a swarm of its own; HIVEPAAS_SWAP=false keeps the install
# from setting vm.swappiness, which is the kernel's and so the host's too.
# HIVEPAAS_E2E_KEEP=1 leaves the container running to look around in.
#
# The commands in single quotes run in the container, where their $ expand:
# shellcheck disable=SC2016

set -euo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)
REPO=$(cd "$HERE/../.." && pwd)
NAME=hivepaas-install-e2e
DOMAIN=hivepaas.e2e.example.com
PASSWORD=e2e-password-123
INSTALL="HIVEPAAS_RELEASE_URL=file:///repo/release.signed.json HIVEPAAS_INSTALL_FILES_DIR=/repo/deployment/release \
HIVEPAAS_SWAP=false bash /repo/deployment/release/install.sh --yes"

cleanup() {
  if [ "${HIVEPAAS_E2E_KEEP:-}" != 1 ]; then docker rm -f "$NAME" >/dev/null 2>&1 || true; fi
}
trap cleanup EXIT

in_dind() {
  docker exec "$NAME" sh -c "$1"
}

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

pass() {
  printf 'ok: %s\n' "$*"
}

# expect NAME WANT GOT
expect() {
  if [ "$2" = "$3" ]; then pass "$1"; else fail "$1: want '$2', got '$3'"; fi
}

docker rm -f "$NAME" >/dev/null 2>&1 || true
# The swarm inside does not pull: every image is pulled below, before the deploy.
docker run -d --name "$NAME" --privileged --platform "linux/$(docker version --format '{{.Server.Arch}}')" \
  -e DOCKER_SERVICE_PREFER_OFFLINE_IMAGE=1 -v "$REPO":/repo:ro docker:dind >/dev/null
for _ in $(seq 60); do
  if in_dind 'docker info' >/dev/null 2>&1; then break; fi
  sleep 1
done
in_dind 'docker info' >/dev/null 2>&1 || fail "the Docker daemon in $NAME did not start"
in_dind 'apk add -q bash jq curl openssl >/dev/null'

# The HivePaaS images are built for amd64 only. Elsewhere they are pulled for
# amd64 and run emulated, and the stack is deployed without resolving images:
# resolving would pin their platform, and no node here has it.
docker exec -i "$NAME" bash -s >/dev/null <<'EOF'
set -e
HIVEPAAS_INSTALL_LIB=1 . /repo/deployment/release/install.sh
release_payload /repo/release.signed.json >/tmp/release.json
app=$(release_field /tmp/release.json beta appImage)
agent=$(release_field /tmp/release.json beta agentImage)
agent=${agent:-$(derive_agent_image "$app")}
platform=()
if [ "$(uname -m)" != x86_64 ]; then
  platform=(--platform linux/amd64)
  cat >/usr/local/sbin/docker <<'SHIM'
#!/bin/sh
if [ "$1" = stack ] && [ "$2" = deploy ]; then
  shift 2
  exec /usr/local/bin/docker stack deploy --resolve-image never "$@"
fi
exec /usr/local/bin/docker "$@"
SHIM
  chmod +x /usr/local/sbin/docker
fi
docker pull -q "${platform[@]}" "$app"
docker pull -q "${platform[@]}" "$agent"
for field in dbImage redisImage traefikImage; do
  docker pull -q "$(release_field /tmp/release.json beta "$field")"
done
EOF
pass "images pulled"

in_dind "HIVEPAAS_ADMIN_EMAIL=admin@example.com HIVEPAAS_ADMIN_PASSWORD=$PASSWORD HIVEPAAS_APP_DOMAIN=$DOMAIN \
  $INSTALL >/root/install.log 2>&1" || {
  in_dind 'tail -n 40 /root/install.log' >&2
  fail "the install"
}
pass "the install"

HOST_IP=$(in_dind "ip route get 1.1.1.1 | awk '{for (i = 1; i < NF; i++) if (\$i == \"src\") print \$(i + 1)}'")
expect "the dashboard by domain" 200 "$(in_dind "curl -sk -o /dev/null -w '%{http_code}' \
  --resolve $DOMAIN:443:127.0.0.1 https://$DOMAIN/_/ping")"
expect "the dashboard by address" 200 "$(in_dind "curl -sk -o /dev/null -w '%{http_code}' https://$HOST_IP/")"
expect "http by address goes to https" "302 https://$HOST_IP/" "$(in_dind "curl -s -o /dev/null \
  -w '%{http_code} %{redirect_url}' http://$HOST_IP/")"
login() {
  in_dind "curl -sk -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -H 'Origin: https://$1' \
    --resolve $DOMAIN:443:127.0.0.1 -d '{\"username\":\"admin\",\"password\":\"$2\"}' \
    https://$1/_/auth/login-with-password"
}
expect "the admin signs in" 200 "$(login "$DOMAIN" "$PASSWORD")"
expect "the admin signs in by address" 200 "$(login "$HOST_IP" "$PASSWORD")"
expect "a wrong password does not" 401 "$(login "$DOMAIN" wrong-password-1)"
expect "the admin's account left the app and worker" 0 "$(in_dind 'for s in app worker; do
  docker service inspect hivepaas_$s --format "{{range .Spec.TaskTemplate.ContainerSpec.Env}}{{println .}}{{end}}"
  done | grep -c HP_USER_ADMIN || true')"
oom_priorities() {
  in_dind 'for s in traefik db redis app worker updater agent; do
    docker service inspect hivepaas_$s --format "{{.Spec.TaskTemplate.ContainerSpec.OomScoreAdj}}"; done' | xargs
}
expect "every system service has OOM priority -500" "-500 -500 -500 -500 -500 -500 -500" "$(oom_priorities)"
expect "install.env is root's alone" "700 600" "$(in_dind 'stat -c %a /etc/hivepaas /etc/hivepaas/install.env' | xargs)"
expect "install.env says installed" 1 "$(in_dind 'grep -c "^HIVEPAAS_INSTALLED=.true.$" /etc/hivepaas/install.env')"
expect "install.env has no password" 0 "$(in_dind 'grep -c ADMIN_PASSWORD /etc/hivepaas/install.env || true')"

versions() {
  in_dind 'for s in traefik db redis app worker updater agent; do
    docker service inspect hivepaas_$s --format "{{.Version.Index}}"; done' | xargs
}
before=$(versions)
in_dind "$INSTALL >/root/install2.log 2>&1" || fail "a second run"
expect "a second run changes no service" "$before" "$(versions)"

if in_dind "HIVEPAAS_APP_SECRET=0123456789abcdef0123456789abcdefXX $INSTALL >/root/install3.log 2>&1"; then
  fail "a run with another app secret went on"
fi
pass "a run with another app secret stops"

in_dind 'mv /etc/hivepaas/install.env /root/install.env.bak'
if in_dind "$INSTALL >/root/install-lost.log 2>&1"; then fail "a run without install.env went on"; fi
in_dind 'grep -q "Put your copy of the file back" /root/install-lost.log' || fail "a run without install.env: no reason given"
in_dind 'mv /root/install.env.bak /etc/hivepaas/install.env'
pass "a run without install.env stops"

in_dind "docker service update --detach --quiet \
  --label-add 'traefik.http.routers.x-custom-router-ip.rule=Host(\`192.0.2.1\`)' hivepaas_app >/dev/null"
in_dind "$INSTALL >/root/install4.log 2>&1" || fail "a run after an address changed"
rule=$(in_dind "docker service inspect hivepaas_app \
  --format '{{index .Spec.Labels \"traefik.http.routers.x-custom-router-ip.rule\"}}'")
case "$rule" in
  *"Host(\`$HOST_IP\`)"*) pass "the route by address follows the address" ;;
  *) fail "the route by address is $rule" ;;
esac

in_dind "$INSTALL --redeploy >/root/install5.log 2>&1" || {
  in_dind 'tail -n 30 /root/install5.log' >&2
  fail "a redeploy"
}
pass "a redeploy"
expect "after it, the app's update stands" "" "$(in_dind 'docker service inspect hivepaas_app \
  --format "{{if .UpdateStatus}}{{.UpdateStatus.State}}{{end}}"' | grep rollback || true)"
expect "and every system service has OOM priority -500 again" "-500 -500 -500 -500 -500 -500 -500" "$(oom_priorities)"
expect "and the dashboard answers" 200 "$(in_dind "curl -sk -o /dev/null -w '%{http_code}' \
  --resolve $DOMAIN:443:127.0.0.1 https://$DOMAIN/_/ping")"

printf 'all end-to-end checks passed\n'
```

- [ ] **Step 2: Wire it into the Makefile**

Replace the `test-installer` target added in Task 1 with:

```make
# The release installer (deployment/release): shellcheck, then the tests of
# its functions under this machine's bash.
test-installer:
	@docker run --rm -v "$(PWD)/deployment/release":/mnt:ro -w /mnt koalaman/shellcheck:stable \
		-x install.sh install_test.sh install_e2e.sh
	@bash deployment/release/install_test.sh

# A whole install in a docker:dind container. Pulls the release's images and
# takes several minutes; nothing outside the container is deployed.
test-installer-e2e:
	@bash deployment/release/install_e2e.sh
```

Run: `make test-installer`
Expected: no shellcheck output, then `339 passed, 0 failed`.

- [ ] **Step 3: Run it**

Run: `make test-installer-e2e` (several minutes; about 1 GB of images the first time)
Expected, in order:

```
ok: images pulled
ok: the install
ok: the dashboard by domain
ok: the dashboard by address
ok: http by address goes to https
ok: the admin signs in
ok: the admin signs in by address
ok: a wrong password does not
ok: the admin's account left the app and worker
ok: every system service has OOM priority -500
ok: install.env is root's alone
ok: install.env says installed
ok: install.env has no password
ok: a second run changes no service
ok: a run with another app secret stops
ok: a run without install.env stops
ok: the route by address follows the address
ok: a redeploy
ok: after it, the app's update stands
ok: and every system service has OOM priority -500 again
ok: and the dashboard answers
all end-to-end checks passed
```

On a `FAIL`, run it again with `HIVEPAAS_E2E_KEEP=1`, then look inside: `docker exec hivepaas-install-e2e sh -c 'cat /root/install.log; docker service ls; docker service ps hivepaas_app --no-trunc; docker service logs --tail 50 hivepaas_app'`. The logs of each run are `/root/install*.log`. When done: `docker rm -f hivepaas-install-e2e`.

Then check nothing is left: `docker ps -a --filter name=hivepaas-install-e2e --format '{{.Names}}'`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add deployment/release/install_e2e.sh Makefile
git commit -m "test(installer): a whole install in docker-in-docker" -m "make test-installer-e2e installs HivePaaS in a docker:dind container with this checkout's release info and stack files, then checks what a person would - the dashboard by domain and by address, signing in, the admin's password gone - and the runs after it: unchanged, stopped, rerouted and redeployed." -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 5: The whole branch, reviewed**

Run: `make test-installer` and the bash 5 command of Task 6 Step 4 once more, and read `git diff main...HEAD --stat`: only the seven files of this plan. Then review the branch against the spec and this plan's Review Focus. The user declined a reviewer subagent before: the review is the executor's own, and the final message says so.

- [ ] **Step 6: Merge locally**

```bash
git checkout main
git merge --no-ff feat/release-installer -m "Merge branch 'feat/release-installer'" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
make test-installer
git branch -d feat/release-installer
```

Expected: the merge succeeds, `339 passed, 0 failed`. Do not push.

- [ ] **Step 7: Hand over what only a real server can show**

Tell the user, in Vietnamese: the install on their Linux server needs `HIVEPAAS_RELEASE_BRANCH=main` until the `release` branch exists; what to try there (an interactive `curl ... | sudo bash`, a silent one with `--yes`, a second run, the dashboard by domain and by IP, the browser's certificate warning); the families whose Docker install was never run here (everything but Alpine); and the two app fixes the spec's last section lists.
