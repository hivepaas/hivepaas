#!/bin/bash
#
# A throwaway HivePaaS for the end-to-end tests, on this machine's Docker,
# touching nothing else on it:
#   - postgres and redis as plain containers on a network of their own;
#   - docker itself in a dind container, made a swarm;
#   - the backend and its agent built from the backend repo and run INSIDE dind;
#   - Traefik, the release's, as the stack's proxy inside dind: the apps'
#     domains answer on this machine at 10180 (HTTP) and 10443 (HTTPS).
#
# The last point is the one that matters. The backend's docker client takes the
# default socket and ignores DOCKER_HOST: run anywhere else, it would work on
# this machine's own swarm - and on its first start it rewrites the routing of
# a `hivepaas_app` service it finds there. Inside dind the default socket is
# dind's. The script checks that before it says the stack is up.
#
# Usage: env/up.sh            (from e2e/; yarn env:up)
#   HP_BACKEND_DIR   the backend repo, default ../../hivepaas beside the dashboard
#   HP_TEMPLATES_SRC the app-templates repo the catalog is read from, default
#                    ../../app-templates beside the dashboard; none if missing
#   HP_E2E_PORT      where the dashboard answers, default 10100
#   HP_E2E_HTTP_PORT, HP_E2E_HTTPS_PORT
#                    where the apps' domains answer, default 10180 and 10443
#   HP_E2E_SKIP_BUILD=1  reuse the last build of the dashboard and the binaries

set -euo pipefail

ENV_DIR="$(cd "$(dirname "$0")" && pwd)"
DASHBOARD_DIR="$(cd "$ENV_DIR/../.." && pwd)"
BACKEND_DIR="$(cd "${HP_BACKEND_DIR:-$DASHBOARD_DIR/../hivepaas}" && pwd)"
TEMPLATES_SRC="${HP_TEMPLATES_SRC:-$DASHBOARD_DIR/../app-templates}"
PORT="${HP_E2E_PORT:-10100}"
HTTP_PORT="${HP_E2E_HTTP_PORT:-10180}"
HTTPS_PORT="${HP_E2E_HTTPS_PORT:-10443}"
BUILD="$ENV_DIR/.build"
NET=hp-e2e-net
DB=hp-e2e-db
REDIS=hp-e2e-redis
DIND=hp-e2e-dind

say() { printf '\033[1m%s\033[0m\n' "$*"; }
fail() {
	printf '\033[31m%s\033[0m\n' "$*" >&2
	docker exec "$DIND" tail -n 20 /hp/app.log >&2 2>/dev/null || true
	"$ENV_DIR/down.sh" >/dev/null 2>&1 || true
	exit 1
}
in_dind() { docker exec "$DIND" "$@"; }

if docker container inspect "$DIND" >/dev/null 2>&1; then
	fail "$DIND exists already: run env/down.sh first."
fi
docker image inspect hivepaas-devtools >/dev/null 2>&1 ||
	fail "No hivepaas-devtools image: run 'make init' in $BACKEND_DIR once."

if [ "${HP_E2E_SKIP_BUILD:-}" != 1 ]; then
	say "Building the dashboard and the backend"
	(cd "$DASHBOARD_DIR" && yarn -s build >/dev/null)
	arch="$(docker version -f '{{.Server.Arch}}')"
	mkdir -p "$BUILD"
	rm -rf "$BUILD/dist-dashboard"
	cp -R "$DASHBOARD_DIR/dist" "$BUILD/dist-dashboard"
	(cd "$BACKEND_DIR" && GOOS=linux GOARCH="$arch" CGO_ENABLED=0 go build -o "$BUILD/hivepaas-app" ./hivepaas_app/cmd/app)
	(cd "$BACKEND_DIR" && GOOS=linux GOARCH="$arch" CGO_ENABLED=0 go build -o "$BUILD/hivepaas-agent" ./hivepaas_app/cmd/agent)
fi
[ -x "$BUILD/hivepaas-app" ] || fail "Nothing built yet: run without HP_E2E_SKIP_BUILD."

