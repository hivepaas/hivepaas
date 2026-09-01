############################################################
# Build dashboard
FROM node:26-alpine AS dashboard_builder
RUN apk update && apk upgrade --no-cache && apk add --update --no-cache git yarn

WORKDIR /build

RUN --mount=type=cache,target=/tmp/yarn-cache \
    git clone --depth=1 https://github.com/hivepaas/hivepaas-dashboard.git && \
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

# ENV vars
ENV CGO_ENABLED=0

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY . .

# Copy and download dependency using go mod
RUN make mod

# Build the application
RUN make build

# Install DB migration tool, then copy the binary to user bin path
# NOTE: the output binary is placed in a directory like `linux_amd64`, but the name depends on
# the actual os arch. So I use `find` command here.
RUN go install github.com/rubenv/sql-migrate/...@latest && \
    find /go/bin/ -name 'sql-migrate' -exec cp "{}" ./sql-migrate \;

############################################################
# Build a small image
FROM alpine:3.24
ARG TARGETARCH

RUN apk update && \
    apk add --no-cache ca-certificates make openssl postgresql-client git git-lfs openssh-client zstd lz4 docker-cli docker-cli-buildx && \
    git lfs install --system && \
    rm -rf /var/cache/apk/*

# Install kopia backup engine (statically linked binary from official releases)
ARG KOPIA_VERSION=0.23.1
RUN case "${TARGETARCH}" in \
        amd64) KOPIA_ARCH="x64" ;; \
        arm64) KOPIA_ARCH="arm64" ;; \
        arm)   KOPIA_ARCH="arm" ;; \
        *)     echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    wget -qO /tmp/kopia.tar.gz \
        "https://github.com/kopia/kopia/releases/download/v${KOPIA_VERSION}/kopia-${KOPIA_VERSION}-linux-${KOPIA_ARCH}.tar.gz" && \
    tar -xzf /tmp/kopia.tar.gz -C /tmp && \
    mv "/tmp/kopia-${KOPIA_VERSION}-linux-${KOPIA_ARCH}/kopia" /usr/local/bin/kopia && \
    chmod +x /usr/local/bin/kopia && \
    rm -rf /tmp/kopia* && \
    kopia --version

WORKDIR /hivepaas

# For running app
COPY --from=builder /build/hivepaas .
COPY --from=builder /build/hivepaas_app/db ./hivepaas_app/db
COPY --from=builder /build/config ./config
COPY --from=builder /build/scripts ./scripts
COPY --from=builder /build/Makefile .
COPY --from=builder /build/sql-migrate /usr/local/bin/sql-migrate
COPY --from=dashboard_builder /dist-dashboard ./dist-dashboard

# Command to run
CMD ["sh","-c","./hivepaas"]
