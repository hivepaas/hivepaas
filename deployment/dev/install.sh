#!/bin/bash

set -eo pipefail

echo "---------------------------------------------------------------"
echo "INSTALL LocalPaaS"
echo "---------------------------------------------------------------"

# Delete all unused data that take the disk space
docker system prune -a -f

LOCALPAAS_DIR=localpaas
LOCALPAAS_SSL_CERTS=$LOCALPAAS_DIR/ssl/certs

mkdir -p $LOCALPAAS_DIR
mkdir -p $LOCALPAAS_SSL_CERTS

TRAEFIK_DYNAMIC=$LOCALPAAS_DIR/traefik/etc/dynamic
TRAEFIK_VAR_LOG=$LOCALPAAS_DIR/traefik/var/log

mkdir -p $TRAEFIK_DYNAMIC
mkdir -p $TRAEFIK_VAR_LOG

# Download traefik conf files
echo "Download traefik config files..."
curl -sL "https://raw.githubusercontent.com/localpaas/localpaas/main/deployment/dev/traefik/dynamic_conf.yml" -o $TRAEFIK_DYNAMIC/dynamic_conf.yml

# Gen self-signed SSL certs
if [ ! -f "$LOCALPAAS_SSL_CERTS/self-signed.key" ]; then
  echo "File '$LOCALPAAS_SSL_CERTS/self-signed.key' does not exist. Generate new file..."
  openssl req -x509 -days 365 -nodes -sha256 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
    -keyout $LOCALPAAS_SSL_CERTS/self-signed.key -out $LOCALPAAS_SSL_CERTS/self-signed.crt \
    -subj "/CN=*.dev.localpaas.com"
fi

# Create overlay network for traefik to discover services
echo "Create overlay network 'localpaas_net'..."
docker network create --driver overlay --attachable localpaas_net || true

# Download dev_project_a.yaml
echo "Download dev_project_a.yaml..."
curl -sL "https://raw.githubusercontent.com/localpaas/localpaas/main/deployment/dev/dev_project_a.yaml" -o dev_project_a.yaml

# Deploy dev_project_a stack
echo "Deploy dev project_a stack..."
docker stack deploy -c dev_project_a.yaml project_a

# Download localpaas.yaml
echo "Download localpaas.yaml..."
curl -sL "https://raw.githubusercontent.com/localpaas/localpaas/main/deployment/dev/localpaas.yaml" -o localpaas.yaml

# Deploy localpaas stack
echo "Deploy localpaas stack..."
docker pull localpaas/localpaas-dev:app-latest # pull latest image
docker stack deploy -c localpaas.yaml localpaas

sleep 10
docker run --net localpaas_internal_net \
  -e LP_PLATFORM=remote -e LP_DB_HOST=db -e LP_DB_PORT=5432 -e LP_DB_DB_NAME=localpaas \
  -e LP_DB_USER=localpaas -e LP_DB_PASSWORD=abc123 -e LP_DB_SSL_MODE=disable \
  -w /app localpaas/localpaas-dev:app-latest \
  make seed-data-with-clear

# Force restart the main app
docker service update --force localpaas_app

sleep 3
# docker restart $(docker ps -a -q -f status=running)
TRAEFIK_CONT_ID=$(docker ps -f "status=running" | grep traefik | awk -F' ' '{print $1}')
if [ -n "$TRAEFIK_CONT_ID" ]; then
  docker container restart "$TRAEFIK_CONT_ID"
fi

echo "---------------------------------------------------------------"
echo "DONE."
echo "---------------------------------------------------------------"
