#!/usr/bin/env bash
#
# End-to-end test of install.sh on a distribution as people get it: its cloud
# image in a virtual machine (QEMU, with KVM), a user with sudo and a password,
# and the installer run the way the docs say, at a terminal:
#
#   1. curl -fsSL <installer> -o install.sh && sudo bash install.sh, answering
#      its questions, for hivepaas.localhost: the dashboard answers, the admin
#      signs in and every service runs;
#   2. curl -fsSL <installer> | sudo bash, which must end - running, or
#      refusing as it does under a sudo-rs before 0.2.14 - and never hang;
#   3. sudo bash install.sh --yes, again;
#   4. a reboot, after which HivePaaS answers again.
#
#   make test-installer-distro DISTRO=ubuntu-26.04
#   (bash deployment/release/distro_e2e.sh ubuntu-26.04)
#
# DISTRO is one of the names in image_url below. The installer is this
# checkout's - its install.sh, release info and stack files, as install_e2e.sh
# takes them - unless HIVEPAAS_DISTRO_INSTALLER is the URL of one, such as
# https://get.hivepaas.com, which is then run with nothing set, as by a person.
#
# Needs qemu-system-x86_64, qemu-img, genisoimage, python3, ssh and /dev/kvm.
# The image is downloaded once, into HIVEPAAS_DISTRO_CACHE
# (~/.cache/hivepaas-distro); the install pulls the release's images. The logs
# go to HIVEPAAS_DISTRO_OUT, a new directory by default, which it prints.
# HIVEPAAS_DISTRO_KEEP=1 leaves the machine running to ssh into.
#
# Commands in single quotes run in the machine, where their $ expand (SC2016).
# The whole of it is read before any of it runs, so an edit to this file while a
# run takes its 30 minutes leaves that run alone; shellcheck then takes the
# functions a trap or wait_until calls for unreachable (SC2317).
# shellcheck disable=SC2016,SC2317
{
set -euo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)
REPO=$(cd "$HERE/../.." && pwd)
DOMAIN=hivepaas.localhost
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=e2e-password-123
# The password of the machine's user `tester`, which sudo asks for.
SUDO_PASSWORD=tester-password-123
CPUS=${HIVEPAAS_DISTRO_CPUS:-2}
MEMORY_MB=${HIVEPAAS_DISTRO_MEMORY_MB:-6144}

DISTROS='ubuntu-24.04 ubuntu-26.04 debian-12 debian-13 rocky-9 rocky-10 almalinux-9 almalinux-10 fedora-44
amazonlinux-2023 opensuse-leap-16.0 arch'

# image_url DISTRO: its cloud image, which comes with cloud-init.
image_url() {
  local al2023
  case "$1" in
    ubuntu-24.04) echo https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img ;;
    ubuntu-26.04) echo https://cloud-images.ubuntu.com/resolute/current/resolute-server-cloudimg-amd64.img ;;
    debian-12) echo https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2 ;;
    debian-13) echo https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2 ;;
    rocky-9) echo https://dl.rockylinux.org/pub/rocky/9/images/x86_64/Rocky-9-GenericCloud-Base.latest.x86_64.qcow2 ;;
    rocky-10) echo https://dl.rockylinux.org/pub/rocky/10/images/x86_64/Rocky-10-GenericCloud-Base.latest.x86_64.qcow2 ;;
    almalinux-9) echo https://repo.almalinux.org/almalinux/9/cloud/x86_64/images/AlmaLinux-9-GenericCloud-latest.x86_64.qcow2 ;;
    almalinux-10) echo https://repo.almalinux.org/almalinux/10/cloud/x86_64/images/AlmaLinux-10-GenericCloud-latest.x86_64.qcow2 ;;
    fedora-44)
      echo https://dl.fedoraproject.org/pub/fedora/linux/releases/44/Cloud/x86_64/images/Fedora-Cloud-Base-Generic-44-1.7.x86_64.qcow2
      ;;
    # The file is named after the release, which `latest` redirects to.
    amazonlinux-2023)
      al2023=$(curl -fsSL -o /dev/null -w '%{url_effective}' https://cdn.amazonlinux.com/al2023/os-images/latest/kvm/) &&
        printf '%s%s\n' "$al2023" "$(curl -fsSL "$al2023" | grep -oE 'al2023-kvm-[A-Za-z0-9._-]+-x86_64[A-Za-z0-9._-]*\.qcow2' | head -n 1)"
      ;;
    opensuse-leap-16.0)
      echo https://download.opensuse.org/distribution/leap/16.0/appliances/Leap-16.0-Minimal-VM.x86_64-Cloud.qcow2
      ;;
    arch) echo https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2 ;;
    *) return 1 ;;
  esac
}

