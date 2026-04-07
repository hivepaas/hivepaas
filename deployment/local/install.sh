#!/bin/bash

set -eo pipefail

echo "---------------------------------------------------------------"
echo "INSTALL LocalPaaS LOCALLY"
echo "---------------------------------------------------------------"

# Delete all unused data that take the disk space
# docker system prune -a -f

LOCALPAAS_DIR=.appdata/localpaas
LOCALPAAS_SSL_CERTS=$LOCALPAAS_DIR/ssl/certs

mkdir -p $LOCALPAAS_DIR
mkdir -p $LOCALPAAS_SSL_CERTS

TRAEFIK_DYNAMIC=$LOCALPAAS_DIR/traefik/etc/dynamic

mkdir -p $TRAEFIK_DYNAMIC

# Copy traefik config files
echo "Copy traefik config files..."
cp deployment/local/traefik/dynamic_conf.yml $TRAEFIK_DYNAMIC/dynamic_conf.yml

# Gen self-signed SSL certs
if [ ! -f "$LOCALPAAS_SSL_CERTS/self-signed.key" ]; then
  echo "File '$LOCALPAAS_SSL_CERTS/self-signed.key' does not exist. Generate new file..."
  openssl req -x509 -days 365 -nodes -sha256 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
    -keyout $LOCALPAAS_SSL_CERTS/self-signed.key -out $LOCALPAAS_SSL_CERTS/self-signed.crt \
    -subj "/CN=*.swarm.localhost"
fi

# Init docker swarm
echo "Init docker swarm..."
docker swarm init || true

# Create overlay network for traefik to discover services
echo "Create overlay network 'localpaas_net'..."
docker network create --driver overlay --attachable localpaas_net || true

# Deploy localpaas stack
echo "Deploy localpaas stack..."
cp deployment/local/localpaas.yaml $LOCALPAAS_DIR/../localpaas.yaml
docker stack deploy -c $LOCALPAAS_DIR/../localpaas.yaml localpaas

sleep 5
make seed-data-with-clear

echo "---------------------------------------------------------------"
echo "DONE."
echo "---------------------------------------------------------------"
