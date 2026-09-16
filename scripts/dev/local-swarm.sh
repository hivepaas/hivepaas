#!/bin/bash
#
# Runs the local build as swarm services. Run through `make local-app-up` and
# friends.
#
# Day to day the backend runs straight from the IDE or `make local-app-run`. This
# is for what that cannot reach: code that reads the app's own swarm service, the
# updater paths, placement constraints, the agent's view of the node it sits on,
# or anything gated on `platform = "remote"`.
#
# Each image is one layer on top of the published dev image, which already carries
# every runtime dependency (kopia, docker-cli, sql-migrate, git-lfs, dashboard).
# Building deployment/dev/Dockerfile instead would clone the dashboard repo and
# run a yarn build to pick up a change to one Go file.
#
# The :local tags exist on this daemon and nowhere else, which is fine for a
# single-node local swarm and is the whole of why this does not work on a
# multi-node one: the other nodes have nothing to pull.
#
# Usage: local-swarm.sh image|up|down|logs app|agent

set -euo pipefail

cd "$(dirname "$0")/../.."

APP_SERVICE="hivepaas_app"
APP_IMAGE="hivepaas/hivepaas-dev:local"
APP_IMAGE_BASE="hivepaas/hivepaas-dev:latest"
APP_CTX="tmp/img-app"

AGENT_SERVICE="hivepaas_agent"
AGENT_IMAGE="hivepaas/hivepaas-agent-dev:local"
AGENT_IMAGE_BASE="hivepaas/hivepaas-agent-dev:latest"
AGENT_CTX="tmp/img-agent"

# Same as PROD_LDFLAGS in the Makefile.
LDFLAGS="-s -w"

# --no-resolve-image: a :local tag exists on this daemon and nowhere else, and
#   swarm's default is to ask a registry to resolve a tag to a digest first.
# --force: the tag never changes, so without it swarm reads the spec as unchanged
#   and leaves the task running the previous binary.
# --detach: an attached update does not return until swarm calls the rollout
#   converged, and converged means update_config's monitor has elapsed - 240s for
#   hivepaas_app, measured at 4m13s for an update whose container was serving
#   within seconds. The monitor is what arms the rollback and is not worth
#   shortening, so tools/swarm/wait-for-task.sh waits for the task instead.
svc_update() {
  docker service update --detach --quiet --force --no-resolve-image "$@" >/dev/null
}

# The daemon's arch, not the host's.
image_arch() {
  docker version --format '{{.Server.Arch}}'
}

running_task() {
  docker service ps "$1" --filter desired-state=running -q | head -1
}

# Remembers the image a service runs before it is replaced with a :local one, so
# `down` can put it back. `docker service rollback` cannot be used for that: run
# `up` twice and the spec it rolls back to is the previous :local one.
remember_image() {
  local service="$1" ctx="$2" img
  img="$(docker service inspect "$service" --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}')"
  case "$img" in
  *:local | *:local@*) ;;
  *) echo "$img" >"$ctx/.previous-image" ;;
  esac
}

# Its own context directory rather than the repo root: .dockerignore only drops
# README.*, so a root context would ship vendor/ and .git/ on every build.
#
# config/ and hivepaas_app/db are copied in so local edits to settings and
# migrations take effect; the base image's copies are from whenever it was built.
# An empty dist-dashboard means the dashboard was never built here - copying it
# would replace the base image's with nothing and serve a blank page, so that
# layer is left out instead.
app_image() {
  mkdir -p "$APP_CTX"
  GOOS=linux GOARCH="$(image_arch)" CGO_ENABLED=0 \
    go build -ldflags="$LDFLAGS" -o "$APP_CTX/hivepaas" ./hivepaas_app/cmd/app/...
  rm -rf "$APP_CTX/config" "$APP_CTX/hivepaas_app" "$APP_CTX/dist-dashboard"
  cp -R config "$APP_CTX/config"
  mkdir -p "$APP_CTX/hivepaas_app"
  cp -R hivepaas_app/db "$APP_CTX/hivepaas_app/db"
  printf 'FROM %s\nWORKDIR /hivepaas\nCOPY hivepaas ./hivepaas\nCOPY config ./config\nCOPY hivepaas_app/db ./hivepaas_app/db\n' \
    "$APP_IMAGE_BASE" >"$APP_CTX/Dockerfile"
  if [ -n "$(ls -A dist-dashboard 2>/dev/null)" ]; then
    cp -R dist-dashboard "$APP_CTX/dist-dashboard"
    printf 'COPY dist-dashboard ./dist-dashboard\n' >>"$APP_CTX/Dockerfile"
  else
    echo "dist-dashboard is empty, keeping the one in $APP_IMAGE_BASE - build yours with 'make local-build-dashboard'"
  fi
  docker build -q -t "$APP_IMAGE" "$APP_CTX" >/dev/null && echo "built $APP_IMAGE"
}

