FROM golang:1.27-alpine

RUN apk update && apk upgrade && apk add --update alpine-sdk postgresql-client protobuf

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.13.0 &&\
    go install github.com/rubenv/sql-migrate/...@latest &&\
    go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5 &&\
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest &&\
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app
