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
	$(DEVTOOLS_CMD) golangci-lint --timeout=3m run -v ./...

lint-local:
	@go run ./tools/goroutinelint .
	@go run ./tools/errcodelint
	# Run this cmd locally once to install golangci-lint binary
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.13.0
	# FASTER golangci-lint --timeout=3m run -v --new-from-rev=HEAD~1
	golangci-lint --timeout=3m run -v ./...

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

# Applies exactly the formatters .golangci.yaml enables - gofmt, and gci with the
# local prefix - so this and `make lint` can never disagree about what formatted
# means. The previous body ran gofmt and goimports as two processes per file:
# identical output to the byte, 25s instead of 1s across 2367 files, and it needed
# goimports on the host, which nothing in this repo installs.
#
# Formatting is not on the honor system. `golangci-lint run` reports an
# unformatted file as an issue of its own, so CI already fails on one. This target
# is the convenience that fixes them, not the gate that catches them.
fmt: ## format all go files
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

ifndef HP_FE_DIR
HP_FE_DIR=../hivepaas-dashboard
endif

local-build-dashboard:
	cd ${HP_FE_DIR} && git pull && yarn install && yarn build
	rm -rf dist-dashboard
	mv ${HP_FE_DIR}/dist dist-dashboard

# ----- Running the local build on the host -----
#
# The everyday way to work: `go run` with the environment the local install needs.
#
# HP_CONFIG_FILE on its own is not enough, and what it misses is silent. The app
# path is resolved before any config file is read - resolveAppPath() reads
# HP_APP_PATH and falls back to /var/lib/hivepaas - and it is what locates
# hivepaas.toml, the file the app writes for itself. That file holds `secret`, the
# key that decrypts every stored secret. Without HP_APP_PATH the process reads a
# hivepaas.toml that is not there, keeps the `secret` from config.local.toml, and
# anything encrypted since the app rotated its own key stops being readable.
# `app_path` in the config file cannot help: by the time it is parsed the lookup
# has already happened.
#
# One consequence worth knowing: with HP_APP_PATH set, a config.toml sitting in
# that directory outranks HP_CONFIG_FILE. There is none by default.
LOCAL_CONFIG := config/config.local.toml
# Absolute, because HP_STORAGE_BIND_SOURCE is handed to docker as the source of a
# bind mount and the daemon resolves it on the host, not against this process's
# working directory. In the stack the two are deliberately different - the
# container sees its data at /var/lib/hivepaas while docker binds it from the host
# path - but running on the host there is no boundary between them.
LOCAL_APP_PATH := $(PWD)/.appdata/hivepaas

local-app-run:
	@mkdir -p $(LOCAL_APP_PATH)
	HP_CONFIG_FILE=$(LOCAL_CONFIG) \
	HP_APP_PATH=$(LOCAL_APP_PATH) \
	HP_STORAGE_BIND_SOURCE=$(LOCAL_APP_PATH) \
		go run ./hivepaas_app/cmd/app/...

# No HP_APP_PATH here, on purpose. Nothing on the agent side decrypts anything:
# the app decrypts and sends plaintext over gRPC, which is what
# HP_AGENT_SECRET_TOKEN exists to protect, so the key in hivepaas.toml is of no
# use to the agent. HP_AGENT_NODE_ID and HP_AGENT_NODE_NAME, which the stack file
# sets, are read by nothing in the Go code at all.
local-agent-run:
	HP_CONFIG_FILE=$(LOCAL_CONFIG) go run ./hivepaas_app/cmd/agent/...

# ----- Run the local build as swarm services -----
#
# Day to day these run straight from the IDE. These targets are for what that
# cannot reach: code that reads the app's own swarm service, the updater paths,
# placement constraints, the agent's view of the node it sits on, or anything
# gated on `platform = "remote"`.
#
# Each image is one layer on top of the published dev image, which already carries
# every runtime dependency (kopia, docker-cli, sql-migrate, git-lfs, dashboard).
# Building deployment/dev/Dockerfile instead would clone the dashboard repo and
# run a yarn build to pick up a change to one Go file.
#
# The :local tags exist on this daemon and nowhere else, which is fine for a
# single-node local swarm and is the whole of why this does not work on a
# multi-node one: the other nodes have nothing to pull.
LOCAL_APP_IMAGE := hivepaas/hivepaas-dev:local
LOCAL_APP_IMAGE_BASE := hivepaas/hivepaas-dev:latest
LOCAL_APP_CTX := tmp/img-app
LOCAL_AGENT_IMAGE := hivepaas/hivepaas-agent-dev:local
LOCAL_AGENT_IMAGE_BASE := hivepaas/hivepaas-agent-dev:latest
LOCAL_AGENT_CTX := tmp/img-agent
# The daemon's arch, not the host's - on a mac they are the same answer spelled
# two different ways, and on a remote daemon they are not the same answer at all.
# Assigned with `=` so `docker version` only runs for the targets below.
LOCAL_IMAGE_ARCH = $(shell docker version --format '{{.Server.Arch}}')

