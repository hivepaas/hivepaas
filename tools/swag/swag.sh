#!/bin/bash
set -eo pipefail

DEVTOOLS_IMAGE=hivepaas-devtools

# Gen swagger.json
docker run --entrypoint "/go/bin/swag" --rm --volume "${PWD}":/app --volume "${HOME}/go/pkg/mod":/go/pkg/mod ${DEVTOOLS_IMAGE} init \
  -g hivepaas_app/interface/api/server/server.go -o docs/openapi --outputTypes json \
  --parseDependencyLevel 1 --requiredByDefault

# Convert swagger.json to OpenAPI v3 format
docker run --user $(id -u) --rm --volume "${PWD}:/app" openapitools/openapi-generator-cli generate \
  --skip-validate-spec -i /app/docs/openapi/swagger.json -g openapi -o /app/tmp/swago && \
  cp tmp/swago/openapi.json docs/openapi/swagger.json && rm -rf tmp/swago
