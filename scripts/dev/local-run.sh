#!/bin/bash
#
# Runs the local build on the host: `go run` with the environment the local install
# needs. The everyday way to work. Run through `make local-app-run` and
# `make local-agent-run`.
#
# Usage: local-run.sh app|agent

set -euo pipefail

cd "$(dirname "$0")/../.."

LOCAL_CONFIG="config/config.local.toml"
# Absolute, because HP_STORAGE_BIND_SOURCE is handed to docker as the source of a
# bind mount and the daemon resolves it on the host, not against this process's
# working directory. In the stack the two are deliberately different - the
# container sees its data at /var/lib/hivepaas while docker binds it from the host
# path - but running on the host there is no boundary between them.
LOCAL_APP_PATH="$(pwd)/.appdata/hivepaas"

case "${1:-}" in
app)
  mkdir -p "$LOCAL_APP_PATH"
  HP_CONFIG_FILE="$LOCAL_CONFIG" \
    HP_APP_PATH="$LOCAL_APP_PATH" \
    HP_STORAGE_BIND_SOURCE="$LOCAL_APP_PATH" \
    exec go run ./hivepaas_app/cmd/app/...
  ;;
agent)
  # No HP_APP_PATH here, on purpose. Nothing on the agent side decrypts anything:
  # the app decrypts and sends plaintext over gRPC.
  HP_CONFIG_FILE="$LOCAL_CONFIG" exec go run ./hivepaas_app/cmd/agent/...
  ;;
*)
  echo "usage: $0 app|agent" >&2
  exit 2
  ;;
esac
