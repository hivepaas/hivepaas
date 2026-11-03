#!/bin/bash
#
# Builds the dashboard (web/) and puts the result where the backend serves it
# from. Run through `make local-build-dashboard`.

set -euo pipefail

cd "$(dirname "$0")/../.."

(cd web && yarn install --frozen-lockfile && yarn build)
rm -rf dist-dashboard
mv web/dist dist-dashboard