DISTRO=${1:-}
IMAGE_URL=$(image_url "$DISTRO") || IMAGE_URL=''
# Empty, or a directory: a name it does not know, or an image it did not find.
if [ -z "${IMAGE_URL##*/}" ]; then
  printf 'usage: %s DISTRO, one of: %s\n' "$0" "$(printf '%s' "$DISTROS" | tr '\n' ' ')" >&2
  exit 2
fi
INSTALLER=${HIVEPAAS_DISTRO_INSTALLER:-checkout}
CACHE=${HIVEPAAS_DISTRO_CACHE:-$HOME/.cache/hivepaas-distro}
OUT=${HIVEPAAS_DISTRO_OUT:-$(mktemp -d -t "hivepaas-distro-$DISTRO.XXXXXX")}
WORK=$(mktemp -d)
mkdir -p "$CACHE" "$OUT"
FAILED=1
HTTP_PID=''

fail() {
  printf 'FAIL [%s]: %s\n' "$DISTRO" "$*" >&2
  exit 1
}

pass() {
  printf 'ok [%s]: %s\n' "$DISTRO" "$*"
}

# expect NAME WANT GOT
expect() {
  if [ "$2" = "$3" ]; then pass "$1"; else fail "$1: want '$2', got '$3'"; fi
}

free_port() {
  python3 -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])'
}

SSH_PORT=$(free_port)
SSH=(ssh -i "$WORK/key" -p "$SSH_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null
  -o LogLevel=ERROR -o ConnectTimeout=5 -o ServerAliveInterval=15)

# probe CMD: CMD in the machine, as `probe`, whose sudo asks for nothing. The
# checks go through it; the scenarios are `tester`'s.
probe() {
  timeout 120 "${SSH[@]}" probe@127.0.0.1 "$1"
}

# as_root CMD: CMD in the machine, as root, read by sh from stdin.
as_root() {
  timeout 300 "${SSH[@]}" probe@127.0.0.1 sudo sh -s <<<"$1"
}

# drive NAME CMD: CMD at a terminal in the machine, as `tester`, answered and
# timed by distro_e2e_drive.py; its output in $OUT/NAME.log. An install may
# say nothing for long - a package manager told to be quiet, on a slow mirror -
# so it may be silent for DRIVE_IDLE_SECONDS, 20 minutes unless set.
drive() {
  local status=0
  DRIVE_IDLE_SECONDS=${DRIVE_IDLE_SECONDS:-1200} DRIVE_SUDO_PASSWORD=$SUDO_PASSWORD \
    DRIVE_ADMIN_EMAIL=$ADMIN_EMAIL DRIVE_ADMIN_PASSWORD=$ADMIN_PASSWORD DRIVE_APP_DOMAIN=$DOMAIN \
    python3 "$HERE/distro_e2e_drive.py" "$OUT/$1.log" -- "${SSH[@]}" -tt tester@127.0.0.1 "$2" || status=$?
  # The terminal's last line can be left open, by its output or a shell's
  # escape sequence.
  printf '\n'
  return "$status"
}

collect() {
  as_root '
    echo "== processes"; ps -eo pid,etime,stat,args --forest 2>/dev/null || ps
    echo "== install.log"; cat /var/lib/hivepaas/install.log
    echo "== docker service ls"; docker service ls
    echo "== docker service ps"; docker service ps --no-trunc $(docker service ls -q)
    for s in app agent worker; do
      echo "== docker service logs hivepaas_$s"; docker service logs --no-trunc --tail 200 "hivepaas_$s"
    done
    echo "== docker ps -a"; docker ps -a --no-trunc
    echo "== journalctl -u docker"; journalctl -u docker --no-pager -n 300
  ' >"$OUT/machine.txt" 2>&1 || true
}

cleanup() {
  if [ "$FAILED" = 1 ] && [ -f "$WORK/qemu.pid" ]; then collect; fi
  if [ -n "$HTTP_PID" ]; then kill "$HTTP_PID" 2>/dev/null || true; fi
  if [ "${HIVEPAAS_DISTRO_KEEP:-}" = 1 ]; then
    printf 'The machine runs on: ssh -i %s -p %s tester@127.0.0.1\n' "$WORK/key" "$SSH_PORT"
  else
    if [ -f "$WORK/qemu.pid" ]; then kill "$(cat "$WORK/qemu.pid")" 2>/dev/null || true; fi
    rm -rf "$WORK"
  fi
  printf 'Logs: %s\n' "$OUT"
}
trap cleanup EXIT

