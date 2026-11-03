#!/bin/bash
# Takes the throwaway HivePaaS down: its containers, their volumes, the network.
docker rm -f -v hp-e2e-db hp-e2e-redis hp-e2e-dind >/dev/null 2>&1
docker network rm hp-e2e-net >/dev/null 2>&1
echo "Down."
