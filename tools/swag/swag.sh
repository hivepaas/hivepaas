#!/bin/bash
set -eo pipefail

DEVTOOLS_IMAGE=hivepaas-devtools

# The converter to OpenAPI 3, by tag and digest: its output is committed, and
# `latest` moved under it - the file then changed on a run that changed nothing.
OPENAPI_GENERATOR_IMAGE=openapitools/openapi-generator-cli:v7.25.0@sha256:2ab0a9680222de65dc9d3baf861aa02b99e1b80c211d8221ebf3ae8f8a102524

# SWAG_BIN runs a swag installed on the host instead of the one in the devtools
# image - CI installs the version tools/docker/Dockerfile pins.
SWAG_BIN=${SWAG_BIN:-}

# Configurable string replacement pairs: "search_pattern" "replacement"
REPLACEMENTS=(
  # Strip common module and usecase prefixes
  "github_com_hivepaas_hivepaas_hivepaas_app_" ""
  "github_com_hivepaas_hivepaas_" ""
  "usecase_" ""
)

apply_replacements() {
  local target_file="$1"
  if [[ ! -f "$target_file" ]]; then
    return 0
  fi

  # Step 1: Apply configured string replacement pairs
  for ((i=0; i<${#REPLACEMENTS[@]}; i+=2)); do
    local src="${REPLACEMENTS[i]}"
    local dst="${REPLACEMENTS[i+1]}"
    if sed --version >/dev/null 2>&1; then
      # GNU sed (Linux / Docker)
      sed -i "s|${src}|${dst}|g" "$target_file"
    else
      # BSD sed (macOS)
      sed -i '' "s|${src}|${dst}|g" "$target_file"
    fi
  done

  # Step 2: Replace any XXXuc_XXXdto with XXXdto (where XXX is identical)
  if command -v perl >/dev/null 2>&1; then
    # Universal Perl regex with backreference support (macOS & Linux)
    perl -i -pe 's/([a-zA-Z0-9]+)uc_\1dto/$1dto/g' "$target_file"
  elif sed --version >/dev/null 2>&1; then
    # GNU sed (Linux)
    sed -i -E 's/([a-zA-Z0-9]+)uc_\1dto/\1dto/g' "$target_file"
  fi
}

# Directories swag must not read: the dev helper's handlers, and the tools - the
# openapi tool's test fixtures are handlers with routes of their own.
EXCLUDE_API_DIRS=(
  "hivepaas_app/interface/api/handler/devhelperhandler"
  "./hivepaas_app/interface/api/handler/devhelperhandler"
  "tools"
  "./tools"
)

# Combine exclusions with both standard and ./ relative prefixes for swag parser
EXCLUDE_ITEMS=$(IFS=,; echo "${EXCLUDE_API_DIRS[*]}")

# Gen swagger.json
SWAG_ARGS=(init
  -g hivepaas_app/interface/api/server/server.go -o docs/openapi --outputTypes json
  --parseDependencyLevel 1 --requiredByDefault --overridesFile .swaggo
  --exclude "${EXCLUDE_ITEMS}")
if [[ -n "${SWAG_BIN}" ]]; then
  "${SWAG_BIN}" "${SWAG_ARGS[@]}"
else
  docker run --entrypoint "/go/bin/swag" --rm --volume "${PWD}":/app --volume "${HOME}/go/pkg/mod":/go/pkg/mod \
    ${DEVTOOLS_IMAGE} "${SWAG_ARGS[@]}"
fi

# Shorten type names in swagger.json before conversion
apply_replacements "docs/openapi/swagger.json"

# Convert swagger.json to OpenAPI v3 format
docker run --user $(id -u) --rm --volume "${PWD}:/app" ${OPENAPI_GENERATOR_IMAGE} generate \
  --skip-validate-spec -i /app/docs/openapi/swagger.json -g openapi -o /app/tmp/swago && \
  cp tmp/swago/openapi.json docs/openapi/swagger.json && rm -rf tmp/swago

# Apply final replacements on generated OpenAPI v3 file
apply_replacements "docs/openapi/swagger.json"

# What swag cannot say about the types - optional, nullable, any value - from
# the Go types themselves. See tools/openapi.
go run ./tools/openapi fix docs/openapi/swagger.json

# The API level a CLI must be built for to write, raised when the request of an
# existing write operation changed: docs/openapi/api-level.json, the spec's
# x-api-level and base.APILevel.
go run ./tools/openapi level docs/openapi/swagger.json
