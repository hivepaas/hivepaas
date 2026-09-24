#!/bin/bash

set -eo pipefail

echo "---------------------------------------------------------------"
echo "INSTALL HivePaaS"
echo "---------------------------------------------------------------"

# Delete all unused data that take the disk space
# docker system prune -a -f

# Swap and earlyoom, so running out of memory kills one user app instead of
# stalling the host. Needs root; the install goes on without it.
echo "Prepare host memory settings..."
curl -sL "https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/dev/host-memory.sh" -o host-memory.sh
sudo bash host-memory.sh || echo "WARNING: host-memory.sh failed, continuing without swap/earlyoom" >&2

# Overlay networks of the swarm this node is in. ingress is the swarm's own:
# it is recreated with the swarm and refuses to be removed while it is up.
swarm_networks() {
  docker network ls --filter driver=overlay --filter scope=swarm --format '{{.Name}}' | grep -vx ingress || true
}

# Removes a network, including the load-balancer endpoint swarm gives every
# overlay network with a service on it. That endpoint belongs to no container,
# so nothing else ever takes it off; while it is there the network cannot go.
remove_network() {
  docker network disconnect -f "$1" "lb-$1" >/dev/null 2>&1 || true
  docker network rm "$1" >/dev/null 2>&1
}

# Reset whole cluster.
#
# Leaving a swarm is not enough to clear it: an overlay network that still had
# a service on it stays behind on this node, holding its subnet. The next swarm
# does not know about it and hands the same subnet to a new network, and every
# task on that one is then refused with "Pool overlaps with other one on this
# address space". So the services go first, then their networks, then the swarm.
if [ "$(docker info --format '{{.Swarm.LocalNodeState}}')" = "active" ]; then
  echo "Remove the services and networks of the current swarm..."
  docker service ls -q | xargs -r docker service rm >/dev/null
  # A network cannot go until the tasks on it have stopped.
  for _ in $(seq 60); do
    [ -z "$(docker ps -q --filter label=com.docker.swarm.service.id)" ] && break
    sleep 1
  done
  for net in $(swarm_networks); do
    remove_network "$net" || echo "WARNING: could not remove network $net" >&2
  done
fi
docker swarm leave --force || true
docker swarm init

# Whatever overlay network is here now came from an earlier swarm - this one has
# only just been made - and would collide with a network this one creates.
for net in $(swarm_networks); do
  echo "Remove network '$net' left behind by an earlier swarm..."
  if ! remove_network "$net"; then
    echo "WARNING: network $net could not be removed and still holds its subnet." >&2
    echo "         Restarting docker clears it: sudo systemctl restart docker" >&2
  fi
done

# Label the current node as control-plane
NODE_ID=$(docker info --format '{{.Swarm.NodeID}}')
if [ -n "$NODE_ID" ]; then
  echo "Labeling node $NODE_ID as control-plane..."
  docker node update --label-add hivepaas.role=control-plane "$NODE_ID"
fi

HIVEPAAS_DIR=hivepaas
HIVEPAAS_SSL_CERTS=$HIVEPAAS_DIR/ssl/certs
HIVEPAAS_FILES=$HIVEPAAS_DIR/files

mkdir -p $HIVEPAAS_DIR
mkdir -p $HIVEPAAS_SSL_CERTS
mkdir -p $HIVEPAAS_FILES

TRAEFIK_DYNAMIC=$HIVEPAAS_DIR/traefik/etc/dynamic
TRAEFIK_VAR_LOG=$HIVEPAAS_DIR/traefik/var/log

mkdir -p $TRAEFIK_DYNAMIC
mkdir -p $TRAEFIK_VAR_LOG

# Create some test files
echo "test" > $HIVEPAAS_FILES/test1.txt
echo "test" > $HIVEPAAS_FILES/test2.txt
echo "test" > $HIVEPAAS_FILES/test3.txt

# Download traefik conf files
echo "Download traefik config files..."
curl -sL "https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/dev/traefik/dynamic_conf.yml" -o $TRAEFIK_DYNAMIC/dynamic_conf.yml

# Gen self-signed SSL certs
if [ ! -f "$HIVEPAAS_SSL_CERTS/self-signed.key" ]; then
  echo "File '$HIVEPAAS_SSL_CERTS/self-signed.key' does not exist. Generate new file..."
  openssl req -x509 -days 365 -nodes -sha256 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
    -keyout $HIVEPAAS_SSL_CERTS/self-signed.key -out $HIVEPAAS_SSL_CERTS/self-signed.crt \
    -subj "/CN=*.dev.hivepaas.com"
fi

# Create overlay network for traefik to discover services
echo "Create overlay network 'hivepaas_net'..."
docker network create \
  --driver overlay \
  --attachable \
  --subnet 10.11.0.0/16 \
  --gateway 10.11.0.1 \
  --opt com.docker.network.driver.mtu=1380 \
  hivepaas_net || true

# Create some volumes
docker volume create test-vol-1
docker volume create test-vol-2

# Download dev_project_a.yaml
echo "Download dev_project_a.yaml..."
curl -sL "https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/dev/dev_project_a.yaml" -o dev_project_a.yaml

# Deploy dev_project_a stack
echo "Deploy dev project_a stack..."
docker stack deploy -c dev_project_a.yaml project_a

# Download hivepaas.yaml
echo "Download hivepaas.yaml..."
curl -sL "https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/dev/hivepaas.yaml" -o hivepaas.yaml

# Deploy hivepaas stack
echo "Deploy hivepaas stack..."
docker pull hivepaas/hivepaas-dev:latest # pull latest image
docker stack deploy -c hivepaas.yaml hivepaas

# Kernel OOM priority for the system services, so a user app is the one killed
# when memory runs out. Not in the stack file: `docker stack deploy` drops
# oom_score_adj, and a later deploy that changes a service resets it to 0 - run
# this loop again after one.
for svc in traefik db redis app worker updater agent; do
  docker service update --detach --quiet --oom-score-adj -500 "hivepaas_$svc" >/dev/null
done

sleep 10
docker run --net hivepaas_local_net \
  -e HP_PLATFORM=remote -e HP_DB_HOST=db -e HP_DB_PORT=5432 -e HP_DB_DB_NAME=hivepaas \
  -e HP_DB_USER=hivepaas -e HP_DB_PASSWORD=abc123 -e HP_DB_SSL_MODE=disable \
  -w /hivepaas hivepaas/hivepaas-dev:latest \
  make seed-data-with-clear

# Force restart the main app
docker service update --force hivepaas_app

#sleep 3
## docker restart $(docker ps -a -q -f status=running)
#TRAEFIK_CONT_ID=$(docker ps -f "status=running" | grep traefik | awk -F' ' '{print $1}')
#if [ -n "$TRAEFIK_CONT_ID" ]; then
#  docker container restart "$TRAEFIK_CONT_ID"
#fi

echo "---------------------------------------------------------------"
echo "DONE."
echo "---------------------------------------------------------------"
