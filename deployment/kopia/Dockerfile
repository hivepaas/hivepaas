# kopia, the backup engine the app and agent images ship, built from source.
# kopia's own release binaries keep the Go and the modules of their day: a flaw
# found since in one of them stays in the binary until kopia releases again.
# This one is built with Go 1.27 and the modules below raised past the
# release's go.mod, and built again - a tag kopia-v<version>-<build> - when
# the weekly scan finds a flaw with a fix (.github/workflows/kopia.yml).
#
# The image holds the binary and its license only. The app's and the agent's
# Dockerfiles copy the binary from it (KOPIA_IMAGE), pinned by digest.

############################################################
# Build kopia
FROM golang:1.27-alpine AS builder

ARG KOPIA_VERSION=0.23.1

# Each module raised has a flaw fixed past the release's version. golang.org/x
# to its latest; grpc and compress, which kopia's server and its compression
# run on, to the first release with the fix only.
ARG KOPIA_MODULES="\
golang.org/x/crypto@v0.57.0 \
golang.org/x/mod@v0.41.0 \
golang.org/x/net@v0.59.0 \
golang.org/x/text@v0.42.0 \
google.golang.org/grpc@v1.83.2 \
github.com/klauspost/compress@v1.18.7"

ENV CGO_ENABLED=0

WORKDIR /build

# A module of its own requires kopia's: the module proxy serves kopia's source
# and the checksum database vouches for it, as for every module raised.
RUN go mod init kopia-build && \
    go get "github.com/kopia/kopia@v${KOPIA_VERSION}" ${KOPIA_MODULES} && \
    go build -trimpath -ldflags "-s -w \
        -X github.com/kopia/kopia/repo.BuildVersion=${KOPIA_VERSION} \
        -X github.com/kopia/kopia/repo.BuildInfo=v${KOPIA_VERSION} \
        -X github.com/kopia/kopia/repo.BuildGitHubRepo=kopia/kopia" \
        -o /out/kopia github.com/kopia/kopia && \
    cp "$(go env GOMODCACHE)/github.com/kopia/kopia@v${KOPIA_VERSION}/LICENSE" /out/LICENSE && \
    /out/kopia --version

############################################################
# The image
FROM scratch

ARG KOPIA_VERSION
ARG REVISION
LABEL org.opencontainers.image.title="kopia for HivePaaS" \
      org.opencontainers.image.description="kopia, built from source with patched Go modules, for the HivePaaS app and agent images" \
      org.opencontainers.image.source="https://github.com/hivepaas/hivepaas" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${KOPIA_VERSION}" \
      org.opencontainers.image.revision="${REVISION}"

COPY --from=builder /out/kopia /kopia
COPY --from=builder /out/LICENSE /LICENSE
