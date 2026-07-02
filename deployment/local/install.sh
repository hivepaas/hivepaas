#!/bin/bash

set -eo pipefail

echo "---------------------------------------------------------------"
echo "INSTALL HivePaaS LOCALLY"
echo "---------------------------------------------------------------"

# Delete all unused data that take the disk space
# docker system prune -a -f

# Reset whole cluster
# docker swarm leave --force || true
# docker swarm init

HIVEPAAS_DIR=.appdata/hivepaas
HIVEPAAS_SSL_CERTS=$HIVEPAAS_DIR/ssl/certs

mkdir -p $HIVEPAAS_DIR
mkdir -p $HIVEPAAS_SSL_CERTS

TRAEFIK_DYNAMIC=$HIVEPAAS_DIR/traefik/etc/dynamic

mkdir -p $TRAEFIK_DYNAMIC

# Copy traefik config files
echo "Copy traefik config files..."
cp deployment/local/traefik/dynamic_conf.yml $TRAEFIK_DYNAMIC/dynamic_conf.yml

# Gen self-signed SSL certs
if [ ! -f "$HIVEPAAS_SSL_CERTS/self-signed.key" ]; then
  echo "File '$HIVEPAAS_SSL_CERTS/self-signed.key' does not exist. Generate new file..."
  openssl req -x509 -days 365 -nodes -sha256 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
    -keyout $HIVEPAAS_SSL_CERTS/self-signed.key -out $HIVEPAAS_SSL_CERTS/self-signed.crt \
    -subj "/CN=*.swarm.localhost"
fi

# Init docker swarm
echo "Init docker swarm..."
docker swarm init || true

# Create overlay network for traefik to discover services
echo "Create overlay network 'hivepaas_net'..."
docker network create --driver overlay --attachable hivepaas_net || true

# Deploy hivepaas stack
echo "Deploy hivepaas stack..."
cp deployment/local/hivepaas.yaml $HIVEPAAS_DIR/../hivepaas.yaml
docker stack deploy -c $HIVEPAAS_DIR/../hivepaas.yaml hivepaas

sleep 5
make seed-data-with-clear

echo "---------------------------------------------------------------"
echo "DONE."
echo "---------------------------------------------------------------"