# --------------------------------------------------------------- The machine

# download URL FILE: going on from where a dropped connection left it.
download() {
  local attempt
  for attempt in 1 2 3 4 5; do
    if curl -fsSL --retry 3 -C - -o "$2.part" "$1"; then
      mv "$2.part" "$2"
      return 0
    fi
    printf 'The download stopped (attempt %s); going on.\n' "$attempt" >&2
    sleep 5
  done
  return 1
}

if [ ! -f "$CACHE/$DISTRO.img" ]; then
  printf 'Downloading %s\n' "$IMAGE_URL"
  download "$IMAGE_URL" "$CACHE/$DISTRO.img" || fail "could not download $IMAGE_URL"
fi
qemu-img create -q -f qcow2 -F qcow2 -b "$CACHE/$DISTRO.img" "$WORK/disk.qcow2" 40G

ssh-keygen -q -t ed25519 -N '' -C hivepaas-distro-e2e -f "$WORK/key"
key=$(cat "$WORK/key.pub")
mkdir -p "$WORK/seed"
# tester is a person's account: sudo asks for its password - or root's, which
# is the same, where sudo asks for the target's (openSUSE's targetpw). probe is
# the test's, with the shell every distribution has.
cat >"$WORK/seed/user-data" <<EOF
#cloud-config
ssh_pwauth: false
chpasswd:
  expire: false
  users:
    - {name: root, password: '$SUDO_PASSWORD', type: text}
users:
  - name: tester
    shell: /bin/bash
    lock_passwd: false
    plain_text_passwd: '$SUDO_PASSWORD'
    sudo: 'ALL=(ALL) ALL'
    ssh_authorized_keys: ['$key']
  - name: probe
    shell: /bin/sh
    sudo: 'ALL=(ALL) NOPASSWD:ALL'
    ssh_authorized_keys: ['$key']
EOF
printf 'instance-id: hivepaas-distro-e2e\nlocal-hostname: hivepaas-e2e\n' >"$WORK/seed/meta-data"
genisoimage -quiet -output "$WORK/seed.iso" -volid cidata -joliet -rock \
  "$WORK/seed/user-data" "$WORK/seed/meta-data"

qemu-system-x86_64 -name "hivepaas-$DISTRO" -machine accel=kvm -cpu host -smp "$CPUS" -m "$MEMORY_MB" \
  -drive "file=$WORK/disk.qcow2,if=virtio,format=qcow2" \
  -drive "file=$WORK/seed.iso,if=virtio,format=raw,readonly=on" \
  -netdev "user,id=net0,hostfwd=tcp:127.0.0.1:$SSH_PORT-:22" -device virtio-net-pci,netdev=net0 \
  -display none -serial "file:$OUT/console.log" -daemonize -pidfile "$WORK/qemu.pid"

# wait_for_ssh SECONDS
wait_for_ssh() {
  local start=$SECONDS
  until "${SSH[@]}" probe@127.0.0.1 true 2>/dev/null; do
    if [ $((SECONDS - start)) -ge "$1" ]; then return 1; fi
    sleep 3
  done
}
wait_for_ssh 600 || fail "the machine did not answer on ssh in 10 minutes (see console.log)"
# 2: done, with errors cloud-init could recover from.
probe 'cloud-init status --wait >/dev/null; s=$?; [ $s -eq 0 ] || [ $s -eq 2 ]' ||
  fail "cloud-init did not finish (see console.log)"
pass "the machine runs: $(probe '. /etc/os-release; echo "$PRETTY_NAME"'), $(probe 'sudo --version | head -n 1')"

# ------------------------------------------------------------- The installer

# SUDO: what the docs' `sudo bash` is here: with this checkout's release info
# and stack files, or as it is for the installer of a URL.
if [ "$INSTALLER" = checkout ]; then
  mkdir -p "$WORK/www"
  cp "$HERE/install.sh" "$WORK/www/install.sh"
  HTTP_PORT=$(free_port)
  python3 -m http.server --bind 127.0.0.1 --directory "$WORK/www" "$HTTP_PORT" >/dev/null 2>&1 &
  HTTP_PID=$!
  # The host, as QEMU's user network shows it to the machine.
  INSTALLER_URL=http://10.0.2.2:$HTTP_PORT/install.sh
  tar -C "$REPO" -cf - deployment/release release.signed.json |
    "${SSH[@]}" tester@127.0.0.1 'mkdir -p hp && tar -C hp -xf -'
  SUDO='sudo env HIVEPAAS_RELEASE_URL=file:///home/tester/hp/release.signed.json'
  SUDO="$SUDO HIVEPAAS_INSTALL_FILES_DIR=/home/tester/hp/deployment/release"
