#!/bin/bash
# Generates a Go client from docs/openapi/swagger.json with oapi-codegen, as the
# CLI does, and builds it. Some faults only show there: an operation id or an
# enum value given twice is a method or a switch case declared twice, and the
# client does not compile. Once hivepaas-cli checks itself against the spec (the
# cli-compat job), this is the same check made by hand.
set -eo pipefail

OAPI_CODEGEN_VERSION=v2.8.0
SPEC="${1:-docs/openapi/swagger.json}"

dir=$(mktemp -d)
trap 'rm -rf "${dir}"' EXIT
mkdir -p "${dir}/api"
cp "${SPEC}" "${dir}/api/openapi.json"
cat > "${dir}/api/cfg.yaml" <<EOF
package: api
generate:
  models: true
  client: true
output: ${dir}/api/client.gen.go
EOF
(cd "${dir}" && go mod init example.com/specclient > /dev/null 2>&1)
go run "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}" \
  -config "${dir}/api/cfg.yaml" "${dir}/api/openapi.json"
(cd "${dir}" && go mod tidy > /dev/null 2>&1 && go build ./...)
echo "the Go client generated from ${SPEC} builds"
