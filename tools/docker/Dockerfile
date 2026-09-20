FROM golang:1.27-alpine

# What the make targets that run in here actually need from the system:
#   postgresql-client  psql, and sql-migrate talking to the database
#   protobuf           protoc, which `make gen-go` runs through go:generate
#
# Nothing here needs a C compiler: the whole repository builds with
# CGO_ENABLED=0, and the tools installed below are pure Go. `alpine-sdk` used to
# be pulled in for it and brought 104 packages with it.
#
# --no-cache fetches the index and leaves nothing behind, which is what `apk
# update` was for. `apk upgrade` is gone: the base image is built from current
# packages already, and upgrading them here costs a minute and changes what the
# image was pinned to.
RUN apk add --no-cache postgresql-client protobuf

# The tools are built without cgo, so they need no compiler either, and this
# keeps `go run` and golangci-lint inside the container consistent with how the
# binaries are built.
ENV CGO_ENABLED=0

# One layer per tool: changing a version rebuilds that one instead of all five.
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
    | sh -s -- -b /go/bin v2.13.0
RUN go install github.com/rubenv/sql-migrate/...@latest
RUN go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app
