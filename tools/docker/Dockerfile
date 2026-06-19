FROM golang:1.26-alpine

RUN apk update && apk upgrade && apk add --update alpine-sdk postgresql-client protobuf

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.12.2 &&\
    go install github.com/rubenv/sql-migrate/...@latest &&\
    go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5 &&\
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0 &&\
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

WORKDIR /app
