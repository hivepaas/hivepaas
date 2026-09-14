#!/bin/bash
#
# Adds one docker-in-docker node to the local swarm.
#
# The node is named <prefix><N>, where N is one past the highest that already
# exists. Both swarm nodes and leftover containers are counted: a container with
# no node, or a node with no container, is a half-removed node and its number is
# not free.
#
# Usage: add-node.sh <prefix> [image]

set -eo pipefail

PREFIX="${1:?usage: $0 <prefix> [image]}"
IMAGE="${2:-docker:dind}"
AGENT_SERVICE=hivepaas_agent

highestIndex() {
  {
    docker node ls --format '{{.Hostname}}'
    docker ps -a --format '{{.Names}}'
  } | sed -n "s/^${PREFIX}\([0-9]\{1,\}\)$/\1/p" | sort -n | tail -1
}

LAST=$(highestIndex)
# The manager counts as node 1, so the first one added is 2.
NODE="${PREFIX}$(( ${LAST:-1} + 1 ))"
echo "adding $NODE"

docker run -d --privileged --name "$NODE" --hostname "$NODE" \
  -e DOCKER_TLS_CERTDIR= "$IMAGE" --host=unix:///var/run/docker.sock > /dev/null

printf 'starting the node daemon'
for _ in $(seq 1 60); do
  if docker exec "$NODE" docker info > /dev/null 2>&1; then
    echo ' up'
    break
  fi
  printf '.'
  sleep 1
done

# Docker Desktop rewrites the stack's /var/run/docker.sock bind to its own
# /run/host-services/docker.proxy.sock, a path that exists only inside the Desktop
# VM. The agent is a global service, so it is scheduled here too and swarm rejects
# the task with "bind source path does not exist", every few seconds, forever.
# Pointing that path at this node's own daemon is both the fix and what an agent
# running here should be talking to in the first place.
docker exec "$NODE" sh -c \
  'mkdir -p /run/host-services && ln -sf /var/run/docker.sock /run/host-services/docker.proxy.sock'

docker exec "$NODE" docker swarm join \
  --token "$(docker swarm join-token -q worker)" \
  "$(docker node inspect self --format '{{.ManagerStatus.Addr}}')" > /dev/null

printf 'waiting for the agent (a new node pulls its image first)'
for _ in $(seq 1 120); do
  STATE=$(docker service ps "$AGENT_SERVICE" --no-trunc \
    --filter "node=$NODE" --filter desired-state=running \
    --format '{{.CurrentState}}|{{.Error}}' | head -1)
  case "$STATE" in
    Running*)
      echo ' running'
      exit 0
      ;;
    Failed* | Rejected*)
      echo " $STATE"
      exit 1
      ;;
  esac
  printf '.'
  sleep 5
done

echo ' timed out'
docker service ps "$AGENT_SERVICE" --filter "node=$NODE" --no-trunc | head -3
exit 1
