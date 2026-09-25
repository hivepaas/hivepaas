#!/usr/bin/env bash
#
# End-to-end test of install.sh, in a docker:dind container: a silent install
# with this checkout's release info and stack files, then what a person would
# check - the dashboard by domain and by address, signing in, the admin's
# password gone - and the runs after it: one that must change nothing, one
# with another app secret and one without install.env that must stop, one
# after an address changed, and a --redeploy.
#
#   make test-installer-e2e     (bash deployment/release/install_e2e.sh)
#
# It pulls the release's images and takes several minutes. The container is
# privileged and runs a swarm of its own; HIVEPAAS_SWAP=false keeps the install
# from setting vm.swappiness, which is the kernel's and so the host's too.
# HIVEPAAS_E2E_KEEP=1 leaves the container running to look around in.
#
# The commands in single quotes run in the container, where their $ expand:
# shellcheck disable=SC2016

set -euo pipefail

HERE=$(cd "$(dirname "$0")" && pwd)
REPO=$(cd "$HERE/../.." && pwd)
NAME=hivepaas-install-e2e
DOMAIN=hivepaas.e2e.example.com
PASSWORD=e2e-password-123
INSTALL="HIVEPAAS_RELEASE_URL=file:///repo/release.signed.json HIVEPAAS_INSTALL_FILES_DIR=/repo/deployment/release \
HIVEPAAS_SWAP=false bash /repo/deployment/release/install.sh --yes"

cleanup() {
  if [ "${HIVEPAAS_E2E_KEEP:-}" != 1 ]; then docker rm -f "$NAME" >/dev/null 2>&1 || true; fi
}
trap cleanup EXIT

in_dind() {
  docker exec "$NAME" sh -c "$1"
}

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

pass() {
  printf 'ok: %s\n' "$*"
}

# expect NAME WANT GOT
expect() {
  if [ "$2" = "$3" ]; then pass "$1"; else fail "$1: want '$2', got '$3'"; fi
}

docker rm -f "$NAME" >/dev/null 2>&1 || true
# The swarm inside does not pull: every image is pulled below, before the deploy.
docker run -d --name "$NAME" --privileged --platform "linux/$(docker version --format '{{.Server.Arch}}')" \
  -e DOCKER_SERVICE_PREFER_OFFLINE_IMAGE=1 -v "$REPO":/repo:ro docker:dind >/dev/null
for _ in $(seq 60); do
  if in_dind 'docker info' >/dev/null 2>&1; then break; fi
  sleep 1
done
in_dind 'docker info' >/dev/null 2>&1 || fail "the Docker daemon in $NAME did not start"
in_dind 'apk add -q bash jq curl openssl >/dev/null'

# The HivePaaS images are built for amd64 only. Elsewhere they are pulled for
# amd64 and run emulated, and the stack is deployed without resolving images:
# resolving would pin their platform, and no node here has it.
docker exec -i "$NAME" bash -s >/dev/null <<'EOF'
set -e
HIVEPAAS_INSTALL_LIB=1 . /repo/deployment/release/install.sh
release_payload /repo/release.signed.json >/tmp/release.json
app=$(release_field /tmp/release.json beta appImage)
agent=$(release_field /tmp/release.json beta agentImage)
agent=${agent:-$(derive_agent_image "$app")}
platform=()
if [ "$(uname -m)" != x86_64 ]; then
  platform=(--platform linux/amd64)
  cat >/usr/local/sbin/docker <<'SHIM'
#!/bin/sh
if [ "$1" = stack ] && [ "$2" = deploy ]; then
  shift 2
  exec /usr/local/bin/docker stack deploy --resolve-image never "$@"
fi
exec /usr/local/bin/docker "$@"
SHIM
  chmod +x /usr/local/sbin/docker
fi
docker pull -q "${platform[@]}" "$app"
docker pull -q "${platform[@]}" "$agent"
for field in dbImage redisImage traefikImage; do
  docker pull -q "$(release_field /tmp/release.json beta "$field")"
done
EOF
pass "images pulled"

in_dind "HIVEPAAS_ADMIN_EMAIL=admin@example.com HIVEPAAS_ADMIN_PASSWORD=$PASSWORD HIVEPAAS_APP_DOMAIN=$DOMAIN \
  $INSTALL >/root/install.log 2>&1" || {
  in_dind 'tail -n 40 /root/install.log' >&2
  fail "the install"
}
pass "the install"

HOST_IP=$(in_dind "ip route get 1.1.1.1 | awk '{for (i = 1; i < NF; i++) if (\$i == \"src\") print \$(i + 1)}'")
expect "the dashboard by domain" 200 "$(in_dind "curl -sk -o /dev/null -w '%{http_code}' \
  --resolve $DOMAIN:443:127.0.0.1 https://$DOMAIN/_/ping")"