# --no-resolve-image: a :local tag exists on this daemon and nowhere else, and
#   swarm's default is to ask a registry to resolve a tag to a digest first.
# --force: the tag never changes, so without it swarm reads the spec as unchanged
#   and leaves the task running the previous binary.
# --detach: an attached update does not return until swarm calls the rollout
#   converged, and converged means update_config's monitor has elapsed - 240s for
#   hivepaas_app, measured at 4m13s for an update whose container was serving
#   within seconds. The monitor is what arms the rollback and is not worth
#   shortening, so tools/swarm/wait-for-task.sh waits for the task instead.
LOCAL_SVC_UPDATE := docker service update --detach --quiet --force --no-resolve-image

# Its own context directory rather than the repo root: .dockerignore only drops
# README.*, so a root context would ship vendor/ and .git/ on every build.
#
# config/ and hivepaas_app/db are copied in so local edits to settings and
# migrations take effect; the base image's copies are from whenever it was built.
# An empty dist-dashboard means the dashboard was never built here - copying it
# would replace the base image's with nothing and serve a blank page, so that
# layer is left out instead.
#
# PROD_LDFLAGS takes the binary from 116MB to 83MB. This layer is rebuilt every
# time, so its size is the loop: about 17s end to end once the base image is
# local, of which roughly 7s is the link and 8s the image build.
local-app-image:
	@mkdir -p $(LOCAL_APP_CTX)
	@GOOS=linux GOARCH=$(LOCAL_IMAGE_ARCH) CGO_ENABLED=0 \
		go build -ldflags="$(PROD_LDFLAGS)" -o $(LOCAL_APP_CTX)/hivepaas ./hivepaas_app/cmd/app/...
	@rm -rf $(LOCAL_APP_CTX)/config $(LOCAL_APP_CTX)/hivepaas_app $(LOCAL_APP_CTX)/dist-dashboard
	@cp -R config $(LOCAL_APP_CTX)/config
	@mkdir -p $(LOCAL_APP_CTX)/hivepaas_app
	@cp -R hivepaas_app/db $(LOCAL_APP_CTX)/hivepaas_app/db
	@printf 'FROM %s\nWORKDIR /hivepaas\nCOPY hivepaas ./hivepaas\nCOPY config ./config\nCOPY hivepaas_app/db ./hivepaas_app/db\n' \
		$(LOCAL_APP_IMAGE_BASE) > $(LOCAL_APP_CTX)/Dockerfile
	@if [ -n "$$(ls -A dist-dashboard 2>/dev/null)" ]; then \
		cp -R dist-dashboard $(LOCAL_APP_CTX)/dist-dashboard; \
		printf 'COPY dist-dashboard ./dist-dashboard\n' >> $(LOCAL_APP_CTX)/Dockerfile; \
	else \
		echo "dist-dashboard is empty, keeping the one in $(LOCAL_APP_IMAGE_BASE) - build yours with 'make local-build-dashboard'"; \
	fi
	@docker build -q -t $(LOCAL_APP_IMAGE) $(LOCAL_APP_CTX) > /dev/null && echo "built $(LOCAL_APP_IMAGE)"

# HP_RUN_MODE=app+worker: what the stack itself runs this service as, so the task
# is the whole thing rather than half of it - the deploy, clone and periodic-job
# executors all live on the worker side. Set explicitly rather than left to
# config.development.toml so `docker service inspect` says what the task is doing.
# Note that the IDE build is usually still up on the same database, and then both
# are pulling from the same task queue; stop one of them when that matters.
#
# The URL is whatever domain the install itself holds, not the `app.dev.localhost`
# in deployment/local/hivepaas.yaml: the app rewrites its own traefik labels from
# the domains in the database, so the stack's static labels are replaced the first
# time it runs. Both builds read the same database, so both answer on the same
# host - `docker service inspect hivepaas_app` shows the labels in force.
#
# The image being replaced is remembered so local-app-down can put it back.
# `docker service rollback` cannot be used for that: run this twice and the spec
# it rolls back to is the previous :local one.
local-app-up: local-app-image
	@prev=$$(docker service ps hivepaas_app --filter desired-state=running -q | head -1); \
	img=$$(docker service inspect hivepaas_app --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}'); \
	case "$$img" in *:local|*:local@*) ;; *) echo "$$img" > $(LOCAL_APP_CTX)/.previous-image ;; esac; \
	$(LOCAL_SVC_UPDATE) --image $(LOCAL_APP_IMAGE) --env-add HP_RUN_MODE=app+worker --replicas 1 hivepaas_app > /dev/null; \
	bash tools/swarm/wait-for-task.sh hivepaas_app "$$prev"
	@echo "hivepaas_app is running the local build - https://localhost"
	@echo "logs: make local-app-logs   stop: make local-app-down"

