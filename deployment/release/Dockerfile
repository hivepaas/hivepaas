# The HivePaaS app image of a release: the backend, and the dashboard of the same
# version, built from the tag the release workflow passes as DASHBOARD_REF.
#
# It differs from deployment/dev/Dockerfile in what makes a release a release:
# - the dashboard is its tag, not whatever main held when the image was built;
# - the binary is a production build (make build-prod);
# - downloads are pinned: sql-migrate and the x/crypto it is built with by
#   version, kopia by checksum;
# - only the configuration files a release runs with are copied: the stack sets
#   HP_CONFIG_FILE to config/config.<beta|production>.toml, and a developer's
#   config.local.toml has no place in a published image.

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
ARG TARGETARCH
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

# kopia, the backup engine. The checksums are those of the release's
# checksums.txt (https://github.com/kopia/kopia/releases/tag/v0.23.1): change
# them with the version.
ARG KOPIA_VERSION=0.23.1
RUN case "${TARGETARCH}" in \
        amd64) KOPIA_ARCH="x64";   KOPIA_SHA256="416d0f84a3dbb321a8b2d8f0997b1a0a6e915babe79ee76fa6e4d2bd1e1c5178" ;; \
        arm64) KOPIA_ARCH="arm64"; KOPIA_SHA256="a4ffbc019e0b0f932e2632054e73ec521dc1e80172a00095369c53ecf4e5a6cb" ;; \
        *)     echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    wget -qO /tmp/kopia.tar.gz \
        "https://github.com/kopia/kopia/releases/download/v${KOPIA_VERSION}/kopia-${KOPIA_VERSION}-linux-${KOPIA_ARCH}.tar.gz" && \
    echo "${KOPIA_SHA256}  /tmp/kopia.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/kopia.tar.gz -C /tmp && \
    mv "/tmp/kopia-${KOPIA_VERSION}-linux-${KOPIA_ARCH}/kopia" /usr/local/bin/kopia && \
    chmod +x /usr/local/bin/kopia && \
    rm -rf /tmp/kopia* && \
    kopia --version

WORKDIR /hivepaas

COPY --from=builder /build/hivepaas .
COPY --from=builder /build/hivepaas_app/db ./hivepaas_app/db
COPY --from=builder /build/config/config.production.toml /build/config/config.beta.toml ./config/
COPY --from=builder /build/sql-migrate /usr/local/bin/sql-migrate
COPY --from=dashboard_builder /dist-dashboard ./dist-dashboard

CMD ["./hivepaas"]
