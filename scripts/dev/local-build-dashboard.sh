#!/bin/bash
#
# Builds the dashboard from its repository and puts the result where the backend
# serves it from. Run through `make local-build-dashboard`.
#
# Environment:
#   HP_FE_DIR  the dashboard checkout (default: ../hivepaas-dashboard)

set -euo pipefail

cd "$(dirname "$0")/../.."

HP_FE_DIR="${HP_FE_DIR:-../hivepaas-dashboard}"

(cd "$HP_FE_DIR" && git pull && yarn install && yarn build)
rm -rf dist-dashboard
mv "$HP_FE_DIR/dist" dist-dashboard
