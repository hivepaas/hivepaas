#!/bin/bash

set -eo pipefail

echo "---------------------------------------------------------------"
echo "INSTALL HivePaaS"
echo "---------------------------------------------------------------"

# Delete all unused data that take the disk space
# docker system prune -a -f

# Reset whole cluster
docker swarm leave --force || true
docker swarm init

HIVEPAAS_DIR=hivepaas
HIVEPAAS_SSL_CERTS=$HIVEPAAS_DIR/ssl/certs

mkdir -p $HIVEPAAS_DIR
mkdir -p $HIVEPAAS_SSL_CERTS

TRAEFIK_DYNAMIC=$HIVEPAAS_DIR/traefik/etc/dynamic
TRAEFIK_VAR_LOG=$HIVEPAAS_DIR/traefik/var/log

mkdir -p $TRAEFIK_DYNAMIC
mkdir -p $TRAEFIK_VAR_LOG

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
docker network create --driver overlay --attachable hivepaas_net || true

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