# Back to the stack's own image, environment and replica count. Nothing waits for
# a task here - there is not going to be one.
#
# With nothing remembered, the image is left exactly as it is rather than set to
# $(LOCAL_APP_IMAGE_BASE): `docker stack deploy` pins this service to a digest and
# the bare tag does not resolve on this daemon at all, so writing it would trade a
# reference that works for one that has to be fetched.
local-app-down:
	@if [ -f $(LOCAL_APP_CTX)/.previous-image ]; then \
		img=$$(cat $(LOCAL_APP_CTX)/.previous-image); \
		$(LOCAL_SVC_UPDATE) --image "$$img" --env-rm HP_RUN_MODE --replicas 0 hivepaas_app > /dev/null; \
		echo "hivepaas_app scaled to 0, image back to $$img"; \
	else \
		$(LOCAL_SVC_UPDATE) --env-rm HP_RUN_MODE --replicas 0 hivepaas_app > /dev/null; \
		echo "hivepaas_app scaled to 0; no remembered image, run 'make local-deploy' to restore the stack's"; \
	fi

local-app-logs:
	@docker service logs -f --tail 100 hivepaas_app

# The agent image holds only the binary and config - no dashboard, no migrations -
# so this context is the binary plus a few kilobytes.
local-agent-image:
	@mkdir -p $(LOCAL_AGENT_CTX)
	@GOOS=linux GOARCH=$(LOCAL_IMAGE_ARCH) CGO_ENABLED=0 \
		go build -ldflags="$(PROD_LDFLAGS)" -o $(LOCAL_AGENT_CTX)/hivepaas-agent ./hivepaas_app/cmd/agent/...
	@rm -rf $(LOCAL_AGENT_CTX)/config
	@cp -R config $(LOCAL_AGENT_CTX)/config
	@printf 'FROM %s\nWORKDIR /hivepaas\nCOPY hivepaas-agent ./hivepaas-agent\nCOPY config ./config\n' \
		$(LOCAL_AGENT_IMAGE_BASE) > $(LOCAL_AGENT_CTX)/Dockerfile
	@docker build -q -t $(LOCAL_AGENT_IMAGE) $(LOCAL_AGENT_CTX) > /dev/null && echo "built $(LOCAL_AGENT_IMAGE)"

# Unlike the app, this service is meant to be up: it is global, the stack starts
# it, and nothing else answers the app's gRPC calls for a node. So there is no
# scaling it to zero - local-agent-down puts the published image back instead.
#
# HP_AGENT_NODE_ID and HP_AGENT_NODE_NAME are swarm templates (`{{.Node.ID}}`),
# and an update carries them across as the templates they are, not as whatever
# this node resolved them to.
local-agent-up: local-agent-image
	@prev=$$(docker service ps hivepaas_agent --filter desired-state=running -q | head -1); \
	img=$$(docker service inspect hivepaas_agent --format '{{.Spec.TaskTemplate.ContainerSpec.Image}}'); \
	case "$$img" in *:local|*:local@*) ;; *) echo "$$img" > $(LOCAL_AGENT_CTX)/.previous-image ;; esac; \
	$(LOCAL_SVC_UPDATE) --image $(LOCAL_AGENT_IMAGE) hivepaas_agent > /dev/null; \
	bash tools/swarm/wait-for-task.sh hivepaas_agent "$$prev"
	@echo "hivepaas_agent is running the local build on every node"
	@echo "logs: make local-agent-logs   stop: make local-agent-down"

# The agent has to be running something, so with nothing remembered this does fall
# back to $(LOCAL_AGENT_IMAGE_BASE) - unlike the app above, that tag does resolve.
local-agent-down:
	@prev=$$(docker service ps hivepaas_agent --filter desired-state=running -q | head -1); \
	img=$$(cat $(LOCAL_AGENT_CTX)/.previous-image 2>/dev/null || echo $(LOCAL_AGENT_IMAGE_BASE)); \
	$(LOCAL_SVC_UPDATE) --image "$$img" hivepaas_agent > /dev/null; \
	bash tools/swarm/wait-for-task.sh hivepaas_agent "$$prev"; \
	echo "hivepaas_agent is back on $$img"

local-agent-logs:
	@docker service logs -f --tail 100 hivepaas_agent

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