say "Starting postgres, redis and dind"
docker network create "$NET" >/dev/null
docker run -d --name "$DB" --network "$NET" -e POSTGRES_USER=hivepaas -e POSTGRES_PASSWORD=abc123 \
	-e POSTGRES_DB=hivepaas postgres:18.6-alpine >/dev/null
docker run -d --name "$REDIS" --network "$NET" redis:8-alpine redis-server --requirepass abc123 >/dev/null
docker run -d --privileged --name "$DIND" --network "$NET" -e DOCKER_TLS_CERTDIR= -p "$PORT:$PORT" \
	-p "$HTTP_PORT:80" -p "$HTTPS_PORT:443" \
	docker:dind >/dev/null
for _ in $(seq 1 60); do docker exec "$DB" pg_isready -U hivepaas >/dev/null 2>&1 && break; sleep 1; done
for _ in $(seq 1 60); do in_dind docker info >/dev/null 2>&1 && break; sleep 1; done
in_dind docker swarm init >/dev/null

say "Migrating and seeding the database"
docker run --rm --network "$NET" -v "$BACKEND_DIR":/app:ro -e HP_DB_HOST="$DB" -e HP_DB_PORT=5432 \
	-e HP_DB_USER=hivepaas -e HP_DB_PASSWORD=abc123 -e HP_DB_DB_NAME=hivepaas hivepaas-devtools \
	sql-migrate up -config=hivepaas_app/db/dbconfig.yml -env=main >/dev/null
docker exec -i "$DB" psql -q -U hivepaas -d hivepaas <"$BACKEND_DIR/hivepaas_app/db/seed/seed.sql" >/dev/null
# The seed's notification channels are real ones - a mail account, Slack,
# Discord and Telegram, all defaults - and every deployment, health check and
# cleanup here would write to them. Gone: a test makes the channels it needs,
# pointed at an app of its own.
docker exec "$DB" psql -q -U hivepaas -d hivepaas \
	-c "DELETE FROM settings WHERE scope = '' AND type IN ('notification', 'im-service', 'email')" >/dev/null

say "Standing in for the stack inside dind"
# What the installer deploys and the backend looks for: its own services
# (stand-ins, never started) and the shared network. Their image is also what
# the backend runs its helper containers on, so it is pulled - into dind.
in_dind docker network create -d overlay --attachable hivepaas_net >/dev/null
in_dind docker pull -q alpine:3 >/dev/null
for app in app worker updater agent; do
	in_dind docker service create -d --no-resolve-image --name "hivepaas_$app" --replicas 0 \
		--network hivepaas_net --label com.docker.stack.namespace=hivepaas \
		--label "hivepaas.app.info={\"name\":\"$app\",\"key\":\"$app\"}" alpine:3 true >/dev/null
done

