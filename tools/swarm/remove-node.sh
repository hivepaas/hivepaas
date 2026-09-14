#!/bin/bash
#
# Removes docker-in-docker nodes from the local swarm.
#
# Usage: remove-node.sh <prefix> [name|--all]
#   no name: the highest-numbered one, so it undoes the last add
#   --all:   every <prefix><N>
#
# Every step tolerates the previous one having already happened, so this also
# cleans up after a container removed by hand, or a node left showing Down.

set -eo pipefail

PREFIX="${1:?usage: $0 <prefix> [name|--all]}"
TARGET="${2:-}"

nodeNames() {
  {
    docker node ls --format '{{.Hostname}}'
    docker ps -a --format '{{.Names}}'
  } | sed -n "s/^\(${PREFIX}[0-9]\{1,\}\)$/\1/p" | sort -u
}

removeOne() {
  local node="$1"
  local id

  docker exec "$node" docker swarm leave --force > /dev/null 2>&1 || true
  # Matched exactly rather than with `--filter name=`, which matches on substring:
  # that filter would see hivepaas-node2 inside hivepaas-node20.
  id=$(docker node ls --format '{{.ID}} {{.Hostname}}' | awk -v h="$node" '$2 == h {print $1}')
  if [ -n "$id" ]; then
    docker node rm --force "$id" > /dev/null
  fi
  docker rm -f "$node" > /dev/null 2>&1 || true
  echo "$node is gone"
}

case "$TARGET" in
  --all)
    FOUND=$(nodeNames)
    if [ -z "$FOUND" ]; then
      echo "no ${PREFIX}* nodes"
      exit 0
    fi
    for node in $FOUND; do
      removeOne "$node"
    done
    ;;
  "")
    INDEX=$(nodeNames | sed -n "s/^${PREFIX}//p" | sort -n | tail -1)
    if [ -z "$INDEX" ]; then
      echo "no ${PREFIX}* nodes"
      exit 0
    fi
    removeOne "${PREFIX}${INDEX}"
    ;;
  *)
    removeOne "$TARGET"
    ;;
esac