expect "the dashboard by address" 200 "$(in_dind "curl -sk -o /dev/null -w '%{http_code}' https://$HOST_IP/")"
expect "http by address goes to https" "302 https://$HOST_IP/" "$(in_dind "curl -s -o /dev/null \
  -w '%{http_code} %{redirect_url}' http://$HOST_IP/")"
login() {
  in_dind "curl -sk -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -H 'Origin: https://$1' \
    --resolve $DOMAIN:443:127.0.0.1 -d '{\"username\":\"admin\",\"password\":\"$2\"}' \
    https://$1/_/auth/login-with-password"
}
expect "the admin signs in" 200 "$(login "$DOMAIN" "$PASSWORD")"
expect "the admin signs in by address" 200 "$(login "$HOST_IP" "$PASSWORD")"
expect "a wrong password does not" 401 "$(login "$DOMAIN" wrong-password-1)"
expect "the admin's account left the app and worker" 0 "$(in_dind 'for s in app worker; do
  docker service inspect hivepaas_$s --format "{{range .Spec.TaskTemplate.ContainerSpec.Env}}{{println .}}{{end}}"
  done | grep -c HP_USER_ADMIN || true')"
oom_priorities() {
  in_dind 'for s in traefik db redis app worker updater agent; do
    docker service inspect hivepaas_$s --format "{{.Spec.TaskTemplate.ContainerSpec.OomScoreAdj}}"; done' | xargs
}
expect "every system service has OOM priority -500" "-500 -500 -500 -500 -500 -500 -500" "$(oom_priorities)"
expect "install.env is root's alone" "700 600" "$(in_dind 'stat -c %a /etc/hivepaas /etc/hivepaas/install.env' | xargs)"
expect "install.env says installed" 1 "$(in_dind 'grep -c "^HIVEPAAS_INSTALLED=.true.$" /etc/hivepaas/install.env')"
expect "install.env has no password" 0 "$(in_dind 'grep -c ADMIN_PASSWORD /etc/hivepaas/install.env || true')"

versions() {
  in_dind 'for s in traefik db redis app worker updater agent; do
    docker service inspect hivepaas_$s --format "{{.Version.Index}}"; done' | xargs
}
before=$(versions)
in_dind "$INSTALL >/root/install2.log 2>&1" || fail "a second run"
expect "a second run changes no service" "$before" "$(versions)"

if in_dind "HIVEPAAS_APP_SECRET=0123456789abcdef0123456789abcdefXX $INSTALL >/root/install3.log 2>&1"; then
  fail "a run with another app secret went on"
fi
pass "a run with another app secret stops"

in_dind 'mv /etc/hivepaas/install.env /root/install.env.bak'
if in_dind "$INSTALL >/root/install-lost.log 2>&1"; then fail "a run without install.env went on"; fi
in_dind 'grep -q "Put your copy of the file back" /root/install-lost.log' || fail "a run without install.env: no reason given"
in_dind 'mv /root/install.env.bak /etc/hivepaas/install.env'
pass "a run without install.env stops"

in_dind "docker service update --detach --quiet \
  --label-add 'traefik.http.routers.x-custom-router-ip.rule=Host(\`192.0.2.1\`)' hivepaas_app >/dev/null"
in_dind "$INSTALL >/root/install4.log 2>&1" || fail "a run after an address changed"
rule=$(in_dind "docker service inspect hivepaas_app \
  --format '{{index .Spec.Labels \"traefik.http.routers.x-custom-router-ip.rule\"}}'")
case "$rule" in
  *"Host(\`$HOST_IP\`)"*) pass "the route by address follows the address" ;;
  *) fail "the route by address is $rule" ;;
esac

in_dind "$INSTALL --redeploy >/root/install5.log 2>&1" || {
  in_dind 'tail -n 30 /root/install5.log' >&2
  fail "a redeploy"
}
pass "a redeploy"
expect "after it, the app's update stands" "" "$(in_dind 'docker service inspect hivepaas_app \
  --format "{{if .UpdateStatus}}{{.UpdateStatus.State}}{{end}}"' | grep rollback || true)"
expect "and every system service has OOM priority -500 again" "-500 -500 -500 -500 -500 -500 -500" "$(oom_priorities)"
expect "and the dashboard answers" 200 "$(in_dind "curl -sk -o /dev/null -w '%{http_code}' \
  --resolve $DOMAIN:443:127.0.0.1 https://$DOMAIN/_/ping")"

printf 'all end-to-end checks passed\n'
