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

# The stack is deployed from HIVEPAAS_ROOT, not from the repo root - see the
# deploy step at the bottom for why that matters.
HIVEPAAS_ROOT=.appdata
HIVEPAAS_DIR=$HIVEPAAS_ROOT/hivepaas
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

# Label the current node as control-plane
NODE_ID=$(docker info --format '{{.Swarm.NodeID}}')
if [ -n "$NODE_ID" ]; then
  echo "Labeling node $NODE_ID as control-plane..."
  docker node update --label-add hivepaas.role=control-plane "$NODE_ID"
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

# Follow the database where an upgrade moved it.
#
# A postgres major upgrade writes the new cluster to a volume of its own and
# repoints the running service at it. The stack file still says what it always
# said, so deploying again without this would send the service back to the old
# volume and, with it, back to the state the database was in before the upgrade.
# The updater records the new values here; a fresh install has no such file.
DB_VOLUME_ENV=$HIVEPAAS_DIR/system/update/db-volume.env
if [ -f "$DB_VOLUME_ENV" ]; then
  echo "Using the database volume recorded by a previous upgrade:"
  cat "$DB_VOLUME_ENV"
  set -a
  # shellcheck disable=SC1090
  . "$DB_VOLUME_ENV"
  set +a
fi

# Deploy hivepaas stack
#
# The deploy runs from $HIVEPAAS_ROOT, which is why the stack file is copied there
# first. The file names host paths two different ways: traefik's volumes are
# relative (./hivepaas/traefik/...) and the app's are ${PWD}/hivepaas, substituted
# from the environment of the shell running the deploy. They only name the same
# directory when that shell is sitting in $HIVEPAAS_ROOT. Deploy from the repo
# root instead and docker accepts it without a word, having bound the app to an
# empty <repo>/hivepaas while traefik reads the config written above.
echo "Deploy hivepaas stack..."
cp deployment/local/hivepaas.yaml $HIVEPAAS_ROOT/hivepaas.yaml
(cd $HIVEPAAS_ROOT && docker stack deploy -c hivepaas.yaml hivepaas)

# Kernel OOM priority for the system services, so a user app is the one killed
# when memory runs out. Not in the stack file: `docker stack deploy` drops
# oom_score_adj, and a later deploy that changes a service resets it to 0 - run
# this loop again after one.
for svc in traefik db redis app worker updater agent; do
  docker service update --detach --quiet --oom-score-adj -500 "hivepaas_$svc" >/dev/null
done

# Assert what the comment above explains, because the failure it describes is
# silent: a stack that comes up, serves nothing it was given, and looks fine.
#
# Matched as a suffix rather than compared: Docker Desktop rewrites bind sources
# to /host_mnt/<path>, so the absolute path is a tail of what the daemon reports.
EXPECTED_DATA_DIR=$(cd $HIVEPAAS_DIR && pwd)
APP_BIND=$(docker service inspect hivepaas_app \
  --format '{{(index .Spec.TaskTemplate.ContainerSpec.Mounts 0).Source}}')
case "$APP_BIND" in
  *"$EXPECTED_DATA_DIR") ;;
  *)
    echo "ERROR: hivepaas_app is bound to '$APP_BIND', expected it to end in" >&2
    echo "       '$EXPECTED_DATA_DIR'. The stack was deployed from the wrong" >&2
    echo "       directory; re-run this script from the repository root." >&2
    exit 1
    ;;
esac

sleep 5
make seed-data-with-clear

echo "---------------------------------------------------------------"
echo "DONE."
echo "---------------------------------------------------------------"
