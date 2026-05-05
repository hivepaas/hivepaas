############################################################
# Build dashboard
FROM node:24-alpine AS dashboard_builder
RUN apk update && apk upgrade --no-cache && apk add --update --no-cache git yarn

WORKDIR /build

RUN --mount=type=cache,target=/tmp/yarn-cache \
    git clone --depth=1 https://github.com/localpaas/localpaas-dashboard.git && \
    cd localpaas-dashboard && \
    yarn config set network-timeout 300000 && \
    yarn config set cache-folder /tmp/yarn-cache && \
    yarn install --frozen-lockfile && \
    yarn build && \
    mv dist /dist-dashboard

############################################################
# Build backend
FROM golang:1.26-alpine AS builder

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
FROM alpine:3.22
WORKDIR /localpaas
RUN apk update && \
    apk add --no-cache ca-certificates make openssl postgresql-client && \
    rm -rf /var/cache/apk/*

# For running app
COPY --from=builder /build/localpaas .
COPY --from=builder /build/localpaas_app/db ./localpaas_app/db
COPY --from=builder /build/config ./config
COPY --from=builder /build/scripts ./scripts
COPY --from=builder /build/Makefile .
COPY --from=builder /build/sql-migrate /usr/local/bin/sql-migrate
COPY --from=dashboard_builder /dist-dashboard ./dist-dashboard

# Command to run
CMD ["sh","-c","./localpaas"]