say "Starting Traefik inside dind"
# The proxy the release deploys, routing what the apps' service labels and the
# files HivePaaS writes ask for. It binds 80 and 443 on dind's own network, which
# up.sh publishes here; this machine's 80 and 443 are left alone.
traefik_image="$(sed -n 's/.*"traefikImage": *"\([^"]*\)".*/\1/p' "$BACKEND_DIR/release.json" | head -n 1)"
[ -n "$traefik_image" ] || fail "No traefikImage in $BACKEND_DIR/release.json."
in_dind docker pull -q "$traefik_image" >/dev/null
in_dind mkdir -p /hp/appdata/traefik/etc/dynamic /hp/appdata/ssl/certs
in_dind docker service create -d --no-resolve-image --name hivepaas_traefik \
	--network name=hivepaas_net,alias=hivepaas_traefik --label com.docker.stack.namespace=hivepaas \
	--label 'hivepaas.app.info={"name":"traefik","key":"traefik"}' --container-label hivepaas.component=traefik \
	--publish mode=host,target=80,published=80 --publish mode=host,target=443,published=443 \
	--mount type=bind,src=/hp/appdata/traefik/etc/dynamic,dst=/etc/traefik/dynamic \
	--mount type=bind,src=/hp/appdata/ssl/certs,dst=/etc/traefik/ssl/certs \
	--mount type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock,readonly \
	"$traefik_image" \
	--providers.swarm=true --providers.swarm.watch=true --providers.swarm.network=hivepaas_net \
	--providers.swarm.exposedbydefault=false --providers.file.directory=/etc/traefik/dynamic \
	--providers.file.watch=true --entrypoints.web.address=:80 --entrypoints.websecure.address=:443 \
	--entrypoints.websecure.http.tls=true --entrypoints.ping.address=127.0.0.1:8082 --ping=true \
	--ping.entrypoint=ping --log.level=INFO --accesslog=true --accesslog.format=json >/dev/null

say "Giving dind what the agent's image carries"
# kopia, the backup engine, from the image the release copies it from; and the
# node's root at /host, where the agent's container sees it and where a backup
# repository on a volume is looked for.
kopia_image="$(sed -n 's/^ARG KOPIA_IMAGE=//p' "$BACKEND_DIR/deployment/release/Dockerfile")"
in_dind docker pull -q "$kopia_image" >/dev/null
in_dind sh -c "c=\$(docker create '$kopia_image' /kopia) && docker cp \$c:/kopia /usr/local/bin/kopia && docker rm \$c" \
	>/dev/null
in_dind ln -s / /host

say "Starting the agent and the backend inside dind"
sed "s/__PORT__/$PORT/" "$ENV_DIR/config.toml" >"$BUILD/config.toml"
in_dind mkdir -p /hp/appdata
docker cp "$BUILD/hivepaas-app" "$DIND:/hp/hivepaas-app" >/dev/null
docker cp "$BUILD/hivepaas-agent" "$DIND:/hp/hivepaas-agent" >/dev/null
docker cp "$BUILD/config.toml" "$DIND:/hp/config.toml" >/dev/null
docker cp "$BUILD/dist-dashboard" "$DIND:/hp/dist-dashboard" >/dev/null
# The catalog, read from a checkout - which only the development environment
# honours - rather than from the signed release this build is not one of.
templates_env=""
if [ -d "$TEMPLATES_SRC/templates" ]; then
	in_dind mkdir -p /hp/app-templates
	tar -C "$TEMPLATES_SRC" --exclude=.git -cf - . | docker exec -i "$DIND" tar -C /hp/app-templates -xf -
	templates_env="HP_TEMPLATES_DIR=/hp/app-templates"
fi
docker exec -d -w /hp "$DIND" sh -c 'HP_CONFIG_FILE=config.toml ./hivepaas-agent >agent.log 2>&1'
docker exec -d -w /hp "$DIND" sh -c "HP_CONFIG_FILE=config.toml HP_APP_PATH=/hp/appdata $templates_env \
	HP_STORAGE_HOST_DIR=/hp/appdata HP_DEV_MODE_FORCE_AGENT_LOCAL=true ./hivepaas-app >app.log 2>&1"

base="http://localhost:$PORT"
for _ in $(seq 1 90); do curl -s -o /dev/null -m 2 "$base/api/auth/login-options" && break; sleep 1; done
token="$(curl -s -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"abc123"}' \
	"$base/api/auth/login-with-password" | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p' || true)"
[ -n "$token" ] || fail "The backend did not start; the end of its log:"

# The backend must see dind's swarm, and nothing else.
want="$(in_dind docker info -f '{{.Swarm.NodeID}}')"
seen="$(curl -s -H "Authorization: Bearer $token" "$base/api/cluster/nodes" |
	grep -o '"id":"[a-z0-9]\{25\}"' | sed 's/"id":"\(.*\)"/\1/' | sort -u | tr '\n' ' ' || true)"
[ "$seen" = "$want " ] || fail "The backend sees nodes '$seen', not dind's '$want': stopped."

say "Up: $base - admin / abc123; the apps' domains at :$HTTP_PORT and :$HTTPS_PORT. Stop it with env/down.sh."
