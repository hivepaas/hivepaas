#!/bin/bash
#
# Waits for a swarm service to be running a task other than the one it was
# running before the update it has just been given.
#
# `docker service update` can be told to wait, but it only returns once swarm
# calls the rollout converged, and converged means update_config.monitor has
# elapsed - 240s for hivepaas_app, long after the container is serving. The
# make targets that use this pass --detach and poll here instead.
#
# Comparing against the previous task id is the point. For a second or two after
# an update the old task is still the one with desired-state=running, so polling
# for "a running task" on its own reports success before anything has changed.
#
# Usage: wait-for-task.sh <service> [previous-task-id] [timeout-seconds]

set -eo pipefail

SERVICE="$1"
PREV_TASK="$2"
TIMEOUT_SECS="${3:-180}"

if [ -z "$SERVICE" ]; then
  echo "usage: $0 <service> [previous-task-id] [timeout-seconds]" >&2
  exit 2
fi

printf 'starting'
DEADLINE=$(($(date +%s) + TIMEOUT_SECS))

while [ "$(date +%s)" -lt "$DEADLINE" ]; do
  TASK=$(docker service ps "$SERVICE" --filter desired-state=running -q | head -1)
  if [ -n "$TASK" ] && [ "$TASK" != "$PREV_TASK" ]; then
    case "$(docker inspect --format '{{.Status.State}}' "$TASK")" in
      running)
        echo " running"
        exit 0
        ;;
      failed | rejected)
        echo " failed"
        docker service ps "$SERVICE" --no-trunc | head -3
        exit 1
        ;;
    esac
  fi
  printf '.'
  sleep 2
done

echo " timed out after ${TIMEOUT_SECS}s"
docker service ps "$SERVICE" --no-trunc | head -3
exit 1
