UID := $(shell id -u)

# ----- Development tools -----
init: build-devtools

DEVTOOLS_IMAGE := hivepaas-devtools
DEVTOOLS_CMD := docker run --user "$(UID)" --rm --volume "$(PWD)":/app --network="host" $(DEVTOOLS_IMAGE)
build-devtools:
	@docker build --file ./tools/docker/Dockerfile --tag ${DEVTOOLS_IMAGE} .

GO_MOD_ENV=GOPRIVATE=github.com/hivepaas/*
mod:
	@$(GO_MOD_ENV) go mod tidy && go mod vendor && go mod verify

lint:
	$(DEVTOOLS_CMD) golangci-lint --timeout=3m run -v ./...

lint-local:
	# Run this cmd locally once to install golangci-lint binary
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.13.0
	# FASTER golangci-lint --timeout=5m run -v --new-from-rev=HEAD~1
	golangci-lint --timeout=5m run -v ./...

test:
	@./scripts/test.sh

nilaway:
	# run once: go install go.uber.org/nilaway/cmd/nilaway@latest
	@nilaway ./hivepaas_app/...

vuln:
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

trivy:
	@if command -v trivy >/dev/null 2>&1; then \
		trivy fs --skip-dirs "vendor,.temp,tmp,temp,.appdata" --severity CRITICAL,HIGH .; \
	else \
		docker run --rm -v "$(PWD)":/app -w /app aquasec/trivy:latest fs --skip-dirs "vendor,.temp,tmp,temp" --severity CRITICAL,HIGH .; \
	fi

# ----- Build flags -----
PROD_LDFLAGS := -s -w
PROD_FLAGS := -trimpath -ldflags="$(PROD_LDFLAGS)"

run:
	@go run ./hivepaas_app/cmd/app/...

# Dev builds
build:
	@go build -o hivepaas ./hivepaas_app/cmd/app/...

build-agent:
	@go build -o hivepaas-agent ./hivepaas_app/cmd/agent/...

# Production builds (Stripped & Trimpath)
build-prod:
	@go build $(PROD_FLAGS) -o hivepaas ./hivepaas_app/cmd/app/...

build-agent-prod:
	@go build $(PROD_FLAGS) -o hivepaas-agent ./hivepaas_app/cmd/agent/...

build-all-prod: build-prod build-agent-prod

# ----- Code generation -----
gen: gen-go gen-proto gen-swag

gen-go:
	$(DEVTOOLS_CMD) env GOCACHE=/tmp/go-cache go generate ./...

gen-proto:
	# may need to install protobuf (mac os: brew install protobuf)
    # go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    # go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go generate ./hivepaas_app/interface/agent/proto/...

gen-swag:
	@./tools/swag/swag.sh

SRC_LOCAL="github.com/hivepaas/hivepaas/"
fmt: ## gofmt and goimports all go files
	@find . -name '*.go' -not -wholename './vendor/*' -not -wholename './.temp/*'  -not -wholename '*.pb.go' -not -wholename '*_gen.go' -not -wholename '*/mock_*.go' | while read -r file; do gofmt -w -s "$$file"; goimports -local ${SRC_LOCAL} -w "$$file"; done

# ----- DB migration -----
DB_MIGRATE_DIR := hivepaas_app/db
DB_CONN_STR := host=localhost port=35432 dbname=hivepaas user=hivepaas password=abc123
DB_MIGRATE_BASE := $(DEVTOOLS_CMD) sql-migrate
DB_MIGRATE_ENV := development
DB_EXEC_BASE := $(DEVTOOLS_CMD) psql -d "$(DB_CONN_STR)"

# This is considered the remote env
ifdef HP_PLATFORM
ifneq ($(HP_PLATFORM), local)
	DB_CONN_STR := host=${HP_DB_HOST} port=${HP_DB_PORT} dbname=${HP_DB_DB_NAME} user=${HP_DB_USER} password=${HP_DB_PASSWORD}
	DB_MIGRATE_BASE := sql-migrate
	DB_MIGRATE_ENV := main
	DB_EXEC_BASE := psql -d "${DB_CONN_STR}"
endif
endif

migrate-setup: build-devtools

migrate-new:
ifndef NAME
	$(error "Please provide migration name, i.e.: make $@ NAME=example_migration")
else
	$(DB_MIGRATE_BASE) new -config=${DB_MIGRATE_DIR}/dbconfig.yml $(NAME)
endif

migrate-status:
	$(DB_MIGRATE_BASE) status -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

migrate-up:
	$(DB_MIGRATE_BASE) up -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

migrate-down:
	$(DB_MIGRATE_BASE) down -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

migrate-redo:
	$(DB_MIGRATE_BASE) redo -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

seed-data:
	make migrate-up
	$(DB_EXEC_BASE) -f ${DB_MIGRATE_DIR}/seed/seed.sql

seed-data-with-clear:
	$(DB_EXEC_BASE) -f ${DB_MIGRATE_DIR}/seed/clear.sql
	make migrate-up
	$(DB_EXEC_BASE) -f ${DB_MIGRATE_DIR}/seed/seed.sql

dev-deploy:
	git checkout main
	git pull
	git push origin --delete dev-v0.1.0 # delete tag in remote
	git tag dev-v0.1.0 --force
	git push origin dev-v0.1.0 --force

local-deploy:
	mkdir -p tmp
	bash deployment/local/install.sh

ifndef HP_FE_DIR
HP_FE_DIR=../hivepaas-dashboard
endif

local-build-fe:
	cd ${HP_FE_DIR} && git pull && yarn install && yarn build
	rm -rf dist-dashboard
	mv ${HP_FE_DIR}/dist dist-dashboard

# ----- Smee.io config -----
smee-run:
	# RUN ONCE go install -v github.com/chmouel/gosmee@latest
	# github app id: 01JAB9XED0GTXBSQDFVYAJ8WJ1
	# webhook id: 01JAB9XED0GTXBSQDFVYAJ8WO1 (github), 01JAB9XED0GTXBSQDFVYAJ8WO2 (gitlab), 01JAB9XED0GTXBSQDFVYAJ8WO3 (gitea)
	gosmee client --saveDir tmp/gosmee/savedreplay https://smee.io/RBNiNjxieUIWZ6Ej http://localhost:10000/_/webhooks/01JAB9XED0GTXBSQDFVYAJ8WJ1