else
  INSTALLER_URL=$INSTALLER
  SUDO=sudo
fi

ping_code() {
  probe "curl -sk -o /dev/null -w '%{http_code}' --max-time 10 --resolve $DOMAIN:443:127.0.0.1 \
    https://$DOMAIN/api/ping" || true
}

login_code() {
  probe "curl -sk -o /dev/null -w '%{http_code}' --max-time 10 -H 'Content-Type: application/json' \
    -H 'Origin: https://$DOMAIN' --resolve $DOMAIN:443:127.0.0.1 \
    -d '{\"username\":\"admin\",\"password\":\"$ADMIN_PASSWORD\"}' https://$DOMAIN/api/auth/login-with-password" ||
    true
}

# services_down: the stack's services that are missing or run fewer tasks than
# they want. The worker and the updater want none until the app scales them.
services_down() {
  as_root 'for s in traefik db redis app worker updater agent; do
    r=$(docker service ls --filter "name=hivepaas_$s" --format "{{.Name}} {{.Replicas}}" |
      awk -v n="hivepaas_$s" "\$1 == n {print \$2}")
    if [ -z "$r" ] || [ "${r%%/*}" != "${r#*/}" ]; then echo "$s"; fi
  done' | tr '\n' ' ' | sed 's/ $//'
}

# wait_until SECONDS CMD WANT: until CMD prints WANT, for up to SECONDS.
wait_until() {
  local start=$SECONDS got
  while :; do
    got=$($2)
    if [ "$got" = "$3" ] || [ $((SECONDS - start)) -ge "$1" ]; then
      printf '%s' "$got"
      return 0
    fi
    sleep 5
  done
}

check_running() {
  expect "$1: the dashboard answers" 200 "$(wait_until 480 ping_code 200)"
  expect "$1: the admin signs in" 200 "$(login_code)"
  expect "$1: every service runs" '' "$(wait_until 180 services_down '')"
}

# reboot_machine: until it runs again, after a boot of its own.
reboot_machine() {
  local boot_id start=$SECONDS
  boot_id=$(probe 'cat /proc/sys/kernel/random/boot_id')
  as_root 'reboot' >/dev/null 2>&1 || true
  until [ "$(probe 'cat /proc/sys/kernel/random/boot_id' 2>/dev/null || true)" != "$boot_id" ] &&
    probe true 2>/dev/null; do
    if [ $((SECONDS - start)) -ge 600 ]; then return 1; fi
    sleep 5
  done
}

# ---------------------------------------------------------------- Scenarios

status=0
drive 1-install "curl -fsSL $INSTALLER_URL -o install.sh && $SUDO bash install.sh" || status=$?
# A kernel upgraded with Docker - Arch's - needs a reboot first, and says so.
if [ "$status" = 1 ] && grep -q 'Reboot, then run the installer again' "$OUT/1-install.log"; then
  pass "1. the install asks for a reboot: the kernel was upgraded with Docker"
  reboot_machine || fail "1. the machine did not come back from its reboot"
  status=0
  drive 1-install-after-reboot "$SUDO bash install.sh" || status=$?
fi
expect "1. curl -o install.sh && sudo bash install.sh, answered" 0 "$status"
check_running "after the install"

# Under a sudo-rs before 0.2.14 the installer refuses a pipe, as it should;
# anywhere else it runs. Either way it ends.
status=0
DRIVE_IDLE_SECONDS=180 drive 2-pipe "curl -fsSL $INSTALLER_URL | $SUDO bash" || status=$?
case "$status" in
  0) pass "2. curl | sudo bash ran" ;;
  1)
    grep -q 'is sudo-rs' "$OUT/2-pipe.log" || fail "2. curl | sudo bash stopped (see 2-pipe.log)"
    pass "2. curl | sudo bash refused: this sudo is a sudo-rs before 0.2.14"
    ;;
  124) fail "2. curl | sudo bash hangs (see 2-pipe.log)" ;;
  *) fail "2. curl | sudo bash ended with $status (see 2-pipe.log)" ;;
esac

status=0
drive 3-again "$SUDO bash install.sh --yes" || status=$?
expect "3. sudo bash install.sh --yes, again" 0 "$status"
check_running "after running it again"

reboot_machine || fail "4. the machine did not come back from its reboot"
pass "4. the machine rebooted"
check_running "after the reboot"

FAILED=0
pass "all good"
exit
}
