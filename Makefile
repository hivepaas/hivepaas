UID := $(shell id -u)

# ----- Development tools -----
init: build-devtools

DEVTOOLS_IMAGE := hivepaas-devtools
DEVTOOLS_CMD := docker run --user "$(UID)" --rm --volume "$(PWD)":/app --network="host" $(DEVTOOLS_IMAGE)
build-devtools:
	@docker build --file ./tools/docker/Dockerfile --tag ${DEVTOOLS_IMAGE} .

GO_MOD_ENV=GOPRIVATE=github.com/hivepaas/*
mod:
	@$(GO_MOD_ENV) go mod tidy && go mod vendor

lint:
	$(DEVTOOLS_CMD) golangci-lint --timeout=3m run -v ./...

lint-local:
	# Run this cmd locally once to install golangci-lint binary
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.2
	golangci-lint --timeout=5m run -v ./...

test:
	@./scripts/test.sh

build:
	@go build -o hivepaas ./hivepaas_app/cmd/app/...

run:
	@go run ./hivepaas_app/cmd/app/...

# ----- Code generation -----
gen: gen-go gen-swag

gen-go:
	$(DEVTOOLS_CMD) env GOCACHE=/tmp/go-cache go generate ./...

gen-swag:
	@./tools/swag/swag.sh

SRC_LOCAL="github.com/hivepaas/hivepaas/"
fmt: ## gofmt and goimports all go files
	@find . -name '*.go' -not -wholename './vendor/*' -not -wholename './.temp/*' -not -wholename '*_gen.go' -not -wholename '*/mock_*.go' | while read -r file; do gofmt -w -s "$$file"; goimports -local ${SRC_LOCAL} -w "$$file"; done

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
smee-connect:
	# RUN ONCE go install -v github.com/chmouel/gosmee@latest
	# Make sure you use correct <setting-id> below
	gosmee client --saveDir tmp/gosmee/savedreplay https://smee.io/RBNiNjxieUIWZ6Ej http://localhost:10000/_/webhooks/01JAB9XED0GTXBSQDFVYAJ8WJ1

# ----- Build local image -----
build-image:
	docker build -f deployment/dev/Dockerfile -t hivepaas:latest .

build-agent-image:
	docker build -f deployment/dev/Dockerfile.agent -t hivepaas-agent:latest .

