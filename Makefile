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
	$(DEVTOOLS_CMD) go run ./tools/goroutinelint .
	$(DEVTOOLS_CMD) go run ./tools/errcodelint
	$(DEVTOOLS_CMD) golangci-lint --timeout=5m run -v ./...

lint-local:
	@go run ./tools/goroutinelint .
	@go run ./tools/errcodelint
	# Run this cmd locally once to install golangci-lint binary
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.13.0
	# FASTER golangci-lint --timeout=5m run -v --new-from-rev=HEAD~1
	golangci-lint --timeout=5m run -v ./...

# The everyday one. Cached: a package nothing touched is not run again.
test:
	@go test ./...

# Before pushing anything that touches concurrency.
test-race:
	@go test -race ./...

# Race plus the coverage report. This is what CI runs.
test-cover:
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

# ----- Release signing -----
# Signs release.json with the offline release keys (ed25519 and ML-DSA-65),
# writing release.signed.json - what installations fetch, from the `release`
# branch (`main` for development builds). The tool is built from RELEASESIGN_SHA, not the
# working tree: moving the pin is how a reviewed change to tools/releasesign
# reaches the keys. See the script.
#   make release-sign KEYS="/offline/2026_ed.key /offline/2026_ml.key"
RELEASESIGN_SHA := 921f4a7263a6819aed7e8d8f69fcc1259af46255
release-sign:
	@RELEASESIGN_SHA="$(RELEASESIGN_SHA)" KEYS="$(KEYS)" IN="$(IN)" OUT="$(OUT)" ./scripts/release-sign.sh

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

fmt:
	golangci-lint fmt ./...

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
	git push origin --delete dev-v0.1.0 || true # delete tag in remote
	git tag dev-v0.1.0 --force
	git push origin dev-v0.1.0 --force

local-deploy:
	mkdir -p tmp
	bash deployment/local/install.sh

local-build-dashboard:
	@HP_FE_DIR="$(HP_FE_DIR)" ./scripts/dev/local-build-dashboard.sh

# ----- Running the local build on the host -----
# The everyday way to work. See scripts/dev/local-run.sh.
local-app-run:
	@./scripts/dev/local-run.sh app

local-agent-run:
	@./scripts/dev/local-run.sh agent

# ----- Run the local build as swarm services -----
# For what running on the host cannot reach: the app's own swarm service, the
# updater, placement, `platform = "remote"`. See scripts/dev/local-swarm.sh.
local-app-image:
	@./scripts/dev/local-swarm.sh image app

local-app-up:
	@./scripts/dev/local-swarm.sh up app

local-app-down:
	@./scripts/dev/local-swarm.sh down app

local-app-logs:
	@./scripts/dev/local-swarm.sh logs app

local-agent-image:
	@./scripts/dev/local-swarm.sh image agent

local-agent-up:
	@./scripts/dev/local-swarm.sh up agent

local-agent-down:
	@./scripts/dev/local-swarm.sh down agent

local-agent-logs:
	@./scripts/dev/local-swarm.sh logs agent

# ----- Extra swarm nodes -----
#
# docker-in-docker: a privileged container running its own dockerd, joined to the
# local swarm as a worker. Enough to exercise anything that reasons about more
# than one node - global services, placement constraints, node labels, a volume
# pinned to nowhere - without a second machine.
#
# Each node runs what the stack gives it, which in practice means the agent: a
# global service, so swarm schedules a task there the moment the node joins and
# pulls the published image for it. The :local tags that local-app-up and
# local-agent-up build exist on the Desktop daemon only and cannot be pulled here.
#
# Nodes are given no labels, deliberately. app, db and traefik are pinned to
# `node.labels.hivepaas.role == control-plane` and have to stay on the Desktop
# node, whose host paths their volumes point at.
#
# Nothing here is persistent: a node's images and data live in its container's
# writable layer and go with it. Fine for placement work; use real VMs (colima,
# lima, multipass) for anything that has to survive a node restart.
LOCAL_NODE_PREFIX := hivepaas-node
LOCAL_NODE_IMAGE := docker:dind

# Adds one, named for the first free number - so running it again adds another.
local-node-up:
	@bash tools/swarm/add-node.sh $(LOCAL_NODE_PREFIX) $(LOCAL_NODE_IMAGE)
	@docker node ls

# Removes the one the last local-node-up added. `make local-node-down
# LOCAL_NODE=hivepaas-node3` removes that one instead.
local-node-down:
	@bash tools/swarm/remove-node.sh $(LOCAL_NODE_PREFIX) $(LOCAL_NODE)
	@docker node ls

local-node-down-all:
	@bash tools/swarm/remove-node.sh $(LOCAL_NODE_PREFIX) --all
	@docker node ls

# ----- Smee.io config -----

smee-run:
	# RUN ONCE go install -v github.com/chmouel/gosmee@latest
	# github app id: 01JAB9XED0GTXBSQDFVYAJ8WJ1
	# webhook id: 01JAB9XED0GTXBSQDFVYAJ8WO1 (github), 01JAB9XED0GTXBSQDFVYAJ8WO2 (gitlab), 01JAB9XED0GTXBSQDFVYAJ8WO3 (gitea)
	gosmee client --saveDir tmp/gosmee/savedreplay https://smee.io/RBNiNjxieUIWZ6Ej http://localhost:10000/_/webhooks/01JAB9XED0GTXBSQDFVYAJ8WJ1
