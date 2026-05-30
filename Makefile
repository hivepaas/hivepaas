UID := $(shell id -u)

# ----- Development tools -----
init: build-devtools

DEVTOOLS_IMAGE := localpaas-devtools
DEVTOOLS_CMD := docker run --user "$(UID)" --rm --volume "$(PWD)":/app --network="host" $(DEVTOOLS_IMAGE)
build-devtools:
	@docker build --file ./tools/docker/Dockerfile --tag ${DEVTOOLS_IMAGE} .

GO_MOD_ENV=GOPRIVATE=github.com/localpaas/*
mod:
	@$(GO_MOD_ENV) go mod tidy && go mod vendor

lint:
	$(DEVTOOLS_CMD) golangci-lint --timeout=3m run -v ./...

lint-local:
	# Run this cmd locally once to install golangci-lint binary
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.9.0
	golangci-lint --timeout=5m run -v ./...

test:
	@./scripts/test.sh

build:
	@go build -o localpaas ./localpaas_app/cmd/app/...

run:
	@go run ./localpaas_app/cmd/app/...

# ----- Code generation -----
gen: gen-go gen-swag

gen-go:
	@go generate ./...

gen-swag:
	@./tools/swag/swag.sh

# ----- DB migration -----
DB_MIGRATE_DIR := localpaas_app/db
DB_CONN_STR := host=localhost port=35432 dbname=localpaas user=localpaas password=abc123
DB_MIGRATE_BASE := $(DEVTOOLS_CMD) sql-migrate
DB_MIGRATE_ENV := development
DB_EXEC_BASE := $(DEVTOOLS_CMD) psql -d "$(DB_CONN_STR)"

# This is considered the remote env
ifdef LP_PLATFORM
ifneq ($(LP_PLATFORM), local)
	DB_CONN_STR := host=${LP_DB_HOST} port=${LP_DB_PORT} dbname=${LP_DB_DB_NAME} user=${LP_DB_USER} password=${LP_DB_PASSWORD}
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

ifndef LP_FE_DIR
LP_FE_DIR=../localpaas-dashboard
endif

local-build-fe:
	cd ${LP_FE_DIR} && git pull && yarn install && yarn build
	rm -rf dist-dashboard
	mv ${LP_FE_DIR}/dist dist-dashboard

# ----- Smee.io config -----
smee-connect:
	# RUN ONCE go install -v github.com/chmouel/gosmee@latest
	# Make sure you use correct <setting-id> below
	gosmee client --saveDir tmp/gosmee/savedreplay https://smee.io/RBNiNjxieUIWZ6Ej http://localhost:10000/_/webhooks/01JAB9XED0GTXBSQDFVYAJ8WO1/abc123

# ----- Build local image -----
build-image:
	docker build -f deployment/dev/Dockerfile -t localpaas:latest .
