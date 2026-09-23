#!/bin/bash
#
# Prepares a host so that running out of memory kills one user app instead of
# stalling the whole machine. Without swap the kernel can only reclaim program
# code and file cache, from every process at once: everything slows to a crawl,
# healthchecks cannot even start, and the kernel OOM killer may not fire for an
# hour. Swap takes the pages idle processes are not using; earlyoom kills the
# largest process that is not a system one before the machine gets that far.
#
# Idempotent. Debian/Ubuntu (apt) and Fedora/RHEL (dnf). Must run as root.
#
#   SWAP_SIZE_MB   swap file size when the host has no swap yet (default 2048)
#   SWAP_FILE      where to create it (default /swapfile)

set -euo pipefail

SWAP_SIZE_MB=${SWAP_SIZE_MB:-2048}
SWAP_FILE=${SWAP_FILE:-/swapfile}

# Process names earlyoom never kills. postgres and redis-server are the HivePaaS
# database and cache, but the names also cover the databases of user apps -
# earlyoom sees process names, not which service a process belongs to.
#
# The second line is what an admin adds later - the log collector (vlagent-prod),
# the log backend (victoria-logs-prod) and the registry (zot-linux-<arch>). The
# kernel cuts a process name at 15 characters, hence the trailing wildcards.
EARLYOOM_AVOID='^(hivepaas|hivepaas-agent|traefik|postgres|redis-server|dockerd|containerd|containerd-shim|sshd|systemd|systemd-.*'
EARLYOOM_AVOID+='|vlagent.*|victoria-logs.*|zot-linux-.*)$'

if [ "$(id -u)" -ne 0 ]; then
  echo "host-memory.sh: must run as root" >&2
  exit 1
fi

echo "== Swap"
if [ -n "$(swapon --show --noheadings)" ]; then
  echo "Swap already active, leaving it as it is:"
  swapon --show
else
  avail_mb=$(df -Pm "$(dirname "$SWAP_FILE")" | awk 'NR==2 {print $4}')
  if [ "$avail_mb" -lt $((SWAP_SIZE_MB + 1024)) ]; then
    echo "Not enough disk for a ${SWAP_SIZE_MB}MB swap file (${avail_mb}MB free), skipping." >&2
  else
    echo "Creating ${SWAP_SIZE_MB}MB swap file at $SWAP_FILE..."
    fallocate -l "${SWAP_SIZE_MB}M" "$SWAP_FILE" 2>/dev/null ||
      dd if=/dev/zero of="$SWAP_FILE" bs=1M count="$SWAP_SIZE_MB" status=none
    chmod 600 "$SWAP_FILE"
    mkswap "$SWAP_FILE" >/dev/null
    swapon "$SWAP_FILE"
    grep -q "^$SWAP_FILE " /etc/fstab || echo "$SWAP_FILE none swap sw 0 0" >> /etc/fstab
  fi
fi

# Swap only what is really idle: at the default of 60 the kernel also swaps out
# memory that services touch regularly, which is its own kind of slow.
cat > /etc/sysctl.d/99-hivepaas-memory.conf <<'EOF'
vm.swappiness = 10
EOF
sysctl -q -p /etc/sysctl.d/99-hivepaas-memory.conf

echo "== earlyoom"
if ! command -v earlyoom >/dev/null; then
  if command -v apt-get >/dev/null; then
    apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq earlyoom
  elif command -v dnf >/dev/null; then
    dnf install -y -q earlyoom
  else
    echo "No apt-get or dnf: install earlyoom yourself, then re-run." >&2
    exit 1
  fi
fi

# Act when available memory is under 5% and free swap under 20%: the swap is
# there to be used, so earlyoom waits until most of it is gone too. The regex is
# unquoted on purpose: systemd splits $EARLYOOM_ARGS on whitespace and would pass
# quotes through as part of the pattern.
cat > /etc/default/earlyoom <<EOF
EARLYOOM_ARGS="-m 5 -s 20 -r 3600 --avoid $EARLYOOM_AVOID"
EOF
systemctl enable earlyoom >/dev/null 2>&1
systemctl restart earlyoom
echo "earlyoom: $(systemctl is-active earlyoom)"

echo "== Done"
free -m
