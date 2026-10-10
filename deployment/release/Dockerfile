# The HivePaaS app image of a release: the backend, and the dashboard of the same
# version, built from the tag the release workflow passes as DASHBOARD_REF.
#
# It differs from deployment/dev/Dockerfile in what makes a release a release:
# - the dashboard is its tag, not whatever main held when the image was built;
# - the binary is a production build (make build-prod);
# - what it ships beside the binary is pinned: sql-migrate and the x/crypto it
#   is built with by version, kopia's image by digest;
# - only the configuration files a release runs with are copied: the stack sets
#   HP_CONFIG_FILE to config/config.<beta|production>.toml, and a developer's
#   config.local.toml has no place in a published image.

# kopia, the backup engine: built from source into an image of its own
# (deployment/kopia/Dockerfile) and pinned by digest - the same one in the
# release's and the dev's Dockerfiles, app and agent (docs/RELEASING.md).
ARG KOPIA_IMAGE=ghcr.io/hivepaas/kopia:0.23.1-2@sha256:b98b6a5f14746f7007acebe0e0b9c0e50b2a59cb209320c94e208380eabb2b36
FROM ${KOPIA_IMAGE} AS kopia

############################################################
# Build dashboard
FROM node:26-alpine AS dashboard_builder
ARG DASHBOARD_REF
RUN test -n "$DASHBOARD_REF" || { echo "DASHBOARD_REF (the dashboard's release tag) is required" >&2; exit 1; }
RUN apk update && apk upgrade --no-cache && apk add --update --no-cache git yarn

WORKDIR /build

RUN --mount=type=cache,target=/tmp/yarn-cache \
    git clone --depth=1 --branch "$DASHBOARD_REF" https://github.com/hivepaas/hivepaas-dashboard.git && \
    cd hivepaas-dashboard && \
    yarn config set network-timeout 300000 && \
    yarn config set cache-folder /tmp/yarn-cache && \
    yarn install --frozen-lockfile && \
    yarn build && \
    mv dist /dist-dashboard

############################################################
# Build backend
FROM golang:1.27-alpine AS builder

RUN apk update && \
    apk add --no-cache -U tzdata make git build-base ca-certificates && \
    rm -rf /var/cache/apk/*

ENV CGO_ENABLED=0

WORKDIR /build

COPY . .

RUN make mod
RUN make build-prod

# The database migration tool the app runs on an update (sysupdateservice).
# Built in a module of its own that raises golang.org/x/crypto past the one
# its release requires: the image scan before a release refuses that one,
# though govulncheck finds none of its flaws called by sql-migrate.
ARG SQL_MIGRATE_VERSION=v1.8.1
ARG SQL_MIGRATE_CRYPTO_VERSION=v0.57.0
RUN mkdir /tmp/sql-migrate && cd /tmp/sql-migrate && \
    go mod init sql-migrate-build && \
    go get "github.com/rubenv/sql-migrate/sql-migrate@${SQL_MIGRATE_VERSION}" \
        "golang.org/x/crypto@${SQL_MIGRATE_CRYPTO_VERSION}" && \
    go build -trimpath -o /build/sql-migrate github.com/rubenv/sql-migrate/sql-migrate && \
    rm -rf /tmp/sql-migrate

############################################################
# The image
FROM alpine:3.24
ARG VERSION
ARG REVISION

LABEL org.opencontainers.image.title="HivePaaS" \
      org.opencontainers.image.source="https://github.com/hivepaas/hivepaas" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}"

RUN apk update && \
    apk add --no-cache ca-certificates make openssl postgresql-client git git-lfs openssh-client zstd lz4 docker-cli docker-cli-buildx && \
    git lfs install --system && \
    rm -rf /var/cache/apk/*

# kopia, the backup engine, and its license.
COPY --from=kopia /kopia /usr/local/bin/kopia
COPY --from=kopia /LICENSE /usr/share/licenses/kopia/LICENSE
RUN kopia --version

WORKDIR /hivepaas

COPY --from=builder /build/hivepaas .
COPY --from=builder /build/hivepaas_app/db ./hivepaas_app/db
COPY --from=builder /build/config/config.production.toml /build/config/config.beta.toml ./config/
COPY --from=builder /build/sql-migrate /usr/local/bin/sql-migrate
COPY --from=dashboard_builder /dist-dashboard ./dist-dashboard

CMD ["./hivepaas"]