# HP_RUN_MODE=app+worker: what the stack itself runs this service as, so the task
# is the whole thing rather than half of it - the deploy, clone and periodic-job
# executors all live on the worker side. Set explicitly rather than left to
# config.development.toml so `docker service inspect` says what the task is doing.
# Note that the IDE build is usually still up on the same database, and then both
# are pulling from the same task queue; stop one of them when that matters.
#
# The URL is whatever domain the install itself holds, not the `app.dev.localhost`
# in deployment/local/hivepaas.yaml: the app rewrites its own traefik labels from
# the domains in the database, so the stack's static labels are replaced the first
# time it runs. Both builds read the same database, so both answer on the same
# host - `docker service inspect hivepaas_app` shows the labels in force.
app_up() {
  app_image
  local prev
  prev="$(running_task "$APP_SERVICE")"
  remember_image "$APP_SERVICE" "$APP_CTX"
  svc_update --image "$APP_IMAGE" --env-add HP_RUN_MODE=app+worker --replicas 1 "$APP_SERVICE"
  bash tools/swarm/wait-for-task.sh "$APP_SERVICE" "$prev"
  echo "$APP_SERVICE is running the local build - https://localhost"
  echo "logs: make local-app-logs   stop: make local-app-down"
}

# Back to the stack's own image, environment and replica count. Nothing waits for
# a task here - there is not going to be one.
#
# With nothing remembered, the image is left exactly as it is rather than set to
# the base image: `docker stack deploy` pins this service to a digest and the bare
# tag does not resolve on this daemon at all, so writing it would trade a
# reference that works for one that has to be fetched.
app_down() {
  if [ -f "$APP_CTX/.previous-image" ]; then
    local img
    img="$(cat "$APP_CTX/.previous-image")"
    svc_update --image "$img" --env-rm HP_RUN_MODE --replicas 0 "$APP_SERVICE"
    echo "$APP_SERVICE scaled to 0, image back to $img"
  else
    svc_update --env-rm HP_RUN_MODE --replicas 0 "$APP_SERVICE"
    echo "$APP_SERVICE scaled to 0; no remembered image, run 'make local-deploy' to restore the stack's"
  fi
}

# The agent image holds only the binary and config - no dashboard, no migrations -
# so this context is the binary plus a few kilobytes.
agent_image() {
  mkdir -p "$AGENT_CTX"
  GOOS=linux GOARCH="$(image_arch)" CGO_ENABLED=0 \
    go build -ldflags="$LDFLAGS" -o "$AGENT_CTX/hivepaas-agent" ./hivepaas_app/cmd/agent/...
  rm -rf "$AGENT_CTX/config"
  cp -R config "$AGENT_CTX/config"
  printf 'FROM %s\nWORKDIR /hivepaas\nCOPY hivepaas-agent ./hivepaas-agent\nCOPY config ./config\n' \
    "$AGENT_IMAGE_BASE" >"$AGENT_CTX/Dockerfile"
  docker build -q -t "$AGENT_IMAGE" "$AGENT_CTX" >/dev/null && echo "built $AGENT_IMAGE"
}

# Unlike the app, this service is meant to be up: it is global, the stack starts
# it, and nothing else answers the app's gRPC calls for a node. So there is no
# scaling it to zero - `down` puts the published image back instead.
agent_up() {
  agent_image
  local prev
  prev="$(running_task "$AGENT_SERVICE")"
  remember_image "$AGENT_SERVICE" "$AGENT_CTX"
  svc_update --image "$AGENT_IMAGE" "$AGENT_SERVICE"
  bash tools/swarm/wait-for-task.sh "$AGENT_SERVICE" "$prev"
  echo "$AGENT_SERVICE is running the local build on every node"
  echo "logs: make local-agent-logs   stop: make local-agent-down"
}

# The agent has to be running something, so with nothing remembered this does fall
# back to the base image - unlike the app above, that tag does resolve.
agent_down() {
  local prev img
  prev="$(running_task "$AGENT_SERVICE")"
  img="$(cat "$AGENT_CTX/.previous-image" 2>/dev/null || echo "$AGENT_IMAGE_BASE")"
  svc_update --image "$img" "$AGENT_SERVICE"
  bash tools/swarm/wait-for-task.sh "$AGENT_SERVICE" "$prev"
  echo "$AGENT_SERVICE is back on $img"
}

usage() {
  echo "usage: $0 image|up|down|logs app|agent" >&2
  exit 2
}

[ $# -eq 2 ] || usage

case "$1:$2" in
image:app) app_image ;;
up:app) app_up ;;
down:app) app_down ;;
logs:app) exec docker service logs -f --tail 100 "$APP_SERVICE" ;;
image:agent) agent_image ;;
up:agent) agent_up ;;
down:agent) agent_down ;;
logs:agent) exec docker service logs -f --tail 100 "$AGENT_SERVICE" ;;
*) usage ;;
esac
